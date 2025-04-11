package eddsa

import (
	cryptotwistededwards "github.com/consensys/gnark-crypto/ecc/twistededwards"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/native/twistededwards"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/std/signature/eddsa"
)

// Circuit defines a EdDSA signature verification circuit
type EdDSAVerifCircuit struct {
	PublicKey eddsa.PublicKey `gnark:",public"`
	Message   frontend.Variable
	Signature eddsa.Signature `gnark:",public"`
}

// Define checks if the signature is valid for the given message and public key
func (c *EdDSAVerifCircuit) Define(api frontend.API) error {
	// Initialize edwards curve
	curve, err := twistededwards.NewEdCurve(api, cryptotwistededwards.BN254)
	if err != nil {
		return err
	}

	// Create hash function for signature verification
	hash, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}

	// Verify signature
	return eddsa.Verify(curve, c.Signature, c.Message, c.PublicKey, &hash)
}
