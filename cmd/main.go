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
	// Dirección Maestra de 512 bits
	creatorAddr := "AXdc3acc7c0b91eb485d0e3bb78059bb58a3999c14b56cfe6ca0428670afc6410c"

	genTx := core.Transaction{Sender: "SYSTEM", Recipient: creatorAddr, Amount: 1000002021}
	genTx.TxID = genTx.CalculateHash()
	genesis := core.Block{
		Index: 0, Timestamp: 1773561600,
		Transactions: []core.Transaction{genTx},
		PrevHash: strings.Repeat("0", 128), Difficulty: Difficulty,
	}
	genesis.Mine()
	Blockchain = append(Blockchain, genesis)

	// API ENDPOINTS
	http.HandleFunc("/api/balance/", func(w http.ResponseWriter, r *http.Request) {
		addr := strings.TrimPrefix(r.URL.Path, "/api/balance/")
		var bal float64
		for _, b := range Blockchain {
			for _, tx := range b.Transactions {
				if tx.Recipient == addr { bal += tx.Amount }
				if tx.Sender == addr { bal -= tx.Amount }
			}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"balance": bal})
	})

	http.HandleFunc("/api/chain", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Blockchain)
	})

	http.HandleFunc("/api/mine", func(w http.ResponseWriter, r *http.Request) {
		if len(Mempool) == 0 { http.Error(w, "Mempool Empty", 400); return }
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

	http.HandleFunc("/api/transactions/new", func(w http.ResponseWriter, r *http.Request) {
		var tx core.Transaction
		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			http.Error(w, "Error", 400); return
		}
		tx.TxID = tx.CalculateHash()
		Mempool = append(Mempool, tx)
		w.WriteHeader(201)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, dashboardHTML)
	})

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	fmt.Printf("🌐 AstraliX Argentum Pro running on port %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>AstraliX | Premium Network OS</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --ax-blue: #003366; --ax-celeste: #74ACDF; --ax-bg: #F8FAFC; --white: #FFFFFF; }
        body { background: var(--ax-bg); color: #334155; font-family: 'Segoe UI', system-ui, sans-serif; overflow-x: hidden; }
        .sidebar { background: var(--ax-blue); height: 100vh; position: fixed; width: 260px; color: white; z-index: 1000; box-shadow: 10px 0 30px rgba(0,0,0,0.05); }
        .content { margin-left: 260px; padding: 40px; min-height: 100vh; }
        .nav-link { color: rgba(255,255,255,0.6); padding: 15px 25px; margin: 10px 15px; border-radius: 12px; font-weight: 600; cursor: pointer; display: flex; align-items: center; text-decoration: none; transition: 0.3s; }
        .nav-link:hover, .nav-link.active { background: var(--ax-celeste); color: white; }
        .nav-link i { width: 30px; font-size: 1.1rem; }
        .card-elite { background: var(--white); border: none; border-radius: 25px; box-shadow: 0 10px 30px rgba(0,0,0,0.03); padding: 30px; margin-bottom: 30px; }
        .hero-balance { background: linear-gradient(135deg, var(--ax-blue) 0%, var(--ax-celeste) 100%); color: white; }
        .btn-ax { background: var(--ax-blue); color: white; border-radius: 15px; padding: 12px 25px; font-weight: 700; border: none; transition: 0.3s; }
        .addr-pill { background: #F1F5F9; padding: 10px 15px; border-radius: 12px; font-family: monospace; font-size: 0.8rem; color: #64748B; border: 1px solid #E2E8F0; word-break: break-all; }
        @media (max-width: 992px) { .sidebar { display: none; } .content { margin-left: 0; padding: 20px; } }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="p-4 mb-4 text-center">
            <h2 class="fw-bold mb-0" style="letter-spacing: -2px; color: white;">ASTRALIX</h2>
            <small class="opacity-50 text-uppercase fw-bold" style="font-size: 10px;">Argentum 512-bit Security</small>
        </div>
        <nav>
            <div class="nav-link active" onclick="nav('dash', this)"><i class="fas fa-home"></i> Resumen</div>
            <div class="nav-link" onclick="nav('wallet', this)"><i class="fas fa-wallet"></i> Billetera</div>
            <div class="nav-link" onclick="nav('explorer', this)"><i class="fas fa-search"></i> Explorador</div>
            <div class="nav-link" onclick="nav('security', this)"><i class="fas fa-key"></i> Seguridad</div>
        </nav>
    </div>

    <div class="content">
        <div id="v-dash" class="view">
            <div class="row g-4">
                <div class="col-lg-8">
                    <div class="card-elite hero-balance">
                        <small class="text-uppercase opacity-75 fw-bold">Saldo Disponible</small>
                        <h1 id="bal-main" class="display-3 fw-bold my-3">0.00 AX</h1>
                        <div id="addr-main" class="addr-pill bg-white bg-opacity-10 border-0 text-white opacity-75">Sin sincronizar</div>
                    </div>
                    <div class="card-elite">
                        <h5 class="fw-bold mb-4">Red en Vivo</h5>
                        <div id="mini-feed"></div>
                    </div>
                </div>
                <div class="col-lg-4">
                    <div class="card-elite">
                        <h5 class="fw-bold mb-3">Acciones</h5>
                        <button class="btn btn-ax w-100 mb-3" onclick="mine()">MINAR BLOQUE</button>
                        <hr>
                        <div class="mt-3 small">
                            <div class="d-flex justify-content-between mb-1"><span>Estado</span><b class="text-success">Online</b></div>
                            <div class="d-flex justify-content-between"><span>Bloques</span><b id="h-count">0</b></div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div id="v-wallet" class="view" style="display:none">
            <div class="card-elite mx-auto" style="max-width: 500px;">
                <h4 class="fw-bold mb-4">Enviar Fondos</h4>
                <input type="text" id="tx-to" class="form-control mb-3 p-3 border-0 bg-light" placeholder="Dirección AX...">
                <input type="number" id="tx-amt" class="form-control mb-4 p-3 border-0 bg-light" placeholder="Monto">
                <button class="btn btn-ax w-100 py-3" onclick="send()">AUTORIZAR ENVÍO</button>
            </div>
        </div>

        <div id="v-explorer" class="view" style="display:none">
            <div class="card-elite">
                <h4 class="fw-bold mb-4">Explorador de la Cadena</h4>
                <div id="full-chain"></div>
            </div>
        </div>

        <div id="v-security" class="view" style="display:none">
            <div class="card-elite">
                <h4 class="fw-bold mb-4">Gestión 512-bit</h4>
                <input type="password" id="i-priv" class="form-control mb-3 p-3 border-0 bg-light" placeholder="Clave Privada">
                <button class="btn btn-ax w-100" onclick="login()">CONECTAR</button>
                <hr class="my-5">
                <button class="btn btn-outline-dark w-100" onclick="gen()">GENERAR SEMILLA</button>
                <div id="g-res" class="mt-4" style="display:none">
                    <small>Llave Privada:</small><div class="addr-pill mb-2" id="g-priv"></div>
                    <small>Dirección AX:</small><div class="addr-pill fw-bold text-primary" id="g-pub"></div>
                </div>
            </div>
        </div>
    </div>

    <script>
        async function derive(priv) {
            const buf = new TextEncoder().encode(priv);
            const hash = await crypto.subtle.digest('SHA-512', buf);
            const hex = Array.from(new Uint8Array(hash)).map(b => b.toString(16).padStart(2, '0')).join('');
            return 'AX' + hex.substring(0, 64);
        }

        let session = JSON.parse(localStorage.getItem('ax_argentum')) || null;

        function nav(id, el) {
            document.querySelectorAll('.view').forEach(v => v.style.display = 'none');
            document.getElementById('v-' + id).style.display = 'block';
            if(el) {
                document.querySelectorAll('.nav-link').forEach(n => n.classList.remove('active'));
                el.classList.add('active');
            }
        }

        async function login() {
            const p = document.getElementById('i-priv').value;
            const pb = await derive(p);
            session = { pub: pb, priv: p };
            localStorage.setItem('ax_argentum', JSON.stringify(session));
            location.reload();
        }

        async function load()
