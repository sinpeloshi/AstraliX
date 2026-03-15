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
	fmt.Printf("🌐 AstraliX Pro Node running on port %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>AstraliX Elite | Web3 Dashboard</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --accent: #00ffa3; --bg: #050508; --card: rgba(255, 255, 255, 0.03); }
        body { background: var(--bg); color: #e0e0e0; font-family: 'Inter', sans-serif; padding-bottom: 90px; }
        
        .main-content { padding: 30px; max-width: 800px; margin: auto; }
        .glass-card { background: var(--card); backdrop-filter: blur(15px); border: 1px solid rgba(255,255,255,0.05); border-radius: 24px; padding: 25px; margin-bottom: 20px; }
        
        .mobile-nav { background: rgba(0,0,0,0.8); backdrop-filter: blur(10px); border-top: 1px solid #222; position: fixed; bottom: 0; width: 100%; height: 75px; display: flex; justify-content: space-around; align-items: center; z-index: 1000; }
        .nav-item { color: #666; text-decoration: none; text-align: center; font-size: 11px; cursor: pointer; }
        .nav-item.active { color: var(--accent); }
        .nav-item i { font-size: 22px; display: block; margin-bottom: 4px; }

        .btn-elite { background: var(--accent); color: #000; font-weight: 800; border-radius: 14px; border: none; padding: 14px; transition: 0.3s; }
        .btn-elite:hover { box-shadow: 0 0 15px rgba(0,255,163,0.5); }
        
        .accent { color: var(--accent); }
        .addr-display { font-family: monospace; font-size: 0.8rem; background: #000; padding: 15px; border-radius: 15px; border: 1px solid #222; word-break: break-all; margin: 10px 0; }
        input { background: #111 !important; border: 1px solid #333 !important; color: white !important; border-radius: 12px !important; padding: 12px !important; }
    </style>
</head>
<body>

    <div class="main-content">
        <div class="d-flex justify-content-between align-items-center mb-4">
            <h2 class="fw-bold accent">AstraliX</h2>
            <span id="status-tag" class="badge bg-success">Online</span>
        </div>

        <div id="view-dash" class="view-section">
            <div class="glass-card text-center">
                <small class="text-muted">Total Balance</small>
                <h1 id="bal-large" class="accent fw-bold my-2">0.00 AX</h1>
                <div id="connected-addr" class="small text-muted text-truncate px-4">No wallet connected</div>
            </div>

            <div class="row g-3">
                <div class="col-6">
                    <button class="btn btn-elite w-100" onclick="switchTab('wallet')"><i class="fas fa-paper-plane me-2"></i> Send</button>
                </div>
                <div class="col-6">
                    <button class="btn btn-dark w-100 border-secondary" style="border-radius:14px; padding:14px" onclick="showReceive()"><i class="fas fa-qrcode me-2"></i> Receive</button>
                </div>
            </div>

            <div class="glass-card mt-4">
                <div class="d-flex justify-content-between align-items-center mb-3">
                    <h5 class="mb-0">Latest Blocks</h5>
                    <button class="btn btn-sm btn-outline-info" onclick="mine()">Mine Now</button>
                </div>
                <div id="feed"></div>
            </div>
        </div>

        <div id="view-wallet" class="view-section" style="display:none">
            <div class="glass-card">
                <h4>Wallet Manager</h4>
                <p class="small text-muted">Connect your AX address to sync your assets.</p>
                <input type="text" id="w-input" class="form-control mb-2" placeholder="Enter AX Address...">
                <button class="btn btn-elite w-100 mb-3" onclick="saveWallet()">Connect & Sync</button>
                
                <hr class="border-secondary">
                
                <h5>New Transaction</h5>
                <input type="text" id="tx-to" class="form-control mb-2" placeholder="Recipient Address">
                <input type="number" id="tx-amt" class="form-control mb-3" placeholder="Amount (AX)">
                <button class="btn btn-outline-light w-100" onclick="sendTx()">Send Transaction</button>
            </div>
        </div>

        <div id="view-explorer" class="view-section" style="display:none">
            <div class="glass-card">
                <h4>Chain Explorer</h4>
                <div id="full-chain"></div>
            </div>
        </div>
    </div>

    <div class="mobile-nav">
        <div class="nav-item active" onclick="switchTab('dash', this)"><i class="fas fa-chart-pie"></i>Assets</div>
        <div class="nav-item" onclick="switchTab('wallet', this)"><i class="fas fa-wallet"></i>Wallet</div>
        <div class="nav-item" onclick="switchTab('explorer', this)"><i class="fas fa-search"></i>Explorer</div>
    </div>

    <script>
        let myAddr = localStorage.getItem('ax_addr') || '';

        function switchTab(id, el) {
            document.querySelectorAll('.view-section').forEach(v => v.style.display = 'none');
            document.getElementById('view-' + id).style.display = 'block';
            document.querySelectorAll('.nav-item').forEach(l => l.classList.remove('active'));
            if(el) el.classList.add('active');
        }

        async function saveWallet() {
            const val = document.getElementById('w-input').value;
            if(val) {
                localStorage.setItem('ax_addr', val);
                myAddr = val;
                alert('Wallet Connected!');
                switchTab('dash');
                loadData();
            }
        }

        function showReceive() {
            if(!myAddr) { alert('Connect a wallet first'); switchTab('wallet'); return; }
            alert('Your Address:\n' + myAddr);
        }

        async function loadData() {
            if(myAddr) {
                const res = await fetch('/api/balance/' + myAddr);
                const data = await res.json();
                document.getElementById('bal-large').innerText = data.balance.toLocaleString() + ' AX';
                document.getElementById('connected-addr').innerText = myAddr;
            }

            const resC = await fetch('/api/chain');
            const chain = await resC.json();
            const feed = document.getElementById('feed');
            feed.innerHTML = '';
            chain.slice(-5).reverse().forEach(b => {
                feed.innerHTML += '<div class="border-bottom border-secondary py-2 small"><span class="accent">#' + b.index + '</span> ' + b.hash.substring(0,30) + '...</div>';
            });
        }

        async function mine() {
            const r = await fetch('/api/mine');
            if(r.ok) { alert('Mined!'); loadData(); } else { alert('Mempool empty'); }
        }

        async function sendTx() {
            const tx = { sender: myAddr || "EXTERNAL", recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            alert("Sent! Waiting for mining.");
        }

        loadData();
        setInterval(loadData, 15000);
    </script>
</body>
</html>
`
