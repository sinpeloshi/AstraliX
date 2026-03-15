package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// Wallet almacena tu clave privada (secreta) y tu clave pública
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  []byte
}

// NewWallet usa criptografía de curva elíptica (igual que Bitcoin) para crear tus llaves
func NewWallet() *Wallet {
	// Generamos la clave privada matemática
	private, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	
	// Derivamos la clave pública a partir de la privada
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return &Wallet{private, pubKey}
}

// GetAddress convierte tu clave pública en una dirección legible para la red
func (w *Wallet) GetAddress() string {
	// Hasheamos la clave pública por seguridad
	pubKeyHash := sha256.Sum256(w.PublicKey)
	
	// Creamos tu dirección única agregando el prefijo "AX" de AstraliX
	address := "AX" + hex.EncodeToString(pubKeyHash[:])[:38]
	return address
}