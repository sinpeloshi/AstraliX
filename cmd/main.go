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

	fmt.Println("--- AstraliX Network Central Node (512-bit Edition) ---")
	
	creatorAddress := "AXdc3acc7c0b91eb485d0e3bb78059bb58a3999c14b56cfe6ca0428670afc6410c"
	fmt.Printf("🏦 Master Address (Genesis): %s\n", creatorAddress)
	fmt.Println("Mining Genesis Block...")

	emptyPrevHash := strings.Repeat("0", 128)
	genesisData := fmt.Sprintf("Genesis: %d AX allocated to master wallet %s", TotalSupply, creatorAddress)

	genesis := &core.Block{
		Index:      0,
		Timestamp:  1773561600, 
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

	// Capturamos el puerto de Railway
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Fallback local
	}

	fmt.Printf("🌐 Preparing to bind API on port %s...\n", port)
	
	// Levantamos el servidor de forma que bloquee el hilo principal
	fmt.Println("⚡ API Active & Listening. Node is live 24/7.")
	err := http.ListenAndServe("0.0.0.0:"+port, nil)
	if err != nil {
		fmt.Printf("CRITICAL ERROR starting server: %v\n", err)
	}
}
