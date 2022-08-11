package gossip

/*
Code Ownership:
Jie - Created the function
Finn - Reviews and Edits
*/
import (
	"CTng/crypto"
	"errors"
	"fmt"
)

//This function is invoked when the gossiper receives an accusation
//It will store all the non-duplicate accusations
//it will deal with PoM generation due to threshold number of accusations by invoking generate_accusation_pom_from_accused_entity from accusation_validation.go
//if a PoM due to threshold number of accusations is generated, the PoM will be returned as a gossip_object to be sent to the monitor
func Process_Accusation(new_acc Gossip_object, accs *AccusationDB, c *crypto.CryptoConfig) (*Gossip_object, bool, error) {
	// Convert signature string
	p_sig, err := crypto.SigFragmentFromString(new_acc.Signature[0])
	if err != nil {
		fmt.Println("partial sig conversion error (from string)")
	}
	key := new_acc.Payload[0]
	if val, ok := (*accs)[key]; ok {
		//check if the accusation is a duplicate (from an existing accuser)
		for _, x := range val.Accusers {
			if x == new_acc.Signer {
				return nil, false, errors.New("Accusation already Stored")
			}
		}
		//update the number of accusations and the list of accusers
		val.Accusers = append(val.Accusers, new_acc.Signer)
		val.Num_acc++
		val.Partial_sigs = append(val.Partial_sigs, p_sig)
		//Case1: PoM has already been generated, no need to generate again, but still worth sending PoM.
		// If no PoM is to be generated, we're done here.
		if val.PoM_status == true || val.Num_acc < c.Threshold {
			return nil, true, nil
		}
		// Otherwise, generate the PoM.
		acc_pom := generate_accusation_pom_from_accused_entity(*val, c)
		// Verify the PoM has been generated successfully.
		err = verify_accusation_pom(*acc_pom, c)
		if err != nil {
			fmt.Println("Generated PoM verification failed")
			// Shouldn't happen, but we shouldn't gossip if it does.
			return nil, false, err
		}
		//Case2: PoM is generated successfully,
		//set the PoM_status to true,
		//return the updated accusationsDB,
		//the generated PoM, and true (we still want to gossip the accusation to other gossipers)
		sig_string, converr := acc_pom.Signature_TH.String()
		if converr != nil {
			fmt.Println(converr)
		}
		val.PoM_status = true
		gossip_obj := new(Gossip_object)
		*gossip_obj = Gossip_object{}
		gossip_obj.Application = "Ctng"
		gossip_obj.Type = GOSSIP_POM
		// gossip_obj.Signer Signers are contained within the signature.
		gossip_obj.Signature[0] = sig_string
		gossip_obj.Timestamp = GetCurrentTimestamp()
		gossip_obj.Payload[0] = acc_pom.Entity_URL
		return gossip_obj, true, nil
	}
	//if the entity is accused the first time
	new_entity := new(EntityAccusations)
	*new_entity = EntityAccusations{
		Accusers:     []string{new_acc.Signer},
		Entity_URL:   new_acc.Payload[0],
		Num_acc:      1,
		Partial_sigs: []crypto.SigFragment{p_sig},
		PoM_status:   false,
	}
	(*accs)[key] = new_entity
	return nil, true, nil
}
