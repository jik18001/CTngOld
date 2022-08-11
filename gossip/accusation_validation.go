package gossip

/*
Code Ownership
Jie - All Functions
*/

import (
	"CTng/crypto"
)

type Accusation_POM struct {
	Entity_URL   string
	Signature_TH crypto.ThresholdSig
}

//verification for PoM due to enough number of accusations happens during the PoM generation
//we are not sending the entire PoM generated by enough number of accusations as gossip object
//therefore, we will verify the combined parital signature at the local level
//This function will be invoked in the Process_accusation function in accusations.go when the number of accusations accumulated after receiving a new one exceeds the threshold
func generate_accusation_pom_from_accused_entity(Ea EntityAccusations, c *crypto.CryptoConfig) *Accusation_POM {
	pom_sig, _ := c.ThresholdAggregate(Ea.Partial_sigs)
	new_Pom02 := new(Accusation_POM)
	new_Pom02.Entity_URL = Ea.Entity_URL
	new_Pom02.Signature_TH = pom_sig
	return new_Pom02
}

//This function is invoked in Process_accusation function in accusations.go after invoking generate_accusation_pom_from_accused_entity to check the validity of the new generated PoM
func verify_accusation_pom(candidatepom Accusation_POM, c *crypto.CryptoConfig) error {
	//generate the msg and verify the signature
	return c.ThresholdVerify(candidatepom.Entity_URL, candidatepom.Signature_TH)
}
