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
		PrevHash: strings.Repeat("0", 128),
		Difficulty: Difficulty,
	}
	genesis.Mine()
	Blockchain = append(Blockchain, genesis)

	// --- API ROUTES ---
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
    <title>AstraliX | Elite Financial Network</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --ax-blue: #003366; --ax-celeste: #74ACDF; --ax-bg: #F4F7F9; --white: #FFFFFF; }
        body { background: var(--ax-bg); color: #334155; font-family: 'Inter', sans-serif; margin: 0; }
        .sidebar { background: var(--ax-blue); height: 100vh; position: fixed; width: 280px; color: white; z-index: 1000; box-shadow: 10px 0 30px rgba(0,0,0,0.05); }
        .content { margin-left: 280px; padding: 40px; min-height: 100vh; }
        .nav-link { color: rgba(255,255,255,0.6); padding: 16px 30px; margin: 8px 20px; border-radius: 16px; font-weight: 600; cursor: pointer; display: flex; align-items: center; text-decoration: none; transition: 0.3s; }
        .nav-link:hover, .nav-link.active { background: rgba(255,255,255,0.1); color: var(--ax-celeste); }
        .nav-link i { width: 35px; font-size: 1.2rem; }
        .card-pro { background: var(--white); border: none; border-radius: 30px; box-shadow: 0 10px 40px rgba(0,51,102,0.04); padding: 35px; margin-bottom: 30px; }
        .hero-balance { background: linear-gradient(135deg, var(--ax-blue) 0%, #001f3d 100%); color: white; position: relative; overflow: hidden; }
        .hero-balance::after { content: ""; position: absolute; top: -50px; right: -50px; width: 200px; height: 200px; background: var(--ax-celeste); filter: blur(100px); opacity: 0.2; }
        .btn-ax { background: var(--ax-blue); color: white; border-radius: 18px; padding: 15px 35px; font-weight: 700; border: none; transition: 0.3s; }
        .btn-ax:hover { background: var(--ax-celeste); transform: translateY(-3px); }
        .addr-pill { background: #F1F5F9; padding: 12px 20px; border-radius: 14px; font-family: monospace; font-size: 0.8rem; color: #4A5568; word-break: break-all; border: 1px solid #E2E8F0; }
        @media (max-width: 992px) { .sidebar { display: none; } .content { margin-left: 0; padding: 20px; } }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="p-5 text-center">
            <h2 class="fw-bold mb-0" style="letter-spacing: -2px; color: var(--ax-celeste);">ASTRALIX</h2>
            <small class="opacity-50 text-uppercase fw-bold" style="font-size: 10px; letter-spacing: 2px;">512-bit Security OS</small>
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
                    <div class="card-pro hero-balance">
                        <small class="text-uppercase opacity-75 fw-bold">Saldo Disponible</small>
                        <h1 id="bal-main" class="display-3 fw-bold my-3">0.00 AX</h1>
                        <div id="addr-main" class="addr-pill bg-white bg-opacity-10 text-white border-0 opacity-75">Sincroniza tu cuenta en Seguridad</div>
                    </div>
                    <div class="card-pro">
                        <h5 class="fw-bold mb-4">Feed de Red</h5>
                        <div id="mini-feed"></div>
                    </div>
                </div>
                <div class="col-lg-4">
                    <div class="card-pro text-center">
                        <h5 class="fw-bold mb-4">Estado del Nodo</h5>
                        <div class="mb-4">
                            <small class="text-muted d-block">Altura de Bloques</small>
                            <h4 id="h-stat" class="fw-bold text-primary">0</h4>
                        </div>
                        <button class="btn btn-ax w-100 py-3" onclick="mine()">MINAR BLOQUE</button>
                    </div>
                </div>
            </div>
        </div>

        <div id="v-wallet" class="view" style="display:none">
            <div class="card-pro mx-auto" style="max-width: 600px;">
                <h4 class="fw-bold mb-4">Enviar Fondos</h4>
                <input type="text" id="tx-to" class="form-control mb-3 p-3 border-0 bg-light rounded-4" placeholder="Dirección del destinatario (AX...)">
                <input type="number" id="tx-amt" class="form-control mb-4 p-3 border-0 bg-light rounded-4" placeholder="Monto AX">
                <button class="btn btn-ax w-100 py-3" onclick="send()">FIRMAR Y ENVIAR</button>
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
                <h4 class="fw-bold mb-3">Acceso 512-bit</h4>
                <p class="small text-muted mb-4">Ingresa tu Clave Privada. Tu dirección se derivará mediante SHA-512.</p>
                <input type="password" id="i-priv" class="form-control mb-3 p-3 border-0 bg-light rounded-4" placeholder="Clave Secreta">
                <button class="btn btn-ax w-100" onclick="login()">CONECTAR BILLETERA</button>
                <hr class="my-5">
                <button class="btn btn-outline-dark w-100 py-3 rounded-4" onclick="gen()">GENERAR NUEVA IDENTIDAD</button>
                <div id="g-res" class="mt-4" style="display:none">
                    <small class="text-muted">Semilla / Privada:</small><div class="addr-pill mb-2" id="g-priv"></div>
                    <small class="text-muted">Dirección AX Generada:</small><div class="addr-pill fw-bold text-primary" id="g-pub"></div>
                </div>
            </div>
        </div>
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
            if(el) {
                document.querySelectorAll('.nav-link').forEach(function(n) { n.classList.remove('active'); });
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
            const mini = document.getElementById('mini-feed');
            const full = document.getElementById('full-chain');
            mini.innerHTML = ''; full.innerHTML = '';
            
            chain.reverse().forEach(function(b) {
                const bHash = b.Hash || b.hash || "";
                const bIdx = (b.Index !== undefined) ? b.Index : b.index;
                mini.innerHTML += '<div class="p-3 border-bottom d-flex justify-content-between"><span>Bloque #' + bIdx + '</span><span class="text-muted">' + bHash.substring(0,25) + '...</span></div>';
                full.innerHTML += '<div class="card-pro border mb-3"><h6>Bloque #' + bIdx + '</h6><div class="addr-pill">' + bHash + '</div></div>';
            });
        }

        async function mine() { await fetch('/api/mine'); alert("¡Bloque Minado!"); load(); }
        async function send() {
            if(!session) return nav('security');
            const tx = { sender: session.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            alert("¡Enviada! Miná el bloque para confirmar."); nav('dash'); load();
        }

        async function gen() {
            const words = ["nube", "chaco", "cable", "red", "fibra", "datos", "seguro", "cripto", "luz", "onda", "pampa", "rio"];
            const seed = [1,2,3,4,5,6,7,8,9,10,11,12].map(function() { return words[Math.floor(Math.random()*words.length)]; }).join(' ');
            const p = btoa(seed + Date.now()).substring(0,64);
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
