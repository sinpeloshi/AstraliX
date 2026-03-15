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

// Global Blockchain state
var Blockchain []core.Block
var Mempool []core.Transaction

func main() {
	const TotalSupply = 1000002021
	const Difficulty = 4 

	fmt.Println("--- AstraliX Network Central Node ---")
	
	// 1. Setup Genesis Block
	creatorAddr := "AXdc3acc7c0b91eb485d0e3bb78059bb58a3999c14b56cfe6ca0428670afc6410c"
	
	genesisTx := core.Transaction{
		Sender:    "SYSTEM",
		Recipient: creatorAddr,
		Amount:    TotalSupply,
	}
	genesisTx.TxID = genesisTx.CalculateHash()

	genesis := core.Block{
		Index:        0,
		Timestamp:    1773561600,
		Transactions: []core.Transaction{genesisTx},
		PrevHash:     strings.Repeat("0", 128),
		Difficulty:   Difficulty,
	}
	genesis.Mine()
	Blockchain = append(Blockchain, genesis)

	fmt.Printf("Genesis Block Online: %s\n", genesis.Hash)

	// API ROUTES
	// Show the whole chain
	http.HandleFunc("/chain", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Blockchain)
	})

	// Show pending transactions
	http.HandleFunc("/mempool", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Mempool)
	})

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	fmt.Printf("🌐 Node running on port %s. Launching AstraliX L1...\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}
