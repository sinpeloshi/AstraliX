package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"astralix/core"
	"astralix/wallet"
)

func main() {
	const TotalSupply = 1000002021
	const Difficulty = 4 

	fmt.Println("--- AstraliX Network Central Node (512-bit Edition) ---")
	
	// 1. Generate a brand new Wallet from scratch
	creatorWallet := wallet.NewWallet()
	creatorAddress := creatorWallet.GetAddress()
	
	// 2. Extract the NEW Private Key
	privKeyBytes := creatorWallet.PrivateKey.D.Bytes()
	privKeyHex := hex.EncodeToString(privKeyBytes)
	
	fmt.Printf("\n⚠️ WARNING: SAVE THIS PRIVATE KEY IN A SAFE PLACE ⚠️\n")
	fmt.Printf("🔑 Private Key (Secret): %s\n", privKeyHex)
	fmt.Printf("🏦 Public Address: %s\n", creatorAddress)
	fmt.Printf("💰 Total Supply: %d AX allocated to this address.\n", TotalSupply)
	fmt.Println("-------------------------------------------------------------------\n")
	
	fmt.Println("Mining Genesis Block...")

	emptyPrevHash := strings.Repeat("0", 128)
	genesisData := fmt.Sprintf("Genesis: %d AX allocated to master wallet %s", TotalSupply, creatorAddress)

	// 3. Create a fresh Genesis Block with the current time
	genesis := &core.Block{
		Index:      0,
		Timestamp:  time.Now().Unix(),
		Data:       genesisData,
		PrevHash:   emptyPrevHash,
		Difficulty: Difficulty,
	}

	start := time.Now()
	genesis.Mine()
	elapsed := time.Since(start)

	fmt.Printf("Genesis Block Mined!\nHash: %s\nTime: %s\n", genesis.Hash, elapsed)
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(genesis)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🌐 API active. Node listening on port %s...\n", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}