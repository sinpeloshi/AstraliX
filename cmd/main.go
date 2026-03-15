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
        
        :root { --bg: #F8FAFC; --card: #FFFFFF; --primary: #0D6EFD; --text: #0F172A; }
        body { background: var(--bg); font-family: 'Outfit', sans-serif; margin: 0; padding-bottom: 110px; color: var(--text); -webkit-font-smoothing: antialiased; }
        
        .header-ax { padding: 35px 20px 15px; text-align: center; }
        .header-ax h5 { font-weight: 800; letter-spacing: 1px; margin: 0; color: #0F172A; font-size: 1.5rem; }
        .status-box { display: inline-flex; align-items: center; background: #ECFDF5; padding: 6px 12px; border-radius: 100px; margin-top: 10px; }
        .status-dot { height: 7px; width: 7px; background: #10B981; border-radius: 50%; margin-right: 8px; box-shadow: 0 0 10px #10B981; }
        .status-text { font-size: 0.65rem; font-weight: 800; color: #059669; letter-spacing: 1px; }

        .view-ax { display: none; flex-direction: column; align-items: center; width: 100%; max-width: 500px; margin: 0 auto; }
        
        .card-ax { background: var(--card); border-radius: 32px; box-shadow: 0 10px 40px rgba(0,0,0,0.04); padding: 30px; margin: 15px 20px; width: calc(100% - 40px); box-sizing: border-box; border: 1px solid rgba(0,0,0,0.03); }
        .card-dark { background: linear-gradient(145deg, #0F172A 0%, #1E293B 100%); color: white; border: none; }
        
        .balance-label { font-size: 0.75rem; text-transform: uppercase; letter-spacing: 2px; opacity: 0.5; font-weight: 600; }
        .balance-amount { font-size: 2.4rem; font-weight: 800; margin: 10px 0 25px; letter-spacing: -1.5px; }
        .pill-address { background: rgba(255,255,255,0.1); padding: 14px; border-radius: 16px; font-family: 'JetBrains Mono', monospace; font-size: 0.55rem; word-break: break-all; color: rgba(255,255,255,0.45); }
        
        .seed-container { background: #F1F5F9; border-radius: 20px; padding: 20px; display: grid; grid-template-columns: 1fr 1fr; gap: 10px; margin-top: 20px; }
        .seed-word { font-size: 0.8rem; background: white; padding: 10px 12px; border-radius: 12px; color: #475569; font-weight: 600; display: flex; align-items: center; border: 1px solid #E2E8F0; }
        .seed-num { color: #94A3B8; font-size: 0.65rem; margin-right: 8px; width: 18px; }

        .btn-ax { background: var(--primary); color: white; border-radius: 20px; padding: 20px; font-weight: 700; border: none; width: calc(100% - 40px); margin: 10px 20px; font-size: 1rem; box-shadow: 0 10px 25px rgba(13, 110, 253, 0.2); cursor: pointer; transition: 0.3s; }
        .btn-outline { background: #F8FAFC; border: 2px solid #E2E8F0; color: #475569; box-shadow: none; margin-top: 5px; }
        
        textarea.form-control { width: 100%; background: #F8FAFC; border: 2px solid #E2E8F0; border-radius: 18px; padding: 18px; font-size: 0.95rem; font-family: 'Outfit', sans-serif; resize: none; color: #1E293B; outline: none; transition: border-color 0.3s; }
        textarea.form-control:focus { border-color: var(--primary); background: #FFFFFF; }

        .bottom-bar { background: rgba(255,255,255,0.9); backdrop-filter: blur(15px); position: fixed; bottom: 0; width: 100%; height: 90px; display: flex; justify-content: space-around; align-items: center; border-top: 1px solid #F1F5F9; z-index: 999; }
        .nav-link-ax { color: #94A3B8; text-decoration: none; flex: 1; font-size: 11px; font-weight: 700; display: flex; flex-direction: column; align-items: center; gap: 6px; cursor: pointer; }
        .nav-link-ax.active { color: var(--primary); }
    </style>
</head>
<body>
    <div class="header-ax">
        <h5>AstraliX Core</h5>
        <div class="status-box"><span class="status-dot"></span><span class="status-text">NODE SYNCHRONIZED</span></div>
    </div>

    <div id="v-dash" class="view-ax" style="display:flex;">
        <div class="card-ax card-dark text-center">
            <span class="balance-label">Total Balance</span>
            <div id="bal-txt" class="balance-amount">0.00 AX</div>
            <div id="addr-txt" class="pill-address text-center">Wallet Not Connected</div>
        </div>
        <div class="card-ax text-center" style="background:#FFFFFF;">
            <span class="balance-label">Treasury Pool</span>
            <div id="pool-txt" class="balance-amount" style="color:var(--primary); font-size:1.8rem;">0.00 AX</div>
            <div class="pill-address" style="background:#F8FAFC; color:#94A3B8; border: 1px solid #F1F5F9;">AXf7ca3d5889ed99de642913af6...</div>
        </div>
        <button class="btn-ax" onclick="mine()">VALIDATE NETWORK (+50.00 AX)</button>
    </div>

    <div id="v-wallet" class="view-ax">
        <div class="card-ax">
            <span class="balance-label" style="display:block; margin-bottom:15px;">Send Assets</span>
            <input type="text" id="tx-to" class="form-control mb-3" style="width:100%; padding:18px; border-radius:18px; border:2px solid #E2E8F0; background:#F8FAFC;" placeholder="Destination Address (AX...)">
            <input type="number" id="tx-amt" class="form-control mb-4" style="width:100%; padding:18px; border-radius:18px; border:2px solid #E2E8F0; background:#F8FAFC;" placeholder="Amount AX">
            <button class="btn-ax" style="width:100%; margin:0;" onclick="send()">CONFIRM TRANSFER</button>
        </div>
    </div>

    <div id="v-sec" class="view-ax">
        <div class="card-ax">
            <span class="balance-label" style="display:block; margin-bottom:10px;">Access your Vault</span>
            <textarea id="i-seed" class="form-control mb-3" rows="3" placeholder="Input your recovery phrase (24 words)..."></textarea>
            <button class="btn-ax" style="width:100%; margin:0;" onclick="login()">RESTORE WALLET</button>
            
            <div style="margin: 30px 0; display: flex; align-items: center; opacity: 0.3;"><hr style="flex:1;"><span style="margin:0 15px; font-weight:800; font-size:0.7rem;">OR</span><hr style="flex:1;"></div>
            
            <button class="btn-ax btn-outline" style="width:100%; margin:0;" onclick="gen()">CREATE NEW 512-BIT IDENTITY</button>
            
            <div id="g-res" style="display:none; margin-top:30px;">
                <span style="font-size:0.75rem; font-weight:800; color:#EF4444; text-transform:uppercase; letter-spacing:1px;"><i class="fas fa-exclamation-triangle"></i> Security Phrase</span>
                <p style="font-size:0.8rem; color:#64748B; margin:5px 0 20px;">Save these 24 words in order. Never share them.</p>
                <div class="seed-container" id="g-seed"></div>
                <div class="mt-4">
                    <span class="balance-label">Public Address</span>
                    <div class="pill-address" style="background:#F1F5F9; color:var(--primary); margin-top:10px; font-weight:700;" id="g-pub"></div>
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
            document.getElementById("g-seed").innerHTML = seed.map((w, i) => '<div class="seed-word"><span class="seed-num">'+(i+1)+'</span>'+w+'</div>').join("");
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
            if(r.ok) { alert("¡Network Validated!"); load(); }
        }
        async function send() {
            const tx = { sender: session.pub, recipient: document.getElementById("tx-to").value, amount: parseFloat(document.getElementById("tx-amt").value) };
            const r = await fetch("/api/transactions/new", { method: "POST", body: JSON.stringify(tx) });
            if(r.ok) { alert("¡Transaction Sent!"); nav('dash'); load(); }
        }
        load(); setInterval(load, 15000);
    </script>
</body>
</html>
`