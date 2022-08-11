package server

/*
Code Ownership:
Marcus
	Established all initial HTTP function declarations and basic functionality
	- initializeGossiperEndpoints
	- homePage
	- created handleGossip, handleOwnerGossip, and StartGossiperServer

Finn
	- bindContext,
	- handleGossipObjectRequest
	- revised handleGossip/ownerGossip
	- revised StartGossiperServer
*/

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
	// Error response to "/". 
	basePage()
	// Error response to "/gossiper/". Can replace with informative home page in the future.
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
	// Sends object to an owner's /monitor/recieve-gossip endpoint.
	sendToOwner()
	// Process JSON object received from HTTP POST requests.
	processData()
	// Create and gossip accusation object.
	accuseEntity()
	// TODO: Erase stored data after gossip Erase Time.
	eraseData()
	// HTTP server function which handles GET and POST requests.
	handleRequests()
}

// Binds the context to the functions we pass to the router.
// Accepts a gossiperContext object and a function that accepts a gossipercontext and 2 HTTP functions
// Returns a version of that function with the given context bound to the function and
// This is used when creating versions of functions that can be passed to a http router's HandleFunc function.
func bindContext(context *gossip.GossiperContext, fn func(context *gossip.GossiperContext, w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(context, w, r)
	}
}

//Given the a gossiperContext, initialize all neccessary HTTP endpoints and begin hosting the server.
// This run as the final step in StartGossiperServer().
func initializeGossiperEndpoints(c *gossip.GossiperContext) {
	// MUX which routes HTTP directories to functions.
	gorillaRouter := mux.NewRouter().StrictSlash(true)

	// basePage returns a 404 error as it is currently not intended to be directly accessed.
	gorillaRouter.HandleFunc("/", basePage)
	// homePage() is ran "/gossip/" is accessed, separated to allow development of an informative homepage in the future.
	// Returns a 404 error as it is currently not intended to be directly accessed.
	gorillaRouter.HandleFunc("/gossip/", homePage)

	// Inter-gossiper endpoints: Post/Get required
	gorillaRouter.HandleFunc("/gossip/push-data", bindContext(c, handleGossip)).Methods("POST")

	// For requesting a gossip object when a PULL model is being utilized.
	gorillaRouter.HandleFunc("/gossip/get-data", bindContext(c, handleGossipObjectRequest)).Methods("GET")

	// Monitor interaction endpoint
	gorillaRouter.HandleFunc("/gossip/gossip-data", bindContext(c, handleOwnerGossip)).Methods("POST")

	// Start the HTTP server.
	http.Handle("/", gorillaRouter)
	fmt.Println(util.BLUE+"Listening on port:", c.Config.Port, util.RESET)
	err := http.ListenAndServe(":"+c.Config.Port, nil)

	// We wont get here unless there's an error, ListenAndServe should run on a loop.
	log.Fatal("ListenAndServe: ", err)
	os.Exit(1)
}

// Reply with a 404 error when "/" is accessed.
func basePage(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Page not found.", http.StatusNotFound)
}

// Reply with a 404 error when "/gossip/" is accessed.
func homePage(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Page not found.", http.StatusNotFound)
}

// handleGossipObjectRequest() is run when /gossip/get-data is accessed or sent a GET request.
// Expects a gossip_object identifier, and returns the gossip object if it has it, 400 otherwise.
// The gossip_object identifier as parameters in the request.
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
		http.Error(w, "Gossip object not found.", http.StatusNotFound)
		return
	}
	// If the object is found, return it to the requester.
	err = json.NewEncoder(w).Encode(gossip_obj)
	if err != nil {
		// Internal Server error
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleGossip is ran when a POST is recieved at /gossip/push-data: the endpoint for gossipers to submit new data to the system.
// It parses the object from the request body, checks the validity of it,
// and then passes it to the appropriate "process____object" function.
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
		// Currently nothing happens when we recieve an invalid object besides writing it in our log.
		// This function is included in case this changes in the future.
		gossip.ProcessInvalidObject(gossip_obj, err)
		http.Error(w, err.Error(), http.StatusOK)
		return
	}
	// Check for a duplicate object by trying to ask the gossiper for the gossip object.
	stored_obj, found := c.GetObject(gossip_obj.GetID(c.Config.Public.Period_interval))
	if found {
		if c.Verbose {
			fmt.Println("Ignoring Duplicate ", gossip_obj.Type)
		}
		err := gossip.ProcessDuplicateObject(c, gossip_obj, stored_obj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusOK)
		}
		// If the object is duplicate, still return OK, but notify the sender.
		http.Error(w, "Recieved Duplicate Object.", http.StatusOK)
		return
	}
	// If not found, it is new.
	fmt.Println(util.GREEN+"Recieved new, valid", gossip.TypeString(gossip_obj.Type), "from "+getSenderURL(r)+".", util.RESET)
	gossip.ProcessValidObject(c, gossip_obj)
	c.SaveStorage()
	http.Error(w, "Gossip object Processed.", http.StatusOK)
}

// handleOwnerGossip runs similarly to handleGossip, but is specifically used for the owner.
// It currently makes only two changes to handleGossip: it checks that the sender is actually the owner,
// and doesn't process invalid objects. Instead, it notifies the owner that it wasn't valid.
// A new function could be created which is called by handleGossip and handleOwnerGossip to reduce redundant code,
// and move more gossiper gode into the gossip package.
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
			http.Error(w, "", http.StatusOK)
		}
		return
	} else {
		// Prints the body of the post request to the server console
		fmt.Println(util.GREEN+"Recieved new, valid", gossip.TypeString(gossip_obj.Type), "from owner.", util.RESET)
		gossip.ProcessValidObject(c, gossip_obj)
		c.SaveStorage()
	}
}

// The entrypoint into running the server: requires a gossiperContext file with a valid config,
// and generated structures stored in each of the fields (except for c.Client, which is handled here).
// to see how this object is made, check ./CTng.go.
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

	// Run the HTTP Server Loop
	initializeGossiperEndpoints(c)
}
