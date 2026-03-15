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
	const Difficulty = 4 
	creatorAddr := "AXdc3acc7c0b91eb485d0e3bb78059bb58a3999c14b56cfe6ca0428670afc6410c"

	// Setup Genesis
	genTx := core.Transaction{Sender: "SYSTEM", Recipient: creatorAddr, Amount: 1000002021}
	genTx.TxID = genTx.CalculateHash()
	genesis := core.Block{
		Index: 0, Timestamp: 1773561600,
		Transactions: []core.Transaction{genTx},
		PrevHash: strings.Repeat("0", 128), Difficulty: Difficulty,
	}
	genesis.Mine()
	Blockchain = append(Blockchain, genesis)

	// API
	http.HandleFunc("/api/chain", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Blockchain)
	})

	http.HandleFunc("/api/mine", func(w http.ResponseWriter, r *http.Request) {
		if len(Mempool) == 0 { http.Error(w, "Empty", 400); return }
		last := Blockchain[len(Blockchain)-1]
		newB := core.Block{
			Index: int64(len(Blockchain)), Timestamp: time.Now().Unix(),
			Transactions: Mempool, PrevHash: last.Hash, Difficulty: Difficulty,
		}
		newB.Mine()
		Blockchain = append(Blockchain, newB)
		Mempool = []core.Transaction{}
		json.NewEncoder(w).Encode(newB)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<h1>AstraliX Node Online</h1><p>Check <a href='/api/chain'>/api/chain</a></p>")
	})

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	fmt.Printf("🌐 AstraliX Live on port %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}
