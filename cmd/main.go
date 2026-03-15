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
	// Reemplaza con tu dirección AX si ya tienes una, o deja esta por defecto
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

	// --- API ENDPOINTS ---
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
	fmt.Printf("🌐 AstraliX Argentum Live on port %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>AstraliX | Core Network OS</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --ax-dark: #002244; --ax-blue: #003366; --ax-celeste: #74ACDF; --ax-bg: #F0F4F8; }
        body { background: var(--ax-bg); color: #334155; font-family: 'Segoe UI', system-ui, sans-serif; overflow-x: hidden; }
        
        /* Layout */
        .sidebar { background: var(--ax-dark); height: 100vh; position: fixed; width: 280px; color: white; z-index: 1000; box-shadow: 5px 0 15px rgba(0,0,0,0.05); }
        .main-content { margin-left: 280px; padding: 40px; min-height: 100vh; transition: 0.3s; }
        
        /* Nav */
        .nav-link { color: rgba(255,255,255,0.6); padding: 15px 25px; margin: 10px 20px; border-radius: 12px; font-weight: 600; cursor: pointer; display: flex; align-items: center; transition: 0.3s; }
        .nav-link:hover, .nav-link.active { background: var(--ax-celeste); color: white; box-shadow: 0 4px 12px rgba(116, 172, 223, 0.3); }
        .nav-link i { width: 30px; font-size: 1.1rem; }

        /* Cards */
        .card-elite { background: white; border: none; border-radius: 25px; box-shadow: 0 10px 30px rgba(0,0,0,0.03); padding: 30px; margin-bottom: 30px; border: 1px solid rgba(0,0,0,0.02); }
        .hero-balance { background: linear-gradient(135deg, var(--ax-blue) 0%, var(--ax-dark) 100%); color: white; }
        
        .btn-ax { background: var(--ax-blue); color: white; border-radius: 15px; padding: 12px 25px; font-weight: 700; border: none; transition: 0.3s; }
        .btn-ax:hover { background: var(--ax-celeste); transform: translateY(-2px); }
        
        .addr-pill { background: #F1F5F9; padding: 10px 15px; border-radius: 12px; font-family: monospace; font-size: 0.8rem; color: #64748B; border: 1px solid #E2E8F0; }
        .status-badge { background: #D1FAE5; color: #065F46; padding: 5px 12px; border-radius: 10px; font-size: 11px; font-weight: 800; }

        @media (max-width: 992px) { .sidebar { display: none; } .main-content { margin-left: 0; padding: 20px; } }
    </style>
</head>
<body>

    <div class="sidebar">
        <div class="p-4 mb-4 text-center">
            <h2 class="fw-bold mb-0" style="letter-spacing: -2px; color: var(--ax-celeste);">ASTRALIX</h2>
            <small class="opacity-50 text-uppercase fw-bold" style="font-size: 10px;">512-bit Security Node</small>
        </div>
        <nav>
            <a class="nav-link active" onclick="nav('dash', this)"><i class="fas fa-home"></i> Inicio</a>
            <a class="nav-link" onclick="nav('wallet', this)"><i class="fas fa-wallet"></i> Billetera</a>
            <a class="nav-link" onclick="nav('explorer', this)"><i class="fas fa-cube"></i> Explorador</a>
            <a class="nav-link" onclick="nav('sec', this)"><i class="fas fa-fingerprint"></i> Seguridad</a>
        </nav>
    </div>

    <div class="main-content">
        <div class="d-flex justify-content-between align-items-center mb-5">
            <div><h3 class="fw-bold m-0">Dashboard Operativo</h3><span class="status-badge">NODO ACTIVO</span></div>
            <button class="btn btn-white shadow-sm rounded-circle p-3" onclick="location.reload()"><i class="fas fa-sync text-primary"></i></button>
        </div>

        <div id="v-dash" class="view">
            <div class="row g-4">
                <div class="col-lg-8">
                    <div class="card-elite hero-balance">
                        <small class="text-uppercase opacity-50 fw-bold">Saldo en Red AstraliX</small>
                        <h1 id="bal-txt" class="display-3 fw-bold my-3">0.00</h1>
                        <div id="addr-txt" class="addr-pill bg-white bg-opacity-10 border-0 text-white opacity-75 text-truncate">Sincroniza tu cuenta en 'Seguridad'</div>
                    </div>
                    <div class="card-elite">
                        <h5 class="fw-bold mb-4">Actividad Reciente</h5>
                        <div id="mini-chain" class="table-responsive"></div>
                    </div>
                </div>
                <div class="col-lg-4">
                    <div class="card-elite">
                        <h5 class="fw-bold mb-4">Resumen de Red</h5>
                        <div class="d-flex justify-content-between mb-2"><span>Supply</span><span class="fw-bold">1.0B AX</span></div>
                        <div class="d-flex justify-content-between mb-2"><span>Bloques</span><span id="h-stat" class="fw-bold">0</span></div>
                        <div class="d-flex justify-content-between mb-4"><span>Dificultad</span><span class="fw-bold text-primary">4 (PoW)</span></div>
                        <button class="btn btn-ax w-100 py-3" onclick="mine()">MINAR SIGUIENTE BLOQUE</button>
                    </div>
                </div>
            </div>
        </div>

        <div id="v-wallet" class="view" style="display:none">
            <div class="card-elite mx-auto" style="max-width: 500px;">
                <h4 class="fw-bold mb-4">Enviar AX</h4>
                <div class="mb-3">
                    <label class="small fw-bold text-muted mb-2">Destinatario (AX...)</label>
                    <input type="text" id="tx-to" class="form-control p-3 border-0 bg-light rounded-4" placeholder="AX...">
                </div>
                <div class="mb-4">
                    <label class="small fw-bold text-muted mb-2">Monto</label>
                    <input type="number" id="tx-amt" class="form-control p-3 border-0 bg-light rounded-4" placeholder="0.00">
                </div>
                <button class="btn btn-ax w-100 py-3" onclick="send()">AUTORIZAR ENVÍO</button>
            </div>
        </div>

        <div id="v-sec" class="view" style="display:none">
            <div class="card-elite">
                <h4 class="fw-bold mb-4">Gestión de Claves 512-bit</h4>
                <p class="small text-muted mb-4">Pega tu Clave Privada para derivar tu dirección y acceder a tus fondos.</p>
                <input type="password" id="i-priv" class="form-control p-3 border-0 bg-light rounded-4 mb-4" placeholder="Clave Secreta">
                <button class="btn btn-ax w-100 mb-3" onclick="login()">CONECTAR BILLETERA</button>
                <button class="btn btn-outline-danger w-100 border-0" onclick="logout()">Cerrar Sesión</button>
                <hr class="my-5">
                <button class="btn btn-light w-100 py-3 border rounded-4" onclick="gen()">GENERAR NUEVA IDENTIDAD</button>
                <div id="g-res" class="mt-3 small" style="display:none">
                    <div class="addr-pill mb-2" id="g-priv"></div>
                    <div class="addr-pill fw-bold text-primary" id="g-pub"></div>
                </div>
            </div>
        </div>

        <div id="v-explorer" class="view" style="display:none">
            <div class="card-elite">
                <h4 class="fw-bold mb-4">Explorador de Bloques</h4>
                <div id="full-chain"></div>
            </div>
        </div>
    </div>

    <script>
        async function derive(priv) {
            const buf = new TextEncoder().encode(priv);
            const hash = await crypto.subtle.digest('SHA-512', buf);
            const hex = Array.from(new Uint8Array(hash)).map(b => b.toString(16).padStart(2,'0')).join('');
            return 'AX' + hex.substring(0, 64);
        }

        let wallet = JSON.parse(localStorage.getItem('ax_argentum')) || null;

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
            wallet = { pub: pb, priv: p };
            localStorage.setItem('ax_argentum', JSON.stringify(wallet));
            location.reload();
        }

        function logout() { localStorage.removeItem('ax_argentum'); location.reload(); }

        async function load() {
            if(wallet) {
                document.getElementById('addr-txt').innerText = wallet.pub;
                const r = await fetch('/api/balance/' + wallet.pub);
                const d = await r.json();
                document.getElementById('bal-txt').innerText = d.balance.toLocaleString() + ' AX';
            }
            const res = await fetch('/api/chain');
            const chain = await res.json();
            document.getElementById('h-stat').innerText = chain.length;
            const mini = document.getElementById('mini-chain');
            const full = document.getElementById('full-chain');
            mini.innerHTML = ''; full.innerHTML = '';
            
            chain.reverse().forEach(b => {
                const h = b.hash.substring(0,24) + '...';
                mini.innerHTML += '<div class="p-3 border-bottom d-flex justify-content-between"><span>Bloque #' + b.index + '</span><span class="text-muted">' + h + '</span></div>';
                full.innerHTML += '<div class="card-elite border mb-3"><h6>Bloque #' + b.index + '</h6><div class="addr-pill mb-2">' + b.hash + '</div><small class="text-muted">Prev: ' + b.prev_hash + '</small></div>';
            });
        }

        async function mine() { await fetch('/api/mine'); alert("Bloque Minado!"); load(); }
        async function send() {
            if(!wallet) return nav('sec');
            const tx = { sender: wallet.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            alert("Enviada!"); nav('dash'); load();
        }

        async function gen() {
            const p = btoa(Math.random()).substring(0,64);
            const pb = await derive(p);
            document.getElementById('g-res').style.display = 'block';
            document.getElementById('g-priv').innerText = 'Priv: ' + p;
            document.getElementById('g-pub').innerText = 'Pub: ' + pb;
        }

        load(); setInterval(load, 15000);
    </script>
</body>
</html>
`
