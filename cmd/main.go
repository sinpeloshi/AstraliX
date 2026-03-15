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

// PURE 512-BIT TREASURY ADDRESS (128 HEX CHARACTERS)
const TREASURY_POOL_ADDR = "AX5def33f67eda5560561837935709169eb17955ffe13c1f112b3a329321bef5400000000000000000000000000000000000000000000000000000000000000000"

func loadChain() {
	file, err := os.ReadFile(DB_FILE)
	if err == nil { json.Unmarshal(file, &Blockchain) }
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
	// ROOT GENESIS (Matches your existing key but displays full 128 chars in frontend)
	rootAddr := "AX5eaba583bf646e0e39f41da6f9d8fa6db929c2e858bd32dffe6ac0cee2e3e97400000000000000000000000000000000000000000000000000000000000000000"

	loadChain()

	if len(Blockchain) == 0 {
		genesisTx := core.Transaction{Sender: "SYSTEM", Recipient: rootAddr, Amount: 1000002021}
		genesisTx.TxID = genesisTx.CalculateHash()
		
		genesisBlock := core.Block{
			Index: 0, Timestamp: 1773561600,
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

	http.HandleFunc("/api/mine", func(w http.ResponseWriter, r *http.Request) {
		miner := r.URL.Query().Get("address")
		if miner == "" || len(Mempool) == 0 { http.Error(w, "Mempool empty", 400); return }

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
		if tx.Sender != "SYSTEM" && tx.Sender != TREASURY_POOL_ADDR && getBalance(tx.Sender) < tx.Amount {
			http.Error(w, "Insufficient balance", 400); return
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
    <title>AX Core | Global Node</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --ax-dark: #0F172A; --ax-main: #2563EB; --bg: #F8FAFC; }
        body { background: var(--bg); font-family: 'Inter', sans-serif; padding-bottom: 90px; margin: 0; }
        .sidebar { background: var(--ax-dark); height: 100vh; position: fixed; width: 280px; color: white; z-index: 1000; box-shadow: 10px 0 40px rgba(0,0,0,0.1); }
        .main-content { margin-left: 280px; padding: 40px; min-height: 100vh; }
        .nav-link-ax { color: #94A3B8; padding: 16px 28px; margin: 8px 15px; border-radius: 12px; font-weight: 600; cursor: pointer; display: flex; align-items: center; text-decoration: none; transition: 0.2s; }
        .nav-link-ax.active { background: var(--ax-main); color: white; }
        .card-ax { background: white; border-radius: 28px; box-shadow: 0 4px 25px rgba(0,0,0,0.02); padding: 30px; margin-bottom: 25px; border: none; }
        .hero { background: linear-gradient(135deg, var(--ax-dark) 0%, var(--ax-main) 100%); color: white; border-radius: 32px; padding: 50px 30px; }
        .pill { background: #F1F5F9; padding: 14px; border-radius: 16px; font-family: 'JetBrains Mono', monospace; font-size: 0.65rem; word-break: break-all; margin-top: 12px; border: 1px solid #E2E8F0; }
        .btn-ax { background: var(--ax-main); color: white; border-radius: 16px; padding: 16px; font-weight: 700; border: none; width: 100%; }
        @media (max-width: 992px) { .sidebar { display: none; } .main-content { margin-left: 0; padding: 20px; } .mobile-nav { display: flex; } }
        .mobile-nav { background: white; position: fixed; bottom: 0; width: 100%; height: 80px; display: none; justify-content: space-around; align-items: center; border-top: 1px solid #E2E8F0; z-index: 2000; }
        .m-nav-item { color: #94A3B8; text-align: center; font-size: 10px; font-weight: 800; cursor: pointer; flex: 1; }
        .m-nav-item.active { color: var(--ax-main); }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="p-5 text-center"><h2 class="fw-bold m-0" style="color:white; letter-spacing:-2px;">AX CORE</h2><small class="opacity-50 fw-bold">GLOBAL L1</small></div>
        <nav>
            <div class="nav-link-ax active" onclick="nav('dash', this)"><i class="fas fa-th-large me-2"></i> Dashboard</div>
            <div class="nav-link-ax" onclick="nav('wallet', this)"><i class="fas fa-wallet me-2"></i> Wallet</div>
            <div class="nav-link-ax" onclick="nav('security', this)"><i class="fas fa-key me-2"></i> Security</div>
        </nav>
    </div>

    <div class="main-content">
        <div id="v-dash" class="view">
            <div class="card-ax hero text-center">
                <small class="text-uppercase fw-bold opacity-75">Network Balance</small>
                <h1 id="bal-txt" class="display-3 fw-bold my-2">0.00</h1>
                <div id="addr-txt" class="pill bg-white bg-opacity-10 border-0 text-white opacity-75">Connect Wallet</div>
            </div>
            <div class="card-ax text-center" style="border: 1px dashed #60A5FA;">
                <small class="fw-bold text-muted">FIXED REWARDS POOL ($2^{512}$)</small>
                <h3 id="pool-txt" class="fw-bold m-0 text-primary">0.00 AX</h3>
                <div class="pill mt-2">AX5def33f67eda5560561837935709169eb17955ffe13c1f112b3a329321bef540...</div>
            </div>
            <button class="btn-ax py-3" onclick="mine()">MINE BLOCK (+50.00 AX)</button>
        </div>

        <div id="v-wallet" class="view" style="display:none">
            <div class="card-ax mx-auto" style="max-width: 600px;">
                <h4 class="fw-bold mb-4">Send AX Assets</h4>
                <input type="text" id="tx-to" class="form-control mb-3 p-3 border-0 bg-light rounded-4" placeholder="Destination Address (128 chars)">
                <input type="number" id="tx-amt" class="form-control mb-4 p-3 border-0 bg-light rounded-4" placeholder="0.00">
                <button class="btn-ax" onclick="send()">CONFIRM TRANSFER</button>
            </div>
        </div>

        <div id="v-security" class="view" style="display:none">
            <div class="card-ax">
                <h4 class="fw-bold mb-4">Identity Sync</h4>
                <input type="password" id="i-priv" class="form-control p-3 border-0 bg-light rounded-4 mb-4" placeholder="Private Key">
                <button class="btn-ax" onclick="login()">CONNECT</button>
                <hr class="my-5">
                <button class="btn btn-outline-dark w-100 py-3 rounded-4" onclick="gen()">GENERATE PURE 512-BIT KEY</button>
                <div id="g-res" class="mt-4" style="display:none">
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
        // PURE 512-BIT DERIVATION (128 HEX CHARS)
        async function derive(priv) {
            const buf = new TextEncoder().encode(priv);
            const hash = await crypto.subtle.digest('SHA-512', buf);
            const hex = Array.from(new Uint8Array(hash)).map(function(b) { return b.toString(16).padStart(2,'0'); }).join('');
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
            // Fetch Rewards Pool using the 128-char treasury address
            const rp = await fetch('/api/balance/AX5def33f67eda5560561837935709169eb17955ffe13c1f112b3a329321bef5400000000000000000000000000000000000000000000000000000000000000000');
            const dp = await rp.json();
            document.getElementById('pool-txt').innerText = dp.balance.toLocaleString() + ' AX';

            const res = await fetch('/api/chain');
            const chain = await res.json();
        }

        async function mine() {
            if(!session) return alert('Sync first');
            const r = await fetch('/api/mine?address=' + session.pub);
            if(r.ok) { alert('Mined!'); load(); } else { alert('Mempool or Treasury issue.'); }
        }

        async function send() {
            const tx = { sender: session.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            const r = await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            if(r.ok) { alert('Broadcasted!'); nav('dash'); load(); } else { alert('Error: Low balance.'); }
        }

        async function gen() {
            const p = btoa(Math.random().toString() + Date.now()).substring(0,64);
            const pb = await derive(p);
            document.getElementById('g-res').style.display = 'block';
            document.getElementById('g-priv').innerText = p;
            document.getElementById('g-pub').innerText = pb;
        }

        load(); setInterval(load, 15000);
    </script>
</body>
</html>
