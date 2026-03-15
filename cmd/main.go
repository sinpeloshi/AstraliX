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
	
	// --- PASO CLAVE ---
	// Primero corre el código, logueate con tu privada, copia la dirección que te de la web 
	// y pegala ACÁ abajo para que el sistema te asigne los tokens a VOS.
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

	http.HandleFunc("/api/transactions/new", func(w http.ResponseWriter, r *http.Request) {
		var tx core.Transaction
		json.NewDecoder(r.Body).Decode(&tx)
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
	fmt.Printf("🌐 AstraliX Argentum Live on %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>AstraliX | Core OS</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --ax-blue: #003366; --ax-celeste: #74ACDF; --bg: #F8FAFC; }
        body { background: var(--bg); font-family: 'Inter', sans-serif; padding-bottom: 90px; }
        .top-bar { background: #fff; padding: 15px 25px; border-bottom: 1px solid #eee; position: sticky; top: 0; z-index: 100; }
        .card-ax { background: #fff; border-radius: 25px; border: none; box-shadow: 0 10px 30px rgba(0,0,0,0.05); padding: 25px; margin-bottom: 20px; }
        .balance-hero { background: linear-gradient(135deg, var(--ax-blue) 0%, var(--ax-celeste) 100%); color: #fff; }
        .bottom-nav { background: #fff; position: fixed; bottom: 0; width: 100%; height: 75px; display: flex; border-top: 1px solid #eee; z-index: 1000; }
        .nav-item { flex: 1; text-align: center; padding: 12px; color: #94A3B8; cursor: pointer; text-decoration: none; font-size: 11px; font-weight: 700; }
        .nav-item.active { color: var(--ax-blue); }
        .nav-item i { font-size: 22px; display: block; margin-bottom: 4px; }
        .btn-ax { background: var(--ax-blue); color: #fff; border-radius: 15px; padding: 15px; border: none; width: 100%; font-weight: 700; }
        .addr-pill { background: rgba(0,0,0,0.05); padding: 10px; border-radius: 12px; font-family: monospace; font-size: 11px; word-break: break-all; margin-top: 10px; }
    </style>
</head>
<body>
    <div class="top-bar d-flex justify-content-between align-items-center">
        <b style="color:var(--ax-blue); font-size: 1.2rem;">ASTRALIX</b>
        <span class="badge bg-success">ONLINE</span>
    </div>

    <div class="container mt-4">
        <div id="v-dash" class="view">
            <div class="card-ax balance-hero text-center">
                <small class="opacity-75">SALDO DISPONIBLE</small>
                <h1 id="bal-txt" class="display-4 fw-bold my-2">0.00 AX</h1>
                <div id="addr-txt" class="addr-pill">No conectado</div>
            </div>
            <div class="card-ax">
                <h6 class="fw-bold mb-3">Últimos Bloques</h6>
                <div id="mini-feed"></div>
            </div>
        </div>

        <div id="v-wallet" class="view" style="display:none">
            <div class="card-ax">
                <h4 class="fw-bold mb-4">Enviar AX</h4>
                <input type="text" id="tx-to" class="form-control mb-3 p-3 border-0 bg-light" placeholder="Dirección Destino">
                <input type="number" id="tx-amt" class="form-control mb-4 p-3 border-0 bg-light" placeholder="Monto">
                <button class="btn-ax" onclick="send()">ENVIAR AHORA</button>
            </div>
            <button class="btn btn-outline-dark w-100 rounded-4 py-3" onclick="mine()">MINAR BLOQUE PENDIENTE</button>
        </div>

        <div id="v-security" class="view" style="display:none">
            <div class="card-ax">
                <h4 class="fw-bold mb-4">Importar / Generar</h4>
                <input type="password" id="i-priv" class="form-control mb-3 p-3 border-0 bg-light" placeholder="Tu Clave Privada">
                <button class="btn-ax" onclick="login()">CONECTAR BILLETERA</button>
                <div id="calc-res" class="mt-4" style="display:none">
                    <p class="small text-muted mb-1">Tu dirección pública calculada es:</p>
                    <div class="addr-pill fw-bold text-primary" id="calc-pub"></div>
                    <p class="text-danger small mt-2">Copia esta dirección y pegala en el creatorAddr de tu código Go para recibir los tokens.</p>
                </div>
                <hr class="my-4">
                <button class="btn btn-link text-danger w-100 text-decoration-none" onclick="logout()">Cerrar Sesión</button>
            </div>
        </div>
    </div>

    <div class="bottom-nav">
        <div class="nav-item active" onclick="nav('dash', this)"><i class="fas fa-home"></i>Inicio</div>
        <div class="nav-item" onclick="nav('wallet', this)"><i class="fas fa-paper-plane"></i>Enviar</div>
        <div class="nav-item" onclick="nav('security', this)"><i class="fas fa-key"></i>Keys</div>
    </div>

    <script>
        // Cálculo SHA-512 Fijo
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
            document.getElementById('calc-res').style.display = 'block';
            document.getElementById('calc-pub').innerText = pb;
            session = { pub: pb, priv: p };
            localStorage.setItem('ax_argentum', JSON.stringify(session));
        }

        function logout() { localStorage.removeItem('ax_argentum'); location.reload(); }

        async function load() {
            if(session) {
                document.getElementById('addr-txt').innerText = session.pub.substring(0,25) + "...";
                const r = await fetch('/api/balance/' + session.pub);
                const d = await r.json();
                document.getElementById('bal-txt').innerText = d.balance.toLocaleString() + ' AX';
            }
            const res = await fetch('/api/chain');
            const chain = await res.json();
            const feed = document.getElementById('mini-feed');
            feed.innerHTML = '';
            chain.reverse().slice(0,5).forEach(function(b) {
                const h = (b.Hash || b.hash || "").substring(0,30) + "...";
                feed.innerHTML += '<div class="small border-bottom py-2 d-flex justify-content-between"><span>#' + b.index + '</span><span class="text-muted">' + h + '</span></div>';
            });
        }

        async function mine() { await fetch('/api/mine'); alert('Bloque Minado!'); load(); }
        async function send() {
            const tx = { sender: session.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            alert('Enviado! Mina el bloque para confirmar.'); nav('dash', document.querySelector('.nav-item')); load();
        }

        load(); setInterval(load, 15000);
    </script>
</body>
</html>
`
