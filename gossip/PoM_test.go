package gossip

// import (
// 	"CTng/crypto"
// 	"testing"
// )

// func TestAccusationUpdates(t *testing.T) {
// 	println("test start\n")
// 	var accs AccusationDB
// 	var cp1 Accusation
// 	var cp2 Accusation
// 	var cp3 Accusation
// 	var cp4 Accusation
// 	cp1.Accuser = "monitor_01_DNS_place_holder"
// 	cp1.Entity_URL = "logger_01_DNS_place_holder"
// 	cp1.Partial_sig = crypto.SigFragment{}
// 	cp2.Accuser = "monitor_02_DNS_place_holder"
// 	cp2.Entity_URL = "logger_02_DNS_place_holder"
// 	cp2.Partial_sig = crypto.SigFragment{}
// 	cp3.Accuser = "monitor_01_DNS_place_holder"
// 	cp3.Entity_URL = "logger_02_DNS_place_holder"
// 	cp3.Partial_sig = crypto.SigFragment{}
// 	cp4.Accuser = "monitor_02_DNS_place_holder"
// 	cp4.Entity_URL = "logger_01_DNS_place_holder"
// 	cp4.Partial_sig = crypto.SigFragment{}
// 	println("data structures initialized")
// 	accs = Process_Accusation(cp1, accs)
// 	accs = Process_Accusation(cp1, accs)
// 	accs = Process_Accusation(cp2, accs)
// 	accs = Process_Accusation(cp3, accs)
// 	println("Accsuation processing went through")
// 	println("_______________________________________")
// 	println("start listing accused entities and the number of accusations")
// 	for i := 0; i < len(accs); i++ {
// 		println("\nThe accused entity DNS is: ", accs[i].Entity_URL)
// 		println("The total number of accusations is", accs[i].Num_acc)
// 		for j := 0; j < len(accs[i].Accusers); j++ {
// 			print("Accusers No.", j+1, "is ")
// 			println(accs[i].Accusers[j])
// 			// print("Partial sigature No.", j+1, "is ")
// 			// println(accs[i].Partial_sigs[j])
// 		}
// 	}

// }
