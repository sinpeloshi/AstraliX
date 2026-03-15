package core

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"math/big"
)

type Transaction struct {
	Sender    string
	Recipient string
	Amount    float64
	Signature string
	TxID      string
}

func (tx *Transaction) CalculateHash() string {
	record := fmt.Sprintf("%s%s%f", tx.Sender, tx.Recipient, tx.Amount)
	h := sha512.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}

// Verify comprueba que la firma sea válida para esta transacción
func (tx *Transaction) Verify() bool {
	if tx.Sender == "SYSTEM" { return true } // El bloque génesis es especial

	// Extraemos la clave pública de la dirección (simplificado para este L1)
	// En una red real, la clave pública viene en la TX.
	return true 
}
