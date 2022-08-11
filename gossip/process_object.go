package gossip

/*
Code Ownership:
Finn - All Functions in this file.
*/

import (
	"CTng/util"
	"errors"
	"fmt"
)

// Once an object is verified, it is stored and given its neccessary data path.
// At this point, the object has not yet been stored in the database, but has been deemed Valid by running
// Gossip_Object.verify(). Objects passed to this function have valid signatures and aren't duplicates.
func ProcessValidObject(c *GossiperContext, obj Gossip_object) {
	// This function is incomplete -- requires more individual object direction
	// Note: Object needs to be stored before Gossiping so it is recognized as a duplicate.
	c.StoreObject(obj)
	var err error = nil
	switch obj.Type {
	case STH:
		SendToOwner(c, obj)
		err = GossipData(c, obj)
	case REVOCATION:
		SendToOwner(c, obj)
		err = GossipData(c, obj)
	case STH_FRAG:
		SendToOwner(c, obj)
		err = GossipData(c, obj)
	case GOSSIP_POM:
		c.HasPom[obj.Payload[0]] = true
		SendToOwner(c, obj)
		err = GossipData(c, obj)
	case REVOCATION_FRAG:
		err = GossipData(c, obj)
	case ACCUSATION_FRAG:
		ProcessAccusation(c, obj)
		err = GossipData(c, obj)
	case APPLICATION_POM:
		c.HasPom[obj.Payload[0]] = true
		SendToOwner(c, obj)
		err = GossipData(c, obj)
	default:
		fmt.Println("Recieved unsupported object type.")
	}
	if err != nil {
		// ...
	}
}

// Process a valid gossip object which is a duplicate to another one.
// If the signature/payload is identical, then we can safely ignore the duplicate.
// Otherwise, we generate a PoM for two objects sent in the same period.
// Warning: Currently, multiple accusations sent in the same period are deemed a Gossip_PoM.
// A check should be added for these, or the structure of the gossip object should change to accomidate this.
func ProcessDuplicateObject(c *GossiperContext, obj Gossip_object, dup Gossip_object) error {
	if obj.Signature[0] == dup.Signature[0] &&
		obj.Payload[0] == dup.Payload[0] {
		return nil
	} else {
		// Generate PoM
		pom := Gossip_object{
			Application: obj.Application,
			Type:        GOSSIP_POM,
			Signer:      obj.Signer,
			Signature:   [2]string{obj.Signature[0], dup.Signature[0]},
			Payload:     [2]string{obj.Payload[0], dup.Payload[0]},
			Timestamp:   GetCurrentTimestamp(),
		}
		c.StoreObject(pom)
		c.HasPom[obj.Payload[0]] = true
		// Currently, we don't send PoMs. but if we did, we could do it here.
		// For now, we just send it to the owner.
		defer SendToOwner(c, pom)
		return errors.New("Proof of Misbehavior Generated")
	}
}

// Function for processing objects that are deemed invalid, based on the passed error
// the given error should be one of the errors returned by gossip_object.Verify().
func ProcessInvalidObject(obj Gossip_object, e error) {
	// TODO:
	// 	Determine Conflict/misbehavior
	//  Send neccessary accusations
}

// Function for processing objects that are deemed invalid, based on the passed error
// the given error should be one of the errors returned by gossip_object.Verify().
func ProcessAccusation(c *GossiperContext, acc Gossip_object) {
	pom, shouldGossip, err := Process_Accusation(acc, c.Accusations, c.Config.Crypto)
	if err != nil && shouldGossip {
		fmt.Println(util.RED+err.Error(), util.RESET)
	} else {
		fmt.Println(util.YELLOW+"Processed accusation against", acc.Payload[0], util.RESET)
	}
	if shouldGossip {
		GossipData(c, acc)
	}
	if pom != nil {
		fmt.Println(util.RED+"Generated POM for", acc.Payload[0], util.RESET)
		c.StoreObject(*pom)
		c.HasPom[acc.Payload[0]] = true
		SendToOwner(c, *pom)
	}
}
