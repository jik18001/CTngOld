package server

import (
	"CTng/gossip"
	"CTng/logger"
	"CTng/revocator"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Binds the context to the functions we pass to the router.
func bindLoggerContext(context *logger.LoggerContext, fn func(context *logger.LoggerContext, w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(context, w, r)
	}
}

func handleLoggerRequests(logger *logger.LoggerContext) {
	// MUX which routes HTTP directories to functions.
	gorillaRouter := mux.NewRouter().StrictSlash(true)

	// POST functions
	gorillaRouter.HandleFunc("/logger/receive-srh", bindLoggerContext(logger, recieveSRH)).Methods("POST")
	gorillaRouter.HandleFunc("/logger/get-srh", bindLoggerContext(logger, getSRH)).Methods("GET")

	// Start the HTTP server.
	http.Handle("/", gorillaRouter)
	// Listen on port set by config until server is stopped.
	fmt.Println("Listening on port", logger.Config.Port)
	log.Fatal(http.ListenAndServe(":"+logger.Config.Port, nil))
}

func getSRH(logger *logger.LoggerContext, w http.ResponseWriter, r *http.Request) {
	fmt.Println("getSRH still not implemented")
	return
}

func recieveSRH(loggerContext *logger.LoggerContext, w http.ResponseWriter, r *http.Request) {
	/// Parse sent object.
	// Converts JSON passed in the body of a POST to a Gossip_object.
	var gossip_obj gossip.Gossip_object
	err := json.NewDecoder(r.Body).Decode(&gossip_obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		fmt.Println(err.Error())
		return
	}
	//  Verifies RSAsig matches payload
	err = gossip_obj.Verify(loggerContext.Config.Crypto)
	if err != nil {
		fmt.Println("Recieved invalid object from " + getSenderURL(r) + ".")
		http.Error(w, err.Error(), http.StatusOK)
		return
	}
	fmt.Println("The signature is valid")
	loggerContext.SRHs = append(loggerContext.SRHs, gossip_obj)
	json_rev_info := gossip_obj.Payload[0]
	var rev_info revocator.Revocation
	json.Unmarshal([]byte(json_rev_info), &rev_info)
	fmt.Println(rev_info)

	// vertify consistency of the deltas
	// add all delta to the current crv vectors
	// generate hash on the the union of the vetors
	// compare the result of the hash to the hash was sent by the ca (rev_info.Vectors_Hash)
	valid, err := logger.VerifyRevocationConsistency(loggerContext, rev_info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		fmt.Println(err.Error())
		return
	}
	if valid == true {
		fmt.Println("valid SRH")
		http.Error(w, "Recieved SRH", http.StatusOK)
		return
	} else {
		fmt.Println("unvalid SRH")
		http.Error(w, "Recieved unvalid SRH", http.StatusForbidden)
	}
}

// Run Logger server
// Note that the monitor configurations must include then Logger's Public key and ID as trusted
func StartLoggerServer(logger *logger.LoggerContext) {
	tr := &http.Transport{}
	logger.Client = &http.Client{
		Transport: tr,
	}
	// Run a go routine to handle tasks that must occur every MRD
	// go logger.PeriodicTasks(logger)
	// Start HTTP server loop on the main thread
	handleLoggerRequests(logger)
}
