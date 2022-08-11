package crypto

/*
Code Ownership:
Finn - Created all Functions
*/

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
)

type RsaSignatures interface {
	NewRSAPrivateKey() (*rsa.PrivateKey, error)
	GetPublicKey(privateKey *rsa.PrivateKey) (*rsa.PublicKey, error)
	Sign(msg []byte, privateKey *rsa.PrivateKey) ([]byte, error)
	//Verify returns an error if the signature couldnt be verified.
	Verify(msg []byte, signature []byte, publicKey []byte, config *CryptoConfig) error
}

// Generates a new RSA private key of length 2048 bits
// This is the length required by the CT standard, and has been adopted in our specification.
func NewRSAPrivateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// Although unsued, this function has been left for clarity
// of where the public key is located for an RSA private key.
func GetPublicKey(privateKey *rsa.PrivateKey) (*rsa.PublicKey, error) {
	return &privateKey.PublicKey, nil
}

// Sign message msg using the private key.
// Returns an RSA signature, which is an object containing the passed id
// and the signature itself. id should ALWAYS be the same as the CTngID of the signer.
func RSASign(msg []byte, privateKey *rsa.PrivateKey, id CTngID) (RSASig, error) {
	// SHA256 = Specification Requirement for RSA signatures
	hash, err := GenerateSHA256(msg)
	if err != nil {
		return RSASig{}, err
	}
	sig, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash)
	return RSASig{
		Sig: sig,
		ID:  id}, err
}

// Verifies the signature of the message against the given public key.
// Returns nil if the signature is valid.
func RSAVerify(msg []byte, signature RSASig, pub *rsa.PublicKey) error {
	hash, err := GenerateSHA256(msg)
	if err != nil {
		return err
	}
	return rsa.VerifyPKCS1v15(pub, crypto.SHA256, hash, signature.Sig)
}
