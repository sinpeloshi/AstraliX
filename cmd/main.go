package main // Corregido: minúscula obligatoria

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

	// 1. Genesis Setup
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

	http.HandleFunc("/api/transactions/new", func(w http.ResponseWriter, r *http.Request) {
		var tx core.Transaction
		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			http.Error(w, "Invalid TX", 400); return
		}
		tx.TxID = tx.CalculateHash()
		Mempool = append(Mempool, tx)
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]string{"status": "pending"})
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
    <title>AstraliX | Elite Network OS</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --primary: #00ffa3; --bg: #050508; --sidebar: #0a0a0c; --glass: rgba(255, 255, 255, 0.03); }
        body { background: var(--bg); color: #e0e0e0; font-family: 'Inter', sans-serif; margin: 0; }
        .sidebar { background: var(--sidebar); height: 100vh; position: fixed; width: 260px; border-right: 1px solid #1a1a1a; padding-top: 20px; }
        .main-content { margin-left: 260px; padding: 40px; }
        .nav-link { color: #888; padding: 12px 25px; border-radius: 12px; margin: 8px 15px; transition: 0.3s; cursor: pointer; display: flex; align-items: center; text-decoration: none; }
        .nav-link:hover, .nav-link.active { background: rgba(0, 255, 163, 0.1); color: var(--primary); }
        .glass-card { background: var(--glass); backdrop-filter: blur(12px); border: 1px solid rgba(255,255,255,0.05); border-radius: 20px; padding: 25px; margin-bottom: 25px; }
        .accent { color: var(--primary); }
        .btn-elite { background: var(--primary); color: #000; font-weight: 800; border-radius: 12px; border: none; padding: 12px 24px; transition: 0.3s; }
        .btn-elite:hover { box-shadow: 0 0 20px rgba(0,255,163,0.4); transform: translateY(-2px); }
        .addr-text { font-family: 'Monaco', monospace; font-size: 0.75rem; background: #000; padding: 12px; border-radius: 10px; border: 1px solid #222; word-break: break-all; }
        @media (max-width: 992px) { .sidebar { display: none; } .main-content { margin-left: 0; padding: 20px; } }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="px-4 mb-5 text-center">
            <h3 class="accent fw-bold mb-0">ASTRALIX</h3>
            <small class="text-muted text-uppercase tracking-widest" style="font-size: 10px;">Layer 1 Protocol</small>
        </div>
        <nav>
            <a class="nav-link active" onclick="switchView('dash')"><i class="fas fa-th-large me-3"></i> Dashboard</a>
            <a class="nav-link" onclick="switchView('wallet')"><i class="fas fa-wallet me-3"></i> My Wallet</a>
            <a class="nav-link" onclick="switchView('explorer')"><i class="fas fa-cube me-3"></i> Explorer</a>
        </nav>
    </div>

    <div class="main-content">
        <div id="view-dash" class="view-section">
            <div class="row g-4 mb-4">
                <div class="col-md-4"><div class="glass-card"><h6>Circulating Supply</h6><h3 class="accent">1,000,021,000 AX</h3></div></div>
                <div class="col-md-4"><div class="glass-card"><h6>Difficulty</h6><h3 class="text-warning">4 (PoW)</h3></div></div>
                <div class="col-md-4"><div class="glass-card"><h6>Network State</h6><h3 class="text-info">Stable</h3></div></div>
            </div>
            <div class="glass-card">
                <div class="d-flex justify-content-between align-items-center mb-4">
                    <h4>Recent Activity</h4>
                    <button class="btn btn-elite btn-sm" onclick="mineBlock()">Mine New Block</button>
                </div>
                <div id="recent-feed"></div>
            </div>
        </div>

        <div id="view-wallet" class="view-section" style="display:none">
            <div class="row g-4">
                <div class="col-lg-6">
                    <div class="glass-card h-100">
                        <h4>Wallet Access</h4>
                        <input type="text" id="wallet-addr" class="form-control bg-dark border-0 text-white p-3 my-3" placeholder="Paste Public Address...">
                        <button class="btn btn-elite w-100 mb-4" onclick="updateBalance()">Refresh Data</button>
                        <div class="text-center p-3">
                            <small class="text-muted d-block">Available Balance</small>
                            <h1 id="balance-val" class="accent">0.00 AX</h1>
                        </div>
                    </div>
                </div>
                <div class="col-lg-6">
                    <div class="glass-card">
                        <h4>Transfer Assets</h4>
                        <input type="text" id="send-to" class="form-control bg-dark border-0 text-white mb-2" placeholder="Recipient Address">
                        <input type="number" id="send-amt" class="form-control bg-dark border-0 text-white mb-3" placeholder="Amount (AX)">
                        <button class="btn btn-elite w-100" onclick="sendTransaction()">Broadcast to Network</button>
                    </div>
                </div>
            </div>
            <div class="mt-4 text-center">
                <button class="btn btn-outline-secondary btn-sm" onclick="generateKeys()">Generate New Secure Keypair</button>
                <div id="keys-output" class="mt-3" style="display:none">
                    <div class="card bg-black p-3 text-start">
                        <small class="text-danger">PRIVATE KEY (STORE OFFLINE):</small>
                        <div id="priv-key" class="addr-text mb-2"></div>
                        <small class="accent">PUBLIC ADDRESS:</small>
                        <div id="pub-key" class="addr-text"></div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        function switchView(id) {
            document.querySelectorAll('.view-section').forEach(v => v.style.display = 'none');
            document.getElementById('view-' + id).style.display = 'block';
            document.querySelectorAll('.nav-link').forEach(l => l.classList.remove('active'));
            event.currentTarget.classList.add('active');
        }

        async function updateBalance() {
            const addr = document.getElementById('wallet-addr').value;
            if(!addr) return;
            const res = await fetch('/api/balance/' + addr);
            const data = await res.json();
            document.getElementById('balance-val').innerText = data.balance.toLocaleString() + ' AX';
        }

        async function loadBlockchain() {
            const res = await fetch('/api/chain');
            const chain = await res.json();
            const feed = document.getElementById('recent-feed');
            feed.innerHTML = '';
            chain.reverse().forEach(block => {
                feed.innerHTML += ` + "`" + `
                    <div class="d-flex justify-content-between border-bottom border-secondary py-3">
                        <div>
                            <span class="accent fw-bold">BLOCK #${block.index}</span><br>
                            <small class="text-muted">${block.hash.substring(0,40)}...</small>
                        </div>
                        <div class="text-end">
                            <small class="text-muted d-block">${new Date(block.timestamp*1000).toLocaleTimeString()}</small>
                            <span class="badge bg-dark border border-secondary">${block.transactions ? block.transactions.length : 0} TXs</span>
                        </div>
                    </div>
                ` + "`" + `;
            });
        }

        async function mineBlock() {
            const res = await fetch('/api/mine');
            if(res.ok) { alert("Mining successful!"); loadBlockchain(); }
            else { alert("Mempool is empty. Send a transaction first."); }
        }

        async function sendTransaction() {
            const tx = {
                sender: "MASTER_CLIENT",
                recipient: document.getElementById('send-to').value,
                amount: parseFloat(document.getElementById('send-amt').value)
            };
            const res = await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            if(res.ok) alert("Transaction broadcasted! Now click 'Mine' to confirm it.");
        }

        function generateKeys() {
            const h = l => [...Array(l)].map(()=>Math.floor(Math.random()*16).toString(16)).join('');
            document.getElementById('keys-output').style.display = 'block';
            document.getElementById('priv-key').innerText = h(128);
            document.getElementById('pub-key').innerText = 'AX' + h(64);
        }

        loadBlockchain();
        setInterval(loadBlockchain, 15000);
    </script>
</body>
</html>
`
