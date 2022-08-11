package gossip

/*
Code Ownership:
Finn - GetID, GetCurrentTimestamp
Isaac - All other Functions: Gossip_object.verify() and all functions called by that.
*/

import (
	"CTng/crypto"
	"errors"
	"fmt"
	"time"
)

// Given a period interval, create a gossip object ID.
// Strips some fields from the object and rounds the timestamp.
func (g Gossip_object) GetID(period_interval int64) Gossip_object_ID {
	// Convert g.Timestamp to a time.Time
	t, err := time.Parse(time.RFC3339, g.Timestamp)
	if err != nil {
		fmt.Println(err)
	}
	// Truncate Floors the time to the nearest period interval
	rounded := t.Truncate(time.Duration(period_interval) * time.Second)
	// Construct the ID
	return Gossip_object_ID{
		Application: g.Application,
		Type:        g.Type,
		Signer:      g.Signer,
		Period:      rounded.String(),
	}
}

// GetCurrentTimestamp returns the current UTC timestamp in RFC3339 format
// This is the standard which we've decided upon in  the specs.
func GetCurrentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// Verify that the gossip_pom is actually a gossip_pom: The two signature fields are valid over the two payloads, signed by the same entity.
// Right now, this function is not working. Originally it assumed thresholdsigs were being used, but this is not true:
// RSASigs could be used when looking at STHs+revocation.
func Verify_gossip_pom(g Gossip_object, c *crypto.CryptoConfig) error {
	// For now, we assume true for now.
	return nil
	if g.Type == GOSSIP_POM {
		//gossip pom refers to Pom generated due to conflicting information
		//From Finn's gossiper design, gossip poms are defaulted to have 2 non empty fields for signature and paypload
		var err1, err2 error
		if g.Signature[0] != g.Signature[1] {
			//that means there are conflicting information
			//the PoM is valid and the verification went through.

			// Next we need to figure out what type of signature is being used.
			// First: try ThresholdSignature
			thresSig1, sigerr1 := crypto.ThresholdSigFromString(g.Signature[0])
			thresSig2, sigerr2 := crypto.ThresholdSigFromString(g.Signature[1])
			// Verify  the signatures were made successfully and no errors were thrown.
			if sigerr1 != nil || sigerr2 != nil && thresSig1.Sign != thresSig2.Sign {
				err1 = c.ThresholdVerify(g.Payload[0], thresSig1)
				err2 = c.ThresholdVerify(g.Payload[1], thresSig2)
			} else {
				// If that didnt work, try SigFragments
				fragsig1, sigerr1 := crypto.SigFragmentFromString(g.Signature[0])
				fragsig2, sigerr2 := crypto.SigFragmentFromString(g.Signature[1])
				// Verify the signatures were made successfully
				if sigerr1 != nil || sigerr2 != nil && !fragsig1.Sign.IsEqual(fragsig2.Sign) {
					err1 = c.FragmentVerify(g.Payload[0], fragsig1)
					err2 = c.FragmentVerify(g.Payload[1], fragsig2)
				} else {
					// If that didn't work, try RSASig.
					rsaSig1, sigerr1 := crypto.RSASigFromString(g.Signature[0])
					rsaSig2, sigerr2 := crypto.RSASigFromString(g.Signature[1])
					// Verify the signatures were made successfully
					if sigerr1 != nil || sigerr2 != nil {
						err1 = c.Verify([]byte(g.Payload[0]), rsaSig1)
						err2 = c.Verify([]byte(g.Payload[1]), rsaSig2)
					}
				}
			}
			if err1 == nil && err2 == nil {
				return nil
			} else {
				return errors.New("Message Signature Mismatch" + fmt.Sprint(sigerr1) + fmt.Sprint(sigerr2))
			}
		} else {
			//if signatures are the same, there are no conflicting information
			return errors.New("This is not a valid gossip pom")
		}
	}
	return errors.New("the input is not an gossip pom")
}

//verifies signature fragments match with payload
func Verify_PayloadFrag(g Gossip_object, c *crypto.CryptoConfig) error {
	if g.Signature[0] != "" && g.Payload[0] != "" {
		sig, _ := crypto.SigFragmentFromString(g.Signature[0])
		err := c.FragmentVerify(g.Payload[0], sig)
		if err != nil {
			return errors.New(No_Sig_Match)
		}
		return nil
	} else {
		return errors.New(Mislabel)
	}
}

//verifies threshold signatures match payload
func Verfiy_PayloadThreshold(g Gossip_object, c *crypto.CryptoConfig) error {
	if g.Signature[0] != "" && g.Payload[0] != "" {
		sig, _ := crypto.ThresholdSigFromString(g.Signature[0])
		err := c.ThresholdVerify(g.Payload[0], sig)
		if err != nil {
			return errors.New(No_Sig_Match)
		}
		return nil
	} else {
		return errors.New(Mislabel)
	}
}

// Verifies RSAsig matches payload, wait.... i think this just works out of the box with what we have
func Verify_RSAPayload(g Gossip_object, c *crypto.CryptoConfig) error {
	if g.Signature[0] != "" && g.Payload[0] != "" {
		sig, err := crypto.RSASigFromString(g.Signature[0])
		if err != nil {
			return errors.New(No_Sig_Match)
		}
		return c.Verify([]byte(g.Payload[0]), sig)

	} else {
		return errors.New(Mislabel)
	}
}

//Verifies Gossip object based on the type:
//STH and Revocations use RSA
//Trusted information Fragments use BLS SigFragments
//PoMs use Threshold signatures
func (g Gossip_object) Verify(c *crypto.CryptoConfig) error {
	// If everything Verified correctly, we return nil
	switch g.Type {
	case GOSSIP_POM:
		return Verify_gossip_pom(g, c)
	case STH:
		return Verify_RSAPayload(g, c)
	case REVOCATION:
		//Adding this comment here, this will check if the RSA signature on the SRH is correct, but not the hash of the CRVs within the SRH
		return Verify_RSAPayload(g, c)
	case STH_FRAG:
		return Verify_PayloadFrag(g, c)
	case REVOCATION_FRAG:
		return Verify_PayloadFrag(g, c)
	case ACCUSATION_FRAG:
		return Verify_PayloadFrag(g, c)
	case APPLICATION_POM:
		return Verfiy_PayloadThreshold(g, c)
	default:
		return errors.New(Invalid_Type)
	}
}
