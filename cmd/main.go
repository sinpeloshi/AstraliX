package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"astralix/core"
	_ "github.com/lib/pq"
)

var Blockchain []core.Block
var Mempool []core.Transaction
var db *sql.DB

const TREASURY_POOL_ADDR = "AXf7ca3d5889ed99de642913af6c5630d6c491732b44180771cba042a4eb5a7109cc3ccde9e1a24d5315947415d5e592123ab90edcc4ea85415c1747fbe1684158"

// ==========================================
// ⚙️ MOTOR BLOCKCHAIN
// ==========================================

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
	_, err := db.Exec(`INSERT INTO chain_state (id, data) VALUES (1, $1) ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data`, string(data))
	if err != nil { log.Println("❌ Error guardando:", err) }
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
	initDB()
	loadChain()
	const Difficulty = 4
	rootAddr := "AXec99e78875c95208706ae0be9b90ca7774bdbf458ebefc4307b66d5426385aefc91b072a68e6d567cfb371d01892d892e51c82113de5644ba4f6a973b7db345d"

	if len(Blockchain) == 0 {
		genesisTx := core.Transaction{Sender: "SYSTEM", Recipient: rootAddr, Amount: 1000002021}
		genesisTx.TxID = genesisTx.CalculateHash()
		genesisBlock := core.Block{Index: 0, Timestamp: 1773561600, Transactions: []core.Transaction{genesisTx}, PrevHash: strings.Repeat("0", 128), Difficulty: Difficulty}
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
		if miner == "" || len(Mempool) == 0 { http.Error(w, "Error", 400); return }
		reward := 50.0
		txs := append(Mempool, core.Transaction{Sender: TREASURY_POOL_ADDR, Recipient: miner, Amount: reward})
		prev := Blockchain[len(Blockchain)-1]
		newBlock := core.Block{Index: int64(len(Blockchain)), Timestamp: time.Now().Unix(), Transactions: txs, PrevHash: prev.Hash, Difficulty: Difficulty}
		newBlock.Mine()
		Blockchain = append(Blockchain, newBlock)
		Mempool = []core.Transaction{}
		saveChain()
		json.NewEncoder(w).Encode(newBlock)
	})
	http.HandleFunc("/api/transactions/new", func(w http.ResponseWriter, r *http.Request) {
		var tx core.Transaction
		json.NewDecoder(r.Body).Decode(&tx)
		if getBalance(tx.Sender) < tx.Amount && tx.Sender != "SYSTEM" { http.Error(w, "Low balance", 400); return }
		tx.TxID = tx.CalculateHash()
		Mempool = append(Mempool, tx)
		w.WriteHeader(201)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" { http.Redirect(w, r, "/", http.StatusSeeOther); return }
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, landingHTML)
	})
	http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, dashboardHTML)
	})
	http.HandleFunc("/whitepaper", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, whitepaperHTML)
	})

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}

// ==========================================
// 🎨 LANDING PAGE (SILICON VALLEY STYLE)
// ==========================================

const landingHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <title>AstraliX | The 512-bit Layer 1 Protocol</title>
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;600;800&family=JetBrains+Mono:wght@400;700&display=swap');
        :root { --bg: #020202; --bg-card: #080808; --prim: #3B82F6; --txt: #FFFFFF; --txt-m: #8899A6; --brd: #1A1A1A; --acc: #10B981; }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: 'Plus Jakarta Sans', sans-serif; background: var(--bg); color: var(--txt); line-height: 1.5; overflow-x: hidden; -webkit-font-smoothing: antialiased; scroll-behavior: smooth; }
        .bg-p { position: fixed; width: 100vw; height: 100vh; background-image: radial-gradient(circle at 1px 1px, #111 1px, transparent 0); background-size: 40px 40px; z-index: -1; }
        .nav { padding: 25px 6%; display: flex; justify-content: space-between; align-items: center; position: sticky; top: 0; background: rgba(2,2,2,0.8); backdrop-filter: blur(20px); z-index: 100; border-bottom: 1px solid var(--brd); }
        .logo { font-weight: 800; font-size: 1.8rem; letter-spacing: -1.5px; color: var(--txt); text-decoration: none; }
        .logo span { color: var(--prim); }
        .btn-core-nav { background: var(--prim); color: white !important; padding: 10px 22px; border-radius: 100px; font-size: 0.75rem; font-weight: 800; text-decoration: none; transition: 0.3s; }
        .hero { text-align: center; padding: 100px 6% 80px; max-width: 1200px; margin: 0 auto; }
        .hero h1 { font-size: clamp(3.2rem, 9vw, 6.2rem); font-weight: 800; letter-spacing: -4px; line-height: 1.1; margin-bottom: 25px; background: linear-gradient(180deg, #FFF 30%, #555 100%); -webkit-background-clip: text; -webkit-text-fill-color: transparent; padding-bottom: 10px; }
        .hero p { font-size: clamp(1.1rem, 2.5vw, 1.4rem); color: var(--txt-m); max-width: 750px; margin: 0 auto 50px; }
        .hero-btns { display: flex; gap: 15px; justify-content: center; flex-wrap: wrap; }
        .btn-p { padding: 20px 45px; border-radius: 100px; font-weight: 700; text-decoration: none; font-size: 1rem; transition: 0.3s; }
        .btn-white { background: #FFF; color: #000; }
        .btn-dark { border: 1px solid var(--brd); color: #FFF; background: rgba(255,255,255,0.03); }
        .sec-q { display: flex; gap: 2px; max-width: 1200px; margin: 80px auto; padding: 0 6%; }
        .q-box { flex: 1; background: var(--bg-card); border: 1px solid var(--brd); padding: 50px 40px; text-align: left; }
        .q-box h3 { font-size: 0.8rem; text-transform: uppercase; letter-spacing: 2px; color: var(--txt-m); margin-bottom: 20px; }
        .q-box .val { font-family: 'JetBrains Mono'; font-size: 2.2rem; font-weight: 800; margin-bottom: 10px; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 2px; max-width: 1200px; margin: 0 auto 100px; padding: 0 6%; }
        .card { background: var(--bg-card); border: 1px solid var(--brd); padding: 60px 40px; text-align: left; transition: 0.3s; }
        .card:hover { border-color: var(--prim); background: #0A0A0A; }
        .card i { color: var(--prim); font-size: 2rem; margin-bottom: 25px; display: block; }
        .roadmap { max-width: 800px; margin: 100px auto; padding: 0 6%; }
        .rm-step { border-left: 1px solid #222; padding: 0 0 50px 40px; position: relative; }
        .rm-step::before { content: ''; position: absolute; left: -5px; top: 0; width: 100%; height: 2px; background: linear-gradient(90deg, var(--prim), transparent); }
        .rm-date { font-weight: 800; color: var(--prim); font-size: 0.8rem; margin-bottom: 10px; text-transform: uppercase; }
        .pre-sale { background: var(--bg-card); border-top: 1px solid var(--brd); padding: 120px 6%; text-align: center; }
        .price-tag { font-size: clamp(4rem, 12vw, 7rem); font-weight: 800; letter-spacing: -6px; margin: 20px 0; }
        .btn-buy { background: var(--acc); color: #000; padding: 22px 60px; border-radius: 100px; font-weight: 800; text-decoration: none; font-size: 1.2rem; display: inline-block; transition: 0.3s; }
        footer { padding: 100px 6% 40px; border-top: 1px solid var(--brd); display: grid; grid-template-columns: 2fr 1fr 1fr; gap: 50px; max-width: 1400px; margin: 0 auto; text-align: left; }
        .f-col h5 { margin-bottom: 25px; font-size: 0.85rem; text-transform: uppercase; letter-spacing: 1px; color: var(--txt); }
        .f-col a { display: block; color: var(--txt-m); text-decoration: none; margin-bottom: 12px; font-size: 0.9rem; }
        @media (max-width: 850px) { footer { grid-template-columns: 1fr; } .sec-q { flex-direction: column; } .nav-links { display: none; } }
    </style>
</head>
<body>
    <div class="bg-p"></div>
    <nav class="nav">
        <a href="/" class="logo"><span>A</span>strali<span>X</span></a>
        <div class="nav-links" style="display:flex; gap:30px; align-items:center;">
            <a href="/whitepaper" style="color:var(--txt-m); text-decoration:none; font-size:0.85rem; font-weight:600;">Protocol</a>
            <a href="#roadmap" style="color:var(--txt-m); text-decoration:none; font-size:0.85rem; font-weight:600;">Roadmap</a>
            <a href="/dashboard" class="btn-core-nav">ENTER DASHBOARD</a>
        </div>
    </nav>
    <header class="hero">
        <div style="background: rgba(59,130,246,0.1); color: var(--prim); padding: 8px 24px; border-radius: 100px; font-size: 0.75rem; font-weight: 800; display: inline-block; margin-bottom: 35px; border: 1px solid rgba(59,130,246,0.2);">ALPHA TESTNET LIVE & OPERATIONAL</div>
        <h1>The 512-bit Era Begins Here.</h1>
        <p>A mission-critical Layer 1 protocol doubling cryptographic security standards for the post-quantum era. Built for absolute digital sovereignty.</p>
        <div class="hero-btns">
            <a href="/dashboard" class="btn-p btn-white"><i class="fas fa-terminal"></i> Launch Dashboard</a>
            <a href="/whitepaper" class="btn-p btn-dark">Read Whitepaper</a>
        </div>
    </header>
    <section class="sec-q">
        <div class="q-box">
            <h3>Legacy Standard (BTC, ETH, SOL)</h3>
            <div class="val" style="color: #EF4444;">256-bit</div>
            <p style="color: var(--txt-m); font-size: 0.9rem; line-height:1.7;">Vulnerable to Shor's Algorithm and next-gen quantum decryption within this decade.</p>
        </div>
        <div class="q-box" style="border-left: none; background: #0A0A0A;">
            <h3>AstraliX Core Standard</h3>
            <div class="val" style="color: var(--acc);">512-bit</div>
            <p style="color: var(--txt-m); font-size: 0.9rem; line-height:1.7;">Mathematically immune to classical and quantum brute-force attacks via ultra-high entropy.</p>
        </div>
    </section>
    <main class="grid">
        <div class="card"><i class="fas fa-microchip"></i><h4>Quantum-Proof</h4><p>SHA-512 architecture provides $2^{512}$ combinations, securing assets for the next century of computing.</p></div>
        <div class="card"><i class="fas fa-bolt"></i><h4>Go-Native Engine</h4><p>Multi-threaded consensus built in Golang for sub-second block finality and massive node concurrency.</p></div>
        <div class="card"><i class="fas fa-fingerprint"></i><h4>Sovereign Vault</h4><p>Local mnemonic derivation. Your private keys never touch a server. Pure, untamperable decentralization.</p></div>
    </main>
    <section class="roadmap" id="roadmap">
        <div style="margin-bottom: 50px;"><h2 style="font-size:2.5rem; font-weight:800;">Strategic Roadmap</h2></div>
        <div class="rm-step"><div class="rm-date">Q1 2026</div><h4>Genesis Alpha</h4><p>Deployment of the core engine and Founder Node allocation program for early backers.</p></div>
        <div class="rm-step" style="border-left: 2px solid var(--prim);"><div class="rm-date" style="background: var(--prim); color: #000; display: inline-block; padding: 4px 12px; border-radius: 4px; font-weight:900;">APRIL 2026</div><h4>Mainnet Launch</h4><p>Official network transition. Token migration 1:1 and full decentralized validator onboarding.</p></div>
    </section>
    <section id="buy" class="pre-sale">
        <div style="text-transform: uppercase; letter-spacing: 4px; font-weight: 800; color: var(--prim); font-size: 0.8rem; margin-bottom:20px;">Founder Node Allocation</div>
        <div class="price-tag">21 USDT</div>
        <div style="background: #000; border: 1px solid #1A1A1A; padding: 40px; border-radius: 30px; max-width: 600px; margin: 40px auto;">
            <div style="color: #F3BA2F; font-size: 0.8rem; font-weight: 800; margin-bottom: 15px;">BINANCE SMART CHAIN (BEP-20)</div>
            <div style="font-family: 'JetBrains Mono'; font-size: 0.9rem; word-break: break-all; color:var(--txt);">0x948a663b1bd1292ded76a8412af2092bf0462d7c</div>
        </div>
        <a href="https://tally.so/r/jaxlL1" target="_blank" class="btn-buy">VERIFY TRANSACTION <i class="fas fa-arrow-right"></i></a>
    </section>
    <footer>
        <div class="f-col">
            <a href="/" class="logo" style="font-size: 1.5rem;"><span>A</span>strali<span>X</span></a>
            <p style="color: var(--txt-m); margin-top: 25px; font-size: 0.9rem; line-height:1.8;">Leading the cryptographic revolution through 512-bit security standards.</p>
            <a href="https://x.com/XAstraliX" target="_blank" style="margin-top:25px; color:#FFF; font-weight:800; display:flex; align-items:center; gap:12px; text-decoration:none;"><i class="fab fa-x-twitter" style="font-size:1.5rem;"></i> @XAstraliX</a>
        </div>
        <div class="f-col"><h5>Protocol</h5><a href="/whitepaper">Whitepaper</a><a href="#roadmap">Roadmap</a></div>
        <div class="f-col"><h5>Resources</h5><a href="/dashboard">Dashboard</a><a href="https://tally.so/r/jaxlL1">Verify Node</a></div>
    </footer>
</body>
</html>
`

// ==========================================
// 📄 WHITEPAPER (EXTENDED TECHNICAL DOC)
// ==========================================

const whitepaperHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Whitepaper | AstraliX Protocol</title>
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;600;800&family=JetBrains+Mono:wght@400&display=swap');
        :root { --bg: #020202; --prim: #3B82F6; --txt: #FFFFFF; --txt-m: #8899A6; --brd: #1A1A1A; }
        body { font-family: 'Plus Jakarta Sans', sans-serif; background: var(--bg); color: var(--txt); line-height: 1.9; }
        .container { max-width: 900px; margin: 0 auto; padding: 100px 6%; }
        h1 { font-size: clamp(3rem, 7vw, 4.5rem); font-weight: 800; letter-spacing: -4px; margin-bottom: 20px; }
        h2 { font-size: 2rem; font-weight: 800; margin: 70px 0 30px; color: var(--prim); border-bottom: 1px solid var(--brd); padding-bottom: 15px; }
        p { margin-bottom: 30px; color: #CCC; font-size: 1.15rem; }
        .tech-box { background: #080808; border: 1px solid var(--brd); padding: 40px; border-radius: 24px; margin: 50px 0; font-family: 'JetBrains Mono'; font-size: 0.9rem; color: var(--txt-m); }
        footer { text-align: center; padding: 100px 0; border-top: 1px solid var(--brd); margin-top: 120px; color: var(--txt-m); }
    </style>
</head>
<body>
    <div class="container">
        <a href="/" style="color:var(--prim); text-decoration:none; font-weight:800; font-size:0.9rem;"><i class="fas fa-arrow-left"></i> BACK TO HOME</a>
        <h1 style="margin-top:50px;">AstraliX Protocol</h1>
        <p style="font-size:0.95rem; color:var(--txt-m);">Version 1.0 (Alpha Genesis) • Lead Architect: Denis Waldemar • March 2026</p>
        
        <h2>1. The Post-Quantum Moat</h2>
        <p>Current blockchains (Bitcoin, Ethereum) rely on 256-bit ECDSA for digital signatures. While mathematically secure against classical computers, Shor's algorithm on a 4000-qubit quantum computer could theoretically solve the discrete logarithm problem in seconds. AstraliX triples the complexity by moving the core protocol to a 512-bit standard.</p>
        
        <div class="tech-box">
            // MATHEMATICAL ENTROPY COMPARISON<br>
            SHA-256 Complexity: 1.15 x 10^77 combinations<br>
            SHA-512 Complexity: 1.34 x 10^154 combinations<br>
            Status: QUANTUM IMMUNE
        </div>

        <h2>2. High-Concurrency Go Engine</h2>
        <p>The AstraliX core is built in Golang, utilizing its native "Goroutines" to handle thousands of state transitions per second without memory collisions. This allows our 512-bit hashes to be processed with the same latency as legacy 256-bit networks.</p>

        <h2>3. Mainnet Launch: April 2026</h2>
        <p>We are currently allocating 100 Founder Nodes. These nodes act as the seed for our Proof-of-Stake transition. On April 2026, the ledger will migrate to a fully decentralized Mainnet, where Founder Nodes will become the primary governance validators.</p>

        <footer>© 2026 AstraliX Foundation • Secured for the Future.</footer>
    </div>
</body>
</html>
`

// ==========================================
// 📱 DASHBOARD (OS STYLE & FULL RESPONSIVE)
// ==========================================

const dashboardHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>Core | AstraliX OS</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;600;800&family=JetBrains+Mono:wght@400&display=swap');
        :root { --bg: #020202; --card: #080808; --prim: #3B82F6; --txt: #FFFFFF; --txt-m: #8899A6; --brd: #1A1A1A; }
        body { background: var(--bg); font-family: 'Plus Jakarta Sans', sans-serif; margin: 0; padding-bottom: 100px; color: var(--txt); overflow-x: hidden; }
        .container { max-width: 500px; margin: 0 auto; padding: 0 20px; width: 100%; }
        .header-ax { padding: 40px 0 20px; text-align: center; }
        .status-box { display: inline-flex; align-items: center; background: rgba(16, 185, 129, 0.1); padding: 8px 16px; border-radius: 100px; margin-top: 12px; border: 1px solid rgba(16, 185, 129, 0.2); }
        .status-dot { height: 8px; width: 8px; background: #10B981; border-radius: 50%; margin-right: 10px; box-shadow: 0 0 12px #10B981; }
        .view-ax { display: none; flex-direction: column; width: 100%; gap: 20px; margin-top: 10px; }
        .card-ax { background: var(--card); border-radius: 28px; padding: 35px; width: 100%; border: 1px solid var(--brd); box-sizing: border-box; }
        .bal-lbl { font-size: 0.75rem; text-transform: uppercase; letter-spacing: 2px; color: var(--txt-m); font-weight: 700; margin-bottom: 12px; display: block; }
        .bal-val { font-size: 2.2rem; font-weight: 800; margin-bottom: 25px; letter-spacing: -1px; word-break: break-word; }
        .pill { background: #000; padding: 18px; border-radius: 20px; font-family: 'JetBrains Mono'; font-size: 0.6rem; word-break: break-all; color: var(--txt-m); border: 1px solid var(--brd); line-height: 1.5; width: 100%; box-sizing: border-box; }
        .btn-ax { background: var(--prim); color: white; border-radius: 20px; padding: 20px; font-weight: 800; border: none; width: 100%; font-size: 1rem; cursor: pointer; transition: 0.3s; }
        .bottom-bar { background: rgba(2,2,2,0.8); backdrop-filter: blur(20px); position: fixed; bottom: 0; width: 100%; height: 90px; display: flex; justify-content: space-around; align-items: center; border-top: 1px solid var(--brd); z-index: 1000; }
        .nav-l { color: #444; text-decoration: none; font-size: 0.65rem; font-weight: 800; display: flex; flex-direction: column; align-items: center; gap: 8px; cursor: pointer; text-transform: uppercase; }
        .nav-l.active { color: var(--prim); }
        .input-ax { width: 100%; padding: 20px; border-radius: 18px; border: 1px solid var(--brd); background: #000; color: #FFF; margin-bottom: 12px; box-sizing: border-box; font-family: inherit; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header-ax"><a href="/" style="color:white; text-decoration:none; font-weight:800; font-size:1.2rem;">AstraliX Core</a><br><div class="status-box"><span class="status-dot"></span><span style="font-size:0.7rem; font-weight:800; color:#10B981; letter-spacing:1px;">ALPHA TESTNET ACTIVE</span></div></div>
        
        <div id="v-dash" class="view-ax" style="display:flex;">
            <div class="card-ax" style="border-color: var(--prim);"><span class="bal-lbl">Personal Ledger</span><div id="bal-txt" class="bal-val">0.00 AX</div><div id="addr-txt" class="pill" style="text-align:center;">VAULT LOCKED</div></div>
            <div class="card-ax"><span class="bal-lbl">Reward Reserve</span><div id="pool-txt" class="bal-val" style="color:#10B981;">0.00 AX</div><div class="pill">AXf7ca3d5889ed99de642913af6c5630d6c491732b44180771cba042a4eb5a7109cc3ccde9e1a24d5315947415d5e592123ab90edcc4ea85415c1747fbe1684158</div></div>
            <button class="btn-ax" onclick="mine()"><i class="fas fa-hammer"></i> VALIDATE NETWORK (+50 AX)</button>
        </div>

        <div id="v-wallet" class="view-ax"><div class="card-ax"><span class="bal-lbl">Secure Asset Transfer</span><input type="text" id="tx-to" class="input-ax" placeholder="Recipient AX Address"><input type="number" id="tx-amt" class="input-ax" placeholder="Amount AX Tokens"><button class="btn-ax" onclick="send()">CONFIRM DEPOSIT</button></div></div>
        <div id="v-explorer" class="view-ax"><span class="bal-lbl">Real-Time Explorer</span><div id="block-list"></div></div>
        <div id="v-sec" class="view-ax"><div class="card-ax"><span class="bal-lbl">Vault Security</span><textarea id="i-seed" class="input-ax" style="height:120px; resize:none;" placeholder="Enter 24-word seed..."></textarea><button class="btn-ax" onclick="login()">RESTORE WALLET</button><div style="text-align:center; margin: 20px 0; color:#333; font-size:0.7rem; font-weight:900;">SECURE ENCRYPTION</div><button class="btn-ax" style="background:transparent; border:1px solid #222; color:#555;" onclick="gen()">GENERATE IDENTITY</button><div id="g-res" style="display:none; margin-top:25px;"><div id="g-seed" style="display:grid; grid-template-columns:1fr 1fr; gap:10px;"></div><span class="bal-lbl" style="margin-top:25px;">Public Identity</span><div class="pill" id="g-pub"></div></div></div></div>
    </div>

    <div class="bottom-bar">
        <a class="nav-l active" id="n-dash" onclick="nav('dash')"><i class="fas fa-chart-pie"></i>DASH</a>
        <a class="nav-l" id="n-wallet" onclick="nav('wallet')"><i class="fas fa-paper-plane"></i>SEND</a>
        <a class="nav-l" id="n-explorer" onclick="nav('explorer')"><i class="fas fa-cubes"></i>CHAIN</a>
        <a class="nav-l" id="n-sec" onclick="nav('sec')"><i class="fas fa-shield-halved"></i>VAULT</a>
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
        async function nav(id) {
            document.querySelectorAll(".view-ax").forEach(v => v.style.display = "none");
            document.getElementById("v-" + id).style.display = "flex";
            document.querySelectorAll(".nav-l").forEach(n => n.classList.remove("active"));
            document.getElementById("n-" + id).classList.add("active");
            if(id === 'explorer') renderExplorer();
            window.scrollTo(0,0);
        }
        async function renderExplorer() {
            const r = await fetch("/api/chain"); const chain = await r.json();
            const list = document.getElementById("block-list"); let html = ""; const revChain = chain.reverse();
            for(let i=0; i<revChain.length; i++) {
                let b = revChain[i]; let idx = b.Index !== undefined ? b.Index : b.index;
                let hash = b.Hash || b.hash;
                html += '<div class="block-card" style="background:#0A0A0A; border:1px solid #1A1A1A; padding:20px; border-radius:15px; margin-bottom:15px;"><span style="background:var(--prim); padding:4px 10px; border-radius:6px; font-size:0.6rem; font-weight:800;">BLOCK #' + idx + '</span><div style="margin-top:12px; font-size:0.55rem; word-break:break-all; font-family:monospace; color:var(--txt-m);">' + hash + '</div></div>';
            }
            list.innerHTML = html;
        }
        async function login() {
            const s = document.getElementById("i-seed").value.trim().toLowerCase(); if(!s) return;
            const keys = await derive(s); session = { pub: keys.pub, priv: keys.priv, seed: s };
            localStorage.setItem("ax_v18_session", JSON.stringify(session)); location.reload();
        }
        async function gen() {
            let seed = []; for(let i=0; i<24; i++) seed.push(words[Math.floor(Math.random()*words.length)]);
            const keys = await derive(seed.join(" ")); document.getElementById("g-res").style.display = "block";
            let sH = ""; for(let i=0; i<seed.length; i++) sH += '<div style="background:#000; padding:10px; border-radius:10px; border:1px solid #111; font-size:0.65rem; color:var(--txt-m);"><span style="color:var(--prim);">'+(i+1)+'</span> '+seed[i]+'</div>';
            document.getElementById("g-seed").innerHTML = sH; document.getElementById("g-pub").innerText = keys.pub;
        }
        async function load() {
            if(session) {
                const r = await fetch("/api/balance/" + session.pub); const d = await r.json();
                document.getElementById("bal-txt").innerText = d.balance.toLocaleString() + " AX";
                document.getElementById("addr-txt").innerText = session.pub;
            }
            const rp = await fetch("/api/balance/AXf7ca3d5889ed99de642913af6c5630d6c491732b44180771cba042a4eb5a7109cc3ccde9e1a24d5315947415d5e592123ab90edcc4ea85415c1747fbe1684158");
            const dp = await rp.json(); document.getElementById("pool-txt").innerText = dp.balance.toLocaleString() + " AX";
        }
        async function mine() {
            if(!session) return alert("Vault Required");
            const r = await fetch("/api/mine?address=" + session.pub);
            if(r.ok) { alert("Validated!"); load(); } else { alert("Mempool empty."); }
        }
        async function send() {
            const tx = { sender: session.pub, recipient: document.getElementById("tx-to").value, amount: parseFloat(document.getElementById("tx-amt").value) };
            const r = await fetch("/api/transactions/new", { method: "POST", body: JSON.stringify(tx) });
            if(r.ok) { alert("Sent!"); nav('dash'); load(); } else { alert("Failed."); }
        }
        load(); setInterval(load, 15000);
    </script>
</body>
</html>
`