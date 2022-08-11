package monitor

/*
CodeOwners:
Jie - Designed Process_valid_object
Finn - Reviewed + Revised
*/

import (
	"CTng/gossip"
	"CTng/util"
	"fmt"
	"time"
)

//This function is called by handle_gossip in monitor_server.go under the server folder
//It will be called if the gossip object is validated
func Process_valid_object(c *MonitorContext, g gossip.Gossip_object) {
	//if the valid object is from the logger in the monitor config logger URL list
	//This handles the STHS
	if IsLogger(c, g.Signer) && g.Type == gossip.STH {

		// Send an unsigned copy to the gossiper
		Send_to_gossiper(c, g)
		// The below function for creates the SIG_FRAG object
		f := func() {
			sig_frag, err := c.Config.Crypto.ThresholdSign(g.Payload[0])
			if err != nil {
				fmt.Println(err.Error())
			}
			pom_err := Check_entity_pom(c, g.Signer)
			//if there is no conflicting information/PoM send the Threshold signed version to the gossiper
			if pom_err == false {
				fmt.Println(util.BLUE, "Signing Revocation of", g.Signer, util.RESET)
				g.Type = gossip.STH_FRAG
				g.Signature[0] = sig_frag.String()
				g.Signer = c.Config.Crypto.SelfID.String()
				Send_to_gossiper(c, g)
			} else {
				fmt.Println(util.RED, "Conflicting information/PoM found, not sending STH_FRAG", util.RESET)
			}

		}
		// Delay the calling of f until gossip_wait_time has passed.
		time.AfterFunc(time.Duration(c.Config.Public.Gossip_wait_time)*time.Second, f)
		return
	}
	//if the object is from a CA, revocation information
	//this handles revocation information
	if IsAuthority(c, g.Signer) && g.Type == gossip.REVOCATION {
		sig_frag, err := c.Config.Crypto.ThresholdSign(g.Payload[0])
		if err != nil {
			fmt.Println(err.Error())
		}
		Send_to_gossiper(c, g)
		f := func() {
			fmt.Println(util.BLUE, "Signing Revocation of", g.Signer, util.RESET)
			pom_err := Check_entity_pom(c, g.Signer)
			if pom_err == false {
				g.Type = gossip.REVOCATION_FRAG
				g.Signature[0] = sig_frag.String()
				g.Signer = c.Config.Crypto.SelfID.String()
				Send_to_gossiper(c, g)
			}

		}
		time.AfterFunc(time.Duration(c.Config.Public.Gossip_wait_time)*time.Second, f)
		return

	}
	// PoMs should be noted, but currently nothing special is done besides this.
	if g.Type == gossip.ACCUSATION_POM || g.Type == gossip.GOSSIP_POM || g.Type == gossip.APPLICATION_POM {
		fmt.Println("Processing POM")
		c.StoreObject(g)
		c.HasPom[g.Payload[0]] = true
		return
	}
	//if the object is from its own gossiper
	c.StoreObject(g)
	return
}
