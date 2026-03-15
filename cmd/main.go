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
	
	// 🎯 DIRECCIÓN MASTER VINCULADA (Denis W. Sanchez)
	creatorAddr := "AX5eaba583bf646e0e39f41da6f9d8fa6db929c2e858bd32dffe6ac0cee2e3e974"

	// Setup de Red
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
    <title>AstraliX | Network Dashboard</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --ax-blue: #003366; --ax-celeste: #74ACDF; --bg: #F4F7F9; --white: #FFFFFF; }
        body { background: var(--bg); color: #334155; font-family: 'Segoe UI', system-ui, sans-serif; margin: 0; padding-bottom: 90px; }
        
        /* Layout Híbrido */
        .sidebar { background: var(--ax-blue); height: 100vh; position: fixed; width: 280px; color: white; z-index: 1000; box-shadow: 10px 0 30px rgba(0,0,0,0.05); }
        .main-content { margin-left: 280px; padding: 40px; min-height: 100vh; transition: 0.3s; }
        
        .nav-link-ax { color: rgba(255,255,255,0.6); padding: 15px 25px; margin: 10px 15px; border-radius: 12px; font-weight: 600; cursor: pointer; display: flex; align-items: center; text-decoration: none; transition: 0.3s; }
        .nav-link-ax:hover, .nav-link-ax.active { background: var(--ax-celeste); color: white; box-shadow: 0 4px 15px rgba(0,0,0,0.1); }
        .nav-link-ax i { width: 30px; font-size: 1.1rem; }

        .mobile-nav { background: var(--white); position: fixed; bottom: 0; width: 100%; height: 80px; display: none; justify-content: space-around; align-items: center; border-top: 1px solid #E2E8F0; z-index: 2000; box-shadow: 0 -5px 20px rgba(0,0,0,0.03); }
        .m-nav-item { color: #94A3B8; text-decoration: none; text-align: center; font-size: 10px; font-weight: 800; cursor: pointer; flex: 1; }
        .m-nav-item.active { color: var(--ax-blue); }
        .m-nav-item i { font-size: 22px; display: block; margin-bottom: 4px; }

        .card-pro { background: var(--white); border: none; border-radius: 25px; box-shadow: 0 10px 30px rgba(0,0,0,0.03); padding: 30px; margin-bottom: 30px; border: 1px solid rgba(0,0,0,0.01); }
        .hero-balance { background: linear-gradient(135deg, var(--ax-blue) 0%, var(--ax-celeste) 100%); color: white; position: relative; overflow: hidden; }
        .hero-balance::after { content: ""; position: absolute; bottom: -50px; right: -50px; width: 200px; height: 200px; background: rgba(255,255,255,0.1); filter: blur(50px); border-radius: 50%; }
        
        .btn-ax { background: var(--ax-blue); color: white; border-radius: 15px; padding: 12px 25px; font-weight: 700; border: none; transition: 0.3s; width: 100%; }
        .btn-ax:hover { background: var(--ax-celeste); transform: translateY(-2px); }
        
        .addr-pill { background: #F1F5F9; padding: 10px 15px; border-radius: 12px; font-family: monospace; font-size: 0.8rem; color: #64748B; border: 1px solid #E2E8F0; word-break: break-all; margin-top: 10px; }
        .status-badge { background: #D1FAE5; color: #065F46; padding: 5px 12px; border-radius: 10px; font-size: 11px; font-weight: 800; letter-spacing: 1px; }

        @media (max-width: 992px) {
            .sidebar { display: none; }
            .main-content { margin-left: 0; padding: 20px; }
            .mobile-nav { display: flex; }
        }
    </style>
</head>
<body>

    <div class="sidebar">
        <div class="p-5 text-center">
            <h2 class="fw-bold mb-0" style="letter-spacing: -2px; color: var(--ax-celeste);">ASTRALIX</h2>
            <small class="opacity-50 text-uppercase fw-bold" style="font-size: 10px;">Enterprise Red L1</small>
        </div>
        <nav>
            <div class="nav-link-ax active" onclick="nav('dash', this)"><i class="fas fa-chart-pie"></i> Resumen</div>
            <div class="nav-link-ax" onclick="nav('wallet', this)"><i class="fas fa-wallet"></i> Billetera</div>
            <div class="nav-link-ax" onclick="nav('explorer', this)"><i class="fas fa-database"></i> Explorador</div>
            <div class="nav-link-ax" onclick="nav('security', this)"><i class="fas fa-shield-halved"></i> Seguridad</div>
        </nav>
    </div>

    <div class="main-content">
        <div class="d-flex justify-content-between align-items-center mb-5">
            <div>
                <h3 class="fw-bold m-0 text-dark">Panel Central</h3>
                <span class="status-badge"><i class="fas fa-circle me-1" style="font-size: 7px;"></i> NODO ACTIVO</span>
            </div>
            <button class="btn btn-white shadow-sm rounded-circle p-3" onclick="location.reload()"><i class="fas fa-sync text-primary"></i></button>
        </div>

        <div id="v-dash" class="view">
            <div class="row g-4">
                <div class="col-lg-8">
                    <div class="card-pro hero-balance text-center py-5">
                        <small class="text-uppercase opacity-75 fw-bold">Saldo Total AX</small>
                        <h1 id="bal-txt" class="display-3 fw-bold my-3">0.00</h1>
                        <div id="addr-txt" class="addr-pill bg-white bg-opacity-10 text-white border-0 opacity-75">Sincroniza tu cuenta en 'Seguridad'</div>
                    </div>
                    <div class="card-pro">
                        <h5 class="fw-bold mb-4">Actividad del Nodo</h5>
                        <div id="mini-feed"></div>
                    </div>
                </div>
                <div class="col-lg-4">
                    <div class="card-pro text-center">
                        <h5 class="fw-bold mb-4">Control de Red</h5>
                        <div class="mb-4">
                            <small class="text-muted d-block">Altura de Bloques</small>
                            <h4 id="h-stat" class="fw-bold text-primary">0</h4>
                        </div>
                        <button class="btn btn-ax py-3" onclick="mine()">MINAR SIGUIENTE BLOQUE</button>
                    </div>
                </div>
            </div>
        </div>

        <div id="v-wallet" class="view" style="display:none">
            <div class="card-pro mx-auto" style="max-width: 600px;">
                <h4 class="fw-bold mb-4">Transferir Fondos</h4>
                <div class="mb-3">
                    <label class="small fw-bold text-muted">Dirección Destino</label>
                    <input type="text" id="tx-to" class="form-control p-3 border-0 bg-light rounded-4" placeholder="AX...">
                </div>
                <div class="mb-4">
                    <label class="small fw-bold text-muted">Monto</label>
                    <input type="number" id="tx-amt" class="form-control p-3 border-0 bg-light rounded-4" placeholder="0.00">
                </div>
                <button class="btn btn-ax py-3" onclick="send()">AUTORIZAR Y ENVIAR</button>
            </div>
        </div>

        <div id="v-explorer" class="view" style="display:none">
            <div class="card-pro">
                <h4 class="fw-bold mb-4">Blockchain Explorer</h4>
                <div id="full-chain"></div>
            </div>
        </div>

        <div id="v-security" class="view" style="display:none">
            <div class="card-pro">
                <h4 class="fw-bold mb-4">Seguridad 512-bit</h4>
                <p class="small text-muted mb-4">Pega tu Clave Privada para sincronizar tu balance desde el Nodo Central.</p>
                <input type="password" id="i-priv" class="form-control p-3 border-0 bg-light rounded-4 mb-4" placeholder="Private Key (SHA-512)">
                <button class="btn btn-ax mb-3" onclick="login()">CONECTAR BILLETERA</button>
                <button class="btn btn-link text-danger w-100 text-decoration-none small" onclick="logout()">Desconectar</button>
                <hr class="my-5">
                <button class="btn btn-outline-dark w-100 py-3 rounded-4" onclick="gen()">GENERAR SEMILLA</button>
                <div id="g-res" class="mt-4" style="display:none">
                    <small>Clave Privada:</small><div class="addr-pill mb-2" id="g-priv"></div>
                    <small>Tu Dirección AX:</small><div class="addr-pill fw-bold text-primary" id="g-pub"></div>
                </div>
            </div>
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
            const hex = Array.from(new Uint8Array(hash)).map(function(b) { return b.toString(16).padStart(2, '0'); }).join('');
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
            if(!p) return alert('Ingresa tu clave');
            const pb = await derive(p);
            session = { pub: pb, priv: p };
            localStorage.setItem('ax_argentum', JSON.stringify(session));
            location.reload();
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
            document.getElementById('h-stat').innerText = chain.length;
            const mini = document.getElementById('mini-feed');
            const full = document.getElementById('full-chain');
            mini.innerHTML = ''; full.innerHTML = '';
            chain.reverse().forEach(function(b) {
                const h = (b.Hash || b.hash || '').substring(0,25) + '...';
                mini.innerHTML += '<div class="p-3 border-bottom d-flex justify-content-between"><span>Bloque #' + b.index + '</span><span class="text-muted">' + h + '</span></div>';
                full.innerHTML += '<div class="card-pro border mb-3"><h6>Bloque #' + b.index + '</h6><div class="addr-pill">' + (b.Hash || b.hash) + '</div></div>';
            });
        }

        async function mine() { await fetch('/api/mine'); alert('¡Minado!'); load(); }
        async function send() {
            if(!session) return nav('security');
            const tx = { sender: session.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            alert('¡Enviada! Mina el bloque para confirmar.'); nav('dash'); load();
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
