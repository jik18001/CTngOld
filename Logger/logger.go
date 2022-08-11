package logger

import (
	"CTng/crypto"
	"CTng/revocator"
	"CTng/util"
	"bytes"
	"fmt"

	"github.com/google/certificate-transparency-go/tls"

	"github.com/Workiva/go-datastructures/bitarray"
)

const PROTOCOL = "http://"

type STH struct {
	Timestamp string
	RootHash  string
	TreeSize  int
}

func VerifyRevocationConsistency(logger *LoggerContext, rev_info revocator.Revocation) (bool, error) {
	var temp_revocators []*revocator.Revocator = logger.Revocators
	array_of_new_deltas := rev_info.Delta_CRV
	if len(array_of_new_deltas) > len(temp_revocators) {
		for i := 0; i < (len(array_of_new_deltas) - len(temp_revocators)); i++ {
			var new_revocator revocator.Revocator = &revocator.CRV{
				Vector:   bitarray.NewBitArray(1),
				DeltaVec: bitarray.NewBitArray(1),
				CASign: tls.DigitallySigned{
					Algorithm: tls.SignatureAndHashAlgorithm{
						Hash:      tls.SHA256,
						Signature: tls.RSA,
					},
					Signature: []byte("0"),
				},
				LoggerSign: tls.DigitallySigned{
					Algorithm: tls.SignatureAndHashAlgorithm{
						Hash:      tls.SHA256,
						Signature: tls.RSA,
					},
					Signature: []byte("0"),
				},
				Length: logger.Config.Public.Length,
			}

			temp_revocators = append(temp_revocators, &new_revocator) // new(Revocator)
		}
	}
	//add the deltas to temp_revocators
	for i, _ := range temp_revocators {
		(*temp_revocators[i]).UpdateChanges(util.BytesToBits(array_of_new_deltas[i]))
	}
	calculated_vectors_hash, err := crypto.GenerateHashOnVectors(temp_revocators)
	if err != nil {
		fmt.Println(err)
		fmt.Println("can not generate calculated hash")
		return false, err
	}
	if bytes.Compare(calculated_vectors_hash, rev_info.Vectors_Hash) == 0 {
		// update the revocators of the loggers
		logger.Revocators = temp_revocators
		return true, nil
	}
	return false, nil
}

/*
func generateSTH(loggerType int) gossip.Gossip_object {
	// Generate a random-ish STH, add to STHS.
	hashmsg := "Root Hash" + fmt.Sprint(currentPeriod+request_count)
	hash, _ := crypto.GenerateSHA256([]byte(hashmsg))
	STH1 := STH{
		Timestamp: gossip.GetCurrentTimestamp(),
		RootHash:  hex.EncodeToString(hash),
		TreeSize:  currentPeriod * 12571285,
	}
	payload, _ := json.Marshal(STH1)
	signature, _ := crypto.RSASign([]byte(payload), &config.Private, crypto.CTngID(config.Signer))
	gossipSTH := gossip.Gossip_object{
		Application: "CTng",
		Type:        gossip.STH,
		Signer:      config.Signer,
		Signature:   [2]string{signature.String(), ""},
		Timestamp:   STH1.Timestamp,
		Payload:     [2]string{string(payload), ""},
	}
	return gossipSTH
}

func PeriodicTasks(logger *LoggerContext) {
	// Immediately queue up the next task to run at next MRD
	f := func() {
		PeriodicTasks(logger)
	}
	// Run the periodic tasks.
	time.AfterFunc(time.Duration(logger.Config.Public.MRD)*time.Second, f)

	// Generate STH and FakeSTH
	fmt.Println("Running Logger Tasks")
	sth1 := generateSTH(logger.Config.LoggerType)
	logger.Request_Count++
	fakeSTH1 := generateSTH(logger.Config.LoggerType)
	logger.STHS = append(logger.STHS, sth1)
	logger.FakeSTHs = append(logger.FakeSTHs, fakeSTH1)
	logger.Current_Period++
}
*/
