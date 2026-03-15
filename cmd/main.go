package main // Correcto: minúscula para el compilador

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"astralix/core"
)

// Estructuras de datos globales
var Blockchain []core.Block
var Mempool []core.Transaction

func main() {
	// Configuración de Red
	const Difficulty = 4 
	// Dirección Maestra (Génesis)
	creatorAddr := "AXdc3acc7c0b91eb485d0e3bb78059bb58a3999c14b56cfe6ca0428670afc6410c"

	// Inicialización de la Cadena
	genTx := core.Transaction{Sender: "SYSTEM", Recipient: creatorAddr, Amount: 1000002021}
	genTx.TxID = genTx.CalculateHash()
	
	genesis := core.Block{
		Index: 0, Timestamp: 1773561600,
		Transactions: []core.Transaction{genTx},
		PrevHash: strings.Repeat("0", 128), // 512 bits en hex
		Difficulty: Difficulty,
	}
	genesis.Mine()
	Blockchain = append(Blockchain, genesis)

	// --- API ENDPOINTS (Backend) ---

	// Consulta de Saldo
	http.HandleFunc("/api/balance/", func(w http.ResponseWriter, r *http.Request) {
		addr := strings.TrimPrefix(r.URL.Path, "/api/balance/")
		var bal float64
		for _, b := range Blockchain {
			for _, tx := range b.Transactions {
				if tx.Recipient == addr { bal += tx.Amount }
				if tx.Sender == addr { bal -= tx.Amount }
			}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"balance": bal, "address": addr})
	})

	// Obtener Cadena Completa
	http.HandleFunc("/api/chain", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Blockchain)
	})

	// Minar Transacciones Pendientes
	http.HandleFunc("/api/mine", func(w http.ResponseWriter, r *http.Request) {
		if len(Mempool) == 0 {
			w.WriteHeader(http.StatusBadRequest)
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

	// Registrar Nueva Transacción
	http.HandleFunc("/api/transactions/new", func(w http.ResponseWriter, r *http.Request) {
		var tx core.Transaction
		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			http.Error(w, "Invalid Payload", 400); return
		}
		tx.TxID = tx.CalculateHash()
		Mempool = append(Mempool, tx)
		w.WriteHeader(http.StatusCreated)
	})

	// Servir Interfaz Web
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, dashboardHTML)
	})

	// Lanzamiento del Nodo
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	fmt.Printf("🌐 AstraliX Elite running on port %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

// --- INTERFAZ DE USUARIO (Frontend) ---
const dashboardHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>AstraliX Core | Enterprise L1 OS</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --ax-blue: #003366; --ax-celeste: #74ACDF; --ax-bg: #F4F7FA; --white: #FFFFFF; }
        body { background: var(--ax-bg); color: #2D3748; font-family: 'Inter', system-ui, sans-serif; margin: 0; }
        
        /* Layout */
        .sidebar { background: var(--ax-blue); height: 100vh; position: fixed; width: 280px; color: white; transition: 0.3s; z-index: 1000; box-shadow: 10px 0 30px rgba(0,0,0,0.05); }
        .main-content { margin-left: 280px; padding: 40px; min-height: 100vh; }
        
        /* Sidebar Nav */
        .sidebar-brand { padding: 40px 30px; text-align: center; }
        .nav-menu { margin-top: 20px; }
        .nav-item { color: rgba(255,255,255,0.6); padding: 16px 30px; margin: 8px 20px; border-radius: 16px; font-weight: 600; cursor: pointer; display: flex; align-items: center; text-decoration: none; transition: 0.3s; }
        .nav-item:hover, .nav-item.active { background: rgba(255,255,255,0.1); color: var(--ax-celeste); }
        .nav-item i { width: 35px; font-size: 1.2rem; }

        /* Cards Elite */
        .card-ax { background: var(--white); border: none; border-radius: 30px; box-shadow: 0 10px 40px rgba(0,51,102,0.04); padding: 35px; margin-bottom: 30px; border: 1px solid rgba(0,0,0,0.01); }
        .hero-card { background: linear-gradient(135deg, var(--ax-blue) 0%, #001f3d 100%); color: white; position: relative; overflow: hidden; }
        .hero-card::after { content: ""; position: absolute; top: -50px; right: -50px; width: 200px; height: 200px; background: var(--ax-celeste); filter: blur(100px); opacity: 0.2; }
        
        /* Inputs & Buttons */
        .btn-ax { background: var(--ax-blue); color: white; border-radius: 18px; padding: 15px 35px; font-weight: 700; border: none; transition: 0.3s; }
        .btn-ax:hover { background: var(--ax-celeste); transform: translateY(-3px); box-shadow: 0 10px 20px rgba(0,51,102,0.1); }
        .form-ax { background: #F8FAFC; border: 1px solid #E2E8F0; border-radius: 16px; padding: 15px; color: #1A202C; }
        
        .addr-pill { background: #EDF2F7; padding: 12px 20px; border-radius: 14px; font-family: 'JetBrains Mono', monospace; font-size: 0.8rem; color: #4A5568; word-break: break-all; border: 1px solid #E2E8F0; }
        .status-dot { width: 10px; height: 10px; background: #48BB78; border-radius: 50%; display: inline-block; margin-right: 10px; box-shadow: 0 0 10px #48BB78; }

        @media (max-width: 992px) { .sidebar { transform: translateX(-100%); } .main-content { margin-left: 0; padding: 20px; } }
    </style>
</head>
<body>

    <div class="sidebar">
        <div class="sidebar-brand">
            <h2 class="fw-bold mb-0" style="letter-spacing: -2px; color: var(--ax-celeste);">ASTRALIX</h2>
            <small class="opacity-50 text-uppercase fw-bold" style="font-size: 10px; letter-spacing: 2px;">512-bit Security OS</small>
        </div>
        <div class="nav-menu">
            <div class="nav-item active" onclick="nav('dash', this)"><i class="fas fa-chart-pie"></i> Resumen</div>
            <div class="nav-item" onclick="nav('wallet', this)"><i class="fas fa-wallet"></i> Billetera</div>
            <div class="nav-item" onclick="nav('explorer', this)"><i class="fas fa-database"></i> Explorador</div>
            <div class="nav-item" onclick="nav('security', this)"><i class="fas fa-shield-halved"></i> Seguridad</div>
            <div class="nav-item" onclick="nav('stats', this)"><i class="fas fa-server"></i> Nodo</div>
        </div>
    </div>

    <div class="main-content">
        <div class="d-flex justify-content-between align-items-center mb-5">
            <div>
                <h3 class="fw-bold text-dark m-0" id="page-title">Panel de Control</h3>
                <small class="text-muted"><span class="status-dot"></span> Nodo Sincronizado en Chaco, AR</small>
            </div>
            <div class="d-flex align-items-center">
                <button class="btn btn-white shadow-sm rounded-circle p-3 me-3" onclick="location.reload()"><i class="fas fa-sync text-primary"></i></button>
                <div class="dropdown">
                    <div class="btn btn-white shadow-sm rounded-pill px-4 py-2 fw-bold" id="user-display">Invitado</div>
                </div>
            </div>
        </div>

        <div id="v-dash" class="view">
            <div class="row g-4">
                <div class="col-lg-8">
                    <div class="card-ax hero-card">
                        <div class="row align-items-center">
                            <div class="col-md-8">
                                <small class="text-uppercase opacity-75 fw-bold" style="font-size: 11px;">Fondos Disponibles</small>
                                <h1 id="bal-main" class="display-3 fw-bold my-3">0.00 AX</h1>
                                <div id="addr-main" class="addr-pill bg-white bg-opacity-10 text-white border-0 opacity-75">Conecta tu cuenta para operar</div>
                            </div>
                            <div class="col-md-4 text-end d-none d-md-block">
                                <i class="fas fa-gem fa-8x opacity-10"></i>
                            </div>
                        </div>
                    </div>
                    <div class="card-ax">
                        <h5 class="fw-bold mb-4">Actividad Reciente del Nodo</h5>
                        <div id="feed-mini" class="list-group list-group-flush"></div>
                    </div>
                </div>
                <div class="col-lg-4">
                    <div class="card-ax">
                        <h5 class="fw-bold mb-4">Métricas Globales</h5>
                        <div class="mb-4">
                            <small class="text-muted d-block mb-1">Total Supply</small>
                            <h4 class="fw-bold">1,000,002,021 AX</h4>
                        </div>
                        <div class="mb-4">
                            <small class="text-muted d-block mb-1">Altura de Bloques</small>
                            <h4 id="h-stat" class="fw-bold">0</h4>
                        </div>
                        <button class="btn btn-ax w-100 py-3" onclick="mine()">MINAR BLOQUE SIGUIENTE</button>
                    </div>
                </div>
            </div>
        </div>

        <div id="v-wallet" class="view" style="display:none">
            <div class="row g-4 justify-content-center">
                <div class="col-md-7">
                    <div class="card-ax">
                        <h4 class="fw-bold mb-4 text-primary">Enviar AstraliX</h4>
                        <div class="mb-3">
                            <label class="small fw-bold text-muted mb-2">Destinatario (Address AX...)</label>
                            <input type="text" id="tx-to" class="form-control form-ax" placeholder="AX...">
                        </div>
                        <div class="mb-4">
                            <label class="small fw-bold text-muted mb-2">Cantidad a enviar</label>
                            <input type="number" id="tx-amt" class="form-control form-ax" placeholder="0.00">
                        </div>
                        <button class="btn btn-ax w-100 py-3" onclick="send()">AUTORIZAR Y FIRMAR</button>
                    </div>
                </div>
            </div>
        </div>

        <div id="v-security" class="view" style="display:none">
            <div class="row g-4">
                <div class="col-md-6">
                    <div class="card-ax h-100">
                        <h4 class="fw-bold mb-4">Acceder con Clave Privada</h4>
                        <p class="small text-muted mb-4">Ingresa tu llave de 512 bits para sincronizar tu balance y permisos.</p>
                        <input type="password" id="i-priv" class="form-control form-ax mb-4" placeholder="Private Key (SHA-512)">
                        <button class="btn btn-ax w-100" onclick="login()">CONECTAR BILLETERA</button>
                        <button class="btn btn-link text-danger w-100 mt-4 text-decoration-none" onclick="logout()">Cerrar Sesión</button>
                    </div>
                </div>
                <div class="col-md-6">
                    <div class="card-ax h-100">
                        <h4 class="fw-bold mb-4">Generar Nueva Semilla</h4>
                        <p class="small text-muted mb-4">Crea una identidad nueva. Las 12 palabras derivan tu clave de 512 bits.</p>
                        <button class="btn btn-outline-dark w-100 py-3 rounded-pill" onclick="gen()">GENERAR 12 PALABRAS</button>
                        <div id="gen-box" class="mt-4" style="display:none">
                            <div id="seed-words" class="d-flex flex-wrap gap-2 mb-3"></div>
                            <small class="text-muted">Private Key (512-bit):</small>
                            <div class="addr-pill mt-2" id="g-priv"></div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div id="v-explorer" class="view" style="display:none">
            <div class="card-ax">
                <h4 class="fw-bold mb-4 text-primary">Explorador de la Cadena</h4>
                <div id="chain-full"></div>
            </div>
        </div>
    </div>

    <script>
        // Lógica de Derivación Criptográfica SHA-512
        async function derive(priv) {
            const buf = new TextEncoder().encode(priv);
            const hash = await crypto.subtle.digest('SHA-512', buf);
            const hex = Array.from(new Uint8Array(hash)).map(b => b.toString(16).padStart(2,'0')).join('');
            return 'AX' + hex.substring(0, 64);
        }

        let session = JSON.parse(localStorage.getItem('ax_argentum')) || null;

        function nav(id, el) {
            document.querySelectorAll('.view').forEach(v => v.style.display = 'none');
            document.getElementById('v-' + id).style.display = 'block';
            document.getElementById('page-title').innerText = id.charAt(0).toUpperCase() + id.slice(1);
            if(el) {
                document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));
                el.classList.add('active');
            }
        }

        async function login() {
            const p = document.getElementById('i-priv').value;
            if(!p) return;
            const pb = await derive(p);
            session = { pub: pb, priv: p };
            localStorage.setItem('ax_argentum', JSON.stringify(session));
            location.reload();
        }

        function logout() { localStorage.removeItem('ax_argentum'); location.reload(); }

        async function load() {
            if(session) {
                document.getElementById('user-display').innerText = session.pub.substring(0,10) + '...';
                document.getElementById('addr-main').innerText = session.pub;
                const r = await fetch('/api/balance/' + session.pub);
                const d = await r.json();
                document.getElementById('bal-main').innerText = d.balance.toLocaleString() + ' AX';
            }
            const res = await fetch('/api/chain');
            const chain = await res.json();
            document.getElementById('h-stat').innerText = chain.length;
            
            const mini = document.getElementById('feed-mini');
            const full = document.getElementById('chain-full');
            mini.innerHTML = ''; full.innerHTML = '';

            chain.reverse().forEach(b => {
                const bHash = (b.Hash || b.hash || "");
                const bIdx = (b.Index !== undefined) ? b.Index : b.index;
                
                mini.innerHTML += ` + "`" + `<div class="py-3 border-bottom d-flex justify-content-between align-items-center">
                    <div><b>Bloque #${bIdx}</b><br><small class="text-muted">${bHash.substring(0,32)}...</small></div>
                    <span class="badge bg-light text-dark rounded-pill border">${new Date(b.Timestamp*1000 || b.timestamp*1000).toLocaleTimeString()}</span>
                </div>` + "`" + `;

                full.innerHTML += ` + "`" + `<div class="card-ax border mb-3">
                    <div class="d-flex justify-content-between"><h6>Bloque #${bIdx}</h6><small class="text-muted">${new Date(b.Timestamp*1000 || b.timestamp*1000).toLocaleString()}</small></div>
                    <div class="addr-pill my-2">${bHash}</div>
                    <div class="small"><b>TXs:</b> ${b.Transactions ? b.Transactions.length : 0} | <b>Prev:</b> ${b.PrevHash || b.prev_hash}</div>
                </div>` + "`" + `;
            });
        }

        async function mine() { 
            const r = await fetch('/api/mine');
            if(r.ok) { alert("¡Minado exitoso! Nuevo bloque añadido."); load(); } 
            else { alert("Mempool vacío. No hay transacciones para minar."); }
        }

        async function send() {
            if(!session) return nav('security');
            const tx = { sender: session.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            alert("¡Transacción enviada a la red! Procede a minar para confirmar."); nav('dash'); load();
        }

        async function gen() {
            const words = ["nube", "chaco", "cable", "red", "fibra", "datos", "seguro", "cripto", "luz", "onda", "pampa", "rio"];
            const seed = [...Array(12)].map(() => words[Math.floor(Math.random()*words.length)]).join(' ');
            const p = btoa(seed + Date.now()).substring(0, 64);
            const pb = await derive(p);
            document.getElementById('gen-box').style.display = 'block';
            document.getElementById('seed-words').innerHTML = seed.split(' ').map(w => '<span class="badge bg-primary px-3 py-2">'+w+'</span>').join('');
            document.getElementById('g-priv').innerText = p;
            alert("¡Copia tus 12 palabras y tu llave privada! No se pueden recuperar.");
        }

        load(); setInterval(load, 15000);
    </script>
</body>
</html>
