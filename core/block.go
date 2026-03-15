package core

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
)

type Block struct {
	Index        int64
	Timestamp    int64
	Transactions []Transaction 
	PrevHash     string
	Hash         string
	Nonce        int
	Difficulty   int
}

func (b *Block) CalculateHash() string {
	txData := fmt.Sprintf("%v", b.Transactions)
	record := fmt.Sprintf("%d%d%s%s%d", b.Index, b.Timestamp, txData, b.PrevHash, b.Nonce)
	h := sha512.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}

func (b *Block) Mine() {
	// Creamos el prefijo de ceros según la dificultad
	target := strings.Repeat("0", b.Difficulty)

	for {
		b.Hash = b.CalculateHash()
		if b.Hash[:b.Difficulty] == target {
			break
		}
		b.Nonce++
	}
}

// Necesitamos importar "strings" para el Repeat
import "strings"
