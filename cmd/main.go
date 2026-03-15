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

	http.HandleFunc("/api/mempool", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Mempool)
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

	http.HandleFunc("/api/transactions/new", func(w http.ResponseWriter, r *http.Request) {
		var tx core.Transaction
		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			http.Error(w, "Invalid", 400); return
		}
		tx.TxID = tx.CalculateHash()
		Mempool = append(Mempool, tx)
		w.WriteHeader(201)
	})

	// SERVIMOS EL DASHBOARD PRO
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
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>AstraliX Core | Network OS</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --bg-deep: #08090b; --card-bg: #111318; --accent: #00ffa3; --accent-dim: rgba(0, 255, 163, 0.15); }
        body { background: var(--bg-deep); color: #e0e0e0; font-family: 'Inter', sans-serif; overflow-x: hidden; }
        .sidebar { background: #000; height: 100vh; position: fixed; width: 240px; border-right: 1px solid #222; z-index: 1000; }
        .main-content { margin-left: 240px; padding: 30px; }
        .nav-link { color: #888; padding: 12px 20px; border-radius: 8px; margin: 4px 15px; transition: 0.3s; }
        .nav-link:hover, .nav-link.active { background: var(--accent-dim); color: var(--accent); }
        .card { background: var(--card-bg); border: 1px solid #222; border-radius: 16px; box-shadow: 0 8px 32px rgba(0,0,0,0.4); }
        .stat-card { padding: 20px; text-align: center; }
        .accent-text { color: var(--accent); }
        .btn-accent { background: var(--accent); color: #000; font-weight: 700; border-radius: 10px; border: none; padding: 10px 20px; }
        .btn-accent:hover { background: #00d689; transform: translateY(-2px); }
        .addr-box { font-family: 'Monaco', monospace; font-size: 0.75rem; background: #000; padding: 10px; border-radius: 8px; border: 1px dashed #444; word-break: break-all; }
        @media (max-width: 768px) { .sidebar { width: 60px; } .nav-text { display: none; } .main-content { margin-left: 60px; } }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="p-4 text-center">
            <h4 class="accent-text fw-bold">AX <span class="text-white nav-text">AstraliX</span></h4>
        </div>
        <nav class="nav flex-column mt-3">
            <a href="#" class="nav-link active" onclick="showSection('dash')"><i class="fas fa-chart-line me-2"></i> <span class="nav-text">Dashboard</span></a>
            <a href="#" class="nav-link" onclick="showSection('wallet')"><i class="fas fa-wallet me-2"></i> <span class="nav-text">My Wallet</span></a>
            <a href="#" class="nav-link" onclick="showSection('explorer')"><i class="fas fa-cubes me-2"></i> <span class="nav-text">Explorer</span></a>
            <a href="#" class="nav-link" onclick="showSection('mine')"><i class="fas fa-microchip me-2"></i> <span class="nav-text">Mining Central</span></a>
        </nav>
    </div>

    <div class="main-content">
        <div id="section-dash" class="section">
            <div class="row g-4 mb-4">
                <div class="col-md-4"><div class="card stat-card"><h6>Circulating Supply</h6><h3 class="accent-text">1.0B AX</h3></div></div>
                <div class="col-md-4"><div class="card stat-card"><h6>Current Blocks</h6><h3 id="stat-blocks">0</h3></div></div>
                <div class="col-md-4"><div class="card stat-card"><h6>Difficulty</h6><h3 class="text-warning">4 (PoW)</h3></div></div>
            </div>
            <div class="card p-4">
                <h5>Recent Network Activity</h5>
                <div id="recent-txs" class="mt-3 small">Loading activity...</div>
            </div>
        </div>

        <div id="section-wallet" class="section" style="display:none">
            <div class="row g-4">
                <div class="col-md-5">
                    <div class="card p-4 h-100">
                        <h3>Wallet Control</h3>
                        <div class="mt-4">
                            <label class="text-muted small">Your Public Address</label>
                            <input type="text" id="w-addr" class="form-control bg-dark text-white border-secondary mb-3" placeholder="Paste AX address...">
                            <button class="btn btn-accent w-100" onclick="checkB()">Check Balance</button>
                        </div>
                        <div class="text-center mt-4">
                            <h1 id="bal-large" class="accent-text">0.00 AX</h1>
                        </div>
                    </div>
                </div>
                <div class="col-md-7">
                    <div class="card p-4">
                        <h3>Send Tokens</h3>
                        <div class="mb-3">
                            <label class="small">Recipient Address</label>
                            <input type="text" id="tx-to" class="form-control bg-black text-white border-secondary">
                        </div>
                        <div class="mb-3">
                            <label class="small">Amount to Send</label>
                            <input type="number" id="tx-amount" class="form-control bg-black text-white border-secondary">
                        </div>
                        <button class="btn btn-accent w-100" onclick="sendTx()">Broadcast Transaction</button>
                    </div>
                </div>
            </div>
            <div class="mt-4 text-center">
                <button class="btn btn-outline-secondary btn-sm" onclick="genW()">Generate New Keypair</button>
                <div id="new-w-box" class="mt-3 d-none"><div class="card p-3 bg-black"><div id="p-key" class="addr-box mb-2 text-danger"></div><div id="a-key" class="addr-box accent-text"></div></div></div>
            </div>
        </div>

        <div id="section-explorer" class="section" style="display:none">
            <div class="card p-4">
                <h3>Blockchain Explorer</h3>
                <div id="full-chain" class="mt-4"></div>
            </div>
        </div>

        <div id="section-mine" class="section" style="display:none">
            <div class="card p-4 text-center">
                <i class="fas fa-server fa-3x mb-3 accent-text"></i>
                <h3>Mining Console</h3>
                <p class="text-muted">Validate pending transactions from the mempool into the next block.</p>
                <div id="mempool-status" class="my-4 addr-box">Scanning for transactions...</div>
                <button class="btn btn-accent btn-lg px-5" onclick="mine()">START MINING PROCESS</button>
            </div>
        </div>
    </div>

    <script>
        function showSection(id) {
            document.querySelectorAll('.section').forEach(s => s.style.display = 'none');
            document.getElementById('section-' + id).style.display = 'block';
            document.querySelectorAll('.nav-link').forEach(l => l.classList.remove('active'));
            event.currentTarget.classList.add('active');
        }

        async function checkB() {
            const a = document.getElementById('w-addr').value;
            const r = await fetch('/api/balance/' + a);
            const d = await r.json();
            document.getElementById('bal-large').innerText = d.balance.toLocaleString() + ' AX';
        }

        async function loadAll() {
            const r = await fetch('/api/chain');
            const c = await r.json();
            document.getElementById('stat-blocks').innerText = c.length;
            
            // Explorer Update
            let exH = '';
            c.reverse().forEach(b => {
                exH += '<div class="card mb-3 p-3 bg-black"><div class="d-flex justify-content-between"><b>Block #' + b.index + '</b><span class="small text-muted">' + new Date(b.timestamp*1000).toLocaleTimeString() + '</span></div><div class="addr-box my-2">' + b.hash + '</div><small>Transactions: ' + b.transactions.length + '</small></div>';
            });
            document.getElementById('full-chain').innerHTML = exH;

            // Recent TXs for Dashboard
            const latest = c[0].transactions;
            let txH = latest.length > 0 ? latest.map(tx => '<div class="d-flex justify-content-between border-bottom border-secondary py-2"><span>' + tx.recipient.substring(0,20) + '...</span><span class="accent-text">+' + tx.amount + ' AX</span></div>').join('') : 'No recent transactions';
            document.getElementById('recent-txs').innerHTML = txH;
        }

        async function mine() {
            const r = await fetch('/api/mine');
            if(r.ok) { alert('SUCCESS: Block minado y añadido a la cadena.'); loadAll(); }
            else { alert('ERROR: El mempool está vacío.'); }
        }

        async function sendTx() {
            const tx = { sender: "USER", recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amount').value) };
            const r = await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            if(r.ok) alert('TX Enviada al Mempool. ¡Recuerda minar el bloque para confirmarla!');
        }

        function genW() {
            const h = l => [...Array(l)].map(()=>Math.floor(Math.random()*16).toString(16)).join('');
            document.getElementById('new-w-box').classList.remove('d-none');
            document.getElementById('p-key').innerText = 'SECRET PRIVKEY: ' + h(128);
            document.getElementById('a-key').innerText = 'PUBLIC ADDRESS: AX' + h(64);
        }

        loadAll();
        setInterval(loadAll, 10000);
    </script>
</body>
</html>
`
