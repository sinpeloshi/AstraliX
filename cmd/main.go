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
const DB_FILE = "blockchain.json"

// Wallet especial que funcionará como el "Pozo" de recompensas
const REWARDS_POOL_ADDR = "AX_REWARDS_TREASURY_POOL_512_BIT_SECURE"

func loadChain() {
	file, err := os.ReadFile(DB_FILE)
	if err == nil {
		json.Unmarshal(file, &Blockchain)
		fmt.Println("📦 Base de datos cargada.")
	}
}

func saveChain() {
	data, _ := json.MarshalIndent(Blockchain, "", "  ")
	os.WriteFile(DB_FILE, data, 0644)
}

// Función para calcular balance de cualquier dirección
func getBalance(addr string) float64 {
	var bal float64
	for _, b := range Blockchain {
		for _, tx := range b.Transactions {
			if tx.Recipient == addr { bal += tx.Amount }
			if tx.Sender == addr { bal -= tx.Amount }
		}
	}
	return bal
}

func main() {
	const Difficulty = 4 
	// Tu dirección maestra
	creatorAddr := "AX5eaba583bf646e0e39f41da6f9d8fa6db929c2e858bd32dffe6ac0cee2e3e974"

	loadChain()

	if len(Blockchain) == 0 {
		// ÚNICA EMISIÓN EN LA HISTORIA: 1,000,002,021 AX
		genTx := core.Transaction{Sender: "SYSTEM", Recipient: creatorAddr, Amount: 1000002021}
		genTx.TxID = genTx.CalculateHash()
		genesis := core.Block{
			Index: 0, Timestamp: 1773561600,
			Transactions: []core.Transaction{genTx},
			PrevHash: strings.Repeat("0", 128), Difficulty: Difficulty,
		}
		genesis.Mine()
		Blockchain = append(Blockchain, genesis)
		saveChain()
	}

	// --- API ---
	http.HandleFunc("/api/balance/", func(w http.ResponseWriter, r *http.Request) {
		addr := strings.TrimPrefix(r.URL.Path, "/api/balance/")
		json.NewEncoder(w).Encode(map[string]interface{}{"balance": getBalance(addr), "address": addr})
	})

	http.HandleFunc("/api/chain", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Blockchain)
	})

	http.HandleFunc("/api/mine", func(w http.ResponseWriter, r *http.Request) {
		minerAddr := r.URL.Query().Get("address")
		if minerAddr == "" { http.Error(w, "Address req", 400); return }

		// Recompensa fija de 50 AX
		reward := 50.0
		
		// VALIDACIÓN CRÍTICA: ¿Tiene el tesoro fondos para pagar?
		treasuryBal := getBalance(REWARDS_POOL_ADDR)
		
		var blockTxs []core.Transaction
		blockTxs = append(blockTxs, Mempool...)

		if treasuryBal >= reward {
			// El pago sale del TESORO, no de SYSTEM.
			cbTx := core.Transaction{Sender: REWARDS_POOL_ADDR, Recipient: minerAddr, Amount: reward}
			cbTx.TxID = cbTx.CalculateHash()
			blockTxs = append(blockTxs, cbTx)
		} else {
			fmt.Println("⚠️ Tesorería sin fondos. Bloque minado sin recompensa.")
		}

		last := Blockchain[len(Blockchain)-1]
		newB := core.Block{
			Index: int64(len(Blockchain)), Timestamp: time.Now().Unix(),
			Transactions: blockTxs, PrevHash: last.Hash, Difficulty: Difficulty,
		}
		newB.Mine()
		Blockchain = append(Blockchain, newB)
		Mempool = []core.Transaction{}
		saveChain()
		json.NewEncoder(w).Encode(newB)
	})

	http.HandleFunc("/api/transactions/new", func(w http.ResponseWriter, r *http.Request) {
		var tx core.Transaction
		json.NewDecoder(r.Body).Decode(&tx)
		// Validación simple: ¿Tiene saldo el emisor?
		if tx.Sender != "SYSTEM" && tx.Sender != REWARDS_POOL_ADDR {
			if getBalance(tx.Sender) < tx.Amount {
				http.Error(w, "Saldo insuficiente", 400); return
			}
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
	fmt.Printf("🌐 AstraliX v13.0 FIXED SUPPLY on %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>AstraliX | Hard Cap Network</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --ax-blue: #003366; --ax-celeste: #74ACDF; --bg: #F4F7F9; }
        body { background: var(--bg); color: #334155; font-family: 'Segoe UI', sans-serif; padding-bottom: 80px; }
        .card-pro { background: #fff; border-radius: 25px; box-shadow: 0 10px 30px rgba(0,0,0,0.05); padding: 25px; margin-bottom: 20px; border: none; }
        .balance-hero { background: linear-gradient(135deg, var(--ax-blue) 0%, var(--ax-celeste) 100%); color: #fff; padding: 40px 20px; }
        .bottom-nav { background: #fff; position: fixed; bottom: 0; width: 100%; height: 75px; display: flex; border-top: 1px solid #eee; z-index: 1000; }
        .nav-item { flex: 1; text-align: center; color: #94A3B8; padding: 12px; font-size: 10px; font-weight: 700; cursor: pointer; text-decoration: none; }
        .nav-item.active { color: var(--ax-blue); }
        .nav-item i { font-size: 20px; display: block; margin-bottom: 4px; }
        .addr-pill { background: rgba(0,0,0,0.05); padding: 10px; border-radius: 12px; font-family: monospace; font-size: 11px; word-break: break-all; margin-top: 10px; }
        .pool-box { border: 2px dashed var(--ax-celeste); background: #f0f7ff; color: var(--ax-blue); font-weight: bold; }
    </style>
</head>
<body>
    <div class="container mt-4">
        <div id="v-dash" class="view">
            <div class="card-pro balance-hero text-center">
                <small class="opacity-75 fw-bold">SALDO PERSONAL</small>
                <h1 id="bal-txt" class="display-4 fw-bold my-2">0.00 AX</h1>
                <div id="addr-txt" class="addr-pill">No sincronizado</div>
            </div>

            <div class="card-pro pool-box text-center">
                <small>POZO DE RECOMPENSAS (TREASURY)</small>
                <h4 id="pool-txt" class="m-0">0.00 AX</h4>
                <div class="small mt-2 opacity-50">Carga fondos aquí para habilitar el minado</div>
                <div class="addr-pill" style="font-size: 9px">AX_REWARDS_TREASURY_POOL_512_BIT_SECURE</div>
            </div>

            <button class="btn btn-dark w-100 py-3 rounded-4 fw-bold mb-4" onclick="mine()">MINAR BLOQUE (+50 AX)</button>
            <div id="mini-feed"></div>
        </div>

        <div id="v-wallet" class="view" style="display:none">
            <div class="card-pro">
                <h4 class="fw-bold mb-4">Transferir AX</h4>
                <label class="small text-muted mb-2">Dirección Destino</label>
                <input type="text" id="tx-to" class="form-control mb-3 p-3 border-0 bg-light" placeholder="AX...">
                <label class="small text-muted mb-2">Monto</label>
                <input type="number" id="tx-amt" class="form-control mb-4 p-3 border-0 bg-light" placeholder="0.00">
                <button class="btn btn-primary w-100 py-3 border-0 rounded-4 fw-bold" style="background:var(--ax-blue)" onclick="send()">FIRMAR Y ENVIAR</button>
            </div>
        </div>

        <div id="v-sec" class="view" style="display:none">
            <div class="card-pro">
                <h4 class="fw-bold mb-4">Keys & Vault</h4>
                <input type="password" id="i-priv" class="form-control mb-3 p-3 border-0 bg-light" placeholder="Private Key">
                <button class="btn btn-dark w-100 py-3 rounded-4 fw-bold" onclick="login()">CONECTAR BILLETERA</button>
                <hr class="my-4">
                <button class="btn btn-outline-primary w-100 py-3 rounded-4" onclick="gen()">GENERAR IDENTIDAD</button>
                <div id="g-res" class="mt-4" style="display:none">
                    <label class="small fw-bold">Privada:</label><div class="addr-pill mb-2" id="g-priv"></div>
                    <label class="small fw-bold text-primary">Dirección AX:</label><div class="addr-pill fw-bold" id="g-pub"></div>
                </div>
            </div>
        </div>
    </div>

    <div class="bottom-nav">
        <div class="nav-item active" onclick="nav('dash', this)"><i class="fas fa-chart-pie"></i>Dash</div>
        <div class="nav-item" onclick="nav('wallet', this)"><i class="fas fa-paper-plane"></i>Enviar</div>
        <div class="nav-item" onclick="nav('sec', this)"><i class="fas fa-key"></i>Keys</div>
    </div>

    <script>
        async function derive(priv) {
            const buf = new TextEncoder().encode(priv);
            const hash = await crypto.subtle.digest('SHA-512', buf);
            const hex = Array.from(new Uint8Array(hash)).map(function(b) { return b.toString(16).padStart(2,'0'); }).join('');
            return 'AX' + hex.substring(0, 64);
        }

        let session = JSON.parse(localStorage.getItem('ax_argentum')) || null;

        function nav(id, el) {
            document.querySelectorAll('.view').forEach(function(v) { v.style.display = 'none'; });
            document.getElementById('v-' + id).style.display = 'block';
            document.querySelectorAll('.nav-item').forEach(function(n) { n.classList.remove('active'); });
            el.classList.add('active');
        }

        async function login() {
            const p = document.getElementById('i-priv').value;
            const pb = await derive(p);
            session = { pub: pb, priv: p };
            localStorage.setItem('ax_argentum', JSON.stringify(session));
            location.reload();
        }

        async function load() {
            if(session) {
                document.getElementById('addr-txt').innerText = session.pub.substring(0,30) + "...";
                const r = await fetch('/api/balance/' + session.pub);
                const d = await r.json();
                document.getElementById('bal-txt').innerText = d.balance.toLocaleString() + ' AX';
            }
            
            // Cargar saldo del Pozo (Treasury)
            const rp = await fetch('/api/balance/AX_REWARDS_TREASURY_POOL_512_BIT_SECURE');
            const dp = await rp.json();
            document.getElementById('pool-txt').innerText = dp.balance.toLocaleString() + ' AX';

            const res = await fetch('/api/chain');
            const chain = await res.json();
            const feed = document.getElementById('mini-feed');
            feed.innerHTML = '';
            chain.reverse().slice(0,5).forEach(function(b) {
                const h = (b.Hash || b.hash || '').substring(0,25) + '...';
                feed.innerHTML += '<div class="card-pro p-3 mb-2 small d-flex justify-content-between shadow-none border"><span>#' + b.index + '</span><span class="text-muted">' + h + '</span></div>';
            });
        }

        async function mine() {
            if(!session) return alert('Login first');
            const r = await fetch('/api/mine?address=' + session.pub);
            if(r.ok) { alert('Mined! Recompensa cobrada desde la Tesorería.'); load(); }
            else { alert('Mempool vacío o Tesorería sin fondos.'); }
        }

        async function send() {
            const tx = { sender: session.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            const r = await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            if(r.ok) { alert('Enviado! Mina un bloque para confirmar.'); nav('dash', document.querySelector('.nav-item')); load(); }
            else { alert('Error: Revisa tu saldo.'); }
        }

        async function gen() {
            const p = btoa(Math.random()).substring(0,64);
            const pb = await derive(p);
            document.getElementById('g-res').style.display = 'block';
            document.getElementById('g-priv').innerText = p;
            document.getElementById('g-pub').innerText = pb;
        }

        load(); setInterval(load, 15000);
    </script>
</body>
</html>
