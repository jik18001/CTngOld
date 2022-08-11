package gossip

import (
	"CTng/util"
	"errors"
	"fmt"
)

// Once an object is verified, it is stored and given its neccessary data path.
// At this point, the object has not yet been stored in the database.
// What we know is that the signature is valid for the provided data.
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
		SendToOwner(c, obj)
		err = GossipData(c, obj)
	case REVOCATION_FRAG:
		err = GossipData(c, obj)
	case ACCUSATION_FRAG:
		ProcessAccusation(c, obj)
		err = GossipData(c, obj)
	case APPLICATION_POM:
		SendToOwner(c, obj)
		err = GossipData(c, obj)
	default:
		fmt.Println("Recieved unsupported object type.")
	}
	if err != nil {
		// ...
	}
}

// Once an object is verified, it is stored and given its neccessary data path.
// At this point, the object has not yet been stored in the database.
// What we know is that the signature is valid for the provided data.
func ProcessValidObjectFromOwner(c *GossiperContext, obj Gossip_object) {
	// This function is incomplete -- requires more individual object direction
	// Note: Object needs to be stored before Gossiping so it is recognized as a duplicate.
	c.StoreObject(obj)
	var err error = nil
	switch obj.Type {
	case STH:
		err = GossipData(c, obj)
	case REVOCATION:
		err = GossipData(c, obj)
	case STH_FRAG:
		err = GossipData(c, obj)
	case GOSSIP_POM:
		err = GossipData(c, obj)
	case REVOCATION_FRAG:
		err = GossipData(c, obj)
	case ACCUSATION_FRAG:
		// TODO: check that ProcessAccusation doesn't send the gossip object back to the owner
		ProcessAccusation(c, obj)
		err = GossipData(c, obj)
	case APPLICATION_POM:
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
func ProcessDuplicateObject(c *GossiperContext, obj Gossip_object, dup Gossip_object) error {
	if obj.Signature == dup.Signature &&
		obj.Payload == dup.Payload {
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
		// Send to owner.
		defer SendToOwner(c, pom)
		return errors.New("Proof of Misbehavior Generated")
	}
}

func ProcessInvalidObject(obj Gossip_object, e error) {
	// TODO:
	// 	Determine Conflict/misbehavior
	//  Send neccessary accusations
}

func ProcessAccusation(c *GossiperContext, acc Gossip_object) {
	pom, shouldGossip, err := Process_Accusation(acc, c.Accusations, c.Config.Crypto)
	if err != nil {
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
		// We do not currently gossip PoMs.
		SendToOwner(c, *pom)
	}
}
