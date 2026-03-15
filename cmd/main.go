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

var Blockchain []core.Block
var Mempool []core.Transaction

func main() {
	const TotalSupply = 1000002021
	const Difficulty = 4 

	fmt.Println("--- AstraliX Network Central Node ---")
	
	// Genesis Setup
	creatorAddr := "AXdc3acc7c0b91eb485d0e3bb78059bb58a3999c14b56cfe6ca0428670afc6410c"
	genesisTx := core.Transaction{Sender: "SYSTEM", Recipient: creatorAddr, Amount: TotalSupply}
	genesisTx.TxID = genesisTx.CalculateHash()

	genesis := core.Block{
		Index: 0, Timestamp: 1773561600,
		Transactions: []core.Transaction{genesisTx},
		PrevHash: strings.Repeat("0", 128),
		Difficulty: Difficulty,
	}
	genesis.Mine()
	Blockchain = append(Blockchain, genesis)

	// --- ROUTES ---

	// 1. Ver la cadena
	http.HandleFunc("/chain", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Blockchain)
	})

	// 2. Enviar Transacción (POST)
	http.HandleFunc("/transactions/new", func(w http.ResponseWriter, r *http.Request) {
		var tx core.Transaction
		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			http.Error(w, "Invalid TX", 400); return
		}
		tx.TxID = tx.CalculateHash()
		Mempool = append(Mempool, tx)
		fmt.Printf("TX Added to Mempool: %s\n", tx.TxID)
		w.WriteHeader(201)
		fmt.Fprintf(w, "Transaction pending in mempool")
	})

	// 3. MINAR (Convertir mempool en Bloque 1, 2, 3...)
	http.HandleFunc("/mine", func(w http.ResponseWriter, r *http.Request) {
		if len(Mempool) == 0 {
			http.Error(w, "No transactions to mine", 400); return
		}

		lastBlock := Blockchain[len(Blockchain)-1]
		newBlock := core.Block{
			Index:        int64(len(Blockchain)),
			Timestamp:    time.Now().Unix(),
			Transactions: Mempool,
			PrevHash:     lastBlock.Hash,
			Difficulty:   Difficulty,
		}

		fmt.Println("Mining new block...")
		newBlock.Mine()
		Blockchain = append(Blockchain, newBlock)
		Mempool = []core.Transaction{} // Limpiamos la sala de espera

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(newBlock)
	})

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	fmt.Printf("🌐 AstraliX L1 Live on port %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}
