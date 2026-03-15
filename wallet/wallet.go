package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"math/big"
)

type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  []byte
}

func NewWallet() *Wallet {
	private, _ := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return &Wallet{private, pubKey}
}

func (w *Wallet) GetAddress() string {
	pubKeyHash := sha512.Sum512(w.PublicKey)
	return "AX" + hex.EncodeToString(pubKeyHash[:])[:64]
}

// Sign firma el hash de una transacción con tu clave privada
func Sign(privKey *ecdsa.PrivateKey, txID string) (string, error) {
	hash, _ := hex.DecodeString(txID)
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash)
	if err != nil {
		return "", err
	}
	signature := append(r.Bytes(), s.Bytes()...)
	return hex.EncodeToString(signature), nil
}
