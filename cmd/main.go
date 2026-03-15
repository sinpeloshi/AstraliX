package main

import (
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
	
	// Generamos tu Billetera Maestra P-521
	creadorWallet := wallet.NewWallet()
	direccionCreador := creadorWallet.GetAddress()
	
	fmt.Printf("Dirección Maestra: %s\n", direccionCreador)
	fmt.Println("Minando Bloque Génesis...")

	// El Hash previo ahora necesita 128 ceros por el SHA-512
	prevHashVacio := strings.Repeat("0", 128)

	// Asignamos el supply directamente a tu dirección en el bloque
	datosGenesis := fmt.Sprintf("Génesis: %d AX asignados a la billetera maestra %s", TotalSupply, direccionCreador)

	genesis := &core.Block{
		Index:      0,
		Timestamp:  time.Now().Unix(),
		Data:       datosGenesis,
		PrevHash:   prevHashVacio,
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🌐 API activa. Nodo Blindado escuchando en el puerto %s...\n", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		fmt.Printf("Error iniciando el servidor: %s\n", err)
	}
}
