package core

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
)

type Block struct {
	Index        int64         `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	PrevHash     string        `json:"prev_hash"`
	Hash         string        `json:"hash"`
	Nonce        int           `json:"nonce"`
	Difficulty   int           `json:"difficulty"`
}

func (b *Block) CalculateHash() string {
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
