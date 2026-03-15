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
	// Asegúrate de poner tu dirección AX generada aquí para ver tus fondos
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
	fmt.Printf("🌐 AstraliX White Node running on port %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>AstraliX | Digital Banking & Network</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --main-blue: #003366; --celeste: #74ACDF; --bg-light: #F4F7F9; --white: #FFFFFF; }
        body { background: var(--bg-light); color: #333; font-family: 'Segoe UI', Roboto, sans-serif; margin: 0; }
        
        /* Layout */
        .sidebar { background: var(--main-blue); height: 100vh; position: fixed; width: 260px; color: white; transition: 0.3s; z-index: 1000; }
        .content { margin-left: 260px; padding: 30px; }
        
        /* Sidebar Nav */
        .nav-link { color: rgba(255,255,255,0.7); padding: 15px 25px; margin: 5px 15px; border-radius: 10px; transition: 0.2s; cursor: pointer; text-decoration: none; display: flex; align-items: center; }
        .nav-link:hover, .nav-link.active { background: var(--celeste); color: white; }
        .nav-link i { width: 25px; font-size: 18px; }

        /* Components */
        .card-custom { background: var(--white); border: none; border-radius: 20px; box-shadow: 0 10px 30px rgba(0,0,0,0.05); padding: 25px; margin-bottom: 25px; }
        .balance-card { background: linear-gradient(135deg, var(--main-blue) 0%, var(--celeste) 100%); color: white; }
        .btn-primary-ax { background: var(--main-blue); color: white; border: none; border-radius: 12px; padding: 12px 25px; font-weight: 600; }
        .btn-celeste { background: var(--celeste); color: white; border: none; border-radius: 12px; padding: 12px 25px; font-weight: 600; }
        
        .status-dot { width: 10px; height: 10px; background: #2ecc71; border-radius: 50%; display: inline-block; margin-right: 8px; }
        .addr-pill { background: #E9ECEF; padding: 8px 15px; border-radius: 10px; font-family: monospace; font-size: 0.85rem; color: #666; word-break: break-all; }

        @media (max-width: 992px) {
            .sidebar { transform: translateX(-100%); }
            .content { margin-left: 0; padding: 15px; }
            .sidebar.active { transform: translateX(0); }
            .mobile-header { display: flex !important; }
        }
    </style>
</head>
<body>

    <div class="sidebar" id="sidebar">
        <div class="p-4 mb-4 text-center">
            <h3 class="fw-bold mb-0">ASTRALIX</h3>
            <small class="opacity-50">Secure 512-bit Protocol</small>
        </div>
        <nav>
            <a class="nav-link active" onclick="showPage('dash', this)"><i class="fas fa-home"></i> <span>Dashboard</span></a>
            <a class="nav-link" onclick="showPage('wallet', this)"><i class="fas fa-wallet"></i> <span>My Wallet</span></a>
            <a class="nav-link" onclick="showPage('explorer', this)"><i class="fas fa-cubes"></i> <span>Explorer</span></a>
            <a class="nav-link" onclick="showPage('stats', this)"><i class="fas fa-chart-line"></i> <span>Network Stats</span></a>
            <a class="nav-link" onclick="showPage('contacts', this)"><i class="fas fa-address-book"></i> <span>Contacts</span></a>
        </nav>
    </div>

    <div class="content">
        <div class="d-flex justify-content-between align-items-center mb-4">
            <div class="mobile-header d-none align-items-center">
                <button class="btn btn-dark me-3" onclick="toggleSidebar()"><i class="fas fa-bars"></i></button>
                <h4 class="mb-0 fw-bold">AstraliX</h4>
            </div>
            <div class="d-none d-lg-block">
                <h4 class="fw-bold text-dark mb-0">Network Overview</h4>
                <small class="text-muted"><span class="status-dot"></span> Node is Synchronized</small>
            </div>
            <div class="d-flex align-items-center">
                <div class="text-end me-3 d-none d-sm-block">
                    <small class="text-muted d-block">Node Port</small>
                    <span class="fw-bold">8080 (Cloud)</span>
                </div>
                <button class="btn btn-light rounded-circle p-3 shadow-sm" onclick="location.reload()"><i class="fas fa-sync"></i></button>
            </div>
        </div>

        <div id="page-dash" class="page">
            <div class="row g-4">
                <div class="col-lg-8">
                    <div class="card-custom balance-card">
                        <div class="row align-items-center">
                            <div class="col-md-7">
                                <h6 class="text-uppercase opacity-75">Available AX Balance</h6>
                                <h1 id="bal-main" class="fw-bold display-4 mb-3">0.00</h1>
                                <div id="connected-addr" class="addr-pill bg-white bg-opacity-10 text-white border-0 text-truncate">Connect your wallet to start</div>
                            </div>
                            <div class="col-md-5 text-end d-none d-md-block">
                                <i class="fas fa-shield-alt fa-8x opacity-25"></i>
                            </div>
                        </div>
                    </div>
                    
                    <div class="card-custom">
                        <h5 class="fw-bold mb-4">Recent Blocks</h5>
                        <div class="table-responsive">
                            <table class="table align-middle">
                                <thead class="text-muted small">
                                    <tr><th>Block #</th><th>Hash</th><th>TXs</th><th>Time</th></tr>
                                </thead>
                                <tbody id="mini-explorer"></tbody>
                            </table>
                        </div>
                    </div>
                </div>
                
                <div class="col-lg-4">
                    <div class="card-custom">
                        <h5 class="fw-bold mb-3">Quick Actions</h5>
                        <button class="btn btn-primary-ax w-100 mb-2" onclick="showPage('wallet')">Send Money</button>
                        <button class="btn btn-outline-dark w-100 mb-4" onclick="mine()">Mine Next Block</button>
                        <hr>
                        <h6 class="text-muted small mb-3">Network Summary</h6>
                        <div class="d-flex justify-content-between mb-2"><span>Supply</span><span class="fw-bold">1.0B AX</span></div>
                        <div class="d-flex justify-content-between mb-2"><span>Height</span><span id="stat-height" class="fw-bold">0</span></div>
                        <div class="d-flex justify-content-between mb-2"><span>Difficulty</span><span class="fw-bold text-primary">4 (PoW)</span></div>
                    </div>
                </div>
            </div>
        </div>

        <div id="page-wallet" class="page" style="display:none">
            <div class="row g-4">
                <div class="col-md-6">
                    <div class="card-custom">
                        <h4 class="fw-bold mb-4">Transfer AX</h4>
                        <div class="mb-3">
                            <label class="form-label small text-muted">Recipient Address</label>
                            <input type="text" id="tx-to" class="form-control p-3 border-0 bg-light" placeholder="AX...">
                        </div>
                        <div class="mb-4">
                            <label class="form-label small text-muted">Amount</label>
                            <input type="number" id="tx-amt" class="form-control p-3 border-0 bg-light" placeholder="0.00">
                        </div>
                        <button class="btn btn-celeste w-100 py-3 fw-bold" onclick="sendTx()">AUTHORIZE TRANSACTION</button>
                    </div>
                </div>
                <div class="col-md-6">
                    <div class="card-custom">
                        <h4 class="fw-bold mb-4">Access Account</h4>
                        <div class="mb-4">
                            <label class="form-label small text-muted">Private Key (512-bit Secret)</label>
                            <input type="password" id="imp-priv" class="form-control p-3 border-0 bg-light" placeholder="Enter Secret Key">
                        </div>
                        <button class="btn btn-primary-ax w-100 mb-3" onclick="importWallet()">CONNECT SESSION</button>
                        <button class="btn btn-link w-100 text-danger text-decoration-none small" onclick="logout()">Disconnect Wallet</button>
                    </div>
                    <div class="card-custom bg-light border-0 shadow-none">
                        <h5 class="fw-bold">Need a New Wallet?</h5>
                        <p class="small text-muted">Generate a secure cryptographic keypair for the network.</p>
                        <button class="btn btn-outline-dark btn-sm rounded-pill" onclick="createNew()">Generate New Identity</button>
                        <div id="keys-box" class="mt-3 small" style="display:none">
                            <div class="addr-pill mb-2 bg-white" id="gen-priv"></div>
                            <div class="addr-pill bg-white text-primary fw-bold" id="gen-pub"></div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div id="page-explorer" class="page" style="display:none">
            <div class="card-custom">
                <h4 class="fw-bold mb-4">Chain Explorer</h4>
                <div id="full-chain"></div>
            </div>
        </div>
    </div>

    <script>
        async function deriveAddress(priv) {
            const msgBuffer = new TextEncoder().encode(priv);
            const hashBuffer = await crypto.subtle.digest('SHA-512', msgBuffer);
            const hashArray = Array.from(new Uint8Array(hashBuffer));
            const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
            return 'AX' + hashHex.substring(0, 64);
        }

        let wallet = JSON.parse(localStorage.getItem('ax_session')) || null;

        function showPage(id, el) {
            document.querySelectorAll('.page').forEach(p => p.style.display = 'none');
            document.getElementById('page-' + id).style.display = 'block';
            if(el) {
                document.querySelectorAll('.nav-link').forEach(l => l.classList.remove('active'));
                el.classList.add('active');
            }
        }

        async function importWallet() {
            const priv = document.getElementById('imp-priv').value;
            const pub = await deriveAddress(priv);
            wallet = { pub, priv };
            localStorage.setItem('ax_session', JSON.stringify(wallet));
            location.reload();
        }

        function logout() { localStorage.removeItem('ax_session'); location.reload(); }

        async function load() {
            if(wallet) {
                document.getElementById('connected-addr').innerText = wallet.pub;
                const r = await fetch('/api/balance/' + wallet.pub);
                const d = await r.json();
                document.getElementById('bal-main').innerText = d.balance.toLocaleString() + ' AX';
            }
            const resC = await fetch('/api/chain');
            const chain = await resC.json();
            document.getElementById('stat-height').innerText = chain.length;

            const mini = document.getElementById('mini-explorer');
            const full = document.getElementById('full-chain');
            mini.innerHTML = ''; full.innerHTML = '';

            chain.slice().reverse().forEach(b => {
                const row = ` + "`" + `<tr><td>#${b.index}</td><td class="text-muted small">${b.hash.substring(0,24)}...</td><td>${b.transactions?b.transactions.length:0}</td><td>${new Date(b.timestamp*1000).toLocaleTimeString()}</td></tr>` + "`" + `;
                mini.innerHTML += row;
                
                full.innerHTML += ` + "`" + `<div class="card-custom border mb-3"><div class="d-flex justify-content-between"><span class="fw-bold text-primary">BLOCK #${b.index}</span><small class="text-muted">${new Date(b.timestamp*1000).toLocaleString()}</small></div><div class="addr-pill my-2">${b.hash}</div><div class="small">Prev Hash: ${b.prev_hash || 'None'}</div></div>` + "`" + `;
            });
        }

        async function mine() {
            const r = await fetch('/api/mine');
            if(r.ok) { alert("Block Successfully Mined!"); load(); } else { alert("Mempool Empty"); }
        }

        async function sendTx() {
            if(!wallet) return showPage('wallet');
            const tx = { sender: wallet.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            alert("Transaction Broadcasted!"); showPage('dash'); load();
        }

        async function createNew() {
            const priv = btoa(Math.random().toString()).substring(0, 128);
            const pub = await deriveAddress(priv);
            document.getElementById('keys-box').style.display = 'block';
            document.getElementById('gen-priv').innerText = 'Priv: ' + priv;
            document.getElementById('gen-pub').innerText = 'Pub: ' + pub;
        }

        function toggleSidebar() { document.getElementById('sidebar').classList.toggle('active'); }

        load();
        setInterval(load, 20000);
    </script>
</body>
</html>
`
