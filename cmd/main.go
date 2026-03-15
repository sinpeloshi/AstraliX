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
		if len(Mempool) == 0 { http.Error(w, "Mempool empty", 400); return }
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, dashboardHTML)
	})

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	fmt.Printf("🌐 AstraliX Elite Node on port %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>AstraliX | Network Dashboard</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --primary: #00ffa3; --bg: #050507; --card: rgba(255, 255, 255, 0.03); }
        body { background: var(--bg); color: #fff; font-family: 'Inter', sans-serif; }
        .sidebar { background: #000; height: 100vh; position: fixed; width: 260px; border-right: 1px solid #1a1a1a; }
        .main { margin-left: 260px; padding: 40px; }
        .glass-card { background: var(--card); backdrop-filter: blur(10px); border: 1px solid rgba(255,255,255,0.05); border-radius: 20px; padding: 25px; transition: 0.3s; }
        .glass-card:hover { border-color: var(--primary); }
        .nav-link { color: #666; padding: 15px 25px; border-radius: 12px; margin: 5px 15px; transition: 0.3s; display: block; text-decoration: none; }
        .nav-link:hover, .nav-link.active { background: rgba(0, 255, 163, 0.1); color: var(--primary); }
        .accent { color: var(--primary); font-weight: bold; }
        .btn-main { background: var(--primary); color: #000; font-weight: 700; border-radius: 12px; border: none; padding: 12px 25px; }
        .btn-main:hover { background: #00d689; box-shadow: 0 0 20px rgba(0,255,163,0.3); }
        .addr-pill { background: #000; padding: 8px 15px; border-radius: 10px; font-family: monospace; font-size: 0.8rem; border: 1px solid #222; overflow: hidden; }
        @media (max-width: 992px) { .sidebar { display: none; } .main { margin-left: 0; padding: 20px; } }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="p-4 mb-4 text-center">
            <h2 class="accent mb-0">AstraliX</h2>
            <small class="text-muted">Network L1 v1.0</small>
        </div>
        <nav>
            <a href="#" class="nav-link active" onclick="show('dash')"><i class="fas fa-grid-2 me-2"></i> Dashboard</a>
            <a href="#" class="nav-link" onclick="show('wallet')"><i class="fas fa-wallet me-2"></i> Wallet</a>
            <a href="#" class="nav-link" onclick="show('explorer')"><i class="fas fa-database me-2"></i> Explorer</a>
        </nav>
    </div>

    <div class="main">
        <div id="dash" class="view">
            <div class="row g-4 mb-4">
                <div class="col-md-4"><div class="glass-card"><h6>Circulating Supply</h6><h3 class="accent">1,000,002,021 AX</h3></div></div>
                <div class="col-md-4"><div class="glass-card"><h6>Network Status</h6><h3 class="text-info">Live</h3></div></div>
                <div class="col-md-4"><div class="glass-card"><h6>Block Time</h6><h3>~120ms</h3></div></div>
            </div>
            <div class="glass-card">
                <div class="d-flex justify-content-between align-items-center mb-4">
                    <h4 class="mb-0">Recent Blocks</h4>
                    <button class="btn btn-main btn-sm" onclick="mine()">Mine Now</button>
                </div>
                <div id="blocks-feed"></div>
            </div>
        </div>

        <div id="wallet" class="view" style="display:none">
            <div class="row g-4">
                <div class="col-lg-6">
                    <div class="glass-card h-100">
                        <h4>Account Overview</h4>
                        <input type="text" id="w-input" class="form-control bg-dark text-white border-0 my-3 p-3" placeholder="Enter AX address...">
                        <button class="btn btn-main w-100 mb-4" onclick="getBal()">Sync Balance</button>
                        <div class="text-center"><small class="text-muted">Total Balance</small><h1 id="bal-text" class="accent">0.00 AX</h1></div>
                    </div>
                </div>
                <div class="col-lg-6">
                    <div class="glass-card">
                        <h4>Create New Keypair</h4>
                        <p class="small text-muted">Generate a new secure address for the AstraliX Network.</p>
                        <button class="btn btn-outline-light w-100" onclick="gen()">Generate Keys</button>
                        <div id="keys-box" class="mt-4" style="display:none">
                            <div class="mb-2"><small class="text-danger">Private Key (Secret):</small><div id="p-res" class="addr-pill"></div></div>
                            <div><small class="accent">Public Address:</small><div id="a-res" class="addr-pill"></div></div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        function show(id) {
            document.querySelectorAll('.view').forEach(v => v.style.display = 'none');
            document.getElementById(id).style.display = 'block';
            document.querySelectorAll('.nav-link').forEach(l => l.classList.remove('active'));
            event.target.classList.add('active');
        }

        async function getBal() {
            const a = document.getElementById('w-input').value;
            const r = await fetch('/api/balance/' + a);
            const d = await r.json();
            document.getElementById('bal-text').innerText = d.balance.toLocaleString() + ' AX';
        }

        async function load() {
            const r = await fetch('/api/chain');
            const c = await r.json();
            const feed = document.getElementById('blocks-feed');
            feed.innerHTML = '';
            c.reverse().forEach(b => {
                feed.innerHTML += ` + "`" + `
                    <div class="d-flex justify-content-between border-bottom border-secondary py-3">
                        <div>
                            <span class="accent fw-bold">BLOCK #${b.index}</span><br>
                            <small class="text-muted">${b.hash.substring(0,32)}...</small>
                        </div>
                        <div class="text-end">
                            <small class="text-muted d-block">${new Date(b.timestamp*1000).toLocaleTimeString()}</small>
                            <span class="badge bg-dark border border-secondary">${b.transactions ? b.transactions.length : 0} TXs</span>
                        </div>
                    </div>
                ` + "`" + `;
            });
        }

        async function mine() {
            const r = await fetch('/api/mine');
            if(r.ok) { alert("Block successfully mined!"); load(); }
            else { alert("Mempool is currently empty."); }
        }

        function gen() {
            const h = l => [...Array(l)].map(()=>Math.floor(Math.random()*16).toString(16)).join('');
            document.getElementById('keys-box').style.display = 'block';
            document.getElementById('p-res').innerText = h(128);
            document.getElementById('a-res').innerText = 'AX' + h(64);
        }

        load();
        setInterval(load, 15000);
    </script>
</body>
</html>
`
