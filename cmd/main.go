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
	// RECUERDA: Pon aquí tu dirección AX generada con tu clave de 512 bits
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

	// --- API ROUTES ---
	http.HandleFunc("/api/balance/", func(w http.ResponseWriter, r *http.Request) {
		addr := strings.TrimPrefix(r.URL.Path, "/api/balance/")
		balance := 0.0
		for _, b := range Blockchain {
			for _, tx := range b.Transactions {
				if tx.Recipient == addr { balance += tx.Amount }
				if tx.Sender == addr { balance -= tx.Amount }
			}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"balance": balance})
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
			http.Error(w, "Invalid", 400); return
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
	fmt.Printf("🌐 AstraliX 512-bit Node running on port %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>AstraliX | Network L1 OS</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --ax-blue: #003366; --ax-celeste: #74ACDF; --ax-cyan: #00ffa3; --bg: #f8f9fc; }
        body { background: var(--bg); color: #333; font-family: 'Inter', system-ui, sans-serif; }
        
        .sidebar { background: var(--ax-blue); height: 100vh; position: fixed; width: 260px; color: white; transition: 0.3s; z-index: 1000; box-shadow: 4px 0 15px rgba(0,0,0,0.1); }
        .content { margin-left: 260px; padding: 40px; }
        
        .nav-link { color: rgba(255,255,255,0.6); padding: 14px 22px; margin: 8px 15px; border-radius: 12px; transition: 0.3s; cursor: pointer; text-decoration: none; display: flex; align-items: center; font-weight: 500; }
        .nav-link:hover, .nav-link.active { background: rgba(255,255,255,0.1); color: white; border-left: 4px solid var(--ax-celeste); }
        .nav-link i { width: 30px; font-size: 18px; }

        .card-elite { background: white; border: none; border-radius: 24px; box-shadow: 0 10px 40px rgba(0,51,102,0.06); padding: 30px; margin-bottom: 25px; transition: 0.3s; }
        .card-elite:hover { transform: translateY(-5px); box-shadow: 0 15px 45px rgba(0,51,102,0.1); }
        
        .balance-hero { background: linear-gradient(135deg, var(--ax-blue) 0%, #001a33 100%); color: white; position: relative; overflow: hidden; }
        .balance-hero::after { content: ""; position: absolute; top: -50%; right: -20%; width: 300px; height: 300px; background: var(--ax-celeste); filter: blur(100px); opacity: 0.2; }
        
        .btn-ax { background: var(--ax-blue); color: white; border-radius: 14px; padding: 14px 28px; font-weight: 700; border: none; transition: 0.3s; }
        .btn-ax:hover { background: var(--ax-celeste); transform: scale(1.02); }
        
        .addr-box { background: #f1f3f5; padding: 12px 18px; border-radius: 12px; font-family: monospace; font-size: 0.85rem; color: #555; word-break: break-all; border: 1px solid #e9ecef; }
        .badge-512 { background: rgba(0, 255, 163, 0.1); color: #00a86b; font-weight: 800; padding: 5px 12px; border-radius: 8px; font-size: 11px; }

        @media (max-width: 992px) { .sidebar { transform: translateX(-100%); } .content { margin-left: 0; padding: 20px; } .sidebar.active { transform: translateX(0); } }
    </style>
</head>
<body>

    <div class="sidebar" id="sidebar">
        <div class="p-4 mb-4 text-center">
            <h2 class="fw-black mb-0" style="letter-spacing: -2px;">ASTRALIX</h2>
            <div class="badge-512">512-BIT SECURE</div>
        </div>
        <nav>
            <a class="nav-link active" onclick="show('dash', this)"><i class="fas fa-th-large"></i> Dashboard</a>
            <a class="nav-link" onclick="show('wallet', this)"><i class="fas fa-wallet"></i> My Assets</a>
            <a class="nav-link" onclick="show('explorer', this)"><i class="fas fa-database"></i> Network Explorer</a>
            <a class="nav-link" onclick="show('security', this)"><i class="fas fa-shield-halved"></i> Key Management</a>
        </nav>
    </div>

    <div class="content">
        <div id="view-dash" class="view">
            <div class="row g-4">
                <div class="col-lg-8">
                    <div class="card-elite balance-hero">
                        <div class="row align-items-center">
                            <div class="col-md-8">
                                <small class="text-uppercase opacity-50 fw-bold">Portfolio Balance</small>
                                <h1 id="bal-main" class="display-3 fw-bold my-2">0.00 AX</h1>
                                <div id="pub-addr" class="small opacity-50 text-truncate">Connect your 512-bit keys to sync</div>
                            </div>
                            <div class="col-md-4 text-end d-none d-md-block">
                                <i class="fas fa-satellite-dish fa-6x opacity-10"></i>
                            </div>
                        </div>
                    </div>
                    
                    <div class="card-elite">
                        <div class="d-flex justify-content-between align-items-center mb-4">
                            <h5 class="fw-bold m-0">Live Network Feed</h5>
                            <button class="btn btn-sm btn-outline-primary rounded-pill px-3" onclick="mine()">Mine Next Block</button>
                        </div>
                        <div id="mini-chain"></div>
                    </div>
                </div>
                
                <div class="col-lg-4">
                    <div class="card-elite">
                        <h5 class="fw-bold mb-4">Quick Actions</h5>
                        <button class="btn btn-ax w-100 mb-3" onclick="show('wallet')">Transfer Tokens</button>
                        <button class="btn btn-light w-100 border py-3 rounded-4" onclick="show('security')">Check Keys</button>
                        <hr class="my-4">
                        <div class="d-flex justify-content-between mb-2"><small>Network Status</small><span class="badge bg-success rounded-pill">Active</span></div>
                        <div class="d-flex justify-content-between mb-2"><small>Protocol</small><b style="font-size: 12px;">SHA-512 (Elite)</b></div>
                    </div>
                </div>
            </div>
        </div>

        <div id="view-wallet" class="view" style="display:none">
            <div class="card-elite max-w-500 mx-auto">
                <h4 class="fw-bold mb-4">New Transaction</h4>
                <div class="mb-3">
                    <label class="small fw-bold text-muted mb-2">Recipient Address</label>
                    <input type="text" id="tx-to" class="form-control p-3 border-0 bg-light" placeholder="AX...">
                </div>
                <div class="mb-4">
                    <label class="small fw-bold text-muted mb-2">Amount (AX)</label>
                    <input type="number" id="tx-amt" class="form-control p-3 border-0 bg-light" placeholder="0.00">
                </div>
                <button class="btn btn-ax w-100 py-3" onclick="sendTx()">BROADCAST TO NETWORK</button>
            </div>
        </div>

        <div id="view-security" class="view" style="display:none">
            <div class="row g-4">
                <div class="col-md-6">
                    <div class="card-elite h-100">
                        <h4 class="fw-bold mb-4">Import Identity</h4>
                        <input type="password" id="imp-priv" class="form-control p-3 border-0 bg-light mb-3" placeholder="Enter Private Key">
                        <button class="btn btn-ax w-100" onclick="importWallet()">SYNC WITH NODE</button>
                        <button class="btn btn-link w-100 text-danger mt-3 text-decoration-none" onclick="logout()">Logout</button>
                    </div>
                </div>
                <div class="col-md-6">
                    <div class="card-elite h-100">
                        <h4 class="fw-bold mb-4">Seed Phrase</h4>
                        <button class="btn btn-outline-dark w-100 py-3" onclick="gen()">Generate 12 Words</button>
                        <div id="seed-res" class="mt-3" style="display:none">
                            <div class="addr-box small" id="gen-priv"></div>
                            <div class="addr-box small mt-2 fw-bold text-primary" id="gen-pub"></div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        async function derive(priv) {
            const msg = new TextEncoder().encode(priv);
            const hash = await crypto.subtle.digest('SHA-512', msg);
            return 'AX' + Array.from(new Uint8Array(hash)).map(b => b.toString(16).padStart(2,'0')).join('').substring(0,64);
        }

        let wallet = JSON.parse(localStorage.getItem('ax_session')) || null;

        function show(id, el) {
            document.querySelectorAll('.view').forEach(v => v.style.display = 'none');
            document.getElementById('view-' + id).style.display = 'block';
            if(el) { document.querySelectorAll('.nav-link').forEach(n => n.classList.remove('active')); el.classList.add('active'); }
        }

        async function importWallet() {
            const priv = document.getElementById('imp-priv').value;
            const pub = await derive(priv);
            wallet = { pub, priv };
            localStorage.setItem('ax_session', JSON.stringify(wallet));
            location.reload();
        }

        function logout() { localStorage.removeItem('ax_session'); location.reload(); }

        async function load() {
            if(wallet) {
                document.getElementById('pub-addr').innerText = wallet.pub;
                const r = await fetch('/api/balance/' + wallet.pub);
                const d = await r.json();
                document.getElementById('bal-main').innerText = d.balance.toLocaleString() + ' AX';
            }
            const rc = await fetch('/api/chain');
            const c = await rc.json();
            const feed = document.getElementById('mini-chain');
            feed.innerHTML = '';
            c.reverse().forEach(b => {
                feed.innerHTML += ` + "`" + `<div class="p-3 border-bottom d-flex justify-content-between align-items-center"><div class="small"><b>Block #${b.index}</b><br><span class="text-muted">${b.hash.substring(0,32)}...</span></div><small class="text-muted">${new Date(b.timestamp*1000).toLocaleTimeString()}</small></div>` + "`" + `;
            });
        }

        async function mine() { await fetch('/api/mine'); alert("Mined!"); load(); }
        async function sendTx() {
            const tx = { sender: wallet.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            alert("Sent!"); show('dash'); load();
        }

        async function gen() {
            const p = btoa(Math.random()).substring(0,64);
            const pb = await derive(p);
            document.getElementById('seed-res').style.display = 'block';
            document.getElementById('gen-priv').innerText = 'Priv: ' + p;
            document.getElementById('gen-pub').innerText = 'Pub: ' + pb;
        }

        load(); setInterval(load, 20000);
    </script>
</body>
</html>
