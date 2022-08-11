package gossip

/*
Code Ownership:
Jie - Accuse
Finn- GossipData, SendToOwner.
*/

import (
	"CTng/util"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// We've decided a gossiper should not Accuse entities. If this ever changes, however, this function still works.
func Accuse(c *GossiperContext, url string) {
	// Create a new Gossip_object for the accusation
	sig, _ := c.Config.Crypto.ThresholdSign(url)
	obj := Gossip_object{
		Application: "CTng",
		Type:        ACCUSATION_FRAG,
		Signer:      c.Config.Crypto.SelfID.String(),
		Signature:   [2]string{sig.String(), ""},
		Timestamp:   GetCurrentTimestamp(),
		Payload:     [2]string{url, ""},
	}
	// If it's not a duplicate, process it.
	// Issue: I believe this may prevent multiple accusations from occuring.
	// These kind of checks are very important, otherwise infinite loops can occur.
	if !c.IsDuplicate(obj) {
		ProcessValidObject(c, obj)
	}
}

// Sends a gossip object to all connected gossipers.
// This function assumes you are passing valid data. ALWAYS CHECK BEFORE CALLING THIS FUNCTION.
func GossipData(c *GossiperContext, gossip_obj Gossip_object) error {
	// Convert gossip object to JSON
	msg, err := json.Marshal(gossip_obj)
	if err != nil {
		return err
	}

	// Send the gossip object to all connected gossipers.
	for _, url := range c.Config.Connected_Gossipers {

		// HTTP POST the data to the url or IP address.
		resp, err := c.Client.Post("http://"+url+"/gossip/push-data", "application/json", bytes.NewBuffer(msg))
		if err != nil {
			if strings.Contains(err.Error(), "Client.Timeout") ||
				strings.Contains(err.Error(), "connection refused") {
				fmt.Println(util.RED+"Connection failed to "+url+".", util.RESET)
				// Don't accuse gossipers for inactivity.
				// defer Accuse(c, url)
			} else {
				fmt.Println(util.RED+err.Error(), "sending to "+url+".", util.RESET)
			}
			continue
		}
		// Close the response, mentioned by http.Post
		// Alernatively, we could return the response from this function.
		defer resp.Body.Close()
		if c.Verbose {
			fmt.Println("Gossiped to " + url + " and recieved " + resp.Status)
		}
	}
	return nil
}

// Sends a gossip object to the owner of the gossiper.
func SendToOwner(c *GossiperContext, obj Gossip_object) {
	// Convert gossip object to JSON
	msg, err := json.Marshal(obj)
	if err != nil {
		fmt.Println(err)
	}
	// Send the gossip object to the owner.
	resp, postErr := c.Client.Post("http://"+c.Config.Owner_URL+"/monitor/recieve-gossip", "application/json", bytes.NewBuffer(msg))
	if postErr != nil {
		fmt.Println("Error sending object to owner: " + postErr.Error())
	} else {
		// Close the response, mentioned by http.Post
		// Alernatively, we could return the response from this function.
		defer resp.Body.Close()
		if c.Verbose {
			fmt.Println("Owner responded with " + resp.Status)
		}
	}
	// Handling errors from owner could go here.
}
