package core

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
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
