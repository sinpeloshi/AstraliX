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
	// Dirección Maestra (Génesis)
	creatorAddr := "AXdc3acc7c0b91eb485d0e3bb78059bb58a3999c14b56cfe6ca0428670afc6410c"

	genTx := core.Transaction{Sender: "SYSTEM", Recipient: creatorAddr, Amount: 1000002021}
	genTx.TxID = genTx.CalculateHash()
	
	genesis := core.Block{
		Index: 0, Timestamp: 1773561600,
		Transactions: []core.Transaction{genTx},
		PrevHash: strings.Repeat("0", 128), // Estándar 512-bit
		Difficulty: Difficulty,
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
		if len(Mempool) == 0 {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": "mempool_empty"})
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
			http.Error(w, "Invalid Payload", 400); return
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
    <title>AstraliX | Digital Asset Management</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --ax-blue: #003366; --ax-celeste: #74ACDF; --ax-bg: #F8FAFC; --white: #FFFFFF; }
        body { background: var(--ax-bg); color: #334155; font-family: 'Inter', system-ui, sans-serif; margin: 0; padding-bottom: 90px; }
        
        /* Desktop Sidebar */
        .sidebar { background: var(--ax-blue); height: 100vh; position: fixed; width: 280px; color: white; transition: 0.3s; z-index: 1000; box-shadow: 10px 0 30px rgba(0,0,0,0.05); padding-top: 30px; }
        .main-content { margin-left: 280px; padding: 40px; transition: 0.3s; }
        
        /* Navigation Links */
        .nav-link-custom { color: rgba(255,255,255,0.6); padding: 16px 30px; margin: 8px 20px; border-radius: 16px; font-weight: 600; cursor: pointer; display: flex; align-items: center; text-decoration: none; transition: 0.3s; }
        .nav-link-custom:hover, .nav-link-custom.active { background: rgba(255,255,255,0.1); color: var(--ax-celeste); }
        .nav-link-custom i { width: 35px; font-size: 1.2rem; }

        /* Mobile Bottom Nav */
        .mobile-nav { background: var(--white); position: fixed; bottom: 0; width: 100%; height: 80px; display: none; justify-content: space-around; align-items: center; border-top: 1px solid #E2E8F0; z-index: 2000; box-shadow: 0 -5px 20px rgba(0,0,0,0.03); }
        .m-nav-item { color: #94A3B8; text-decoration: none; text-align: center; font-size: 10px; font-weight: 800; cursor: pointer; flex: 1; }
        .m-nav-item.active { color: var(--ax-blue); }
        .m-nav-item i { font-size: 24px; display: block; margin-bottom: 4px; }

        /* Cards Elite */
        .card-pro { background: var(--white); border: none; border-radius: 30px; box-shadow: 0 10px 40px rgba(0,51,102,0.04); padding: 35px; margin-bottom: 25px; transition: 0.3s; }
        .balance-hero { background: linear-gradient(135deg, var(--ax-blue) 0%, var(--ax-celeste) 100%); color: white; position: relative; overflow: hidden; }
        .balance-hero::after { content: ""; position: absolute; bottom: -50px; right: -50px; width: 200px; height: 200px; background: rgba(255,255,255,0.1); filter: blur(50px); border-radius: 50%; }
        
        /* Inputs & Buttons */
        .btn-ax { background: var(--ax-blue); color: white; border-radius: 18px; padding: 16px 30px; font-weight: 700; border: none; transition: 0.3s; width: 100%; }
        .btn-ax:hover { background: var(--ax-celeste); transform: translateY(-3px); box-shadow: 0 10px 20px rgba(0,0,0,0.1); }
        .form-ax { background: #F1F5F9; border: 1px solid #E2E8F0; border-radius: 16px; padding: 15px; width: 100%; color: #1E293B; margin-bottom: 15px; }
        
        .addr-pill { background: #F1F5F9; padding: 12px 18px; border-radius: 14px; font-family: monospace; font-size: 0.8rem; color: #64748B; word-break: break-all; border: 1px solid #E2E8F0; }
        .status-badge { background: #D1FAE5; color: #065F46; padding: 6px 14px; border-radius: 10px; font-size: 11px; font-weight: 800; letter-spacing: 1px; }

        @media (max-width: 992px) {
            .sidebar { display: none; }
            .main-content { margin-left: 0; padding: 20px; }
            .mobile-nav { display: flex; }
        }
    </style>
</head>
<body>

    <div class="sidebar">
        <div class="px-4 mb-5 text-center">
            <h2 class="fw-bold mb-0" style="letter-spacing: -2px; color: var(--ax-celeste);">ASTRALIX</h2>
            <small class="opacity-50 text-uppercase fw-bold" style="font-size: 10px;">Enterprise Red L1</small>
        </div>
        <nav>
            <div class="nav-link-custom active" onclick="nav('dash', this)"><i class="fas fa-chart-pie"></i> Resumen</div>
            <div class="nav-link-custom" onclick="nav('wallet', this)"><i class="fas fa-wallet"></i> Billetera</div>
            <div class="nav-link-custom" onclick="nav('explorer', this)"><i class="fas fa-database"></i> Explorador</div>
            <div class="nav-link-custom" onclick="nav('sec', this)"><i class="fas fa-shield-halved"></i> Seguridad</div>
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

        <div id="v-dash" class="view-section">
            <div class="row g-4">
                <div class="col-lg-8">
                    <div class="card-pro balance-hero">
                        <small class="text-uppercase opacity-75 fw-bold">Saldo Total AX</small>
                        <h1 id="bal-main" class="display-3 fw-bold my-3">0.00</h1>
                        <div id="addr-main" class="addr-pill bg-white bg-opacity-10 text-white border-0 opacity-75">Sin cuenta conectada</div>
                    </div>
                    <div class="card-pro">
                        <h5 class="fw-bold mb-4">Monitor de Red</h5>
                        <div id="mini-chain" class="list-group list-group-flush"></div>
                    </div>
                </div>
                <div class="col-lg-4">
                    <div class="card-pro text-center">
                        <h5 class="fw-bold mb-4">Control del Nodo</h5>
                        <div class="mb-4">
                            <small class="text-muted d-block">Altura de Bloques</small>
                            <h4 id="h-stat" class="fw-bold text-primary">0</h4>
                        </div>
                        <button class="btn btn-ax py-3" onclick="mine()">MINAR BLOQUE</button>
                    </div>
                </div>
            </div>
        </div>

        <div id="v-wallet" class="view-section" style="display:none">
            <div class="card-pro mx-auto" style="max-width: 500px;">
                <h4 class="fw-bold mb-4">Transferir Activos</h4>
                <input type="text" id="tx-to" class="form-ax" placeholder="Dirección AX Destino">
                <input type="number" id="tx-amt" class="form-ax" placeholder="Monto AX">
                <button class="btn btn-ax py-3" onclick="send()">FIRMAR Y ENVIAR</button>
            </div>
        </div>

        <div id="v-explorer" class="view-section" style="display:none">
            <div class="card-pro">
                <h4 class="fw-bold mb-4">Explorador de la Cadena</h4>
                <div id="full-chain"></div>
            </div>
        </div>

        <div id="v-sec" class="view-section" style="display:none">
            <div class="card-pro">
                <h4 class="fw-bold mb-4">Acceso Seguro 512-bit</h4>
                <p class="small text-muted mb-4">Importa tu Clave Privada para sincronizar tu balance desde el Nodo Central.</p>
                <input type="password" id="i-priv" class="form-ax" placeholder="Private Key (SHA-512)">
                <button class="btn btn-ax mb-3" onclick="login()">CONECTAR BILLETERA</button>
                <button class="btn btn-link text-danger w-100 text-decoration-none small" onclick="logout()">Cerrar Sesión</button>
            </div>
            <div class="card-pro">
                <h5 class="fw-bold mb-3">Generar Identidad</h5>
                <button class="btn btn-outline-dark w-100 py-3 rounded-4" onclick="gen()">GENERAR 12 PALABRAS</button>
                <div id="g-res" class="mt-4" style="display:none">
                    <small class="text-muted">Clave Privada:</small><div class="addr-pill mb-2" id="g-priv"></div>
                    <small class="text-muted">Dirección AX:</small><div class="addr-pill fw-bold text-primary" id="g-pub"></div>
                </div>
            </div>
        </div>
    </div>

    <div class="mobile-nav">
        <div class="m-nav-item active" onclick="nav('dash', this)"><i class="fas fa-chart-pie"></i>Activos</div>
        <div class="m-nav-item" onclick="nav('wallet', this)"><i class="fas fa-wallet"></i>Billetera</div>
        <div class="m-nav-item" onclick="nav('explorer', this)"><i class="fas fa-database"></i>Explorar</div>
        <div class="m-nav-item" onclick="nav('sec', this)"><i class="fas fa-shield-halved"></i>Seguridad</div>
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
            document.querySelectorAll('.view-section').forEach(function(v) { v.style.display = 'none'; });
            document.getElementById('v-' + id).style.display = 'block';
            document.querySelectorAll('.nav-link-custom, .m-nav-item').forEach(function(n) { n.classList.remove('active'); });
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
                document.getElementById('addr-main').innerText = session.pub;
                const r = await fetch('/api/balance/' + session.pub);
                const d = await r.json();
                document.getElementById('bal-main').innerText = d.balance.toLocaleString() + ' AX';
            }
            const res = await fetch('/api/chain');
            const chain = await res.json();
            document.getElementById('h-stat').innerText = chain.length;
            const mini = document.getElementById('mini-chain');
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
            if(!session) return nav('sec');
            const tx = { sender: session.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            alert('¡Enviado! Mina el bloque para confirmar.'); nav('dash'); load();
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
