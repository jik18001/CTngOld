package server

import (
	"CTng/gossip"
	"CTng/util"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

type Gossiper interface {

	// Response to entering the 'base page' of a gossiper.
	// TODO: Create informational landing page
	homePage()

	// HTTP POST request, receive a JSON object from another gossiper or connected monitor.
	// /gossip/push-data
	handleGossip(w http.ResponseWriter, r *http.Request)

	// Respond to HTTP GET request.
	// /gossip/get-data
	handleGossipObjectRequest(w http.ResponseWriter, r *http.Request)

	// Push JSON object to connected network from this gossiper via HTTP POST.
	// /gossip/gossip-data
	gossipData()

	// TODO: Push JSON object to connected 'owner' (monitor) from this gossiper via HTTP POST.
	// Sends to an owner's /monitor/recieve-gossip endpoint.
	sendToOwner()

	// Process JSON object received from HTTP POST requests.
	processData()

	// Create and gossip accusation object.
	accuseEntity()

	// TODO: Erase stored data after one MMD.
	eraseData()

	// HTTP server function which handles GET and POST requests.
	handleRequests()
}

// Binds the context to the functions we pass to the router.
func bindContext(context *gossip.GossiperContext, fn func(context *gossip.GossiperContext, w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(context, w, r)
	}
}

func handleRequests(c *gossip.GossiperContext) {

	// MUX which routes HTTP directories to functions.
	gorillaRouter := mux.NewRouter().StrictSlash(true)

	// homePage() is ran when base directory is accessed.
	gorillaRouter.HandleFunc("/gossip/", homePage)

	// Inter-gossiper endpoints
	gorillaRouter.HandleFunc("/gossip/push-data", bindContext(c, handleGossip)).Methods("POST")

	gorillaRouter.HandleFunc("/gossip/get-data", bindContext(c, handleGossipObjectRequest)).Methods("GET")

	// Monitor interaction endpoint
	gorillaRouter.HandleFunc("/gossip/gossip-data", bindContext(c, handleOwnerGossip)).Methods("POST")

	// Start the HTTP server.
	http.Handle("/", gorillaRouter)
	fmt.Println(util.BLUE+"Listening on port:", c.Config.Port, util.RESET)
	err := http.ListenAndServe(":"+c.Config.Port, nil)
	// We wont get here unless there's an error.
	log.Fatal("ListenAndServe: ", err)
	os.Exit(1)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the base page for the CTng gossiper.")
}

// handleGossipObjectRequest() is run when /gossip/get-data is accessed or sent a GET request.
// Expects a gossip_object identifier, and returns the gossip object if it has it, 400 otherwise.
func handleGossipObjectRequest(c *gossip.GossiperContext, w http.ResponseWriter, r *http.Request) {
	// Verify the user sent a valid gossip_object identifier.
	gossipID, err := gossipIDFromParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	gossip_obj, found := c.GetObject(gossipID)
	// If the object is not found, return a 404.
	if found {
		// Reply with error message.
		http.Error(w, "Gossip object not found.", 404)
		return
	}
	// If the object is found, return it to the requester.
	err = json.NewEncoder(w).Encode(gossip_obj)
	if err != nil {
		// Internal Server error
		http.Error(w, "Internal Server Error", 500)
	}
}

// handleGossip() is ran when POST is recieved at /gossip/push-data.
// It should verify the Gossip object and then send it to the network.
func handleGossip(c *gossip.GossiperContext, w http.ResponseWriter, r *http.Request) {
	// Parse sent object.
	// Converts JSON passed in the body of a POST to a Gossip_object.
	var gossip_obj gossip.Gossip_object
	err := json.NewDecoder(r.Body).Decode(&gossip_obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Verify the object is valid.

	err = gossip_obj.Verify(c.Config.Crypto)
	if err != nil {
		fmt.Println("Recieved invalid object from " + getSenderURL(r) + ".")
		gossip.ProcessInvalidObject(gossip_obj, err)
		http.Error(w, err.Error(), http.StatusOK)
		return
	}

	// Check for duplicate object.
	stored_obj, found := c.GetObject(gossip_obj.GetID(c.Config.Public.Period_interval))
	if found {
		// If the object is already stored, still return OK.{
		fmt.Println("Duplicate:", gossip_obj.Type, getSenderURL(r)+".")
		err := gossip.ProcessDuplicateObject(c, gossip_obj, stored_obj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusOK)
		}
		http.Error(w, "Recieved Duplicate Object.", http.StatusOK)
		return
	} else {
		fmt.Println(util.GREEN+"Recieved new, valid", gossip_obj.Type, "from "+getSenderURL(r)+".", util.RESET)
		gossip.ProcessValidObject(c, gossip_obj)
		c.SaveStorage()
	}
	http.Error(w, "Gossip object Processed.", http.StatusOK)
}

// Runs when /gossip/gossip-data is sent a POST request.
// Should verify gossip object and then send it to the network
// With the exception of not handling invalidObjects, this feels identical to gossipObject..
func handleOwnerGossip(c *gossip.GossiperContext, w http.ResponseWriter, r *http.Request) {
	var gossip_obj gossip.Gossip_object
	// Verify sender is an owner.
	if !isOwner(c.Config.Owner_URL, getSenderURL(r)) {
		http.Error(w, "Not an owner.", http.StatusForbidden)
		return
	}
	// Parses JSON from body of the request into gossip_obj
	err := json.NewDecoder(r.Body).Decode(&gossip_obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = gossip_obj.Verify(c.Config.Crypto)
	if err != nil {
		// Might not want to handle invalid object for our owner: Just warn them.
		// gossip.ProcessInvalidObject(gossip_obj, err)
		fmt.Println(util.RED+"Owner sent invalid object.", util.RESET)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	stored_obj, found := c.GetObject(gossip_obj.GetID(c.Config.Public.Period_interval))
	if found {
		// If the object is already stored, still return OK.{
		fmt.Println("Recieved duplicate object from Owner.")
		err := gossip.ProcessDuplicateObject(c, gossip_obj, stored_obj)
		if err != nil {
			http.Error(w, "Duplicate Object recieved!", http.StatusOK)
		} else {
			// TODO: understand how duplicate POM works
			http.Error(w, "error", http.StatusOK)
		}
		return

	} else {
		// Prints the body of the post request to the server console
		fmt.Println(util.GREEN+"Recieved new, valid", gossip_obj.Type, "from owner.", util.RESET)
		gossip.ProcessValidObjectFromOwner(c, gossip_obj)
		c.SaveStorage()
	}
}

func StartGossiperServer(c *gossip.GossiperContext) {
	// Check if the storage file exists in this directory
	err := c.LoadStorage()
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			// Storage File doesn't exit. Create new, empty json file.
			err = c.SaveStorage()
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	// Create the http client to be used.
	// This is thorough and allows for HTTP client configuration,
	// although we don't need it yet.
	tr := &http.Transport{}
	c.Client = &http.Client{
		Transport: tr,
	}

	// HTTP Server Loop
	handleRequests(c)
}
