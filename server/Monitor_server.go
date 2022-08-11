package server

/*
Code Ownership
Marcus
	- initializeMonitorEndpoints
	- Stubs for all functions + design
	- StartMonitorServer functionality
Finn
	- bindMonitorContext
Jie
	- handle_gossip
	- revised StartMonitorServer
*/

import (
	"CTng/gossip"
	"CTng/monitor"
	"CTng/util"
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
// Accepts a gossiperContext object and a function that accepts a gossipercontext and 2 HTTP functions
// Returns a version of that function with the given context bound to the function and
// This is used when creating versions of functions that can be passed to a http router's HandleFunc function.
func bindMonitorContext(context *monitor.MonitorContext, fn func(context *monitor.MonitorContext, w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(context, w, r)
	}
}

// Given a monitor context, binds the context to each endpoint and begins running HTTP files
// Note that this OR the periodic tasks of the monitor need to occur in their own Goroutine(thread)
// so they can occur ininterrupted.
func initializeMonitorEndpoints(c *monitor.MonitorContext) {

	// MUX which routes HTTP directories to functions.
	gorillaRouter := mux.NewRouter().StrictSlash(true)

	// POST functions
	gorillaRouter.HandleFunc("/monitor/recieve-gossip", bindMonitorContext(c, handle_gossip)).Methods("POST")

	// For recieving PoMs from a relying party that identifies conflicting objects.
	gorillaRouter.HandleFunc("/submit-pom", bindMonitorContext(c, receivePOM)).Methods("POST")

	// GET functions
	gorillaRouter.HandleFunc("/full-revocations", bindMonitorContext(c, getRevocations)).Methods("GET")
	gorillaRouter.HandleFunc("/revocations/", bindMonitorContext(c, getRevocation)).Methods("GET")
	gorillaRouter.HandleFunc("/sths/", bindMonitorContext(c, getSTH)).Methods("GET")
	gorillaRouter.HandleFunc("/pom/", bindMonitorContext(c, getPOM)).Methods("GET")

	// Start the HTTP server.
	http.Handle("/", gorillaRouter)
	// Listen on port set by config until server is stopped.
	log.Fatal(http.ListenAndServe(":"+c.Config.Port, nil))
}

//This function handles gossip object received by the monitor
//Note: This function does not handle inactive loggers/CAs
//see monitor folder for handling inactive loggers/CAs
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
		fmt.Println(util.RED+"Recieved invalid object from "+getSenderURL(r)+".", util.RESET)
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
		fmt.Println("Recieved new, valid", gossip.TypeString(gossip_obj.Type), "from "+getSenderURL(r)+".")
		monitor.Process_valid_object(c, gossip_obj)
		c.SaveStorage()
	}
	http.Error(w, "Gossip object Processed.", http.StatusOK)
}

//Recieve a PoM from a relying party. Should be packaged as a Gossip_PoM Gossip object in the body of the request.
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

// Response to requests for fully-signed revocations for all CAs.
func getRevocations(c *monitor.MonitorContext, w http.ResponseWriter, r *http.Request) {

	// if no revocation data found, return a 404
	http.Error(w, "Revocation information not found.", 404)

	// if revocations found, send to requester
}

// Response to requests for fully-signed for all CAs for a given day.
func getRevocation(c *monitor.MonitorContext, w http.ResponseWriter, r *http.Request) {

	// Get {date} from the end of the URL
	date := path.Base(r.URL.Path)
	fmt.Println(date) //temp

	// if no revocation data found for specified day, return a 404
	http.Error(w, "Revocation information not found.", 404)

	//TODO: if REVOCATION_FULL found, send to requester
}

// Response to requests for fully-signed sths for all loggers for a given day.
func getSTH(c *monitor.MonitorContext, w http.ResponseWriter, r *http.Request) {

	// Get {date} from the end of the URL
	date := path.Base(r.URL.Path)
	fmt.Println(date) //temp

	// if no STH found for specified day, return a 404
	http.Error(w, "STH object not found.", 404)

	// TODO: if STH_FULL found, send them to requester.
}

// Response to requests for PoMs of entities for a given day.
func getPOM(c *monitor.MonitorContext, w http.ResponseWriter, r *http.Request) {

	// Get {date} from the end of the URL
	date := path.Base(r.URL.Path)
	fmt.Println(date)

	// if no POM found for specified day, return a 404
	http.Error(w, "PoM not found.", 404)

	// if POMs found, send to requester
}

// Exported function for starting the monitor
// Similar to StartGossiperServer, all fields must have initialized values except c.client.
// This function begins the periodic tasks of a monitor and starts the outward-facing server of that monitor.
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
	initializeMonitorEndpoints(c)
}
