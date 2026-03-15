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

	// Genesis
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
		if len(Mempool) == 0 {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": "empty"})
			return 
		}
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
			http.Error(w, "Bad Request", 400); return
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
	fmt.Printf("🌐 AstraliX Live on port %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>AstraliX | Digital Assets</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --ax-blue: #003366; --ax-celeste: #74ACDF; --ax-bg: #F4F7F9; --white: #FFFFFF; }
        body { background: var(--ax-bg); color: #334155; font-family: 'Segoe UI', system-ui, sans-serif; margin: 0; padding-bottom: 90px; }
        
        .sidebar { background: var(--ax-blue); height: 100vh; position: fixed; width: 280px; color: white; z-index: 1000; box-shadow: 10px 0 30px rgba(0,0,0,0.05); display: block; }
        .main-content { margin-left: 280px; padding: 40px; transition: 0.3s; }
        
        .nav-link-ax { color: rgba(255,255,255,0.6); padding: 15px 25px; margin: 8px 15px; border-radius: 12px; font-weight: 600; cursor: pointer; display: flex; align-items: center; text-decoration: none; }
        .nav-link-ax:hover, .nav-link-ax.active { background: rgba(255,255,255,0.1); color: var(--ax-celeste); }
        .nav-link-ax i { width: 30px; font-size: 1.1rem; }

        .mobile-nav { background: var(--white); position: fixed; bottom: 0; width: 100%; height: 80px; display: none; justify-content: space-around; align-items: center; border-top: 1px solid #E2E8F0; z-index: 2000; box-shadow: 0 -5px 20px rgba(0,0,0,0.03); }
        .m-nav-item { color: #94A3B8; text-decoration: none; text-align: center; font-size: 10px; font-weight: 800; cursor: pointer; flex: 1; }
        .m-nav-item.active { color: var(--ax-blue); }
        .m-nav-item i { font-size: 22px; display: block; margin-bottom: 4px; }

        .card-pro { background: var(--white); border: none; border-radius: 28px; box-shadow: 0 10px 30px rgba(0,51,102,0.05); padding: 25px; margin-bottom: 25px; }
        .hero-balance { background: linear-gradient(135deg, var(--ax-blue) 0%, var(--ax-celeste) 100%); color: white; }
        .btn-ax { background: var(--ax-blue); color: white; border-radius: 16px; padding: 15px; font-weight: 700; border: none; width: 100%; }
        .addr-pill { background: #F1F5F9; padding: 10px 15px; border-radius: 12px; font-family: monospace; font-size: 0.8rem; color: #64748B; word-break: break-all; border: 1px solid #E2E8F0; }

        @media (max-width: 992px) {
            .sidebar { display: none; }
            .main-content { margin-left: 0; padding: 20px; }
            .mobile-nav { display: flex; }
        }
    </style>
</head>
<body>

    <div class="sidebar">
        <div class="p-4 mb-4 text-center">
            <h2 class="fw-bold mb-0" style="letter-spacing: -2px; color: var(--ax-celeste);">ASTRALIX</h2>
            <small class="opacity-50 fw-bold" style="font-size: 10px; text-transform: uppercase;">Argentum L1 Node</small>
        </div>
        <nav>
            <div class="nav-link-ax active" onclick="nav('dash', this)"><i class="fas fa-chart-pie"></i> Resumen</div>
            <div class="nav-link-ax" onclick="nav('wallet', this)"><i class="fas fa-wallet"></i> Billetera</div>
            <div class="nav-link-ax" onclick="nav('explorer', this)"><i class="fas fa-database"></i> Explorador</div>
            <div class="nav-link-ax" onclick="nav('sec', this)"><i class="fas fa-shield-halved"></i> Seguridad</div>
        </nav>
    </div>

    <div class="main-content">
        <div id="v-dash" class="view">
            <div class="card-pro hero-balance text-center py-5">
                <small class="text-uppercase fw-bold opacity-75">Saldo Disponible</small>
                <h1 id="bal-txt" class="display-3 fw-bold my-2">0.00</h1>
                <div id="addr-txt" class="addr-pill bg-white bg-opacity-10 border-0 text-white opacity-75">Desconectado</div>
            </div>
            <div class="card-pro">
                <h5 class="fw-bold mb-4">Feed de Bloques</h5>
                <div id="mini-feed"></div>
            </div>
        </div>

        <div id="v-wallet" class="view" style="display:none">
            <div class="card-pro mx-auto" style="max-width: 500px;">
                <h4 class="fw-bold mb-4">Enviar AX</h4>
                <input type="text" id="tx-to" class="form-control mb-3 p-3 border-0 bg-light rounded-4" placeholder="Dirección Destino">
                <input type="number" id="tx-amt" class="form-control mb-4 p-3 border-0 bg-light rounded-4" placeholder="Monto">
                <button class="btn-ax" onclick="send()">FIRMAR TRANSACCIÓN</button>
            </div>
        </div>

        <div id="v-explorer" class="view" style="display:none">
            <div class="card-pro">
                <h4 class="fw-bold mb-4">Blockchain Explorer</h4>
                <div id="full-chain"></div>
            </div>
        </div>

        <div id="v-sec" class="view" style="display:none">
            <div class="card-pro">
                <h4 class="fw-bold mb-4">Seguridad 512-bit</h4>
                <input type="password" id="i-priv" class="form-control mb-3 p-3 border-0 bg-light rounded-4" placeholder="Clave Privada">
                <button class="btn-ax mb-3" onclick="login()">CONECTAR</button>
                <hr class="my-4">
                <button class="btn btn-outline-dark w-100 py-3 rounded-4" onclick="gen()">GENERAR SEMILLA</button>
                <div id="g-res" class="mt-3" style="display:none">
                    <small>Privada:</small><div class="addr-pill mb-2" id="g-priv"></div>
                    <small>Dirección AX:</small><div class="addr-pill text-primary" id="g-pub"></div>
                </div>
            </div>
        </div>
    </div>

    <div class="mobile-nav">
        <div class="m-nav-item active" onclick="nav('dash', this)"><i class="fas fa-home"></i>Dash</div>
        <div class="m-nav-item" onclick="nav('wallet', this)"><i class="fas fa-paper-plane"></i>Enviar</div>
        <div class="m-nav-item" onclick="nav('explorer', this)"><i class="fas fa-cube"></i>Red</div>
        <div class="m-nav-item" onclick="nav('sec', this)"><i class="fas fa-key"></i>Keys</div>
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
                document.getElementById('addr-txt').innerText = session.pub;
                const r = await fetch('/api/balance/' + session.pub);
                const d = await r.json();
                document.getElementById('bal-txt').innerText = d.balance.toLocaleString() + ' AX';
            }
            const res = await fetch('/api/chain');
            const chain = await res.json();
            const mini = document.getElementById('mini-feed');
            const full = document.getElementById('full-chain');
            mini.innerHTML = ''; full.innerHTML = '';
            chain.reverse().forEach(function(b) {
                const h = (b.Hash || b.hash || '').substring(0,25) + '...';
                mini.innerHTML += '<div class="p-2 border-bottom d-flex justify-content-between"><span>Bloque #' + b.index + '</span><span>' + h + '</span></div>';
                full.innerHTML += '<div class="card-pro border mb-2"><h6>Bloque #' + b.index + '</h6><div class="addr-pill">' + (b.Hash || b.hash) + '</div></div>';
            });
        }

        async function mine() { await fetch('/api/mine'); alert('Mined!'); load(); }
        async function send() {
            const tx = { sender: session.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            alert('Sent!'); nav('dash'); load();
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
`
