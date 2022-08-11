package server

import (
	"CTng/gossip"
	"CTng/monitor"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/gorilla/mux"
)

// Binds the context to the functions we pass to the router.
func bindMonitorContext(context *monitor.MonitorContext, fn func(context *monitor.MonitorContext, w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(context, w, r)
	}
}

func handleMonitorRequests(c *monitor.MonitorContext) {

	// MUX which routes HTTP directories to functions.
	gorillaRouter := mux.NewRouter().StrictSlash(true)

	// POST functions
	gorillaRouter.HandleFunc("/monitor/recieve-gossip", bindMonitorContext(c, handle_gossip)).Methods("POST")

	// For Relying party poms
	gorillaRouter.HandleFunc("/submit-pom", bindMonitorContext(c, receivePOM)).Methods("POST")

	// GET functions
	gorillaRouter.HandleFunc("/full-revocations", bindMonitorContext(c, getRevocationsFromMonitor)).Methods("GET")
	gorillaRouter.HandleFunc("/revocations/", bindMonitorContext(c, getRevocation)).Methods("GET")
	gorillaRouter.HandleFunc("/sths/", bindMonitorContext(c, getSTH)).Methods("GET")
	gorillaRouter.HandleFunc("/pom/", bindMonitorContext(c, getPOM)).Methods("GET")

	// Start the HTTP server.
	http.Handle("/", gorillaRouter)
	// Listen on port set by config until server is stopped.
	log.Fatal(http.ListenAndServe(":"+c.Config.Port, nil))
}

func receiveGossip(c *monitor.MonitorContext, w http.ResponseWriter, r *http.Request) {
	// Post request, parse sent object.
	body, err := ioutil.ReadAll(r.Body)

	// If there is an error, post the error and terminate.
	if err != nil {
		panic(err)
	}

	// Converts JSON passed in the body of a POST to a Gossip_object.
	var gossip_obj gossip.Gossip_object
	err = json.NewDecoder(r.Body).Decode(&gossip_obj)
	// Prints the body of the post request to the server console
	log.Println(string(body))

	// Use a mapped empty interface to store the JSON object.
	var postData map[string]interface{}
	// Decode the JSON object stored in the body
	err = json.Unmarshal(body, &postData)

	// If there is an error, post the error and terminate.
	if err != nil {
		panic(err)
	}

	// TODO - Validate, parse, and store postData
}

func receivePOM(c *monitor.MonitorContext, w http.ResponseWriter, r *http.Request) {
	// Post request, parse sent object.
	body, err := ioutil.ReadAll(r.Body)

	// If there is an error, post the error and terminate.
	if err != nil {
		panic(err)
	}

	PoM := string(body)
	fmt.Println("PoM Received: " + PoM) //temp

	// TODO - Validate, process and save PoM
}

func getRevocationsFromMonitor(c *monitor.MonitorContext, w http.ResponseWriter, r *http.Request) {

	// if no revocation data found, return a 404
	http.Error(w, "Revocation information not found.", 404)

	// if revocations found, send to requester
}

func getRevocation(c *monitor.MonitorContext, w http.ResponseWriter, r *http.Request) {

	// Get {date} from the end of the URL
	date := path.Base(r.URL.Path)
	fmt.Println(date) //temp

	// if no revocation data found for specified day, return a 404
	http.Error(w, "Revocation information not found.", 404)

	// if revocations found, send to requester
}

func getSTH(c *monitor.MonitorContext, w http.ResponseWriter, r *http.Request) {

	// Get {date} from the end of the URL
	date := path.Base(r.URL.Path)
	fmt.Println(date) //temp

	// if no STH found for specified day, return a 404
	http.Error(w, "STH object not found.", 404)

	// if STH found, send to requester
}

func getPOM(c *monitor.MonitorContext, w http.ResponseWriter, r *http.Request) {

	// Get {date} from the end of the URL
	date := path.Base(r.URL.Path)
	fmt.Println(date)

	// if no POM found for specified day, return a 404
	http.Error(w, "PoM not found.", 404)

	// if POM found, send to requester
}

func handle_gossip(c *monitor.MonitorContext, w http.ResponseWriter, r *http.Request) {
	// Parse sent object.
	// Converts JSON passed in the body of a POST to a Gossip_object.
	var gossip_obj gossip.Gossip_object
	err := json.NewDecoder(r.Body).Decode(&gossip_obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	// Verify the object is valid.
	err = gossip_obj.Verify(c.Config.Crypto)
	if err != nil {
		fmt.Println("Recieved invalid object from " + getSenderURL(r) + ".")
		monitor.AccuseEntity(c, gossip_obj.Signer)
		http.Error(w, err.Error(), http.StatusOK)
		return
	}
	// Check for duplicate object.
	_, found := c.GetObject(gossip_obj.GetID(int64(c.Config.Public.Gossip_wait_time)))
	if found {
		// If the object is already stored, still return OK.{
		fmt.Println("Duplicate:", gossip_obj.Type, getSenderURL(r)+".")
		http.Error(w, "Gossip object already stored.", http.StatusOK)
		// processDuplicateObject(c, gossip_obj, stored_obj)
		return
	} else {
		fmt.Println("Recieved new, valid", gossip_obj.Type, "from "+getSenderURL(r)+".")
		monitor.Process_valid_object(c, gossip_obj)
		c.SaveStorage()
	}
	http.Error(w, "Gossip object Processed.", http.StatusOK)
}

func StartMonitorServer(c *monitor.MonitorContext) {
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
	tr := &http.Transport{}
	c.Client = &http.Client{
		Transport: tr,
	}
	// Run a go routine to handle tasks that must occur every MMD
	go monitor.PeriodicTasks(c)
	// Start HTTP server loop on the main thread
	handleMonitorRequests(c)
}
