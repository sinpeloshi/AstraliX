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
	// MASTER ADDRESS (Asegúrate de que esta sea la que genera tu Private Key de 512 bits)
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

	// --- API ---
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
	fmt.Printf("🌐 AstraliX Argentum Live on port %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>AstraliX | Digital Asset Management</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --ax-blue: #003366; --ax-celeste: #74ACDF; --ax-light: #F8FAFC; --ax-white: #FFFFFF; }
        body { background: var(--ax-light); color: #334155; font-family: 'Inter', sans-serif; overflow-x: hidden; }
        
        /* Layout */
        .sidebar { background: var(--ax-blue); height: 100vh; position: fixed; width: 280px; color: white; transition: 0.3s; z-index: 1000; padding: 30px 0; }
        .main-content { margin-left: 280px; padding: 40px; min-height: 100vh; }
        
        /* Navigation */
        .nav-link { color: rgba(255,255,255,0.6); padding: 15px 30px; margin: 8px 20px; border-radius: 15px; font-weight: 500; cursor: pointer; display: flex; align-items: center; transition: 0.3s; text-decoration: none; }
        .nav-link:hover, .nav-link.active { background: rgba(255,255,255,0.1); color: var(--ax-celeste); }
        .nav-link i { width: 30px; font-size: 1.2rem; }

        /* Cards & UI */
        .card-pro { background: var(--ax-white); border: none; border-radius: 24px; box-shadow: 0 10px 30px rgba(0,51,102,0.05); padding: 30px; margin-bottom: 30px; }
        .hero-balance { background: linear-gradient(135deg, var(--ax-blue) 0%, #001f3f 100%); color: white; position: relative; overflow: hidden; }
        .hero-balance::after { content: ""; position: absolute; bottom: -50px; right: -50px; width: 200px; height: 200px; background: var(--ax-celeste); filter: blur(80px); opacity: 0.3; }
        
        .btn-ax { background: var(--ax-blue); color: white; border-radius: 16px; padding: 14px 30px; font-weight: 700; border: none; transition: 0.3s; }
        .btn-ax:hover { background: var(--ax-celeste); transform: translateY(-3px); box-shadow: 0 10px 20px rgba(116, 172, 223, 0.4); }
        
        .addr-pill { background: #f1f5f9; padding: 12px 20px; border-radius: 12px; font-family: monospace; font-size: 0.85rem; color: #64748b; border: 1px solid #e2e8f0; display: block; word-break: break-all; }
        .status-badge { background: #dcfce7; color: #166534; padding: 6px 14px; border-radius: 10px; font-size: 11px; font-weight: 800; text-transform: uppercase; letter-spacing: 1px; }

        @media (max-width: 992px) { .sidebar { display: none; } .main-content { margin-left: 0; padding: 20px; } }
    </style>
</head>
<body>

    <div class="sidebar">
        <div class="px-4 mb-5">
            <h2 class="fw-bold mb-0" style="letter-spacing: -2px; color: var(--ax-celeste);">ASTRALIX</h2>
            <small class="opacity-50 text-uppercase fw-bold" style="font-size: 10px;">512-bit Security Layer</small>
        </div>
        <nav>
            <a class="nav-link active" onclick="nav('dash', this)"><i class="fas fa-chart-pie"></i> Resumen</a>
            <a class="nav-link" onclick="nav('wallet', this)"><i class="fas fa-wallet"></i> Mi Billetera</a>
            <a class="nav-link" onclick="nav('explorer', this)"><i class="fas fa-list-ul"></i> Explorador</a>
            <a class="nav-link" onclick="nav('security', this)"><i class="fas fa-fingerprint"></i> Seguridad</a>
        </nav>
    </div>

    <div class="main-content">
        <div class="d-flex justify-content-between align-items-center mb-5">
            <div>
                <h3 class="fw-bold mb-1" id="page-title">Bienvenido al Nodo Central</h3>
                <span class="status-badge"><i class="fas fa-circle me-1" style="font-size: 8px;"></i> Online</span>
            </div>
            <div class="d-flex align-items-center">
                <div class="text-end me-3">
                    <small class="text-muted d-block">Protocolo</small>
                    <span class="fw-bold text-primary">SHA-512 V. Argentum</span>
                </div>
                <button class="btn btn-white shadow-sm rounded-circle p-3" onclick="location.reload()"><i class="fas fa-sync-alt text-primary"></i></button>
            </div>
        </div>

        <div id="view-dash" class="view">
            <div class="row g-4">
                <div class="col-lg-8">
                    <div class="card-pro hero-balance">
                        <div class="row align-items-center">
                            <div class="col-md-8">
                                <h6 class="opacity-75 text-uppercase fw-bold" style="font-size: 11px; letter-spacing: 1px;">Saldo Total Disponible</h6>
                                <h1 id="bal-main" class="display-3 fw-bold my-3">0.00</h1>
                                <div id="pub-display" class="addr-pill bg-white bg-opacity-10 border-0 text-white opacity-75">No hay billetera sincronizada</div>
                            </div>
                            <div class="col-md-4 text-end d-none d-md-block">
                                <i class="fas fa-shield-halved fa-8x opacity-10"></i>
                            </div>
                        </div>
                    </div>
                    
                    <div class="card-pro">
                        <h5 class="fw-bold mb-4"><i class="fas fa-history me-2 text-primary"></i> Actividad de la Red</h5>
                        <div id="mini-chain" class="table-responsive"></div>
                    </div>
                </div>
                
                <div class="col-lg-4">
                    <div class="card-pro">
                        <h5 class="fw-bold mb-4">Acciones Rápidas</h5>
                        <button class="btn btn-ax w-100 mb-3 py-3" onclick="nav('wallet')">Enviar AX</button>
                        <button class="btn btn-light w-100 border py-3 rounded-4 mb-4" onclick="mine()">Minar Bloque</button>
                        <hr>
                        <div class="mt-4">
                            <small class="text-muted d-block mb-2">Estado del Supply</small>
                            <div class="progress mb-2" style="height: 6px;">
                                <div class="progress-bar" style="width: 100%; background: var(--ax-celeste);"></div>
                            </div>
                            <div class="d-flex justify-content-between small"><span class="fw-bold">1.0B AX</span><span>Circulante</span></div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div id="view-wallet" class="view" style="display:none">
            <div class="row g-4 justify-content-center">
                <div class="col-md-7">
                    <div class="card-pro">
                        <h4 class="fw-bold mb-4 text-primary">Enviar Fondos</h4>
                        <div class="mb-3">
                            <label class="small fw-bold text-muted mb-2">Dirección del Destinatario (AX...)</label>
                            <input type="text" id="tx-to" class="form-control p-3 border-0 bg-light rounded-4" placeholder="AX...">
                        </div>
                        <div class="mb-4">
                            <label class="small fw-bold text-muted mb-2">Monto a Transferir</label>
                            <input type="number" id="tx-amt" class="form-control p-3 border-0 bg-light rounded-4" placeholder="0.00">
                        </div>
                        <button class="btn btn-ax w-100 py-3 rounded-4" onclick="sendTx()">CONFIRMAR Y FIRMAR</button>
                    </div>
                </div>
            </div>
        </div>

        <div id="view-security" class="view" style="display:none">
            <div class="row g-4">
                <div class="col-md-6">
                    <div class="card-pro h-100">
                        <h4 class="fw-bold mb-4"><i class="fas fa-key me-2 text-primary"></i> Importar Llave</h4>
                        <p class="small text-muted mb-4">Ingresa tu Clave Privada de 512 bits. Tu dirección se calculará automáticamente mediante SHA-512.</p>
                        <input type="password" id="imp-priv" class="form-control p-3 border-0 bg-light rounded-4 mb-4" placeholder="Pega tu clave secreta aquí">
                        <button class="btn btn-ax w-100 py-3" onclick="login()">SINCRONIZAR CUENTA</button>
                        <button class="btn btn-link text-danger w-100 mt-4 text-decoration-none small" onclick="logout()">Desconectar Billetera</button>
                    </div>
                </div>
                <div class="col-md-6">
                    <div class="card-pro h-100">
                        <h4 class="fw-bold mb-4">Generar Nueva Identidad</h4>
                        <p class="small text-muted mb-4">Crea una billetera nueva totalmente segura y anónima.</p>
                        <button class="btn btn-outline-dark w-100 py-3 rounded-4" onclick="gen()">GENERAR 512-BIT KEY</button>
                        <div id="gen-res" class="mt-4" style="display:none">
                            <div class="addr-pill mb-2" id="gen-priv"></div>
                            <div class="addr-pill fw-bold text-primary" id="gen-pub"></div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div id="view-explorer" class="view" style="display:none">
            <div class="card-pro">
                <h4 class="fw-bold mb-4">Historial Completo de la Cadena</h4>
                <div id="full-chain"></div>
            </div>
        </div>
    </div>

    <script>
        // Función SHA-512 Real
        async function derive(priv) {
            const buf = new TextEncoder().encode(priv);
            const hash = await crypto.subtle.digest('SHA-512', buf);
            const hex = Array.from(new Uint8Array(hash)).map(b => b.toString(16).padStart(2, '0')).join('');
            return 'AX' + hex.substring(0, 64);
        }

        let wallet = JSON.parse(localStorage.getItem('ax_argentum')) || null;

        function nav(id, el) {
            document.querySelectorAll('.view').forEach(v => v.style.display = 'none');
            document.getElementById('view-' + id).style.display = 'block';
            document.getElementById('page-title').innerText = id.toUpperCase();
            if(el) {
                document.querySelectorAll('.nav-link').forEach(n => n.classList.remove('active'));
                el.classList.add('active');
            }
        }

        async function login() {
            const p = document.getElementById('imp-priv').value;
            const pb = await derive(p);
            wallet = { pub: pb, priv: p };
            localStorage.setItem('ax_argentum', JSON.stringify(wallet));
            location.reload();
        }

        function logout() { localStorage.removeItem('ax_argentum'); location.reload(); }

        async function load() {
            if(wallet) {
                document.getElementById('pub-display').innerText = wallet.pub;
                const r = await fetch('/api/balance/' + wallet.pub);
                const d = await r.json();
                document.getElementById('bal-main').innerText = d.balance.toLocaleString() + ' AX';
            }
            const res = await fetch('/api/chain');
            const chain = await res.json();
            const mini = document.getElementById('mini-chain');
            const full = document.getElementById('full-chain');
            mini.innerHTML = ''; full.innerHTML = '';
            
            chain.reverse().forEach(b => {
                const html = ` + "`" + `<div class="p-3 border-bottom d-flex justify-content-between"><div><b>Bloque #${b.index}</b><br><small class="text-muted">${b.hash.substring(0,32)}...</small></div><div class="text-end small"><b>${b.transactions?b.transactions.length:0} TXs</b><br>${new Date(b.timestamp*1000).toLocaleTimeString()}</div></div>` + "`" + `;
                mini.innerHTML += html;
                full.innerHTML += `<div class="card-pro border mb-3"><h6>Bloque #${b.index}</h6><div class="addr-pill mb-2">${b.hash}</div><small class="text-muted">Anterior: ${b.prev_hash}</small></div>`;
            });
        }

        async function mine() { await fetch('/api/mine'); alert("Bloque Confirmado!"); load(); }
        async function sendTx() {
            if(!wallet) return nav('security');
            const tx = { sender: wallet.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            alert("Enviada! Miná el bloque para confirmar."); nav('dash'); load();
        }

        async function gen() {
            const p = btoa(Math.random()).substring(0,64);
            const pb = await derive(p);
            document.getElementById('gen-res').style.display = 'block';
            document.getElementById('gen-priv').innerText = 'Privada: ' + p;
            document.getElementById('gen-pub').innerText = 'Pública: ' + pb;
        }

        load(); setInterval(load, 15000);
    </script>
</body>
</html>
`
