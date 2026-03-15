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

	http.HandleFunc("/api/mempool", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Mempool)
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
	fmt.Printf("🌐 AstraliX Pro Node running on port %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>AstraliX Elite | Network Dashboard</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --accent: #00ffa3; --bg: #050508; --card: rgba(255, 255, 255, 0.03); }
        body { background: var(--bg); color: #e0e0e0; font-family: 'Inter', sans-serif; padding-bottom: 80px; }
        
        /* Desktop Sidebar */
        .sidebar { background: #000; height: 100vh; position: fixed; width: 260px; border-right: 1px solid #1a1a1a; padding-top: 20px; }
        .main-content { margin-left: 260px; padding: 40px; }
        
        /* Mobile Navigation */
        .mobile-nav { display: none; background: #000; border-top: 1px solid #222; position: fixed; bottom: 0; width: 100%; z-index: 1000; height: 70px; }
        .nav-link { color: #888; padding: 12px 25px; border-radius: 12px; margin: 8px 15px; transition: 0.3s; cursor: pointer; text-decoration: none; display: flex; align-items: center; }
        .nav-link:hover, .nav-link.active { background: rgba(0, 255, 163, 0.1); color: var(--accent); }
        
        /* Glass Style */
        .glass-card { background: var(--card); backdrop-filter: blur(12px); border: 1px solid rgba(255,255,255,0.05); border-radius: 20px; padding: 25px; margin-bottom: 25px; }
        .accent { color: var(--accent); }
        .btn-elite { background: var(--accent); color: #000; font-weight: 800; border-radius: 12px; border: none; padding: 12px 24px; }
        .addr-box { font-family: monospace; font-size: 0.75rem; background: #000; padding: 12px; border-radius: 10px; border: 1px solid #222; word-break: break-all; }

        @media (max-width: 992px) {
            .sidebar { display: none; }
            .main-content { margin-left: 0; padding: 20px; }
            .mobile-nav { display: flex; justify-content: space-around; align-items: center; }
            .nav-link { margin: 0; padding: 10px; flex-direction: column; font-size: 10px; }
            .nav-link i { font-size: 20px; margin-bottom: 4px; }
        }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="px-4 mb-5 text-center"><h3 class="accent fw-bold">ASTRALIX</h3></div>
        <nav>
            <a class="nav-link active" onclick="switchTab('dash', this)"><i class="fas fa-th-large me-3"></i> Dashboard</a>
            <a class="nav-link" onclick="switchTab('wallet', this)"><i class="fas fa-wallet me-3"></i> Wallet</a>
            <a class="nav-link" onclick="switchTab('explorer', this)"><i class="fas fa-cube me-3"></i> Explorer</a>
        </nav>
    </div>

    <div class="mobile-nav">
        <a class="nav-link active" onclick="switchTab('dash', this)"><i class="fas fa-th-large"></i> Dash</a>
        <a class="nav-link" onclick="switchTab('wallet', this)"><i class="fas fa-wallet"></i> Wallet</a>
        <a class="nav-link" onclick="switchTab('explorer', this)"><i class="fas fa-cube"></i> Explorer</a>
    </div>

    <div class="main-content">
        <div id="view-dash" class="view-section">
            <div class="row g-3 mb-4">
                <div class="col-6 col-md-4"><div class="glass-card text-center"><h6>Supply</h6><h4 class="accent">1.0B AX</h4></div></div>
                <div class="col-6 col-md-4"><div class="glass-card text-center"><h6>Blocks</h6><h4 id="stat-blocks" class="text-info">1</h4></div></div>
            </div>
            <div class="glass-card">
                <div class="d-flex justify-content-between align-items-center mb-3">
                    <h4>Recent Activity</h4>
                    <button class="btn btn-elite btn-sm" onclick="mineBlock()">Mine</button>
                </div>
                <div id="feed"></div>
            </div>
        </div>

        <div id="view-wallet" class="view-section" style="display:none">
            <div class="glass-card">
                <h4>My Assets</h4>
                <input type="text" id="w-addr" class="form-control bg-dark border-0 text-white my-3" placeholder="AX Address...">
                <button class="btn btn-elite w-100 mb-3" onclick="getBal()">Sync Balance</button>
                <div class="text-center"><h1 id="bal-val" class="accent">0.00 AX</h1></div>
            </div>
            <div class="glass-card">
                <h4>Transfer</h4>
                <input type="text" id="tx-to" class="form-control bg-dark border-0 text-white mb-2" placeholder="Recipient AX...">
                <input type="number" id="tx-amt" class="form-control bg-dark border-0 text-white mb-3" placeholder="Amount">
                <button class="btn btn-outline-light w-100" onclick="sendTx()">Send AX</button>
            </div>
        </div>

        <div id="view-explorer" class="view-section" style="display:none">
            <div class="glass-card">
                <h4>Block Explorer</h4>
                <div id="full-chain"></div>
            </div>
        </div>
    </div>

    <script>
        function switchTab(id, el) {
            document.querySelectorAll('.view-section').forEach(v => v.style.display = 'none');
            document.getElementById('view-' + id).style.display = 'block';
            document.querySelectorAll('.nav-link').forEach(l => l.classList.remove('active'));
            el.classList.add('active');
        }

        async function getBal() {
            const addr = document.getElementById('w-addr').value;
            const res = await fetch('/api/balance/' + addr);
            const data = await res.json();
            document.getElementById('bal-val').innerText = data.balance.toLocaleString() + ' AX';
        }

        async function load() {
            const res = await fetch('/api/chain');
            const chain = await res.json();
            document.getElementById('stat-blocks').innerText = chain.length;
            const feed = document.getElementById('feed');
            const full = document.getElementById('full-chain');
            feed.innerHTML = ''; full.innerHTML = '';
            chain.reverse().forEach(b => {
                const html = '<div class="border-bottom border-secondary py-2"><small class="accent">BLOCK #' + b.index + '</small><br><small class="text-muted">' + b.hash.substring(0,25) + '...</small></div>';
                feed.innerHTML += html;
                full.innerHTML += html;
            });
        }

        async function mineBlock() {
            const res = await fetch('/api/mine');
            if(res.ok) { alert("Block Mined!"); load(); }
            else { alert("Mempool empty"); }
        }

        async function sendTx() {
            const tx = { sender: "USER", recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            alert("Broadcasted! Mine to confirm.");
        }

        load();
        setInterval(load, 15000);
    </script>
</body>
</html>
`
