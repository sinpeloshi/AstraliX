package core

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"time"
)

type Block struct {
	Index        int64
	Timestamp    int64
	Transactions []Transaction // Now we store real data!
	PrevHash     string
	Hash         string
	Nonce        int
	Difficulty   int
}

func (b *Block) CalculateHash() string {
	// We include transactions in the hash calculation to make them immutable
	txData := fmt.Sprintf("%v", b.Transactions)
	record := fmt.Sprintf("%d%d%s%s%d", b.Index, b.Timestamp, txData, b.PrevHash, b.Nonce)
	h := sha512.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}

func (b *Block) Mine() {
	target := string(make([]byte, b.Difficulty))
	for i := 0; i < b.Difficulty; i++ {
		target = target[:i] + "0" + target[i+1:]
	}

	for {
		b.Hash = b.CalculateHash()
		if b.Hash[:b.Difficulty] == target {
			break
		}
		b.Nonce++
	}
}
