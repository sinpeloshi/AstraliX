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

// Global variables for chain state
var Blockchain []core.Block
var Mempool []core.Transaction

// Constants for system configuration
const DB_FILE = "blockchain_data.json"
const TREASURY_POOL_ADDR = "AX_GLOBAL_REWARDS_TREASURY_POOL_SECURE_512"

// loadChain reads the existing blockchain from local storage
func loadChain() {
	file, err := os.ReadFile(DB_FILE)
	if err == nil {
		json.Unmarshal(file, &Blockchain)
		fmt.Println("Blockchain successfully loaded from disk.")
	}
}

// saveChain persists the current state to a JSON file
func saveChain() {
	data, _ := json.MarshalIndent(Blockchain, "", "  ")
	os.WriteFile(DB_FILE, data, 0644)
}

// getBalance calculates the current balance of any specific address
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
	// Root authority address derived from your private key
	rootAddr := "AX5eaba583bf646e0e39f41da6f9d8fa6db929c2e858bd32dffe6ac0cee2e3e974"

	loadChain()

	// Initialize Genesis Block if chain is empty
	if len(Blockchain) == 0 {
		// FIXED SUPPLY ISSUANCE: 1,000,002,021 AX
		genesisTx := core.Transaction{Sender: "SYSTEM", Recipient: rootAddr, Amount: 1000002021}
		genesisTx.TxID = genesisTx.CalculateHash()
		
		genesisBlock := core.Block{
			Index: 0, 
			Timestamp: 1773561600,
			Transactions: []core.Transaction{genesisTx},
			PrevHash: strings.Repeat("0", 128), // 512-bit standard
			Difficulty: Difficulty,
		}
		genesisBlock.Mine()
		Blockchain = append(Blockchain, genesisBlock)
		saveChain()
	}

	// --- API Endpoints ---

	// GET /api/balance/{address}
	http.HandleFunc("/api/balance/", func(w http.ResponseWriter, r *http.Request) {
		addr := strings.TrimPrefix(r.URL.Path, "/api/balance/")
		json.NewEncoder(w).Encode(map[string]interface{}{"balance": getBalance(addr), "address": addr})
	})

	// GET /api/chain
	http.HandleFunc("/api/chain", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Blockchain)
	})

	// GET /api/mine?address={miner_address}
	http.HandleFunc("/api/mine", func(w http.ResponseWriter, r *http.Request) {
		miner := r.URL.Query().Get("address")
		if miner == "" || len(Mempool) == 0 { 
			http.Error(w, "Invalid request or empty mempool", 400); return 
		}

		reward := 50.0
		treasuryBalance := getBalance(TREASURY_POOL_ADDR)
		
		var transactions []core.Transaction
		transactions = append(transactions, Mempool...)

		// Allocate rewards only if treasury has sufficient funds
		if treasuryBalance >= reward {
			rewardTx := core.Transaction{Sender: TREASURY_POOL_ADDR, Recipient: miner, Amount: reward}
			rewardTx.TxID = rewardTx.CalculateHash()
			transactions = append(transactions, rewardTx)
		}

		prevBlock := Blockchain[len(Blockchain)-1]
		newBlock := core.Block{
			Index: int64(len(Blockchain)), Timestamp: time.Now().Unix(),
			Transactions: transactions, PrevHash: prevBlock.Hash, Difficulty: Difficulty,
		}
		newBlock.Mine()
		Blockchain = append(Blockchain, newBlock)
		Mempool = []core.Transaction{}
		saveChain()
		json.NewEncoder(w).Encode(newBlock)
	})

	// POST /api/transactions/new
	http.HandleFunc("/api/transactions/new", func(w http.ResponseWriter, r *http.Request) {
		var tx core.Transaction
		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			http.Error(w, "Bad Payload", 400); return
		}
		// Balance validation for non-system addresses
		if tx.Sender != "SYSTEM" && tx.Sender != TREASURY_POOL_ADDR {
			if getBalance(tx.Sender) < tx.Amount {
				http.Error(w, "Insufficient funds", 400); return
			}
		}
		tx.TxID = tx.CalculateHash()
		Mempool = append(Mempool, tx)
		w.WriteHeader(201)
	})

	// Root Dashboard
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, dashboardHTML)
	})

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	fmt.Printf("AX Core Node online on port %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>AX Core | Global Network</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --ax-main: #002D5B; --ax-accent: #6CB2EB; --bg: #F1F5F9; --white: #FFFFFF; }
        body { background: var(--bg); color: #2D3748; font-family: 'Inter', -apple-system, sans-serif; margin: 0; padding-bottom: 90px; }
        
        .sidebar { background: var(--ax-main); height: 100vh; position: fixed; width: 280px; color: white; z-index: 1000; box-shadow: 10px 0 30px rgba(0,0,0,0.05); }
        .main-content { margin-left: 280px; padding: 40px; min-height: 100vh; transition: 0.3s; }
        
        .nav-link-ax { color: rgba(255,255,255,0.6); padding: 15px 30px; margin: 10px 15px; border-radius: 12px; font-weight: 600; cursor: pointer; display: flex; align-items: center; text-decoration: none; transition: 0.2s; }
        .nav-link-ax:hover, .nav-link-ax.active { background: var(--ax-accent); color: white; }
        .nav-link-ax i { width: 30px; font-size: 1.1rem; }

        .mobile-nav { background: var(--white); position: fixed; bottom: 0; width: 100%; height: 80px; display: none; justify-content: space-around; align-items: center; border-top: 1px solid #E2E8F0; z-index: 2000; box-shadow: 0 -5px 20px rgba(0,0,0,0.03); }
        .m-nav-item { color: #A0AEC0; text-align: center; font-size: 10px; font-weight: 800; cursor: pointer; flex: 1; }
        .m-nav-item.active { color: var(--ax-main); }
        .m-nav-item i { font-size: 24px; display: block; margin-bottom: 4px; }

        .card-ax { background: var(--white); border-radius: 24px; border: none; box-shadow: 0 4px 20px rgba(0,0,0,0.02); padding: 30px; margin-bottom: 25px; }
        .hero { background: linear-gradient(135deg, var(--ax-main) 0%, var(--ax-accent) 100%); color: white; }
        .btn-ax { background: var(--ax-main); color: white; border-radius: 14px; padding: 14px; font-weight: 700; border: none; width: 100%; transition: 0.2s; }
        .btn-ax:hover { opacity: 0.9; transform: translateY(-1px); }
        
        .pill { background: #EDF2F7; padding: 12px; border-radius: 12px; font-family: monospace; font-size: 0.8rem; color: #4A5568; word-break: break-all; margin-top: 10px; }
        .word-tag { display: inline-block; background: #E2E8F0; color: var(--ax-main); padding: 6px 12px; border-radius: 8px; margin: 4px; font-size: 13px; font-weight: 700; }

        @media (max-width: 992px) { .sidebar { display: none; } .main-content { margin-left: 0; padding: 20px; } .mobile-nav { display: flex; } }
    </style>
</head>
<body>

    <div class="sidebar">
        <div class="p-5 text-center"><h2 class="fw-bold m-0" style="color:white; letter-spacing:-2px;">AX CORE</h2><small class="opacity-50 fw-bold">GLOBAL NETWORK</small></div>
        <nav>
            <div class="nav-link-ax active" onclick="nav('dash', this)"><i class="fas fa-th-large"></i> Dashboard</div>
            <div class="nav-link-ax" onclick="nav('wallet', this)"><i class="fas fa-wallet"></i> Wallet</div>
            <div class="nav-link-ax" onclick="nav('explorer', this)"><i class="fas fa-network-wired"></i> Explorer</div>
            <div class="nav-link-ax" onclick="nav('security', this)"><i class="fas fa-user-shield"></i> Security</div>
        </nav>
    </div>

    <div class="main-content">
        <div id="v-dash" class="view">
            <div class="card-ax hero text-center">
                <small class="text-uppercase fw-bold opacity-75">Available Balance</small>
                <h1 id="bal-txt" class="display-3 fw-bold my-2">0.00</h1>
                <div id="addr-txt" class="pill bg-white bg-opacity-10 border-0 text-white opacity-75">Connect your wallet</div>
            </div>
            
            <div class="card-ax text-center" style="border: 1px dashed var(--ax-accent);">
                <small class="fw-bold text-muted">TREASURY REWARDS POOL</small>
                <h3 id="pool-txt" class="fw-bold m-0 text-primary">0.00 AX</h3>
            </div>

            <button class="btn-ax py-3 mb-4 shadow-sm" onclick="mine()">MINE BLOCK (+50.00 AX)</button>
            <div id="mini-feed"></div>
        </div>

        <div id="v-wallet" class="view" style="display:none">
            <div class="card-ax mx-auto" style="max-width: 500px;">
                <h4 class="fw-bold mb-4">Send Assets</h4>
                <label class="small text-muted mb-2">Recipient Address</label>
                <input type="text" id="tx-to" class="form-control mb-3 p-3 border-0 bg-light rounded-4" placeholder="AX Address">
                <label class="small text-muted mb-2">Amount</label>
                <input type="number" id="tx-amt" class="form-control mb-4 p-3 border-0 bg-light rounded-4" placeholder="0.00">
                <button class="btn-ax" onclick="send()">AUTHORIZE TRANSFER</button>
            </div>
        </div>

        <div id="v-security" class="view" style="display:none">
            <div class="card-ax">
                <h4 class="fw-bold mb-3">Secure Access</h4>
                <p class="text-muted small">Enter your 512-bit private key to sync your account.</p>
                <input type="password" id="i-priv" class="form-control p-3 border-0 bg-light rounded-4 mb-4" placeholder="Private Key">
                <button class="btn-ax mb-3" onclick="login()">CONNECT WALLET</button>
                <button class="btn btn-link text-danger w-100 text-decoration-none small" onclick="logout()">Disconnect Account</button>
                <hr class="my-5">
                <h4 class="fw-bold mb-3">Identity Generator</h4>
                <button class="btn btn-outline-dark w-100 py-3 rounded-4" onclick="gen()">GENERATE NEW SEED</button>
                <div id="g-res" class="mt-4" style="display:none">
                    <label class="small fw-bold">BIP-39 Seed Words (English):</label>
                    <div id="seed-words" class="my-2"></div>
                    <label class="small fw-bold">Private Key (512-bit):</label>
                    <div class="pill mb-2" id="g-priv"></div>
                    <label class="small fw-bold text-primary">Public AX Address:</label>
                    <div class="pill fw-bold" id="g-pub"></div>
                </div>
            </div>
        </div>

        <div id="v-explorer" class="view" style="display:none">
            <div class="card-ax"><h4 class="fw-bold mb-4">Blockchain Explorer</h4><div id="full-chain"></div></div>
        </div>
    </div>

    <div class="mobile-nav">
        <div class="m-nav-item active" onclick="nav('dash', this)"><i class="fas fa-home"></i>Dash</div>
        <div class="m-nav-item" onclick="nav('wallet', this)"><i class="fas fa-paper-plane"></i>Send</div>
        <div class="m-nav-item" onclick="nav('explorer', this)"><i class="fas fa-database"></i>Node</div>
        <div class="m-nav-item" onclick="nav('security', this)"><i class="fas fa-key"></i>Keys</div>
    </div>

    <script>
        // Standard SHA-512 derivation
        async function derive(priv) {
            const buf = new TextEncoder().encode(priv);
            const hash = await crypto.subtle.digest('SHA-512', buf);
            const hex = Array.from(new Uint8Array(hash)).map(function(b) { return b.toString(16).padStart(2,'0'); }).join('');
            return 'AX' + hex.substring(0, 64);
        }

        let session = JSON.parse(localStorage.getItem('ax_core_session')) || null;

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
            localStorage.setItem('ax_core_session', JSON.stringify(session));
            location.reload();
        }

        function logout() { localStorage.removeItem('ax_core_session'); location.reload(); }

        async function load() {
            if(session) {
                document.getElementById('addr-txt').innerText = session.pub.substring(0,25) + "...";
                const r = await fetch('/api/balance/' + session.pub);
                const d = await r.json();
                document.getElementById('bal-txt').innerText = d.balance.toLocaleString() + ' AX';
            }
            const rp = await fetch('/api/balance/AX_GLOBAL_REWARDS_TREASURY_POOL_SECURE_512');
            const dp = await rp.json();
            document.getElementById('pool-txt').innerText = dp.balance.toLocaleString() + ' AX';

            const res = await fetch('/api/chain');
            const chain = await res.json();
            const mini = document.getElementById('mini-feed');
            const full = document.getElementById('full-chain');
            mini.innerHTML = ''; full.innerHTML = '';
            chain.reverse().forEach(function(b) {
                const h = (b.Hash || b.hash || '').substring(0,25) + '...';
                mini.innerHTML += '<div class="p-2 border-bottom d-flex justify-content-between small"><span>#' + b.index + '</span><span>' + h + '</span></div>';
                full.innerHTML += '<div class="card-ax border mb-2 small"><b>Block #' + b.index + '</b><div class="pill">' + (b.Hash || b.hash) + '</div></div>';
            });
        }

        async function mine() {
            if(!session) return alert('Sync your wallet first');
            const r = await fetch('/api/mine?address=' + session.pub);
            if(r.ok) { alert('Mined! Reward accredited.'); load(); } 
            else { alert('Nothing to mine or Treasury empty.'); }
        }

        async function send() {
            const tx = { sender: session.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            const r = await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            if(r.ok) { alert('Transaction sent!'); nav('dash'); load(); } else { alert('Insufficient funds'); }
        }

        async function gen() {
            const words = ["abandon", "ability", "able", "about", "above", "absent", "absorb", "abstract", "absurd", "abuse", "access", "accident", "account", "accuse", "achieve", "acid", "acoustic", "acquire", "across", "act", "action", "actor", "actress", "actual", "adapt", "add", "addict", "address", "adjust", "admit", "adult", "advance", "advice", "aerobic", "affair", "afford", "afraid", "again", "age", "agent"];
            const seed = Array.from({length: 12}, () => words[Math.floor(Math.random()*words.length)]).join(' ');
            const p = btoa(seed).substring(0,64);
            const pb = await derive(p);
            document.getElementById('g-res').style.display = 'block';
            document.getElementById('seed-words').innerHTML = seed.split(' ').map(function(w) { return '<span class="word-tag">' + w + '</span>'; }).join('');
            document.getElementById('g-priv').innerText = p;
            document.getElementById('g-pub').innerText = pb;
        }

        load(); setInterval(load, 15000);
    </script>
</body>
</html>
`
