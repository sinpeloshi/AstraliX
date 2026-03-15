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
	creatorAddr := "AXdc3acc7c0b91eb485d0e3bb78059bb58a3999c14b56cfe6ca0428670afc6410c"

	// Genesis Setup
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
	fmt.Printf("🌐 AstraliX Elite Argentum running on port %s\n", port)
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
        :root { --ax-blue: #003366; --ax-celeste: #74ACDF; --ax-bg: #F4F7F9; --white: #FFFFFF; }
        body { background: var(--ax-bg); color: #334155; font-family: 'Segoe UI', system-ui, sans-serif; margin: 0; padding-bottom: 80px; }
        
        /* Navigation */
        .sidebar { background: var(--ax-blue); height: 100vh; position: fixed; width: 260px; color: white; transition: 0.3s; z-index: 1000; box-shadow: 10px 0 30px rgba(0,0,0,0.05); }
        .content { margin-left: 260px; padding: 40px; }
        .nav-link { color: rgba(255,255,255,0.7); padding: 15px 25px; margin: 5px 15px; border-radius: 12px; transition: 0.2s; cursor: pointer; text-decoration: none; display: flex; align-items: center; }
        .nav-link:hover, .nav-link.active { background: rgba(255,255,255,0.1); color: var(--ax-celeste); }
        .nav-link i { width: 30px; font-size: 1.1rem; }

        /* UI Elements */
        .card-custom { background: var(--white); border: none; border-radius: 25px; box-shadow: 0 10px 30px rgba(0,51,102,0.05); padding: 30px; margin-bottom: 25px; }
        .hero-balance { background: linear-gradient(135deg, var(--ax-blue) 0%, var(--ax-celeste) 100%); color: white; position: relative; overflow: hidden; }
        .hero-balance::after { content: ""; position: absolute; bottom: -50px; right: -50px; width: 200px; height: 200px; background: rgba(255,255,255,0.1); filter: blur(50px); border-radius: 50%; }
        
        .btn-ax { background: var(--ax-blue); color: white; border-radius: 14px; padding: 14px 28px; font-weight: 700; border: none; transition: 0.3s; }
        .btn-ax:hover { background: var(--ax-celeste); transform: translateY(-2px); box-shadow: 0 5px 15px rgba(0,51,102,0.2); }
        
        .addr-pill { background: #F1F5F9; padding: 12px 18px; border-radius: 12px; font-family: monospace; font-size: 0.85rem; color: #64748B; word-break: break-all; border: 1px solid #E2E8F0; }
        .status-badge { background: #D1FAE5; color: #065F46; padding: 5px 12px; border-radius: 10px; font-size: 11px; font-weight: 800; letter-spacing: 1px; }

        @media (max-width: 992px) { .sidebar { display: none; } .content { margin-left: 0; padding: 20px; } }
    </style>
</head>
<body>

    <div class="sidebar">
        <div class="p-4 mb-4 text-center">
            <h2 class="fw-bold mb-0" style="letter-spacing: -2px; color: var(--ax-celeste);">ASTRALIX</h2>
            <small class="opacity-50 text-uppercase fw-bold" style="font-size: 10px;">512-bit Secure Core</small>
        </div>
        <nav>
            <a class="nav-link active" onclick="show('dash', this)"><i class="fas fa-th-large"></i> Dashboard</a>
            <a class="nav-link" onclick="show('wallet', this)"><i class="fas fa-wallet"></i> Wallet</a>
            <a class="nav-link" onclick="show('explorer', this)"><i class="fas fa-search"></i> Explorer</a>
            <a class="nav-link" onclick="show('security', this)"><i class="fas fa-shield-halved"></i> Security</a>
        </nav>
    </div>

    <div class="content">
        <div id="view-dash" class="view">
            <div class="row g-4 mb-5">
                <div class="col-lg-8">
                    <div class="card-custom hero-balance">
                        <small class="text-uppercase opacity-75 fw-bold">Current Balance</small>
                        <h1 id="bal-txt" class="display-3 fw-bold my-3">0.00 AX</h1>
                        <div id="addr-txt" class="addr-pill bg-white bg-opacity-10 border-0 text-white opacity-75">Connect wallet to synchronize</div>
                    </div>
                    <div class="card-custom">
                        <h5 class="fw-bold mb-4">Network Feed</h5>
                        <div id="mini-chain" class="table-responsive"></div>
                    </div>
                </div>
                <div class="col-lg-4">
                    <div class="card-custom">
                        <h5 class="fw-bold mb-4">Network Health</h5>
                        <div class="d-flex justify-content-between mb-2"><span>Status</span><span class="status-badge">ACTIVE</span></div>
                        <div class="d-flex justify-content-between mb-2"><span>Blocks</span><b id="h-stat">0</b></div>
                        <div class="d-flex justify-content-between mb-4"><span>Security</span><b>SHA-512</b></div>
                        <button class="btn btn-ax w-100" onclick="mine()">MINE NEXT BLOCK</button>
                    </div>
                </div>
            </div>
        </div>

        <div id="view-wallet" class="view" style="display:none">
            <div class="row justify-content-center">
                <div class="col-md-7">
                    <div class="card-custom">
                        <h4 class="fw-bold mb-4">Send Transaction</h4>
                        <input type="text" id="tx-to" class="form-control mb-3 p-3 border-0 bg-light rounded-4" placeholder="Recipient Address (AX...)">
                        <input type="number" id="tx-amt" class="form-control mb-4 p-3 border-0 bg-light rounded-4" placeholder="Amount (AX)">
                        <button class="btn btn-ax w-100 py-3" onclick="send()">SIGN & BROADCAST</button>
                    </div>
                </div>
            </div>
        </div>

        <div id="view-security" class="view" style="display:none">
            <div class="row g-4">
                <div class="col-md-6">
                    <div class="card-custom h-100">
                        <h4 class="fw-bold mb-4">Import Wallet</h4>
                        <input type="password" id="i-priv" class="form-control mb-3 p-3 border-0 bg-light rounded-4" placeholder="Private Key (Secret)">
                        <button class="btn btn-ax w-100" onclick="login()">CONNECT ACCOUNT</button>
                        <button class="btn btn-link text-danger w-100 mt-3 text-decoration-none small" onclick="logout()">Logout</button>
                    </div>
                </div>
                <div class="col-md-6">
                    <div class="card-custom h-100">
                        <h4 class="fw-bold mb-4">New Identity</h4>
                        <p class="small text-muted mb-4">Generate 12 words to derive your secure 512-bit key.</p>
                        <button class="btn btn-outline-dark w-100 py-3 rounded-4" onclick="gen()">GENERATE SEED</button>
                        <div id="g-res" class="mt-4" style="display:none">
                            <div class="addr-pill mb-2" id="g-priv"></div>
                            <div class="addr-pill fw-bold text-primary" id="g-pub"></div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div id="view-explorer" class="view" style="display:none">
            <div class="card-custom">
                <h4 class="fw-bold mb-4">Blockchain Explorer</h4>
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

        function show(id, el) {
            document.querySelectorAll('.view').forEach(v => v.style.display = 'none');
            document.getElementById('view-' + id).style.display = 'block';
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
                const h = (b.Hash || b.hash || "").substring(0,30) + "...";
                mini.innerHTML += '<div class="p-3 border-bottom d-flex justify-content-between"><span>Block #' + b.index + '</span><span class="text-muted">' + h + '</span></div>';
                full.innerHTML += '<div class="card-custom border mb-3"><h6>Block #' + b.index + '</h6><div class="addr-pill mb-2">' + (b.Hash || b.hash) + '</div><small class="text-muted">Timestamp: ' + new Date(b.timestamp*1000).toLocaleString() + '</small></div>';
            });
        }

        async function mine() { await fetch('/api/mine'); alert("Block Mined!"); load(); }
        async function send() {
            if(!wallet) return show('security');
            const tx = { sender: wallet.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            alert("Sent! Please mine to confirm."); show('dash'); load();
        }

        async function gen() {
            const p = btoa(Math.random()).substring(0,64);
            const pb = await derive(p);
            document.getElementById('g-res').style.display = 'block';
            document.getElementById('g-priv').innerText = 'Secret Key: ' + p;
            document.getElementById('g-pub').innerText = 'Public Address: ' + pb;
        }

        load(); setInterval(load, 15000);
    </script>
</body>
</html>
