package ca

import (
	"CTng/config"
	"CTng/crypto"
	"CTng/gossip"
	"CTng/revocator"
	"net/http"

	"github.com/google/certificate-transparency-go/x509"
)

type CAContext struct {
	Config            *config.CA_config
	SRHs              []gossip.Gossip_object
	Revocators        []*revocator.Revocator // array of all the CRV's
	Certificates      *crypto.CertPool       //pool of all the certificates the CA generated
	IssuerCertificate x509.Certificate       // the certificate that the ca can sign on pther certifiactes with
	Request_Count     int
	Current_Period    int
	Client            *http.Client
}

type Place struct {
	Vector int
	Index  int
}
