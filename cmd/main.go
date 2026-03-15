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

const DB_FILE = "blockchain_data.json"
// PURE 512-BIT TREASURY ADDRESS
const TREASURY_POOL_ADDR = "AX5def33f67eda5560561837935709169eb17955ffe13c1f112b3a329321bef540"

func loadChain() {
	file, err := os.ReadFile(DB_FILE)
	if err == nil {
		json.Unmarshal(file, &Blockchain)
	}
}

func saveChain() {
	data, _ := json.MarshalIndent(Blockchain, "", "  ")
	os.WriteFile(DB_FILE, data, 0644)
}

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
	// ROOT GENESIS ADDRESS
	rootAddr := "AX5eaba583bf646e0e39f41da6f9d8fa6db929c2e858bd32dffe6ac0cee2e3e974"

	loadChain()

	if len(Blockchain) == 0 {
		genesisTx := core.Transaction{Sender: "SYSTEM", Recipient: rootAddr, Amount: 1000002021}
		genesisTx.TxID = genesisTx.CalculateHash()
		
		genesisBlock := core.Block{
			Index: 0, 
			Timestamp: 1773561600,
			Transactions: []core.Transaction{genesisTx},
			PrevHash: strings.Repeat("0", 128),
			Difficulty: Difficulty,
		}
		genesisBlock.Mine()
		Blockchain = append(Blockchain, genesisBlock)
		saveChain()
	}

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
		json.NewDecoder(r.Body).Decode(&tx)
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
    <title>AX Core | L1 Network Console</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --ax-dark: #0F172A; --ax-blue: #2563EB; --ax-sky: #60A5FA; --bg: #F8FAFC; }
        body { background: var(--bg); color: #1E293B; font-family: 'Inter', system-ui, sans-serif; margin: 0; padding-bottom: 90px; }
        
        .sidebar { background: var(--ax-dark); height: 100vh; position: fixed; width: 280px; color: white; z-index: 1000; box-shadow: 10px 0 40px rgba(0,0,0,0.1); }
        .main-content { margin-left: 280px; padding: 40px; min-height: 100vh; }
        
        .nav-link-ax { color: #94A3B8; padding: 16px 28px; margin: 10px 15px; border-radius: 14px; font-weight: 600; cursor: pointer; display: flex; align-items: center; text-decoration: none; transition: 0.2s; }
        .nav-link-ax:hover, .nav-link-ax.active { background: var(--ax-blue); color: white; }
        .nav-link-ax i { width: 30px; font-size: 1.1rem; }

        .mobile-nav { background: white; position: fixed; bottom: 0; width: 100%; height: 80px; display: none; justify-content: space-around; align-items: center; border-top: 1px solid #E2E8F0; z-index: 2000; box-shadow: 0 -5px 20px rgba(0,0,0,0.03); }
        .m-nav-item { color: #94A3B8; text-align: center; font-size: 10px; font-weight: 800; cursor: pointer; flex: 1; transition: 0.2s; }
        .m-nav-item.active { color: var(--ax-blue); }
        .m-nav-item i { font-size: 24px; display: block; margin-bottom: 4px; }

        .card-ax { background: white; border-radius: 28px; border: none; box-shadow: 0 4px 25px rgba(0,0,0,0.03); padding: 35px; margin-bottom: 30px; }
        .hero { background: linear-gradient(135deg, var(--ax-dark) 0%, var(--ax-blue) 100%); color: white; border-radius: 32px; }
        .pill { background: #F1F5F9; padding: 14px; border-radius: 16px; font-family: 'JetBrains Mono', monospace; font-size: 0.75rem; word-break: break-all; margin-top: 12px; line-height: 1.4; border: 1px solid #E2E8F0; }
        .btn-ax { background: var(--ax-blue); color: white; border-radius: 16px; padding: 16px; font-weight: 700; border: none; width: 100%; box-shadow: 0 4px 15px rgba(37, 99, 235, 0.2); }
        
        @media (max-width: 992px) { .sidebar { display: none; } .main-content { margin-left: 0; padding: 20px; } .mobile-nav { display: flex; } }
    </style>
</head>
<body>

    <div class="sidebar">
        <div class="p-5 text-center"><h2 class="fw-bold m-0" style="color:white; letter-spacing:-2px;">AX CORE</h2><small class="opacity-50 fw-bold" style="font-size: 10px;">L1 ENTERPRISE</small></div>
        <nav>
            <div class="nav-link-ax active" onclick="nav('dash', this)"><i class="fas fa-grid-2"></i> Dashboard</div>
            <div class="nav-link-ax" onclick="nav('wallet', this)"><i class="fas fa-wallet"></i> Wallet</div>
            <div class="nav-link-ax" onclick="nav('security', this)"><i class="fas fa-shield-check"></i> Identity</div>
        </nav>
    </div>

    <div class="main-content">
        <div id="v-dash" class="view">
            <div class="card-ax hero text-center py-5">
                <small class="text-uppercase fw-bold opacity-75" style="letter-spacing: 1px;">Network Balance</small>
                <h1 id="bal-txt" class="display-3 fw-bold my-2">0.00</h1>
                <div id="addr-txt" class="pill bg-white bg-opacity-10 border-0 text-white opacity-75">Connect Vault</div>
            </div>
            <div class="card-ax text-center" style="border: 1px dashed var(--ax-sky);">
                <small class="fw-bold text-muted">TREASURY REWARDS (FIXED)</small>
                <h3 id="pool-txt" class="fw-bold m-0 text-primary">0.00 AX</h3>
                <div class="pill mt-2" style="font-size: 0.65rem;">` + TREASURY_POOL_ADDR + `</div>
            </div>
            <button class="btn-ax py-3 mb-4 shadow-sm" onclick="mine()">VALIDATE PENDING BLOCK (+50.00 AX)</button>
            <div id="mini-feed"></div>
        </div>

        <div id="v-wallet" class="view" style="display:none">
            <div class="card-ax mx-auto" style="max-width: 600px;">
                <h4 class="fw-bold mb-4">Send Assets</h4>
                <label class="small text-muted mb-2">Recipient $2^{512}$ Address</label>
                <input type="text" id="tx-to" class="form-control mb-3 p-3 border-0 bg-light rounded-4" placeholder="AX Address (128 characters)">
                <label class="small text-muted mb-2">Amount</label>
                <input type="number" id="tx-amt" class="form-control mb-4 p-3 border-0 bg-light rounded-4" placeholder="0.00">
                <button class="btn-ax" onclick="send()">SIGN & BROADCAST</button>
            </div>
        </div>

        <div id="v-security" class="view" style="display:none">
            <div class="card-ax">
                <h4 class="fw-bold mb-4">Identity Sync</h4>
                <input type="password" id="i-priv" class="form-control p-3 border-0 bg-light rounded-4 mb-4" placeholder="Secret Key">
                <button class="btn-ax" onclick="login()">SYNC WALLET</button>
                <hr class="my-5">
                <button class="btn btn-outline-dark w-100 py-3 rounded-4" onclick="gen()">GENERATE PURE 512-BIT KEY</button>
                <div id="g-res" class="mt-4" style="display:none">
                    <small class="fw-bold text-danger">Identity Generated. Backup the 128-char Address.</small>
                    <div class="pill mb-2" id="g-priv"></div>
                    <div class="pill fw-bold text-primary" id="g-pub" style="font-size: 0.65rem;"></div>
                </div>
            </div>
        </div>
    </div>

    <div class="mobile-nav">
        <div class="m-nav-item active" onclick="nav('dash', this)"><i class="fas fa-th-large"></i>Dash</div>
        <div class="m-nav-item" onclick="nav('wallet', this)"><i class="fas fa-paper-plane"></i>Send</div>
        <div class="m-nav-item" onclick="nav('security', this)"><i class="fas fa-key"></i>Identity</div>
    </div>

    <script>
        // PURE 512-BIT DERIVATION (128 HEX CHARS)
        async function derive(priv) {
            const buf = new TextEncoder().encode(priv);
            const hash = await crypto.subtle.digest('SHA-512', buf);
            const hex = Array.from(new Uint8Array(hash)).map(function(b) { 
                return b.toString(16).padStart(2,'0'); 
            }).join('');
            return 'AX' + hex; 
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
            if(!p) return;
            const pb = await derive(p);
            session = { pub: pb, priv: p };
            localStorage.setItem('ax_core_v16_session', JSON.stringify(session));
            location.reload();
        }

        async function load() {
            if(session) {
                const r = await fetch('/api/balance/' + session.pub);
                const d = await r.json();
                document.getElementById('bal-txt').innerText = d.balance.toLocaleString() + ' AX';
                document.getElementById('addr-txt').innerText = session.pub.substring(0,40) + "...";
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
            if(!session) return alert('Identity required');
            const r = await fetch('/api/mine?address=' + session.pub);
            if(r.ok) { alert('Success!'); load(); } else { alert('Network Busy or Treasury Issue.'); }
        }

        async function send() {
            const tx = { sender: session.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            const r = await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            if(r.ok) { alert('Broadcasting!'); nav('dash'); load(); } else { alert('Error: Check balance.'); }
        }

        async function gen() {
            const p = btoa(Math.random().toString() + Date.now()).substring(0,64);
            const pb = await derive(p);
            document.getElementById('g-res').style.display = 'block';
            document.getElementById('g-priv').innerText = p;
            document.getElementById('g-pub').innerText
