package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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
		Data:       "AstraliX Genesis Block", // <-- Texto limpio aquí
		PrevHash:   "0000000000000000000000000000000000000000000000000000000000000000",
		Difficulty: Difficulty,
	}

	start := time.Now()
	genesis.Mine()
	elapsed := time.Since(start)

	fmt.Printf("¡Bloque Génesis Minado!\nHash: %s\nTiempo: %s\n", genesis.Hash, elapsed)
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(genesis)
	})

	fmt.Println("🌐 API de AstraliX activa. Nodo escuchando en el puerto 8080...")
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		fmt.Printf("Error iniciando el servidor: %s\n", err)
	}
}
