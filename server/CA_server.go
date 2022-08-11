package server

import (
	"CTng/ca"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Binds the context to the functions we pass to the router.
func bindCAContext(context *ca.CAContext, fn func(context *ca.CAContext, w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(context, w, r)
	}
}

func handleCARequests(ca *ca.CAContext) {
	// MUX which routes HTTP directories to functions.
	gorillaRouter := mux.NewRouter().StrictSlash(true)

	// POST functions
	gorillaRouter.HandleFunc("/ca/add-certificate", bindCAContext(ca, addCert)).Methods("GET")
	gorillaRouter.HandleFunc("/ca/revoke-certificate", bindCAContext(ca, revCert)).Methods("GET")
	gorillaRouter.HandleFunc("/ca/get-revocations", bindCAContext(ca, getRevocations)).Methods("GET")
	gorillaRouter.HandleFunc("/ca/get-period", bindCAContext(ca, getCAPeriod)).Methods("GET")

	// Start the HTTP server.
	http.Handle("/", gorillaRouter)
	// Listen on port set by config until server is stopped.
	fmt.Println("Listening on port", ca.Config.Port)
	log.Fatal(http.ListenAndServe(":"+ca.Config.Port, nil))
}

func getCAPeriod(c *ca.CAContext, w http.ResponseWriter, req *http.Request) {
	fmt.Println("get CA's current period")
	// Convert array of gossip objects to JSON
	msg, err := json.Marshal(c.Current_Period)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(msg)
	return
}

func getRevocations(c *ca.CAContext, w http.ResponseWriter, req *http.Request) {
	fmt.Println("get revocations")
	startPeriod, ok := req.URL.Query()["startPeriod"]
	if !ok {
		http.Error(w, "missing start period argument", http.StatusBadRequest)
		return
	}
	numStartPeriod, err := strconv.Atoi(startPeriod[0])
	if err != nil || numStartPeriod < 0 {
		http.Error(w, "start period argument should be a positive number", http.StatusBadRequest)
		return
	}

	endPeriod, ok := req.URL.Query()["endPeriod"]
	if !ok {
		http.Error(w, "missing end period argument", http.StatusBadRequest)
		return
	}
	numEndPeriod, err := strconv.Atoi(endPeriod[0])
	if err != nil || numEndPeriod < 0 || numEndPeriod < numStartPeriod || numEndPeriod >= c.Current_Period {
		http.Error(w, "end period argument should be a positive number bigger than the start period argument", http.StatusBadRequest)
		return
	}
	srhsToSend := c.SRHs[numStartPeriod : numEndPeriod+1]
	// Convert array of gossip objects to JSON
	msg, err := json.Marshal(srhsToSend)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(msg)
	return
}

//  msg, err := json.Marshal(g)
//	resp, postErr := c.Client.Post(PROTOCOL+c.Config.Gossiper_URL+"/gossip/gossip-data", "application/json", bytes.NewBuffer(msg))
func addCert(c *ca.CAContext, w http.ResponseWriter, req *http.Request) {
	fmt.Println("add certificate")
	fmt.Println(c.Revocators[0])
	domain, ok := req.URL.Query()["domain"]
	if !ok || len(domain[0]) < 1 {
		http.Error(w, "no input", http.StatusOK)
		return
	}
	result := ca.AddCertificateToRevocator(c, domain[0])
	http.Error(w, result.Error(), http.StatusOK)
}

func revCert(c *ca.CAContext, w http.ResponseWriter, req *http.Request) {
	domain, ok := req.URL.Query()["domain"]
	if !ok || len(domain[0]) < 1 {
		fmt.Fprintf(w, "error")
		return
	}
	result := ca.RevokeCertificate(c, domain[0])
	http.Error(w, result.Error(), http.StatusOK)
}

// Run CA server
// Note that the monitor configurations must include then CA's Public key and ID as trusted
func StartCAServer(c *ca.CAContext) {
	tr := &http.Transport{}
	c.Client = &http.Client{
		Transport: tr,
	}
	// ca.AddCertificateTest(c)
	// Run a go routine to handle tasks that must occur every MRD
	go ca.PeriodicTasks(c)
	// Start HTTP server loop on the main thread
	handleCARequests(c)
}
