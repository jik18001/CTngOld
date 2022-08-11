package revocator

import (
	"CTng/util"
	"github.com/Workiva/go-datastructures/bitarray"
	//"github.com/golang-collections/go-datastructures/bitarray"
	"github.com/google/certificate-transparency-go/tls"
	"github.com/google/certificate-transparency-go/x509"
)

type CRV struct {
	SRH        string
	Vector     bitarray.BitArray
	DeltaVec   bitarray.BitArray
	CASign     tls.DigitallySigned // maybe this should be in every Revocator, then we can change Revocator to
	LoggerSign tls.DigitallySigned // abstract class or something..
	Length     uint64
	Timestamp  string
}

func (crv *CRV) GetRevInfo() bitarray.BitArray {
	return crv.Vector
}

func (crv *CRV) RevokeCertificate(crt *x509.Certificate) {
	revPlace := util.FindRevokePlace(crt)
	crv.DeltaVec.SetBit(uint64(revPlace.Index)) // update DeltaVec
}

func (crv *CRV) IsRevoked(crt *x509.Certificate) (bool, error) {
	revPlace := util.FindRevokePlace(crt)
	return crv.Vector.GetBit(uint64(revPlace.Index))
}

func (crv *CRV) GetDelta() bitarray.BitArray {
	crv.Vector = crv.Vector.Or(crv.DeltaVec)
	temp := bitarray.NewBitArray(1)
	temp = temp.Or(crv.DeltaVec)
	crv.DeltaVec.Reset()
	return temp
}

func (crv *CRV) GetDeltaVector() bitarray.BitArray {
	return crv.DeltaVec
}

func (crv *CRV) GetVector() bitarray.BitArray {
	return crv.Vector
}

func (crv *CRV) CalculateChanges(deltaVec bitarray.BitArray) bitarray.BitArray {
	return crv.Vector.Or(deltaVec)
}

func (crv *CRV) UpdateChanges(deltaVec bitarray.BitArray) {
	crv.Vector = crv.Vector.Or(deltaVec)
}

func (crv *CRV) UpdateCASign(sign tls.DigitallySigned) {
	crv.CASign = sign
}

func (crv *CRV) UpdateLoggerSign(sign tls.DigitallySigned) {
	crv.LoggerSign = sign
}
