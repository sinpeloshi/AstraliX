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
	// 🎯 DIRECCIÓN MAESTRA SINCRONIZADA
	creatorAddr := "AX5eaba583bf646e0e39f41da6f9d8fa6db929c2e858bd32dffe6ac0cee2e3e974"

	genTx := core.Transaction{Sender: "SYSTEM", Recipient: creatorAddr, Amount: 1000002021}
	genTx.TxID = genTx.CalculateHash()
	genesis := core.Block{
		Index: 0, Timestamp: 1773561600,
		Transactions: []core.Transaction{genTx},
		PrevHash: strings.Repeat("0", 128), Difficulty: Difficulty,
	}
	genesis.Mine()
	Blockchain = append(Blockchain, genesis)

	// API Handlers
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
			http.Error(w, "Err", 400); return
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
	fmt.Printf("🌐 AstraliX Argentum Live on %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>AstraliX | Digital Assets OS</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --ax-blue: #003366; --ax-celeste: #74ACDF; --bg: #F4F7F9; --white: #FFFFFF; }
        body { background: var(--bg); color: #334155; font-family: 'Segoe UI', sans-serif; margin: 0; padding-bottom: 90px; }
        
        .top-bar { background: var(--white); padding: 15px 25px; border-bottom: 1px solid #E2E8F0; position: sticky; top: 0; z-index: 1000; }
        .logo-txt { font-weight: 900; color: var(--ax-blue); letter-spacing: -1.5px; font-size: 1.4rem; }

        .bottom-nav { background: var(--white); position: fixed; bottom: 0; width: 100%; height: 75px; display: flex; border-top: 1px solid #E2E8F0; z-index: 2000; box-shadow: 0 -5px 15px rgba(0,0,0,0.03); }
        .nav-item { flex: 1; text-align: center; color: #94A3B8; cursor: pointer; text-decoration: none; font-size: 10px; font-weight: 700; padding: 12px; transition: 0.3s; }
        .nav-item.active { color: var(--ax-blue); }
        .nav-item i { font-size: 20px; display: block; margin-bottom: 4px; }

        .card-pro { background: var(--white); border-radius: 28px; border: none; box-shadow: 0 10px 30px rgba(0,51,102,0.05); padding: 25px; margin-bottom: 20px; }
        .hero-card { background: linear-gradient(135deg, var(--ax-blue) 0%, var(--ax-celeste) 100%); color: white; padding: 35px 25px; }
        
        .btn-ax { background: var(--ax-blue); color: white; border-radius: 16px; padding: 15px; font-weight: 700; border: none; width: 100%; transition: 0.3s; }
        .form-ax { background: #F1F5F9; border: 1px solid #E2E8F0; border-radius: 16px; padding: 15px; width: 100%; margin-bottom: 15px; }
        
        .addr-pill { background: rgba(0,0,0,0.05); padding: 12px; border-radius: 14px; font-family: monospace; font-size: 11px; word-break: break-all; margin-top: 10px; position: relative; }
        .copy-btn { position: absolute; top: 5px; right: 5px; color: var(--ax-celeste); background: none; border: none; font-size: 12px; }
        
        .word-badge { display: inline-block; background: var(--ax-blue); color: white; padding: 5px 12px; border-radius: 10px; margin: 4px; font-size: 12px; font-weight: 600; }
        .status-badge { background: #DCFCE7; color: #166534; padding: 4px 10px; border-radius: 20px; font-size: 10px; font-weight: 800; }
    </style>
</head>
<body>

    <div class="top-bar d-flex justify-content-between align-items-center">
        <div class="logo-txt">ASTRALIX</div>
        <span class="status-badge"><i class="fas fa-circle me-1" style="font-size: 7px;"></i> NODO ACTIVO</span>
    </div>

    <div class="container mt-4">
        <div id="v-dash" class="view">
            <div class="card-pro hero-card text-center">
                <small class="text-uppercase fw-bold opacity-75">Saldo Disponible</small>
                <h1 id="bal-txt" class="display-4 fw-bold my-2">0.00 AX</h1>
                <div id="addr-txt" class="addr-pill">Sin sincronizar</div>
            </div>
            <div class="card-pro">
                <h6 class="fw-bold mb-3">Última Actividad de Red</h6>
                <div id="mini-feed"></div>
            </div>
        </div>

        <div id="v-wallet" class="view" style="display:none">
            <div class="card-pro">
                <h4 class="fw-bold mb-4">Enviar AX</h4>
                <input type="text" id="tx-to" class="form-ax" placeholder="Dirección del destinatario">
                <input type="number" id="tx-amt" class="form-ax" placeholder="Monto">
                <button class="btn-ax" onclick="send()">FIRMAR Y ENVIAR</button>
            </div>
            <div class="card-pro bg-light border-0 text-center">
                <h6 class="fw-bold mb-3">Acción de Red</h6>
                <button class="btn btn-outline-dark w-100 rounded-4 py-3" onclick="mine()">MINAR BLOQUE PENDIENTE</button>
            </div>
        </div>

        <div id="v-security" class="view" style="display:none">
            <div class="card-pro">
                <h4 class="fw-bold mb-3">Login Seguro</h4>
                <p class="small text-muted mb-4">Ingresa tu clave de 512 bits para sincronizar tu balance.</p>
                <input type="password" id="i-priv" class="form-ax" placeholder="Private Key">
                <button class="btn-ax mb-3" onclick="login()">CONECTAR</button>
                <hr class="my-4">
                <button class="btn btn-link text-danger w-100 text-decoration-none small" onclick="logout()">Cerrar Sesión</button>
            </div>

            <div class="card-pro">
                <h4 class="fw-bold mb-3">Generar Semilla (Seed)</h4>
                <p class="small text-muted mb-4">Genera una nueva identidad protegida por 12 palabras.</p>
                <button class="btn btn-outline-primary w-100 py-3 rounded-4" onclick="gen()">GENERAR NUEVAS LLAVES</button>
                
                <div id="g-res" class="mt-4" style="display:none">
                    <label class="small fw-bold text-muted mb-2 d-block">12 Palabras de Seguridad:</label>
                    <div id="seed-words" class="mb-4"></div>
                    
                    <label class="small fw-bold text-muted d-block">Clave Privada (512-bit):</label>
                    <div class="addr-pill mb-3" id="g-priv"></div>
                    
                    <label class="small fw-bold text-muted d-block">Dirección Pública AX:</label>
                    <div class="addr-pill fw-bold text-primary" id="g-pub"></div>
                    
                    <div class="alert alert-warning small mt-3 border-0 rounded-4">
                        <i class="fas fa-exclamation-triangle me-1"></i> Guarda esto en papel. No se puede recuperar.
                    </div>
                </div>
            </div>
        </div>
        
        <div id="v-explorer" class="view" style="display:none">
            <div class="card-pro">
                <h4 class="fw-bold mb-4">Blockchain Explorer</h4>
                <div id="full-chain"></div>
            </div>
        </div>
    </div>

    <div class="bottom-nav">
        <div class="nav-item active" onclick="nav('dash', this)"><i class="fas fa-home"></i>Inicio</div>
        <div class="nav-item" onclick="nav('wallet', this)"><i class="fas fa-paper-plane"></i>Enviar</div>
        <div class="nav-item" onclick="nav('explorer', this)"><i class="fas fa-database"></i>Red</div>
        <div class="nav-item" onclick="nav('security', this)"><i class="fas fa-shield-halved"></i>Keys</div>
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
            document.querySelectorAll('.nav-item').forEach(function(n) { n.classList.remove('active'); });
            if(el) el.classList.add('active');
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
                document.getElementById('addr-txt').innerText = session.pub.substring(0,25) + "...";
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
                const bHash = (b.Hash || b.hash || "");
                mini.innerHTML += '<div class="small border-bottom py-2 d-flex justify-content-between"><span>#' + b.index + '</span><span class="text-muted">' + bHash.substring(0,25) + '...</span></div>';
                full.innerHTML += '<div class="card-pro border mb-3 small"><b>Bloque #' + b.index + '</b><div class="addr-pill">' + bHash + '</div></div>';
            });
        }

        async function mine() { await fetch('/api/mine'); alert('¡Mined!'); load(); }
        async function send() {
            if(!session) return nav('security');
            const tx = { sender: session.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            alert('¡Enviado! Mina el bloque para confirmar.'); nav('dash'); load();
        }

        async function gen() {
            const words = ["nube", "chaco", "cable", "red", "fibra", "datos", "seguro", "cripto", "luz", "onda", "pampa", "rio", "noche", "viento", "sol", "fuego"];
            const seed = [1,2,3,4,5,6,7,8,9,10,11,12].map(function() { return words[Math.floor(Math.random()*words.length)]; }).join(' ');
            const p = btoa(seed + Date.now()).substring(0,64);
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
