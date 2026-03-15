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
const TREASURY_POOL_ADDR = "AXf7ca3d5889ed99de642913af6c5630d6c491732b44180771cba042a4eb5a7109cc3ccde9e1a24d5315947415d5e592123ab90edcc4ea85415c1747fbe1684158"

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
	rootAddr := "AX5eaba583bf646e0e39f41da6f9d8fa6db929c2e858bd32dffe6ac0cee2e3e974dc3ff66cb2d73bdabdc9a49279bea46da35d10d925aaf71416e5e351a3f74b56"

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

	http.HandleFunc("/api/chain", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Blockchain)
	})

	http.HandleFunc("/api/mine", func(w http.ResponseWriter, r *http.Request) {
		miner := r.URL.Query().Get("address")
		if miner == "" || len(Mempool) == 0 { 
			http.Error(w, "Mempool empty", 400); return 
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
		if tx.Sender != "SYSTEM" && tx.Sender != TREASURY_POOL_ADDR && getBalance(tx.Sender) < tx.Amount {
			http.Error(w, "Low balance", 400); return
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
    <title>AX Core | OS</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Outfit:wght@400;600;800&display=swap');
        @import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400&display=swap');
        
        :root { --bg: #F4F7F9; --card: #FFFFFF; --primary: #0D6EFD; --text: #1E293B; }
        body { background: var(--bg); font-family: 'Outfit', sans-serif; margin: 0; padding-bottom: 100px; color: var(--text); -webkit-font-smoothing: antialiased; }
        
        .header-ax { padding: 30px 20px 10px; text-align: center; }
        .header-ax h5 { font-weight: 800; letter-spacing: 1.5px; margin: 0; color: #0F172A; font-size: 1.2rem; }
        
        .view-ax { display: flex; flex-direction: column; align-items: center; width: 100%; max-width: 500px; margin: 0 auto; }
        
        .card-ax { background: var(--card); border-radius: 28px; box-shadow: 0 12px 35px rgba(0,0,0,0.03); padding: 30px; margin: 15px 20px; width: calc(100% - 40px); box-sizing: border-box; }
        .card-dark { background: linear-gradient(145deg, #0B1120 0%, #1E293B 100%); color: white; border: 1px solid rgba(255,255,255,0.05); }
        .card-light { background: #FFFFFF; border: 1px solid #E2E8F0; }
        
        .balance-label { font-size: 0.7rem; text-transform: uppercase; letter-spacing: 2px; opacity: 0.6; font-weight: 600; display: block; margin-bottom: 5px; }
        .balance-amount { font-size: 2.2rem; font-weight: 800; margin: 5px 0 20px; letter-spacing: -1px; }
        .text-primary { color: var(--primary); }
        
        .pill-address { background: rgba(0,0,0,0.15); padding: 12px 14px; border-radius: 14px; font-family: 'JetBrains Mono', monospace; font-size: 0.55rem; word-break: break-all; color: rgba(255,255,255,0.5); line-height: 1.5; text-align: left; }
        .card-light .pill-address { background: #F8FAFC; color: #94A3B8; border: 1px solid #F1F5F9; }
        
        .btn-ax { background: var(--primary); color: white; border-radius: 18px; padding: 18px; font-weight: 600; border: none; width: calc(100% - 40px); margin: 10px 20px; font-size: 1rem; box-shadow: 0 8px 20px rgba(13, 110, 253, 0.25); transition: 0.2s; cursor: pointer; font-family: 'Outfit', sans-serif; }
        .btn-ax:active { transform: translateY(2px); box-shadow: 0 2px 10px rgba(13, 110, 253, 0.2); }
        .btn-outline { background: transparent; border: 2px solid #E2E8F0; color: var(--text); box-shadow: none; margin-top: 5px; }
        
        .bottom-bar { background: rgba(255,255,255,0.85); backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px); position: fixed; bottom: 0; width: 100%; height: 85px; display: flex; justify-content: space-around; align-items: center; border-top: 1px solid rgba(0,0,0,0.05); z-index: 999; }
        .nav-link-ax { color: #94A3B8; text-align: center; text-decoration: none; flex: 1; font-size: 10px; font-weight: 600; cursor: pointer; display: flex; flex-direction: column; align-items: center; gap: 6px; transition: color 0.2s; }
        .nav-link-ax.active { color: var(--primary); }
        .nav-link-ax i { font-size: 22px; }
        
        .form-group { text-align: left; margin-bottom: 20px; }
        .form-label { font-size: 0.75rem; font-weight: 600; color: #64748B; margin-bottom: 8px; display: block; }
        input.form-control { width: 100%; background: #F8FAFC; border: 1px solid #E2E8F0; border-radius: 14px; padding: 16px; font-size: 0.9rem; font-family: 'JetBrains Mono', monospace; box-sizing: border-box; color: var(--text); transition: border-color 0.2s; outline: none; }
        input.form-control:focus { border-color: var(--primary); background: #FFFFFF; }
        
        hr { border: 0; height: 1px; background: #E2E8F0; margin: 25px 0; }
    </style>
</head>
<body>

    <div class="header-ax">
        <h5>AX CORE</h5>
    </div>

    <div id="v-dash" class="view-ax">
        <div class="card-ax card-dark text-center">
            <span class="balance-label">Personal Balance</span>
            <div id="bal-txt" class="balance-amount">0.00 AX</div>
            <div id="addr-txt" class="pill-address text-center">Wallet Not Synced</div>
        </div>

        <div class="card-ax card-light text-center">
            <span class="balance-label">Treasury Rewards Pool</span>
            <div id="pool-txt" class="balance-amount text-primary">0.00 AX</div>
            <div class="pill-address">AXf7ca3d5889ed99de642913af6c5630d6c491732b44180771cba042a4eb5a7109cc3ccde9e1a24d5315947415d5e592123ab90edcc4ea85415c1747fbe1684158</div>
        </div>

        <button class="btn-ax" onclick="mine()">VALIDATE NETWORK (+50.00 AX)</button>
    </div>

    <div id="v-wallet" class="view-ax" style="display:none">
        <div class="card-ax card-light">
            <div class="form-group">
                <span class="form-label">Recipient Address (128 chars)</span>
                <input type="text" id="tx-to" class="form-control" placeholder="AX...">
            </div>
            <div class="form-group">
                <span class="form-label">Amount (AX)</span>
                <input type="number" id="tx-amt" class="form-control" style="font-family: 'Outfit', sans-serif;" placeholder="0.00">
            </div>
        </div>
        <button class="btn-ax" onclick="send()">CONFIRM TRANSFER</button>
    </div>

    <div id="v-sec" class="view-ax" style="display:none">
        <div class="card-ax card-light">
            <div class="form-group">
                <span class="form-label">Sync Identity</span>
                <input type="password" id="i-priv" class="form-control" placeholder="Enter Private Key">
            </div>
            <button class="btn-ax" style="width:100%; margin:0;" onclick="login()">CONNECT WALLET</button>
            
            <hr>
            
            <button class="btn-ax btn-outline" style="width:100%; margin:0;" onclick="gen()">GENERATE NEW KEYS</button>
            <div id="g-res" class="mt-4 text-left" style="display:none">
                <span class="form-label">Private Key (Save securely):</span>
                <div class="pill-address mb-3" id="g-priv" style="background: #F1F5F9; color: #334155; border: none;"></div>
                <span class="form-label text-primary">Public Address:</span>
                <div class="pill-address text-primary" id="g-pub" style="background: rgba(13, 110, 253, 0.05); border: 1px solid rgba(13, 110, 253, 0.1);"></div>
            </div>
        </div>
    </div>

    <div class="bottom-bar">
        <a class="nav-link-ax active" id="n-dash" onclick="nav('dash')"><i class="fas fa-chart-pie"></i>Overview</a>
        <a class="nav-link-ax" id="n-wallet" onclick="nav('wallet')"><i class="fas fa-paper-plane"></i>Transfer</a>
        <a class="nav-link-ax" id="n-sec" onclick="nav('sec')"><i class="fas fa-shield-halved"></i>Vault</a>
    </div>

    <script>
        async function derive(priv) {
            const buf = new TextEncoder().encode(priv);
            const hash = await crypto.subtle.digest("SHA-512", buf);
            const hex = Array.from(new Uint8Array(hash)).map(b => b.toString(16).padStart(2,"0")).join("");
            return "AX" + hex;
        }
        let session = JSON.parse(localStorage.getItem("ax_v18_session")) || null;
        function nav(id) {
            document.querySelectorAll(".view-ax").forEach(v => v.style.display = "none");
            document.getElementById("v-" + id).style.display = "flex";
            document.querySelectorAll(".nav-link-ax").forEach(n => n.classList.remove("active"));
            document.getElementById("n-" + id).classList.add("active");
            window.scrollTo(0,0);
        }
        async function login() {
            const p = document.getElementById("i-priv").value;
            if(!p) return;
            const pb = await derive(p);
            session = { pub: pb, priv: p };
            localStorage.setItem("ax_v18_session", JSON.stringify(session));
            location.reload();
        }
        async function load() {
            if(session) {
                const r = await fetch("/api/balance/" + session.pub);
                const d = await r.json();
                document.getElementById("bal-txt").innerText = d.balance.toLocaleString() + " AX";
                document.getElementById("addr-txt").innerText = session.pub;
            }
            const rp = await fetch("/api/balance/AXf7ca3d5889ed99de642913af6c5630d6c491732b44180771cba042a4eb5a7109cc3ccde9e1a24d5315947415d5e592123ab90edcc4ea85415c1747fbe1684158");
            const dp = await rp.json();
            document.getElementById("pool-txt").innerText = dp.balance.toLocaleString() + " AX";
        }
        async function mine() {
            if(!session) return alert("Sync required");
            const r = await fetch("/api/mine?address=" + session.pub);
            if(r.ok) { alert("Block Mined!"); load(); } else { alert("Error: Mempool or Treasury."); }
        }
        async function send() {
            const tx = { sender: session.pub, recipient: document.getElementById("tx-to").value, amount: parseFloat(document.getElementById("tx-amt").value) };
            const r = await fetch("/api/transactions/new", { method: "POST", body: JSON.stringify(tx) });
            if(r.ok) { alert("Sent!"); nav("dash"); load(); } else { alert("Check funds."); }
        }
        async function gen() {
            const array = new Uint8Array(64);
            window.crypto.getRandomValues(array);
            const p = btoa(String.fromCharCode.apply(null, array));
            const pb = await derive(p);
            document.getElementById("g-res").style.display = "block";
            document.getElementById("g-priv").innerText = p;
            document.getElementById("g-pub").innerText = pb;
        }
        load(); setInterval(load, 15000);
    </script>
</body>
</html>
`
