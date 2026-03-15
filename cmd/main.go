package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"astralix/core"
	"astralix/wallet"
)

func main() {
	const TotalSupply = 1000002021
	const Difficulty = 4 

	fmt.Println("--- AstraliX Network Central Node ---")
	
	creadorWallet := wallet.NewWallet()
	direccionCreador := creadorWallet.GetAddress()
	
	fmt.Printf("Dirección Oficial del Creador: %s\n", direccionCreador)
	fmt.Printf("Supply Total: %d AX asignados a esta red.\n", TotalSupply)
	fmt.Println("Minando Bloque Génesis...")

	genesis := &core.Block{
		Index:      0,
		Timestamp:  time.Now().Unix(),
		Data:       "AstraliX Genesis Block",
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

	// Capturamos el puerto que Railway nos impone, o usamos 8080 por defecto
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🌐 API de AstraliX activa. Nodo escuchando en el puerto %s...\n", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		fmt.Printf("Error iniciando el servidor: %s\n", err)
	}
}
