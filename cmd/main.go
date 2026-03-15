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
    <title>AstraliX Elite | Digital Assets</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --accent: #00ffa3; --bg: #050508; --card: rgba(255, 255, 255, 0.03); }
        body { background: var(--bg); color: #e0e0e0; font-family: 'Inter', sans-serif; padding-bottom: 100px; }
        .main-content { padding: 25px; max-width: 600px; margin: auto; }
        .glass-card { background: var(--card); backdrop-filter: blur(15px); border: 1px solid rgba(255,255,255,0.05); border-radius: 28px; padding: 25px; margin-bottom: 20px; }
        .mobile-nav { background: rgba(0,0,0,0.9); backdrop-filter: blur(15px); border-top: 1px solid #222; position: fixed; bottom: 0; width: 100%; height: 80px; display: flex; justify-content: space-around; align-items: center; z-index: 1000; }
        .nav-item { color: #555; text-decoration: none; text-align: center; font-size: 11px; cursor: pointer; transition: 0.3s; }
        .nav-item.active { color: var(--accent); }
        .nav-item i { font-size: 26px; display: block; margin-bottom: 5px; }
        .btn-elite { background: var(--accent); color: #000; font-weight: 800; border-radius: 18px; border: none; padding: 16px; width: 100%; transition: 0.3s; }
        .btn-elite:hover { box-shadow: 0 0 25px rgba(0,255,163,0.5); transform: translateY(-2px); }
        .accent { color: var(--accent); }
        input { background: #000 !important; border: 1px solid #333 !important; color: #fff !important; border-radius: 16px !important; padding: 16px !important; margin-bottom: 15px; }
        .key-box { font-family: 'Monaco', monospace; font-size: 0.7rem; background: #000; padding: 15px; border-radius: 15px; border: 1px solid #333; word-break: break-all; margin-top: 5px; position: relative; }
        .btn-copy { position: absolute; top: 5px; right: 5px; font-size: 10px; color: var(--accent); cursor: pointer; }
    </style>
</head>
<body>

    <div class="main-content">
        <div class="d-flex justify-content-between align-items-center mb-4">
            <div>
                <h1 class="fw-bold accent mb-0" style="letter-spacing: -1px;">AstraliX</h1>
                <small id="wallet-status" class="text-muted"><i class="fas fa-circle-nodes me-1"></i> Decentralized Node</small>
            </div>
            <div class="dropdown">
                <button class="btn btn-dark border-secondary rounded-circle" style="width:45px; height:45px;" onclick="location.reload()"><i class="fas fa-sync-alt"></i></button>
            </div>
        </div>

        <div id="view-assets" class="view-section">
            <div class="glass-card text-center mb-4" style="background: linear-gradient(145deg, rgba(0,255,163,0.05) 0%, rgba(255,255,255,0.02) 100%);">
                <small class="text-muted text-uppercase fw-bold" style="font-size: 10px; letter-spacing: 2px;">Total Portfolio</small>
                <h1 id="bal-large" class="accent fw-bold my-2" style="font-size: 3rem;">0.00 AX</h1>
                <div id="pub-display" class="small text-muted text-truncate px-3">Sync your wallet to start</div>
            </div>

            <div class="row g-3">
                <div class="col-6"><button class="btn btn-elite" onclick="switchTab('send')"><i class="fas fa-paper-plane me-2"></i>Send</button></div>
                <div class="col-6"><button class="btn btn-dark w-100 border-secondary py-3 rounded-4" onclick="showReceive()"><i class="fas fa-qrcode me-2"></i>Receive</button></div>
            </div>

            <div class="mt-5">
                <div class="d-flex justify-content-between align-items-center mb-3">
                    <h5 class="fw-bold mb-0">Network Explorer</h5>
                    <button class="btn btn-sm btn-outline-info rounded-pill px-3" onclick="mine()">Mine Block</button>
                </div>
                <div id="recent-feed"></div>
            </div>
        </div>

        <div id="view-send" class="view-section" style="display:none">
            <div class="glass-card">
                <h4 class="fw-bold mb-4">Send Assets</h4>
                <label class="small text-muted mb-2">Recipient Public Address</label>
                <input type="text" id="tx-to" class="form-control" placeholder="AX...">
                <label class="small text-muted mb-2">Amount (AX)</label>
                <input type="number" id="tx-amt" class="form-control" placeholder="0.00">
                <button class="btn btn-elite mt-3" onclick="processSend()">Confirm Transaction</button>
                <button class="btn btn-link text-muted w-100 mt-2 text-decoration-none" onclick="switchTab('assets')">Cancel</button>
            </div>
        </div>

        <div id="view-wallet" class="view-section" style="display:none">
            <div class="glass-card mb-4" style="border: 1px dashed var(--accent);">
                <h4 class="fw-bold accent">Create New Wallet</h4>
                <p class="small text-muted">Generate a new identity on the AstraliX Network.</p>
                <button class="btn btn-outline-light w-100 py-3 rounded-4" onclick="createNew()">Generate New Keys</button>
                
                <div id="keys-output" class="mt-4" style="display:none">
                    <div class="mb-3">
                        <small class="text-danger fw-bold">PRIVATE KEY (SECRET):</small>
                        <div class="key-box"><span id="priv-key"></span><i class="fas fa-copy btn-copy" onclick="copyText('priv-key')"></i></div>
                    </div>
                    <div>
                        <small class="accent fw-bold">PUBLIC ADDRESS:</small>
                        <div class="key-box"><span id="pub-key"></span><i class="fas fa-copy btn-copy" onclick="copyText('pub-key')"></i></div>
                    </div>
                    <p class="small text-warning mt-2"><i class="fas fa-exclamation-triangle me-1"></i> Save these keys! They cannot be recovered.</p>
                </div>
            </div>

            <div class="glass-card">
                <h4 class="fw-bold">Import Existing</h4>
                <label class="small text-muted mt-3 mb-1">Public Address (AX...)</label>
                <input type="text" id="imp-pub" class="form-control" placeholder="Enter Address">
                <label class="small text-muted mb-1">Private Key (Secret)</label>
                <input type="password" id="imp-priv" class="form-control" placeholder="Enter Secret Key">
                <button class="btn btn-elite" onclick="importWallet()">Login to Wallet</button>
                <button class="btn btn-link text-danger w-100 mt-2 text-decoration-none" onclick="logout()">Disconnect Current</button>
            </div>
        </div>
    </div>

    <div class="mobile-nav">
        <div class="nav-item active" id="nav-assets" onclick="switchTab('assets', this)"><i class="fas fa-chart-bar"></i>Assets</div>
        <div class="nav-item" id="nav-wallet" onclick="switchTab('wallet', this)"><i class="fas fa-fingerprint"></i>Manage</div>
        <div class="nav-item" id="nav-explorer" onclick="switchTab('explorer', this)"><i class="fas fa-globe"></i>Explorer</div>
    </div>

    <script>
        let wallet = JSON.parse(localStorage.getItem('ax_session')) || null;

        function switchTab(id, el) {
            document.querySelectorAll('.view-section').forEach(v => v.style.display = 'none');
            document.getElementById('view-' + id).style.display = 'block';
            document.querySelectorAll('.nav-item').forEach(l => l.classList.remove('active'));
            if(el) el.classList.add('active');
            else document.getElementById('nav-' + id).classList.add('active');
        }

        function createNew() {
            const h = l => [...Array(l)].map(()=>Math.floor(Math.random()*16).toString(16)).join('');
            const priv = h(128);
            const pub = 'AX' + h(64);
            document.getElementById('keys-output').style.display = 'block';
            document.getElementById('priv-key').innerText = priv;
            document.getElementById('pub-key').innerText = pub;
            // Fill import fields automatically for convenience
            document.getElementById('imp-priv').value = priv;
            document.getElementById('imp-pub').value = pub;
        }

        function copyText(id) {
            const text = document.getElementById(id).innerText;
            navigator.clipboard.writeText(text);
            alert("Copied to clipboard!");
        }

        function importWallet() {
            const pub = document.getElementById('imp-pub').value;
            const priv = document.getElementById('imp-priv').value;
            if(!pub || !priv) { alert("Please enter both keys"); return; }
            wallet = { pub, priv };
            localStorage.setItem('ax_session', JSON.stringify(wallet));
            location.reload();
        }

        function logout() {
            localStorage.removeItem('ax_session');
            location.reload();
        }

        function showReceive() {
            if(!wallet) { alert("Please connect your wallet first"); switchTab('wallet'); return; }
            prompt("Share your Public Address:", wallet.pub);
        }

        async function processSend() {
            if(!wallet) { alert("Secret key required to sign!"); switchTab('wallet'); return; }
            const tx = {
                sender: wallet.pub,
                recipient: document.getElementById('tx-to').value,
                amount: parseFloat(document.getElementById('tx-amt').value),
                signature: "SIG_" + wallet.priv.substring(0,10) // Simulated signature
            };
            const r = await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            if(r.ok) { alert("Signed & Broadcasted!"); switchTab('assets'); load(); }
        }

        async function load() {
            if(wallet) {
                document.getElementById('pub-display').innerText = wallet.pub;
                const r = await fetch('/api/balance/' + wallet.pub);
                const d = await r.json();
                document.getElementById('bal-large').innerText = d.balance.toLocaleString() + ' AX';
            }

            const resC = await fetch('/api/chain');
            const chain = await resC.json();
            const feed = document.getElementById('recent-feed');
            feed.innerHTML = '';
            chain.slice().reverse().forEach(b => {
                feed.innerHTML += '<div class="glass-card p-3 mb-2" style="border-radius:15px; font-size:12px;"><div class="d-flex justify-content-between"><span class="accent fw-bold">BLOCK #' + b.index + '</span><span class="text-muted">' + new Date(b.timestamp*1000).toLocaleTimeString() + '</span></div><div class="text-truncate text-muted">' + b.hash + '</div></div>';
            });
        }

        async function mine() {
            const r = await fetch('/api/mine');
            if(r.ok) { alert("New Block Mined!"); load(); } else { alert("Mempool empty"); }
        }

        load();
        setInterval(load, 20000);
    </script>
</body>
</html>
`
