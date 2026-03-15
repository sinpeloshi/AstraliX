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
	// Dirección Maestra Denis W. Sanchez
	creatorAddr := "AX5eaba583bf646e0e39f41da6f9d8fa6db929c2e858bd32dffe6ac0cee2e3e974"

	// Bloque Génesis
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

	// NUEVA LÓGICA DE MINADO CON RECOMPENSA Y HALVING
	http.HandleFunc("/api/mine", func(w http.ResponseWriter, r *http.Request) {
		minerAddr := r.URL.Query().Get("address")
		if minerAddr == "" { http.Error(w, "Address required", 400); return }

		// Cálculo de Recompensa (50 AX con halving cada 210.000 bloques)
		height := int64(len(Blockchain))
		halvings := height / 210000
		reward := 50.0 / float64(int(1)<<halvings)

		// Transacción de Recompensa
		cbTx := core.Transaction{Sender: "SYSTEM", Recipient: minerAddr, Amount: reward}
		cbTx.TxID = cbTx.CalculateHash()

		// Unir recompensas + mempool
		blockTxs := append(Mempool, cbTx)

		last := Blockchain[len(Blockchain)-1]
		newB := core.Block{
			Index: height, Timestamp: time.Now().Unix(),
			Transactions: blockTxs, PrevHash: last.Hash, Difficulty: Difficulty,
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
	fmt.Printf("🌐 AstraliX Argentum v11.5 on %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>AstraliX | Network L1</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --ax-blue: #003366; --ax-celeste: #74ACDF; --bg: #F8FAFC; --white: #FFFFFF; }
        body { background: var(--bg); color: #334155; font-family: 'Segoe UI', system-ui, sans-serif; margin: 0; padding-bottom: 90px; }
        
        /* Layout */
        .sidebar { background: var(--ax-blue); height: 100vh; position: fixed; width: 280px; color: white; z-index: 1000; box-shadow: 10px 0 30px rgba(0,0,0,0.05); }
        .main-content { margin-left: 280px; padding: 40px; min-height: 100vh; transition: 0.3s; }
        
        .nav-link-ax { color: rgba(255,255,255,0.6); padding: 15px 25px; margin: 8px 15px; border-radius: 12px; font-weight: 600; cursor: pointer; display: flex; align-items: center; text-decoration: none; }
        .nav-link-ax:hover, .nav-link-ax.active { background: var(--ax-celeste); color: white; }
        .nav-link-ax i { width: 30px; font-size: 1.1rem; }

        .mobile-nav { background: var(--white); position: fixed; bottom: 0; width: 100%; height: 80px; display: none; justify-content: space-around; align-items: center; border-top: 1px solid #E2E8F0; z-index: 2000; box-shadow: 0 -5px 20px rgba(0,0,0,0.03); }
        .m-nav-item { color: #94A3B8; text-align: center; font-size: 10px; font-weight: 800; cursor: pointer; flex: 1; }
        .m-nav-item.active { color: var(--ax-blue); }
        .m-nav-item i { font-size: 24px; display: block; margin-bottom: 4px; }

        .card-pro { background: var(--white); border-radius: 28px; box-shadow: 0 10px 30px rgba(0,51,102,0.05); padding: 25px; margin-bottom: 25px; border: none; }
        .hero-balance { background: linear-gradient(135deg, var(--ax-blue) 0%, var(--ax-celeste) 100%); color: white; padding: 40px 25px; }
        
        .btn-ax { background: var(--ax-blue); color: white; border-radius: 16px; padding: 16px; font-weight: 700; border: none; width: 100%; transition: 0.3s; }
        .btn-ax:hover { background: var(--ax-celeste); transform: translateY(-2px); }
        
        .addr-pill { background: rgba(0,0,0,0.05); padding: 12px; border-radius: 14px; font-family: monospace; font-size: 0.8rem; word-break: break-all; margin-top: 10px; }
        .status-badge { background: #D1FAE5; color: #065F46; padding: 5px 12px; border-radius: 10px; font-size: 10px; font-weight: 800; }

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
            <small class="opacity-50 fw-bold" style="font-size: 10px; text-transform: uppercase;">Enterprise Node</small>
        </div>
        <nav>
            <div class="nav-link-ax active" onclick="nav('dash', this)"><i class="fas fa-home"></i> Inicio</div>
            <div class="nav-link-ax" onclick="nav('wallet', this)"><i class="fas fa-wallet"></i> Billetera</div>
            <div class="nav-link-ax" onclick="nav('explorer', this)"><i class="fas fa-database"></i> Red</div>
            <div class="nav-link-ax" onclick="nav('sec', this)"><i class="fas fa-shield-halved"></i> Seguridad</div>
        </nav>
    </div>

    <div class="main-content">
        <div class="d-flex justify-content-between align-items-center mb-5">
            <div><h3 class="fw-bold m-0 text-dark">Consola de Red</h3><span class="status-badge">ACTIVO</span></div>
            <button class="btn btn-white shadow-sm rounded-circle p-3" onclick="location.reload()"><i class="fas fa-sync text-primary"></i></button>
        </div>

        <div id="v-dash" class="view">
            <div class="card-pro hero-balance text-center">
                <small class="text-uppercase fw-bold opacity-75">Saldo Disponible</small>
                <h1 id="bal-txt" class="display-3 fw-bold my-2">0.00 AX</h1>
                <div id="addr-txt" class="addr-pill bg-white bg-opacity-10 text-white border-0 opacity-75">Sincroniza en 'Seguridad'</div>
            </div>
            <div class="card-pro">
                <h5 class="fw-bold mb-4">Últimos Movimientos</h5>
                <div id="mini-feed"></div>
            </div>
        </div>

        <div id="v-wallet" class="view" style="display:none">
            <div class="card-pro mx-auto" style="max-width: 500px;">
                <h4 class="fw-bold mb-4 text-primary">Transferir AX</h4>
                <input type="text" id="tx-to" class="form-control mb-3 p-3 border-0 bg-light rounded-4" placeholder="Dirección AX...">
                <input type="number" id="tx-amt" class="form-control mb-4 p-3 border-0 bg-light rounded-4" placeholder="Monto">
                <button class="btn btn-ax py-3" onclick="send()">FIRMAR Y ENVIAR</button>
            </div>
            <div class="card-pro bg-light border-0 text-center">
                <h6 class="fw-bold mb-3">Recompensa de Bloque: 50 AX</h6>
                <button class="btn btn-outline-dark w-100 py-3 rounded-4" onclick="mine()">MINAR BLOQUE AHORA</button>
            </div>
        </div>

        <div id="v-sec" class="view" style="display:none">
            <div class="card-pro">
                <h4 class="fw-bold mb-4">Acceso 512-bit</h4>
                <input type="password" id="i-priv" class="form-control mb-3 p-3 border-0 bg-light rounded-4" placeholder="Private Key Secreta">
                <button class="btn btn-ax mb-3" onclick="login()">CONECTAR</button>
                <hr class="my-5">
                <button class="btn btn-outline-dark w-100 py-3 rounded-4" onclick="gen()">GENERAR SEMILLA</button>
                <div id="g-res" class="mt-4" style="display:none">
                    <label class="small fw-bold">12 Palabras:</label>
                    <div id="seed-words" class="mb-3"></div>
                    <label class="small fw-bold">Privada:</label><div class="addr-pill mb-2" id="g-priv"></div>
                    <label class="small fw-bold text-primary">Dirección AX:</label><div class="addr-pill fw-bold" id="g-pub"></div>
                </div>
            </div>
        </div>

        <div id="v-explorer" class="view" style="display:none">
            <div class="card-pro">
                <h4 class="fw-bold mb-4 text-primary">Blockchain Explorer</h4>
                <div id="full-chain"></div>
            </div>
        </div>
    </div>

    <div class="mobile-nav">
        <div class="m-nav-item active" onclick="nav('dash', this)"><i class="fas fa-home"></i>Dash</div>
        <div class="m-nav-item" onclick="nav('wallet', this)"><i class="fas fa-paper-plane"></i>Enviar</div>
        <div class="m-nav-item" onclick="nav('explorer', this)"><i class="fas fa-database"></i>Red</div>
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
                const h = (b.Hash || b.hash || '').substring(0,25) + '...';
                mini.innerHTML += '<div class="p-3 border-bottom d-flex justify-content-between small"><span>#' + b.index + '</span><span class="text-muted">' + h + '</span></div>';
                full.innerHTML += '<div class="card-pro border mb-3"><h6>Bloque #' + b.index + '</h6><div class="addr-pill">' + (b.Hash || b.hash) + '</div></div>';
            });
        }

        async function mine() {
            if(!session) return alert('Sincroniza tu billetera para cobrar la recompensa.');
            const r = await fetch('/api/mine?address=' + session.pub);
            if(r.ok) { alert('¡Mined! Ganaste la recompensa de bloque.'); load(); }
            else { alert('Nada para minar todavía.'); }
        }

        async function send() {
            if(!session) return nav('sec');
            const tx = { sender: session.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            alert('¡Transacción enviada!'); nav('dash'); load();
        }

        async function gen() {
            const words = ["nube", "chaco", "cable", "red", "fibra", "datos", "seguro", "cripto", "luz", "onda", "pampa", "rio"];
            const seed = [1,2,3,4,5,6,7,8,9,10,11,12].map(function() { return words[Math.floor(Math.random()*words.length)]; }).join(' ');
            const p = btoa(seed + Date.now()).substring(0,64);
            const pb = await derive(p);
            document.getElementById('g-res').style.display = 'block';
            document.getElementById('seed-words').innerHTML = seed.split(' ').map(function(w) { return '<span class="badge bg-primary m-1">'+w+'</span>'; }).join('');
            document.getElementById('g-priv').innerText = p;
            document.getElementById('g-pub').innerText = pb;
        }

        load(); setInterval(load, 15000);
    </script>
</body>
</html>
`
