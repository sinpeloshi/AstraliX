package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"astralix/core"
)

func main() {
	const TotalSupply = 1000002021
	const Difficulty = 4 

	fmt.Println("--- AstraliX Network Central Node ---")
	
	// 1. Hardcode your Official Public Address forever (Genesis Wallet)
	creatorAddress := "AX1c1769524d7a291e259ee96dd11e76c76b1fdb64de732d035bbcff5bbef71471"
	
	fmt.Printf("🏦 Master Address (Genesis): %s\n", creatorAddress)
	fmt.Println("Mining Genesis Block...")

	// 128 zeros for SHA-512 Previous Hash
	emptyPrevHash := strings.Repeat("0", 128)
	
	// Genesis block data allocation
	genesisData := fmt.Sprintf("Genesis: %d AX allocated to master wallet %s", TotalSupply, creatorAddress)

	genesis := &core.Block{
		Index:      0,
		// 2. Fix the creation date (Unix Time) to make the Hash immutable
		Timestamp:  1742025600, 
		Data:       genesisData,
		PrevHash:   emptyPrevHash,
		Difficulty: Difficulty,
	}

	start := time.Now()
	genesis.Mine()
	elapsed := time.Since(start)

	fmt.Printf("Genesis Block Mined!\nHash: %s\nTime: %s\n", genesis.Hash, elapsed)
	
	// Web server to expose the blockchain state
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(genesis)
	})

	// Dynamic port routing for Railway
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🌐 API active. Node listening on port %s...\n", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}