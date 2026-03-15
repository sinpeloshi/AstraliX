package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
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
	h := sha256.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}
