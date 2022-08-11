package crypto

/*
Code Ownership:
Finn - All Functions
*/

import (
	"CTng/util"
	"crypto/rsa"
	"encoding/json"
	"fmt"
)

type CryptoStorage interface {
	GenerateCryptoCryptoConfigs([]CTngID, int) error
	SaveCryptoFiles(string, []CryptoConfig) error
	ReadCryptoConfig(file string) (*CryptoConfig, error)
}

// Generate a list of cryptoconfigs from a list of entity names.
// The threshold determines the k value in "k-of-n" threshold signing.
// Every entity configuration is given a bls keypair and an rsa keypair.
// Currently, entities such as Loggers and CAs (which only sign with rsa)
// need to have their keypairs manually copy-pasted into the config mappings.
func GenerateEntityCryptoConfigs(entityIDs []CTngID, threshold int) ([]CryptoConfig, error) {
	// Define some CTng Constant values
	ThresholdScheme := "bls"
	SignScheme := "rsa"
	HashScheme := SHA256
	configs := make([]CryptoConfig, len(entityIDs))

	// Generate Threshold Keys
	blsPubMap, blsPrivMap, err := GenerateThresholdKeypairs(entityIDs, threshold)
	if err != nil {
		panic(err)
	}

	// Generate RSA Keys for each entity and populate the mappings with them.
	rsaPrivMap := make(map[CTngID]rsa.PrivateKey)
	rsaPubMap := make(RSAPublicMap)
	for _, entity := range entityIDs {
		priv, err := NewRSAPrivateKey()
		if err != nil {
			panic(err)
		}
		pub := priv.PublicKey
		rsaPrivMap[entity] = *priv
		rsaPubMap[entity] = pub
	}

	//Generate configs, assigning a different entityId to each config.
	for i := range configs {
		configs[i] = CryptoConfig{
			ThresholdScheme:    ThresholdScheme,
			SignScheme:         SignScheme,
			HashScheme:         HashScheme,
			Threshold:          threshold,
			N:                  len(entityIDs),
			SelfID:             entityIDs[i],
			SignaturePublicMap: rsaPubMap,
			RSAPrivateKey:      rsaPrivMap[entityIDs[i]],
			ThresholdPublicMap: blsPubMap,
			ThresholdSecretKey: blsPrivMap[entityIDs[i]],
		}
	}
	return configs, nil
}

// Saves a list of cryptoconfigs to files in a given.
// Each file is named the SelfID of the corresponding config.
func SaveCryptoFiles(directory string, configs []CryptoConfig) error {
	for _, config := range configs {
		// Sets the name to the ID of the entity using C-like printf.
		file := fmt.Sprintf("%s/%s.test.json", directory, config.SelfID)
		err := util.WriteData(file, *NewStoredCryptoConfig(&config))
		if err != nil {
			return err
		}
	}
	return nil
}

// Read a Storedcryptoconfig from a file, convert it to a usable cryptoconfig and return a pointer to it.
func ReadCryptoConfig(file string) (*CryptoConfig, error) {
	scc := new(StoredCryptoConfig)
	bytes, err := util.ReadByte(file)
	json.Unmarshal(bytes, scc)
	if err != nil {
		return nil, err
	}
	cc, err := NewCryptoConfig(scc)
	return cc, err
}

// Converts a CryptoConfig to a marshal-able format.
// This serializes any neccessary fields in the process.
func NewStoredCryptoConfig(c *CryptoConfig) *StoredCryptoConfig {
	scc := new(StoredCryptoConfig)
	scc = &StoredCryptoConfig{
		Threshold:          c.Threshold,
		N:                  c.N,
		SignScheme:         c.SignScheme,
		ThresholdScheme:    c.ThresholdScheme,
		HashScheme:         int(c.HashScheme),
		SelfID:             c.SelfID,
		SignaturePublicMap: c.SignaturePublicMap,
		RSAPrivateKey:      c.RSAPrivateKey,
	}
	scc.ThresholdPublicMap = (&c.ThresholdPublicMap).Serialize()
	scc.ThresholdSecretKey = (&c.ThresholdSecretKey).Serialize()
	return scc
}

// Creates a cryptoconfig from a stored one.
// This is used for reading a stored file cryptoconfig.
// Deserializes any neccessary fields.
func NewCryptoConfig(scc *StoredCryptoConfig) (*CryptoConfig, error) {
	c := new(CryptoConfig)
	c = &CryptoConfig{
		Threshold:          scc.Threshold,
		N:                  scc.N,
		SignScheme:         scc.SignScheme,
		ThresholdScheme:    scc.ThresholdScheme,
		HashScheme:         HashAlgorithm(scc.HashScheme),
		SelfID:             scc.SelfID,
		RSAPrivateKey:      scc.RSAPrivateKey,
		SignaturePublicMap: scc.SignaturePublicMap,
		ThresholdPublicMap: make(BlsPublicMap),
	}
	err := (&c.ThresholdPublicMap).Deserialize(scc.ThresholdPublicMap)
	if err != nil {
		return c, err
	}
	err = (&c.ThresholdSecretKey).Deserialize(scc.ThresholdSecretKey)
	if err != nil {
		return c, err
	}
	return c, nil
}
