package main

import (
	"fmt"
	"time"
	"astralix/core"
)

func main() {
	const TotalSupply = 1000002021
	const Difficulty = 4 

	fmt.Println("--- AstraliX Network Central Node ---")
	fmt.Printf("Supply Total: %d AX\n", TotalSupply)
	fmt.Println("Minando Bloque Génesis...")

	genesis := &core.Block{
		Index:      0,
		Timestamp:  time.Now().Unix(),
		Data:       "AstraliX Genesis Block - Chaco, Argentina 2026",
		PrevHash:   "0000000000000000000000000000000000000000000000000000000000000000",
		Difficulty: Difficulty,
	}

	start := time.Now()
	genesis.Mine()
	elapsed := time.Since(start)

	fmt.Printf("¡Bloque Génesis Minado!\nHash: %s\nTiempo: %s\n", genesis.Hash, elapsed)
	
	// Mantiene el nodo encendido para Railway
	select {}
}
