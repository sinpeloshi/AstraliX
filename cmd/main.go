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

// Dirección de Tesorería para Recompensas
const TREASURY_ADDR = "AX_TREASURY_REWARDS_POOL_SYSTEM_512"

func loadChain() {
	file, err := os.ReadFile(DB_FILE)
	if err == nil { json.Unmarshal(file, &Blockchain) }
}

func saveChain() {
	data, _ := json.MarshalIndent(Blockchain, "", "  ")
	os.WriteFile(DB_FILE, data, 0644)
}

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
	creatorAddr := "AX5eaba583bf646e0e39f41da6f9d8fa6db929c2e858bd32dffe6ac0cee2e3e974"

	loadChain()

	if len(Blockchain) == 0 {
		// ÚNICA EMISIÓN EN LA HISTORIA
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

	http.HandleFunc("/api/balance/", func(w http.ResponseWriter, r *http.Request) {
		addr := strings.TrimPrefix(r.URL.Path, "/api/balance/")
		json.NewEncoder(w).Encode(map[string]interface{}{"balance": getBalance(addr)})
	})

	http.HandleFunc("/api/chain", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Blockchain)
	})

	http.HandleFunc("/api/mine", func(w http.ResponseWriter, r *http.Request) {
		miner := r.URL.Query().Get("address")
		if miner == "" || len(Mempool) == 0 { http.Error(w, "Error", 400); return }

		reward := 50.0
		treasuryBal := getBalance(TREASURY_ADDR)
		
		var txs []core.Transaction
		txs = append(txs, Mempool...)

		if treasuryBal >= reward {
			cbTx := core.Transaction{Sender: TREASURY_ADDR, Recipient: miner, Amount: reward}
			cbTx.TxID = cbTx.CalculateHash()
			txs = append(txs, cbTx)
		}

		last := Blockchain[len(Blockchain)-1]
		newB := core.Block{
			Index: int64(len(Blockchain)), Timestamp: time.Now().Unix(),
			Transactions: txs, PrevHash: last.Hash, Difficulty: Difficulty,
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
		if tx.Sender != "SYSTEM" && tx.Sender != TREASURY_ADDR && getBalance(tx.Sender) < tx.Amount {
			http.Error(w, "No funds", 400); return
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
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>AstraliX | Enterprise Asset OS</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --ax-blue: #003366; --ax-celeste: #74ACDF; --bg: #F8FAFC; --white: #FFFFFF; }
        body { background: var(--bg); color: #334155; font-family: 'Segoe UI', sans-serif; margin: 0; padding-bottom: 90px; }
        
        /* Layout Híbrido */
        .sidebar { background: var(--ax-blue); height: 100vh; position: fixed; width: 280px; color: white; z-index: 1000; box-shadow: 10px 0 30px rgba(0,0,0,0.05); }
        .main-content { margin-left: 280px; padding: 40px; min-height: 100vh; transition: 0.3s; }
        
        .nav-link-ax { color: rgba(255,255,255,0.6); padding: 15px 25px; margin: 10px 15px; border-radius: 12px; font-weight: 600; cursor: pointer; display: flex; align-items: center; text-decoration: none; transition: 0.3s; }
        .nav-link-ax:hover, .nav-link-ax.active { background: var(--ax-celeste); color: white; }

        .mobile-nav { background: var(--white); position: fixed; bottom: 0; width: 100%; height: 80px; display: none; justify-content: space-around; align-items: center; border-top: 1px solid #E2E8F0; z-index: 2000; box-shadow: 0 -5px 20px rgba(0,0,0,0.03); }
        .m-nav-item { color: #94A3B8; text-align: center; font-size: 10px; font-weight: 800; cursor: pointer; flex: 1; }
        .m-nav-item.active { color: var(--ax-blue); }
        .m-nav-item i { font-size: 24px; display: block; margin-bottom: 4px; }

        .card-pro { background: var(--white); border: none; border-radius: 28px; box-shadow: 0 10px 40px rgba(0,51,102,0.04); padding: 30px; margin-bottom: 25px; }
        .hero-card { background: linear-gradient(135deg, var(--ax-blue) 0%, var(--ax-celeste) 100%); color: white; padding: 40px 25px; }
        .btn-ax { background: var(--ax-blue); color: white; border-radius: 16px; padding: 16px; font-weight: 700; border: none; width: 100%; transition: 0.3s; }
        .form-ax { background: #F1F5F9; border: 1px solid #E2E8F0; border-radius: 16px; padding: 15px; width: 100%; margin-bottom: 15px; }
        .addr-pill { background: #F1F5F9; padding: 12px; border-radius: 14px; font-family: monospace; font-size: 0.8rem; color: #64748B; word-break: break-all; border: 1px solid #E2E8F0; margin-top: 10px; }
        
        .word-badge { display: inline-block; background: var(--ax-blue); color: white; padding: 6px 14px; border-radius: 10px; margin: 4px; font-size: 12px; font-weight: 600; }

        @media (max-width: 992px) {
            .sidebar { display: none; }
            .main-content { margin-left: 0; padding: 20px; }
            .mobile-nav { display: flex; }
        }
    </style>
</head>
<body>

    <div class="sidebar">
        <div class="p-5 text-center"><h2 class="fw-bold m-0" style="color:var(--ax-celeste)">ASTRALIX</h2><small class="opacity-50 fw-bold">Enterprise Node</small></div>
        <nav>
            <div class="nav-link-ax active" onclick="nav('dash', this)"><i class="fas fa-chart-pie"></i> Resumen</div>
            <div class="nav-link-ax" onclick="nav('wallet', this)"><i class="fas fa-wallet"></i> Billetera</div>
            <div class="nav-link-ax" onclick="nav('explorer', this)"><i class="fas fa-database"></i> Red</div>
            <div class="nav-link-ax" onclick="nav('security', this)"><i class="fas fa-key"></i> Seguridad</div>
        </nav>
    </div>

    <div class="main-content">
        <div id="v-dash" class="view">
            <div class="card-pro hero-card text-center">
                <small class="text-uppercase fw-bold opacity-75">Saldo Disponible</small>
                <h1 id="bal-txt" class="display-3 fw-bold my-2">0.00</h1>
                <div id="addr-txt" class="addr-pill bg-white bg-opacity-10 border-0 text-white opacity-75">Conecta tu cuenta en 'Seguridad'</div>
            </div>
            <div class="card-pro" style="border: 2px dashed var(--ax-celeste); background: #f0f9ff;">
                <small class="fw-bold text-primary">POZO DE RECOMPENSAS (TREASURY)</small>
                <h3 id="pool-txt" class="fw-bold m-0">0.00 AX</h3>
            </div>
            <button class="btn-ax py-3 mb-4" onclick="mine()">MINAR BLOQUE (+50 AX)</button>
            <div id="mini-feed"></div>
        </div>

        <div id="v-wallet" class="view" style="display:none">
            <div class="card-pro mx-auto" style="max-width: 500px;">
                <h4 class="fw-bold mb-4">Transferir Fondos</h4>
                <input type="text" id="tx-to" class="form-ax" placeholder="AX Dirección Destino">
                <input type="number" id="tx-amt" class="form-ax" placeholder="Monto">
                <button class="btn-ax" onclick="send()">AUTORIZAR ENVÍO</button>
            </div>
        </div>

        <div id="v-security" class="view" style="display:none">
            <div class="card-pro">
                <h4 class="fw-bold mb-3">Login Seguro</h4>
                <input type="password" id="i-priv" class="form-ax" placeholder="Private Key">
                <button class="btn-ax mb-3" onclick="login()">CONECTAR BILLETERA</button>
                <hr class="my-5">
                <h4 class="fw-bold mb-3">Nueva Identidad</h4>
                <button class="btn btn-outline-dark w-100 py-3 rounded-4" onclick="gen()">GENERAR SEMILLA</button>
                <div id="g-res" class="mt-4" style="display:none">
                    <label class="small fw-bold">12 Palabras:</label><div id="seed-words" class="my-2"></div>
                    <label class="small fw-bold">Privada (512-bit):</label><div class="addr-pill mb-2" id="g-priv"></div>
                    <label class="small fw-bold text-primary">Dirección AX:</label><div class="addr-pill fw-bold" id="g-pub"></div>
                </div>
            </div>
        </div>

        <div id="v-explorer" class="view" style="display:none">
            <div class="card-pro"><h4 class="fw-bold mb-4">Explorador de Red</h4><div id="full-chain"></div></div>
        </div>
    </div>

    <div class="mobile-nav">
        <div class="m-nav-item active" onclick="nav('dash', this)"><i class="fas fa-home"></i>Dash</div>
        <div class="m-nav-item" onclick="nav('wallet', this)"><i class="fas fa-paper-plane"></i>Enviar</div>
        <div class="m-nav-item" onclick="nav('explorer', this)"><i class="fas fa-database"></i>Red</div>
        <div class="m-nav-item" onclick="nav('security', this)"><i class="fas fa-key"></i>Keys</div>
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
            document.querySelectorAll('.nav-link-ax, .m-nav-item').forEach(function(n) { n.classList.remove('active'); });
            if(el) el.classList.add('active');
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
                document.getElementById('addr-txt').innerText = session.pub.substring(0,25) + "...";
                const r = await fetch('/api/balance/' + session.pub);
                const d = await r.json();
                document.getElementById('bal-txt').innerText = d.balance.toLocaleString() + ' AX';
            }
            const rp = await fetch('/api/balance/AX_TREASURY_REWARDS_POOL_SYSTEM_512');
            const dp = await rp.json();
            document.getElementById('pool-txt').innerText = dp.balance.toLocaleString() + ' AX';

            const res = await fetch('/api/chain');
            const chain = await res.json();
            const mini = document.getElementById('mini-feed');
            const full = document.getElementById('full-chain');
            mini.innerHTML = ''; full.innerHTML = '';
            chain.reverse().forEach(function(b) {
                const h = (b.Hash || b.hash || '').substring(0,25) + '...';
                mini.innerHTML += '<div class="p-2 border-bottom d-flex justify-content-between small"><span>#' + b.index + '</span><span>' + h + '</span></div>';
                full.innerHTML += '<div class="card-pro border mb-2 small"><b>Bloque #' + b.index + '</b><div class="addr-pill">' + (b.Hash || b.hash) + '</div></div>';
            });
        }

        async function mine() {
            if(!session) return alert('Login first');
            const r = await fetch('/api/mine?address=' + session.pub);
            if(r.ok) { alert('Mined!'); load(); } else { alert('Mempool empty or Treasury empty.'); }
        }

        async function send() {
            const tx = { sender: session.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            const r = await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            if(r.ok) { alert('Sent!'); nav('dash'); load(); } else { alert('Check funds'); }
        }

        async function gen() {
            const words = ["nube", "chaco", "cable", "red", "fibra", "datos", "seguro", "cripto", "luz", "onda", "pampa", "rio", "sol", "viento"];
            const seed = [1,2,3,4,5,6,7,8,9,10,11,12].map(function() { return words[Math.floor(Math.random()*words.length)]; }).join(' ');
            const p = btoa(seed).substring(0,64);
            const pb = await derive(p);
            document.getElementById('g-res').style.display = 'block';
            document.getElementById('seed-words').innerHTML = seed.split(' ').map(function(w) { return '<span class="word-badge">' + w + '</span>'; }).join('');
            document.getElementById('g-priv').innerText = p;
            document.getElementById('g-pub').innerText = pb;
        }

        load(); setInterval(load, 15000);
    </script>
</body>
</html>
`
