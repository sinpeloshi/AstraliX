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

// Global state variables
var Blockchain []core.Block
var Mempool []core.Transaction

// Network configuration constants
const DB_FILE = "blockchain_data.json"
const TREASURY_POOL_ADDR = "AX5def33f67eda5560561837935709169eb17955ffe13c1f112b3a329321bef540"

// loadChain initializes the chain from local storage
func loadChain() {
	file, err := os.ReadFile(DB_FILE)
	if err == nil {
		json.Unmarshal(file, &Blockchain)
		fmt.Println("Network history loaded.")
	}
}

// saveChain persists current state to disk
func saveChain() {
	data, _ := json.MarshalIndent(Blockchain, "", "  ")
	os.WriteFile(DB_FILE, data, 0644)
}

// getBalance scans the chain for a specific address balance
func getBalance(addr string) float64 {
	var balance float64
	for _, block := range Blockchain {
		for _, tx := range block.Transactions {
			if tx.Recipient == addr { balance += tx.Amount }
			if tx.Sender == addr { balance -= tx.Amount }
		}
	}
	return balance
}

func main() {
	const Difficulty = 4 
	// Genesis Root Authority Address
	rootAddr := "AX5eaba583bf646e0e39f41da6f9d8fa6db929c2e858bd32dffe6ac0cee2e3e974"

	loadChain()

	// Establish Genesis Block if network is new
	if len(Blockchain) == 0 {
		// INITIAL SUPPLY ISSUANCE
		genesisTx := core.Transaction{Sender: "SYSTEM", Recipient: rootAddr, Amount: 1000002021}
		genesisTx.TxID = genesisTx.CalculateHash()
		
		genesisBlock := core.Block{
			Index: 0, 
			Timestamp: 1773561600,
			Transactions: []core.Transaction{genesisTx},
			PrevHash: strings.Repeat("0", 128), // 512-bit Zero-Hash
			Difficulty: Difficulty,
		}
		genesisBlock.Mine()
		Blockchain = append(Blockchain, genesisBlock)
		saveChain()
		fmt.Println("Genesis Block successfully generated.")
	}

	// --- API Handlers ---

	http.HandleFunc("/api/balance/", func(w http.ResponseWriter, r *http.Request) {
		addr := strings.TrimPrefix(r.URL.Path, "/api/balance/")
		json.NewEncoder(w).Encode(map[string]interface{}{"balance": getBalance(addr)})
	})

	http.HandleFunc("/api/chain", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Blockchain)
	})

	http.HandleFunc("/api/mine", func(w http.ResponseWriter, r *http.Request) {
		miner := r.URL.Query().Get("address")
		if miner == "" || len(Mempool) == 0 { 
			http.Error(w, "Nothing to validate", 400); return 
		}

		reward := 50.0
		treasuryBalance := getBalance(TREASURY_POOL_ADDR)
		
		var txs []core.Transaction
		txs = append(txs, Mempool...)

		// Reward allocation from Fixed Treasury Pool
		if treasuryBalance >= reward {
			rewardTx := core.Transaction{Sender: TREASURY_POOL_ADDR, Recipient: miner, Amount: reward}
			rewardTx.TxID = rewardTx.CalculateHash()
			txs = append(txs, rewardTx)
		}

		prev := Blockchain[len(Blockchain)-1]
		newBlock := core.Block{
			Index: int64(len(Blockchain)), Timestamp: time.Now().Unix(),
			Transactions: txs, PrevHash: prev.Hash, Difficulty: Difficulty,
		}
		newBlock.Mine()
		Blockchain = append(Blockchain, newBlock)
		Mempool = []core.Transaction{}
		saveChain()
		json.NewEncoder(w).Encode(newBlock)
	})

	http.HandleFunc("/api/transactions/new", func(w http.ResponseWriter, r *http.Request) {
		var tx core.Transaction
		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			http.Error(w, "Payload Error", 400); return
		}
		// Basic balance check
		if tx.Sender != "SYSTEM" && tx.Sender != TREASURY_POOL_ADDR {
			if getBalance(tx.Sender) < tx.Amount {
				http.Error(w, "Low balance", 400); return
			}
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
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>AX Core | Network OS</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --ax-main: #002D5B; --ax-accent: #6CB2EB; --bg: #F8FAFC; }
        body { background: var(--bg); color: #2D3748; font-family: 'Inter', sans-serif; margin: 0; padding-bottom: 90px; }
        .sidebar { background: var(--ax-main); height: 100vh; position: fixed; width: 280px; color: white; z-index: 1000; box-shadow: 10px 0 30px rgba(0,0,0,0.05); }
        .main-content { margin-left: 280px; padding: 40px; min-height: 100vh; }
        .nav-link-ax { color: rgba(255,255,255,0.6); padding: 15px 30px; margin: 10px 15px; border-radius: 12px; font-weight: 600; cursor: pointer; display: flex; align-items: center; text-decoration: none; }
        .nav-link-ax.active { background: var(--ax-accent); color: white; }
        .mobile-nav { background: white; position: fixed; bottom: 0; width: 100%; height: 80px; display: none; justify-content: space-around; align-items: center; border-top: 1px solid #E2E8F0; z-index: 2000; }
        .m-nav-item { color: #A0AEC0; text-align: center; font-size: 10px; font-weight: 800; cursor: pointer; flex: 1; }
        .m-nav-item.active { color: var(--ax-main); }
        .card-ax { background: white; border-radius: 24px; border: none; box-shadow: 0 4px 20px rgba(0,0,0,0.02); padding: 30px; margin-bottom: 25px; }
        .hero { background: linear-gradient(135deg, var(--ax-main) 0%, var(--ax-accent) 100%); color: white; }
        .pill { background: #F1F5F9; padding: 12px; border-radius: 12px; font-family: monospace; font-size: 0.8rem; word-break: break-all; margin-top: 10px; }
        .btn-ax { background: var(--ax-main); color: white; border-radius: 14px; padding: 14px; font-weight: 700; border: none; width: 100%; transition: 0.2s; }
        @media (max-width: 992px) { .sidebar { display: none; } .main-content { margin-left: 0; padding: 20px; } .mobile-nav { display: flex; } }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="p-5 text-center"><h2 class="fw-bold m-0" style="color:white; letter-spacing:-2px;">AX CORE</h2><small class="opacity-50 fw-bold">NETWORK CONSOLE</small></div>
        <nav>
            <div class="nav-link-ax active" onclick="nav('dash', this)"><i class="fas fa-th-large me-2"></i> Overview</div>
            <div class="nav-link-ax" onclick="nav('wallet', this)"><i class="fas fa-wallet me-2"></i> Wallet</div>
            <div class="nav-link-ax" onclick="nav('security', this)"><i class="fas fa-user-shield me-2"></i> Identity</div>
        </nav>
    </div>

    <div class="main-content">
        <div id="v-dash" class="view">
            <div class="card-ax hero text-center py-5">
                <small class="text-uppercase fw-bold opacity-75">Personal Balance</small>
                <h1 id="bal-txt" class="display-3 fw-bold my-2">0.00</h1>
                <div id="addr-txt" class="pill bg-white bg-opacity-10 border-0 text-white opacity-75">Connect Identity</div>
            </div>
            <div class="card-ax text-center" style="border: 1px dashed var(--ax-accent);">
                <small class="fw-bold text-muted">TREASURY REWARDS POOL</small>
                <h3 id="pool-txt" class="fw-bold m-0 text-primary">0.00 AX</h3>
                <div class="pill mt-2" style="font-size: 0.6rem;">` + TREASURY_POOL_ADDR + `</div>
            </div>
            <button class="btn-ax py-3 mb-4" onclick="mine()">MINE PENDING BLOCK (+50.00 AX)</button>
            <div id="mini-feed"></div>
        </div>

        <div id="v-wallet" class="view" style="display:none">
            <div class="card-ax mx-auto" style="max-width: 500px;">
                <h4 class="fw-bold mb-4">Transfer Assets</h4>
                <input type="text" id="tx-to" class="form-control mb-3 p-3 border-0 bg-light rounded-4" placeholder="Destination AX Address">
                <input type="number" id="tx-amt" class="form-control mb-4 p-3 border-0 bg-light rounded-4" placeholder="Amount">
                <button class="btn-ax" onclick="send()">CONFIRM TRANSFER</button>
            </div>
        </div>

        <div id="v-security" class="view" style="display:none">
            <div class="card-ax">
                <h4 class="fw-bold mb-4">Vault Access</h4>
                <input type="password" id="i-priv" class="form-control p-3 border-0 bg-light rounded-4 mb-4" placeholder="Secret Private Key">
                <button class="btn-ax" onclick="login()">CONNECT</button>
                <hr class="my-5">
                <button class="btn btn-outline-dark w-100 py-3 rounded-4" onclick="gen()">GENERATE SEED PHRASE</button>
                <div id="g-res" class="mt-4" style="display:none">
                    <small class="fw-bold">New keys generated. Save these words safely.</small>
                    <div id="seed-words" class="my-2"></div>
                    <div class="pill mb-2" id="g-priv"></div>
                    <div class="pill fw-bold text-primary" id="g-pub"></div>
                </div>
            </div>
        </div>
    </div>

    <div class="mobile-nav">
        <div class="m-nav-item active" onclick="nav('dash', this)"><i class="fas fa-home"></i>Dash</div>
        <div class="m-nav-item" onclick="nav('wallet', this)"><i class="fas fa-paper-plane"></i>Send</div>
        <div class="m-nav-item" onclick="nav('security', this)"><i class="fas fa-key"></i>Keys</div>
    </div>

    <script>
        async function derive(priv) {
            const buf = new TextEncoder().encode(priv);
            const hash = await crypto.subtle.digest('SHA-512', buf);
            const hex = Array.from(new Uint8Array(hash)).map(function(b) { return b.toString(16).padStart(2,'0'); }).join('');
            return 'AX' + hex.substring(0, 64);
        }

        let session = JSON.parse(localStorage.getItem('ax_core_v16_session')) || null;

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
            localStorage.setItem('ax_core_v16_session', JSON.stringify(session));
            location.reload();
        }

        function logout() { localStorage.removeItem('ax_core_v16_session'); location.reload(); }

        async function load() {
            if(session) {
                const r = await fetch('/api/balance/' + session.pub);
                const d = await r.json();
                document.getElementById('bal-txt').innerText = d.balance.toLocaleString() + ' AX';
                document.getElementById('addr-txt').innerText = session.pub.substring(0,30) + "...";
            }
            const rp = await fetch('/api/balance/` + TREASURY_POOL_ADDR + `');
            const dp = await rp.json();
            document.getElementById('pool-txt').innerText = dp.balance.toLocaleString() + ' AX';

            const res = await fetch('/api/chain');
            const chain = await res.json();
            const mini = document.getElementById('mini-feed');
            mini.innerHTML = '';
            chain.reverse().slice(0,5).forEach(function(b) {
                const h = (b.Hash || b.hash || '').substring(0,25) + '...';
                mini.innerHTML += '<div class="p-3 border-bottom d-flex justify-content-between small"><span>#' + b.index + '</span><span class="text-muted">' + h + '</span></div>';
            });
        }

        async function mine() {
            if(!session) return alert('Login required');
            const r = await fetch('/api/mine?address=' + session.pub);
            if(r.ok) { alert('Block mined!'); load(); } else { alert('Mempool or Treasury issue.'); }
        }

        async function send() {
            const tx = { sender: session.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            const r = await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            if(r.ok) { alert('Authorized!'); nav('dash'); load(); } else { alert('Error: Check funds.'); }
        }

        async function gen() {
            const words = ["abandon", "ability", "able", "about", "above", "absent", "absorb", "abstract", "absurd", "abuse", "access", "accident"];
            const seed = Array.from({length: 12}, () => words[Math.floor(Math.random()*words.length)]).join(' ');
            const p = btoa(seed).substring(0,64);
            const pb = await derive(p);
            document.getElementById('g-res').style.display = 'block';
            document.getElementById('g-priv').innerText = p;
            document.getElementById('g-pub').innerText = pb;
        }

        load(); setInterval(load, 15000);
    </script>
</body>
</html>
`
