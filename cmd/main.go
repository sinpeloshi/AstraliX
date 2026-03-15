package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"context"

	"astralix/core"
	_ "github.com/lib/pq"
)

var Blockchain []core.Block
var Mempool []core.Transaction
var db *sql.DB

const TREASURY_POOL_ADDR = "AXf7ca3d5889ed99de642913af6c5630d6c491732b44180771cba042a4eb5a7109cc3ccde9e1a24d5315947415d5e592123ab90edcc4ea85415c1747fbe1684158"

func initDB() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" { log.Fatal("❌ ERROR: DATABASE_URL vacía.") }
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil { log.Fatal("❌ ERROR de Driver:", err) }
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil { log.Fatal("❌ ERROR de conexión a la DB:", err) }
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS chain_state (id INT PRIMARY KEY, data TEXT)`)
	if err != nil { log.Fatal("❌ ERROR creando tablas:", err) }
}

func loadChain() {
	var data string
	err := db.QueryRow("SELECT data FROM chain_state WHERE id = 1").Scan(&data)
	if err == nil { json.Unmarshal([]byte(data), &Blockchain) }
}

func saveChain() {
	data, _ := json.Marshal(Blockchain)
	db.Exec(`INSERT INTO chain_state (id, data) VALUES (1, $1) ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data`, string(data))
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
	initDB(); loadChain()
	const Difficulty = 4 
	rootAddr := "AX5eaba583bf646e0e39f41da6f9d8fa6db929c2e858bd32dffe6ac0cee2e3e974dc3ff66cb2d73bdabdc9a49279bea46da35d10d925aaf71416e5e351a3f74b56"
	if len(Blockchain) == 0 {
		genesisTx := core.Transaction{Sender: "SYSTEM", Recipient: rootAddr, Amount: 1000002021}
		genesisTx.TxID = genesisTx.CalculateHash()
		genesisBlock := core.Block{Index: 0, Timestamp: 1773561600, Transactions: []core.Transaction{genesisTx}, PrevHash: strings.Repeat("0", 128), Difficulty: Difficulty}
		genesisBlock.Mine(); Blockchain = append(Blockchain, genesisBlock); saveChain()
	}

	http.HandleFunc("/api/balance/", func(w http.ResponseWriter, r *http.Request) {
		addr := strings.TrimPrefix(r.URL.Path, "/api/balance/")
		json.NewEncoder(w).Encode(map[string]interface{}{"balance": getBalance(addr)})
	})
	http.HandleFunc("/api/mine", func(w http.ResponseWriter, r *http.Request) {
		miner := r.URL.Query().Get("address")
		if miner == "" || len(Mempool) == 0 { http.Error(w, "Mempool empty", 400); return }
		reward := 50.0
		if getBalance(TREASURY_POOL_ADDR) >= reward {
			rewardTx := core.Transaction{Sender: TREASURY_POOL_ADDR, Recipient: miner, Amount: reward}
			rewardTx.TxID = rewardTx.CalculateHash()
			Mempool = append(Mempool, rewardTx)
		}
		prev := Blockchain[len(Blockchain)-1]
		newBlock := core.Block{Index: int64(len(Blockchain)), Timestamp: time.Now().Unix(), Transactions: Mempool, PrevHash: prev.Hash, Difficulty: Difficulty}
		newBlock.Mine(); Blockchain = append(Blockchain, newBlock); Mempool = []core.Transaction{}; saveChain()
		json.NewEncoder(w).Encode(newBlock)
	})
	http.HandleFunc("/api/transactions/new", func(w http.ResponseWriter, r *http.Request) {
		var tx core.Transaction
		json.NewDecoder(r.Body).Decode(&tx)
		if getBalance(tx.Sender) < tx.Amount && tx.Sender != "SYSTEM" { http.Error(w, "Saldo insuficiente", 400); return }
		tx.TxID = tx.CalculateHash(); Mempool = append(Mempool, tx); w.WriteHeader(201)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8"); fmt.Fprint(w, dashboardHTML)
	})
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>AstraliX Core | OS</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Outfit:wght@400;600;800&display=swap');
        @import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400&display=swap');
        :root { --bg: #F4F7F9; --card: #FFFFFF; --primary: #0D6EFD; --text: #1E293B; }
        body { background: var(--bg); font-family: 'Outfit', sans-serif; margin: 0; padding-bottom: 110px; color: var(--text); -webkit-font-smoothing: antialiased; }
        .header-ax { padding: 30px 20px 10px; text-align: center; }
        .header-ax h5 { font-weight: 800; letter-spacing: 1.5px; margin: 0; color: #0F172A; font-size: 1.4rem; }
        .status-text { font-size: 0.7rem; font-weight: 600; color: #10B981; letter-spacing: 1.2px; text-transform: uppercase; }
        .status-dot { height: 8px; width: 8px; background-color: #10B981; border-radius: 50%; display: inline-block; margin-right: 5px; box-shadow: 0 0 8px #10B981; }
        .view-ax { display: none; flex-direction: column; align-items: center; width: 100%; max-width: 500px; margin: 0 auto; }
        .card-ax { background: var(--card); border-radius: 28px; box-shadow: 0 12px 35px rgba(0,0,0,0.03); padding: 25px; margin: 15px 20px; width: calc(100% - 40px); box-sizing: border-box; border: 1px solid #E2E8F0; }
        .card-dark { background: linear-gradient(145deg, #0B1120 0%, #1E293B 100%); color: white; border: none; }
        .balance-label { font-size: 0.7rem; text-transform: uppercase; letter-spacing: 2px; opacity: 0.6; font-weight: 600; display: block; margin-bottom: 5px; }
        .balance-amount { font-size: 2.2rem; font-weight: 800; margin: 5px 0 20px; letter-spacing: -1px; }
        .pill-address { background: rgba(0,0,0,0.15); padding: 12px; border-radius: 14px; font-family: 'JetBrains Mono', monospace; font-size: 0.55rem; word-break: break-all; color: rgba(255,255,255,0.4); }
        .seed-box { background: #F8FAFC; border: 2px dashed #CBD5E1; padding: 15px; border-radius: 18px; display: grid; grid-template-columns: 1fr 1fr; gap: 8px; margin-top: 15px; }
        .seed-word { font-size: 0.75rem; color: #475569; font-weight: 600; }
        .btn-ax { background: var(--primary); color: white; border-radius: 18px; padding: 18px; font-weight: 600; border: none; width: calc(100% - 40px); margin: 10px 20px; font-size: 1rem; box-shadow: 0 8px 20px rgba(13, 110, 253, 0.25); cursor: pointer; }
        .btn-outline { background: transparent; border: 2px solid #E2E8F0; color: var(--text); box-shadow: none; margin-top: 5px; }
        .bottom-bar { background: rgba(255,255,255,0.85); backdrop-filter: blur(12px); position: fixed; bottom: 0; width: 100%; height: 85px; display: flex; justify-content: space-around; align-items: center; border-top: 1px solid rgba(0,0,0,0.05); z-index: 999; }
        .nav-link-ax { color: #94A3B8; text-align: center; text-decoration: none; flex: 1; font-size: 10px; font-weight: 600; display: flex; flex-direction: column; align-items: center; gap: 6px; cursor: pointer; }
        .nav-link-ax.active { color: var(--primary); }
        input.form-control { width: 100%; background: #F8FAFC; border: 1px solid #E2E8F0; border-radius: 14px; padding: 16px; font-size: 0.9rem; box-sizing: border-box; }
        .form-label { font-size: 0.75rem; font-weight: 600; color: #64748B; margin-bottom: 8px; display: block; }
    </style>
</head>
<body>
    <div class="header-ax">
        <h5>AstraliX Core</h5>
        <div><span class="status-dot"></span><span class="status-text">NODE SYNCHRONIZED</span></div>
    </div>

    <div id="v-dash" class="view-ax" style="display:flex;">
        <div class="card-ax card-dark text-center">
            <span class="balance-label">Personal Balance</span>
            <div id="bal-txt" class="balance-amount">0.00 AX</div>
            <div id="addr-txt" class="pill-address text-center">Wallet Not Synced</div>
        </div>
        <div class="card-ax text-center">
            <span class="balance-label">Treasury Rewards Pool</span>
            <div id="pool-txt" class="balance-amount" style="color:var(--primary)">0.00 AX</div>
            <div class="pill-address" style="background:#F8FAFC; color:#94A3B8;">AXf7ca3d5889ed99de642913af6c5630d6c491732...</div>
        </div>
        <button class="btn-ax" onclick="mine()">VALIDATE NETWORK (+50.00 AX)</button>
    </div>

    <div id="v-wallet" class="view-ax">
        <div class="card-ax">
            <span class="form-label">Recipient Address</span>
            <input type="text" id="tx-to" class="form-control mb-3" placeholder="AX...">
            <span class="form-label">Amount (AX)</span>
            <input type="number" id="tx-amt" class="form-control mb-4" placeholder="0.00">
            <button class="btn-ax" style="width:100%; margin:0;" onclick="send()">CONFIRM TRANSFER</button>
        </div>
    </div>

    <div id="v-sec" class="view-ax">
        <div class="card-ax">
            <span class="form-label">Sync with Seed Phrase (24 words)</span>
            <textarea id="i-seed" class="form-control mb-3" rows="3" placeholder="word1 word2 ..."></textarea>
            <button class="btn-ax" style="width:100%; margin:0;" onclick="login()">RESTORE WALLET</button>
            <hr style="border:0; border-top:1px solid #EEE; margin:20px 0;">
            <button class="btn-ax btn-outline" style="width:100%; margin:0;" onclick="gen()">CREATE NEW 512-BIT WALLET</button>
            
            <div id="g-res" style="display:none; margin-top:20px;">
                <span class="form-label text-danger">⚠️ WRITE DOWN YOUR SEED PHRASE:</span>
                <div class="seed-box" id="g-seed"></div>
                <div class="mt-3">
                    <span class="form-label text-primary">Your Public Address:</span>
                    <div class="pill-address text-primary" style="background:rgba(13,110,253,0.05);" id="g-pub"></div>
                </div>
            </div>
        </div>
    </div>

    <div class="bottom-bar">
        <a class="nav-link-ax active" id="n-dash" onclick="nav('dash')"><i class="fas fa-chart-pie"></i>Overview</a>
        <a class="nav-link-ax" id="n-wallet" onclick="nav('wallet')"><i class="fas fa-paper-plane"></i>Transfer</a>
        <a class="nav-link-ax" id="n-sec" onclick="nav('sec')"><i class="fas fa-shield-halved"></i>Vault</a>
    </div>

    <script>
        // Mini Dictionary for 512-bit Derivation
        const words = ["alpha","bravo","cipher","delta","echo","falcon","ghost","hazard","iron","joker","knight","lunar","matrix","nexus","omega","phantom","quantum","radar","sigma","titan","ultra","vector","wolf","xray","yield","zenith","astral","block","chain","data","edge","fiber","grid","hash","index","joint","kern","link","mine","node","open","peer","root","seed","tech","unit","vault","web","zone"];
        
        async function derive(seed) {
            const buf = new TextEncoder().encode(seed);
            const hash = await crypto.subtle.digest("SHA-512", buf);
            const hex = Array.from(new Uint8Array(hash)).map(b => b.toString(16).padStart(2,"0")).join("");
            return { priv: btoa(hex).substring(0,88), pub: "AX" + hex };
        }

        let session = JSON.parse(localStorage.getItem("ax_v18_session")) || null;
        function nav(id) {
            document.querySelectorAll(".view-ax").forEach(v => v.style.display = "none");
            document.getElementById("v-" + id).style.display = "flex";
            document.querySelectorAll(".nav-link-ax").forEach(n => n.classList.remove("active"));
            document.getElementById("n-" + id).classList.add("active");
        }

        async function login() {
            const s = document.getElementById("i-seed").value.trim().toLowerCase();
            if(!s) return;
            const keys = await derive(s);
            session = { pub: keys.pub, priv: keys.priv, seed: s };
            localStorage.setItem("ax_v18_session", JSON.stringify(session));
            location.reload();
        }

        async function gen() {
            let seed = [];
            for(let i=0; i<24; i++) seed.push(words[Math.floor(Math.random()*words.length)]);
            const seedStr = seed.join(" ");
            const keys = await derive(seedStr);
            
            document.getElementById("g-res").style.display = "block";
            document.getElementById("g-seed").innerHTML = seed.map((w, i) => '<div class="seed-word">'+(i+1)+'. '+w+'</div>').join("");
            document.getElementById("g-pub").innerText = keys.pub;
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
            if(r.ok) { alert("¡Mined!"); load(); }
        }

        async function send() {
            const tx = { sender: session.pub, recipient: document.getElementById("tx-to").value, amount: parseFloat(document.getElementById("tx-amt").value) };
            const r = await fetch("/api/transactions/new", { method: "POST", body: JSON.stringify(tx) });
            if(r.ok) { alert("¡Sent!"); nav('dash'); load(); }
        }

        load(); setInterval(load, 15000);
    </script>
</body>
</html>
`