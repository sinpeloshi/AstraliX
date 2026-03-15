package core

import "strings"

func (b *Block) Mine() {
	target := strings.Repeat("0", b.Difficulty)
	for {
		b.Hash = b.CalculateHash()
		if strings.HasPrefix(b.Hash, target) {
			break
		}
		b.Nonce++
	}
}
