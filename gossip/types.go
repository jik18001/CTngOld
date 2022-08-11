package gossip

/*
Code Ownership:
Jie:
	- EntityAccusations and AccusationDB
Isaac:
	- Error types when parsing a gossip object
Finn:
	- Gossip_object
	- Gossip_object_ID
	- Gossip Storage
	-GossiperContext
	- All associated gossipercontext methods
*/

import (
	"CTng/config"
	"CTng/crypto"
	"CTng/util"
	"encoding/json"
	"net/http"
)

// The only valid application type
const CTNG_APPLICATION = "CTng"

// Identifiers for different types of gossip that can be sent.
const (
	GOSSIP_POM      = "http://ctng.uconn.edu/001"
	STH             = "http://ctng.uconn.edu/101"
	REVOCATION      = "http://ctng.uconn.edu/102"
	STH_FRAG        = "http://ctng.uconn.edu/201"
	REVOCATION_FRAG = "http://ctng.uconn.edu/202"
	ACCUSATION_FRAG = "http://ctng.uconn.edu/203"
	// Full support in the future for sending signatures with more than 2 signers: The current data structure doesn't support it
	STH_FULL        = "http://ctng.uconn.edu/301" // Does not need to be sent.
	REVOCATION_FULL = "http://ctng.uconn.edu/302" // Does not need to be sent
	ACCUSATION_POM  = "http://ctng.uconn.edu/303" // Does not need to be sent, only accusations do.
	APPLICATION_POM = "http://ctng.uconn.edu/304"
)

// This function prints the "name string" of each Gossip object type. It's used when printing this info to console.
func TypeString(t string) string {
	switch t {
	case GOSSIP_POM:
		return "GOSSIP_POM"
	case STH:
		return "STH"
	case REVOCATION:
		return "REVOCATION"
	case STH_FRAG:
		return "STH_FRAG"
	case REVOCATION_FRAG:
		return "REVOCATION_FRAG"
	case ACCUSATION_FRAG:
		return "ACCUSATION_FRAG"
	case STH_FULL:
		return "STH_FULL"
	case REVOCATION_FULL:
		return "REVOCATION_FULL"
	case ACCUSATION_POM:
		return "ACCUSATION_POM"
	case APPLICATION_POM:
		return "APPLICATION_POM"
	default:
		return "UNKNOWN"
	}
}

// Types of errors that can occur when parsing a Gossip_object
const (
	No_Sig_Match = "Signatures don't match"
	Mislabel     = "Fields mislabeled"
	Invalid_Type = "Invalid Type"
)

// Gossip_object representations of these types can be utilized in many places, as opposed to
// converting them back and forth from an intermediate representation.
type Gossip_object struct {
	Application string `json:"application"`
	Type        string `json:"type"`
	// Signer TODO: Figure out case of sending threshold signatures,
	// which have many signers? What will be present here?
	// Maybe scrap fully-signed signatures and just send partials for now.
	Signer string `json:"signer"`
	// Multiple types of signatures can be sent, but we need to be able to send them all as srings.
	// Therefore a signature should just be a string, which is converted to a signature based on the value of Type()
	// Multiple fields for case of PoM: if unneeded, the second field will be empty.
	// In practice, the user can completely ignore the second field.
	Signature [2]string `json:"signature"`
	// Timestamp is a UTC RFC3339 string
	Timestamp string `json:"timestamp"`
	// String-ified JSON Payloads (the string representation of the payload is what has been signed to create Signature[]).
	// 2 fields in case there are conflicting objects.
	Payload [2]string `json:"payload,omitempty"`
}

// The identifier for a Gossip Object is the (Application,Type,Signer,Period) tuple.
// Gossip_object.GetID(time Period) returns the ID of an object, accepting a period to be used for conversion.
type Gossip_object_ID struct {
	Application string `json:"application"`
	Type        string `json:"type"`
	Signer      string `json:"signer"`
	Period      string `json:"period"`
}

//Simple mapping of object IDs to objects.
type Gossip_Storage map[Gossip_object_ID]Gossip_object

// Gossiper Context
// Ths type represents the current state of a gossiper HTTP server.
// This is the state of a gossiper server. It contains:
// The gossiper Configuration,
// Storage utilized by the gossiper,
// Any objects needed throughout the gossiper's lifetime (such as the http client).
type GossiperContext struct {
	Config      *config.Gossiper_config
	Storage     *Gossip_Storage
	Accusations *AccusationDB
	StorageFile string // Where storage can be stored.
	// Client: used for HTTP connections, allows for timeouts
	// and more control over the connections we make.
	Client  *http.Client
	HasPom  map[string]bool
	Verbose bool
}

// Saves the Storage object to the value in c.StorageFile.
func (c *GossiperContext) SaveStorage() error {
	// Turn the gossipStorage into a list, and save the list.
	// This is slow as the size of the DB increases, but since we want to clear the DB each Period it will not infinitely grow..
	storageList := []Gossip_object{}
	for _, gossipObject := range *c.Storage {
		storageList = append(storageList, gossipObject)
	}
	err := util.WriteData(c.StorageFile, storageList)
	return err
}

// Read every gossip object from c.StorageFile.
// Store all files in c.Storage by their ID.
func (c *GossiperContext) LoadStorage() error {
	// Get the array that has been written to the storagefile.
	storageList := []Gossip_object{}
	period := c.Config.Public.Period_interval
	bytes, err := util.ReadByte(c.StorageFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &storageList)
	if err != nil {
		return err
	}
	// Store the objects by their ID, based on the current period defined in the gossiper context.
	// Note that if the period changes (particularly, increases) between loads of the gossiper, some objects may be overwritten/lost.
	// So careful!
	for _, gossipObject := range storageList {
		(*c.Storage)[gossipObject.GetID(period)] = gossipObject
	}
	return nil
}

// Stores an object in storage by its ID. Note that the ID utilizes Config.Public.Period_interval.
func (c *GossiperContext) StoreObject(o Gossip_object) {
	(*c.Storage)[o.GetID(c.Config.Public.Period_interval)] = o
}

// Returns 2 fields: the object, and whether or not the object was successfully found.
// If the object isn't found then all fields of the Gossip_object will also be empty.
func (c *GossiperContext) GetObject(id Gossip_object_ID) (Gossip_object, bool) {
	obj := (*c.Storage)[id]
	if obj == (Gossip_object{}) {
		return obj, false
	}
	return obj, true
}

// Given a gossip object, check if the an object with the same ID exists in the storage.
func (c *GossiperContext) IsDuplicate(g Gossip_object) bool {
	id := g.GetID(c.Config.Public.Period_interval)
	_, exists := c.GetObject(id)
	return exists
}

// Accusation DBs for Jie's Accusation functions
type EntityAccusations struct {
	Accusers     []string
	Entity_URL   string
	Num_acc      int
	Partial_sigs []crypto.SigFragment
	PoM_status   bool
}
type AccusationDB map[string]*EntityAccusations
