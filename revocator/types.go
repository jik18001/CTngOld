package revocator

import (
	"github.com/google/certificate-transparency-go/x509"

	"github.com/Workiva/go-datastructures/bitarray"
	"github.com/google/certificate-transparency-go/tls"
)

type Revocator interface {
	GetRevInfo() bitarray.BitArray                                 // returns current updated revocation information (newest CRV)
	RevokeCertificate(crt *x509.Certificate)                       // set the recieved certificate as revoked
	IsRevoked(crt *x509.Certificate) (bool, error)                 //
	GetDelta() bitarray.BitArray                                   // returns the current delta vector
	CalculateChanges(deltaVec bitarray.BitArray) bitarray.BitArray // gets delta vector and returns the new vector
	UpdateChanges(deltaVec bitarray.BitArray)                      // gets delta vector and updates the current vector with it
	UpdateCASign(sign tls.DigitallySigned)                         // gets CA signature on the vector????????
	UpdateLoggerSign(sign tls.DigitallySigned)                     // gets Logger signature on the vector and saves it
	GetDeltaVector() bitarray.BitArray
	GetVector() bitarray.BitArray
}

type Revocation struct {
	Signer       string
	Delta_CRV    [][]byte
	Vectors_Hash []byte
	Timestamp    string
	Period       int
}
