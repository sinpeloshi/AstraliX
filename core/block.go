package core

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
)

type Block struct {
	Index      int64
	Timestamp  int64
	Data       string
	PrevHash   string
	Hash       string
	Nonce      int
	Difficulty int
}

func (b *Block) CalculateHash() string {
	record := fmt.Sprintf("%d%d%s%s%d", b.Index, b.Timestamp, b.Data, b.PrevHash, b.Nonce)
	// ¡Upgrade a SHA-512!
	h := sha512.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}
