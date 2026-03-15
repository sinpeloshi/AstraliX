package core

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
)

type Transaction struct {
	Sender    string  // Address of the sender
	Recipient string  // Address of the receiver
	Amount    float64 // Amount of AX to send
	Signature string  // Digital signature (proof of ownership)
	TxID      string  // Unique hash of this transaction
}

// CalculateHash creates a unique ID for the transaction
func (tx *Transaction) CalculateHash() string {
	record := fmt.Sprintf("%s%s%f", tx.Sender, tx.Recipient, tx.Amount)
	h := sha512.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}
