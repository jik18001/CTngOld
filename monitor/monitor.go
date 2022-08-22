package monitor

/*
Code Owners
Marcus
	- Initial QueryLoggers and QueryAuthorities functions
	-
Finn
	- Revisions and comments to QueryLoggers + QueryAuthorities
	- AccuseEntity
	- IsLogger + IsAuthority
	- PeriodicTasks
Jie
	- Send_to_gossiper
*/

import (
	"CTng/gossip"
	"CTng/util"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const PROTOCOL = "http://"

// Queries all loggers for their STH.
// Currently, will only grab the latest STH, as our fakeLogger doesn't have date-handling implemented.
// It also Accuses each logger if it doesnt give correct data or any issues occur.
// Obtaining certificate entries from a logger is unimplemented and left as an exercise for the reader ;)
func QueryLoggers(c *MonitorContext) {
	for _, logger := range c.Config.Logger_URLs {

		// For when dates are possible: Get today's STH from logger.
		// Get today's date in format YYYY-MM-DD
		// (Used when querying individual days)
		// var today = time.Now().UTC().Format(time.RFC3339)[0:10]

		sthResp, err := http.Get(PROTOCOL + logger + "/ctng/v2/get-sth/")
		if err != nil {
			log.Println(util.RED+"Query Logger Failed: "+err.Error(), util.RESET)
			AccuseEntity(c, logger)
			continue
		}

		sthBody, err := ioutil.ReadAll(sthResp.Body)
		var STH gossip.Gossip_object
		err = json.Unmarshal(sthBody, &STH)
		if err != nil {
			log.Println(util.RED+err.Error(), util.RESET)
			AccuseEntity(c, logger)
			continue
		}
		err = STH.Verify(c.Config.Crypto)
		if err != nil {
			log.Println(util.RED+"STH signature verification failed", err.Error(), util.RESET)
			AccuseEntity(c, logger)
		} else {

			Process_valid_object(c, STH)
		}
		// Get today's entries from logger. Currently unimplemented in both storage + executation.
		// entriesResp, err := http.Get(logger + "/ctng/v1/get-entries/" + today)
		// if err != nil {
		// 	log.Println(util.RED+err.Error(), util.RESET)
		// }

		// entiresBody, err := ioutil.ReadAll(entriesResp.Body)
		// if err != nil {
		// 	log.Println(util.RED+err.Error(), util.RESET)
		// }
		// entries := string(entiresBody)
		// fmt.Printf("Entries from logger " + logger + ": " + entries + "\n") //temp
	}

}

// Queries CAs for revocation information
// The revocation datapath hasn't been very fleshed out currently, nor has this function.
func QueryAuthorities(c *MonitorContext) {
	for _, CA := range c.Config.CA_URLs {

		// Get today's revocation information from CA.
		// Get today's date in format YYYY-MM-DD
		// var today = time.Now().UTC().Format(time.RFC3339)[0:10]

		revResp, err := http.Get(PROTOCOL + CA + "/ctng/v2/get-revocations/")
		if err != nil {
			log.Println(util.RED+"Query CA failed: "+err.Error(), util.RESET)
			AccuseEntity(c, CA)
			continue
		}

		revBody, err := ioutil.ReadAll(revResp.Body)
		if err != nil {
			log.Println(util.RED+err.Error(), util.RESET)
			AccuseEntity(c, CA)
		}
		rev := string(revBody)
		fmt.Println("Revocation information from CA " + CA + ": " + rev + "\n")

		// TODO - process revocation data
		// Our plan was to have the SRH in payload[0] of the object and the dCRV in payload[1].
		// Thus, it will pass the RSA gossip object verification
		// and later functions can verify the SRH is actually accurate.
		// Some of these design decisions exist in fakeCA.go, but nowhere else in the code.
		var REV gossip.Gossip_object
		err = json.Unmarshal(revBody, &REV)
		if err != nil {
			log.Println(util.RED+err.Error(), util.RESET)
			AccuseEntity(c, CA)
			continue
		}
		err = REV.Verify(c.Config.Crypto)
		if err != nil {
			log.Println(util.RED+"Revocation information signature verification failed", err.Error(), util.RESET)
			AccuseEntity(c, CA)
		} else {
			Process_valid_object(c, REV)
		}
	}

}

//This function accuses the entity if the domain name is provided
//It is called when the gossip object received is not valid, or the monitor didn't get response when querying the logger or the CA
//Accused = Domain name of the accused entity (logger etc.)
func AccuseEntity(c *MonitorContext, Accused string) {
	// verify we havent accused during this MMD
	// if we have, dont accuse again. This is a temporary fix.
	// Currently, if two accusations are sent in the same MMD is is caught as "conflicting objects",
	// and they cannot be stored in the gossip_object database.
	// If we have a PoM for this party we also shouldn't accuse again.
	if (c.HasAccused!= nil||c.HasPom!=nil){
	if (c.HasAccused[Accused] || c.HasPom[Accused]){
		fmt.Println(util.BLUE+"Entity has accusations/pom on file."+util.RESET)
		// Could 'queue up' the accusation to try again next period with this line of code:
		// time.AfterFunc(time.Duration(c.Config.Public.MMD)*time.Second, func() { AccuseEntity(c, Accused) })
		// for now, to prevent flooding with accusations, we will just not accuse again.
		return
	}}

	msg := Accused
	signature, _ := c.Config.Crypto.ThresholdSign(msg)
	var sigarray [2]string
	sigarray[0] = signature.String()
	sigarray[1] = ""
	var payloadarray [2]string
	payloadarray[0] = msg
	payloadarray[1] = ""
	accusation := gossip.Gossip_object{
		Application: "CTng",
		Type:        gossip.ACCUSATION_FRAG,
		Signer:      c.Config.Crypto.SelfID.String(),
		Signature:   sigarray,
		Timestamp:   gossip.GetCurrentTimestamp(),
		Payload:     payloadarray,
	}
	c.HasAccused[Accused] = true
	fmt.Println(util.BLUE+"New accusation generated, Sending to gossiper"+util.RESET)
	Send_to_gossiper(c, accusation)
}

//Send the input gossip object to its gossiper
func Send_to_gossiper(c *MonitorContext, g gossip.Gossip_object) {
	// Convert gossip object to JSON
	msg, err := json.Marshal(g)
	if err != nil {
		fmt.Println(err)
	}
	// Send the gossip object to the gossiper.
	resp, postErr := c.Client.Post(PROTOCOL+c.Config.Gossiper_URL+"/gossip/gossip-data", "application/json", bytes.NewBuffer(msg))
	if postErr != nil {
		fmt.Println(util.RED+"Error sending object to Gossiper: ", postErr.Error(),util.RESET)
	} else {
		// Close the response, mentioned by http.Post
		// Alernatively, we could return the response from this function.
		defer resp.Body.Close()
		fmt.Println(util.BLUE+"Sent", gossip.TypeString(g.Type), "to Gossiper, Recieved "+resp.Status, util.RESET)
	}

}

//this function takes the name of the entity as input and check if there is a POM against it
//this should be invoked after the monitor receives the information from its loggers and CAs prior to threshold signning it
func Check_entity_pom(c *MonitorContext, name string) bool {
	return c.HasPom[name]
}

func IsLogger(c *MonitorContext, loggerURL string) bool {
	for _, url := range c.Config.Public.All_Logger_URLs {
		if url == loggerURL {
			return true
		}
	}
	return false
}

func IsAuthority(c *MonitorContext, authURL string) bool {
	for _, url := range c.Config.Public.All_CA_URLs {
		if url == authURL {
			return true
		}
	}
	return false
}

func PeriodicTasks(c *MonitorContext) {
	// Immediately queue up the next task to run at next MMD.
	// Doing this first means: no matter how long the rest of the function takes,
	// the next call will always occur after the correct amount of time.
	f := func() {
		PeriodicTasks(c)
	}
	time.AfterFunc(time.Duration(c.Config.Public.MMD)*time.Second, f)

	// Reset accusations
	/*
	for k := range c.HasAccused {
		delete(c.HasAccused, k)
	}
	*/
	// Run the periodic tasks.
	fmt.Println(util.GREEN + "Querying Loggers" + util.RESET)
	QueryLoggers(c)
    fmt.Println(util.GREEN + "Querying CAs" + util.RESET)
	QueryAuthorities(c)

	c.SaveStorage()
	// TODO: Switch storage directory to a new folder for the next day's STHs.
	// However, we also still need to accept STH_FULL and REV_FULL for the previous day's data. maybe we need storage for the previous day too?
	// not sure.

}
