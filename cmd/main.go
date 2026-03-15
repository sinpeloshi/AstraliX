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
	
	// Generamos tu Billetera Maestra
	creadorWallet := wallet.NewWallet()
	direccionCreador := creadorWallet.GetAddress()
	
	// EXTRAEMOS TU CLAVE PRIVADA PARA QUE LA GUARDE
	privKeyBytes := creadorWallet.PrivateKey.D.Bytes()
	privKeyHex := hex.EncodeToString(privKeyBytes)
	
	fmt.Printf("\n⚠️ ATENCIÓN: GUARDA ESTA CLAVE PRIVADA EN UN LUGAR SEGURO ⚠️\n")
	fmt.Printf("🔑 Clave Privada (Secreta): %s\n", privKeyHex)
	fmt.Printf("🏦 Dirección Pública: %s\n", direccionCreador)
	fmt.Printf("💰 Supply Total: %d AX asignados a esta dirección.\n", TotalSupply)
	fmt.Println("-------------------------------------------------------------------\n")
	
	fmt.Println("Minando Bloque Génesis...")

	prevHashVacio := strings.Repeat("0", 128)
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

	// Ruteo dinámico de puertos para que Railway no apague el contenedor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🌐 API activa. Nodo escuchando en el puerto %s...\n", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		fmt.Printf("Error iniciando el servidor: %s\n", err)
	}
}
