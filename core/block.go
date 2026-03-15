package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
)

type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  []byte
}

func NewWallet() *Wallet {
	// ¡Upgrade a la curva P-521 (Nivel Máximo del NIST)!
	private, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		panic(err)
	}
	
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return &Wallet{private, pubKey}
}

func (w *Wallet) GetAddress() string {
	// Hasheamos la pública con SHA-512
	pubKeyHash := sha512.Sum512(w.PublicKey)
	
	// Generamos la nueva dirección AstraliX ultra-segura (64 caracteres)
	address := "AX" + hex.EncodeToString(pubKeyHash[:])[:64]
	return address
}
