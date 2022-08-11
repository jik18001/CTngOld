package fakeCA

/*
Code Ownership:
Isaac - Responsible for all functions
Finn - Helped with review+implementation ideas
*/
import (
	"CTng/GZip"
	"CTng/crypto"
	"CTng/gossip"
	"CTng/util"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var CA_SIZE int

type CAConfig struct {
	Signer              string
	Port                string
	NRevoke             int
	MRD                 int
	Private             rsa.PrivateKey
	CRVs                [][]byte //should be array of CRVs
	Day                 int      //I use int so I don't have to round and convert timestamps but that would be ideal
	MisbehaviorInterval int
}

type Revocation struct {
	day       int
	delta_CRV []byte
	Timestamp string
}

//Caution: this file is plagued with Global Variables for conciseness.
var config CAConfig
var SRHs []gossip.Gossip_object
var fakeSRHs []gossip.Gossip_object
var request_count int
var currentPeriod int
var caType int

func generateCRVs(CA CAConfig, miss int) gossip.Gossip_object {
	// Generate delta CRV and then compress it
	first_arr := CA.CRVs[CA.Day] //this assumes we never have CRV of len 0 (fresh CA)
	CA.Day += 1
	CA.CRVs[CA.Day] = make([]byte, CA_SIZE, CA_SIZE)

	var delta_crv = make([]byte, CA_SIZE, CA_SIZE)
	// Make the dCRV here by randomly flipping Config.NRevoke bits
	for i := 0; i < CA.NRevoke; i++ {
		change := rand.Intn(len(delta_crv))
		flip := byte(1)
		flip = flip << uint(rand.Intn(8))
		delta_crv[change] = flip
	}

	// creates the new CRV from the old one+dCRV
	for i, _ := range first_arr {
		CA.CRVs[CA.Day][i] = first_arr[i] ^ delta_crv[i]
	} //this is scuffed/slow for giant CRVs O(n), also I am assuming CRVs are same size, can modify for different sizes
	sec_arr := CA.CRVs[CA.Day]

	delta_crv = GZip.Compress(delta_crv) //should work...

	//Hash the current day CRV
	hash_CRV, err := crypto.GenerateMD5(sec_arr)
	if err != nil {
		fmt.Println("Error Hashing", err)

	}

	//we hash the delta CRV (compressed version)
	hash_dCRV, err := crypto.GenerateMD5(delta_crv)
	if err != nil {
		fmt.Println("Error Hashing", err)

	}

	//Appends byte of day, hash of CRV and hash of deltaCRV (lovely looking line of code)
	//Added (CA.Day-miss) to produce incorrect SRHs when needed
	var inter []byte = make([]byte, 0, 4096)
	sign := append(inter, byte(CA.Day-miss))
	inter = append(inter, hash_CRV...)
	inter = append(inter, hash_dCRV...)

	REV := Revocation{
		day:       CA.Day,
		delta_CRV: delta_crv,
		Timestamp: gossip.GetCurrentTimestamp(),
	}

	payload, _ := json.Marshal(REV)
	signature, _ := crypto.RSASign([]byte(sign), &CA.Private, crypto.CTngID(CA.Signer))

	gossipREV := gossip.Gossip_object{
		Application: "CTng",
		Type:        gossip.REVOCATION,
		Signer:      CA.Signer,
		Signature:   [2]string{signature.String(), ""},
		Timestamp:   REV.Timestamp,
		Payload:     [2]string{string(sign), string(payload)},
	}
	return gossipREV
}

func periodicTasks() {
	// Queue the next tasks to occur at next MRD.
	time.AfterFunc(time.Duration(config.MRD)*time.Second, periodicTasks)
	// Generate CRV and SRH
	fmt.Println("Running Tasks")
	Rev1 := generateCRVs(config, caType-request_count)
	request_count++
	fakeRev1 := generateCRVs(config, caType-request_count) //Should be incorrect SRH
	SRHs = append(SRHs, Rev1)
	fakeSRHs = append(fakeSRHs, fakeRev1)
	currentPeriod++
}

func requestSRH(w http.ResponseWriter, r *http.Request) {
	//Disconnecting CA:
	request_count++
	if caType == 3 && currentPeriod%config.MisbehaviorInterval == 0 {
		// No response or any bad request response should trigger the accusation
		return
	}
	// Split-World CA
	if caType == 2 && request_count%2 == 0 && currentPeriod%config.MisbehaviorInterval == 0 {
		json.NewEncoder(w).Encode(fakeSRHs[currentPeriod-1])
		return
	}
	json.NewEncoder(w).Encode(SRHs[currentPeriod-1])
}

func getCAType() {
	fmt.Println("What type of CA would you like to use?")
	fmt.Println("1. Normal, behaving CA (default)")
	fmt.Println("2. Split-World (Two different SRHs on every", config.MisbehaviorInterval, "MRD)")
	fmt.Println("3. Disconnecting CA (unresponsive every", config.MisbehaviorInterval, "MRD)")
	fmt.Println("4. Invalid SRH on every ", config.MisbehaviorInterval, "MRD) (CURRENTLY UNIMPLEMENTED)")
	fmt.Scanln(&caType)
}

// Runs a fake CA server with the ability to act roguely.
func RunFakeCA(configFile string) {
	// Global Variable initialization
	CA_SIZE = 1024
	caType = 1
	currentPeriod = 0
	request_count = 0
	SRHs = make([]gossip.Gossip_object, 0, 20)
	fakeSRHs = make([]gossip.Gossip_object, 0, 20)
	// Read the config file
	config = CAConfig{}
	configBytes, err := util.ReadByte(configFile)
	if err != nil {
		fmt.Println("Error reading config file: ", err)
		return
	}
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		fmt.Println("Error reading config file: ", err)
	}

	config.CRVs = make([][]byte, 999, 999)
	config.CRVs[0] = make([]byte, CA_SIZE, CA_SIZE)
	config.Day = 0
	// getCAType()
	caType = 1
	// MUX which routes HTTP directories to functions.
	gorillaRouter := mux.NewRouter().StrictSlash(true)
	gorillaRouter.HandleFunc("/ctng/v2/get-revocation", requestSRH).Methods("GET")
	http.Handle("/", gorillaRouter)
	fmt.Println("Listening on port", config.Port)
	go periodicTasks()
	http.ListenAndServe(":"+config.Port, nil)
}
