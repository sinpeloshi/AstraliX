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

	// RUTAS API
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

	// RUTAS WEB
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
// 🎨 LANDING PAGE
// ==========================================

const landingHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <title>AstraliX | Quantum-Resistant Layer 1</title>
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;600;800&family=JetBrains+Mono:wght@400;700&display=swap');
        :root { --bg: #050505; --bg-c: #0A0A0A; --prim: #3B82F6; --prim-g: rgba(59, 130, 246, 0.25); --acc: #10B981; --txt: #FFFFFF; --txt-m: #94A3B8; --brd: #1F2937; }
        * { box-sizing: border-box; }
        body { font-family: 'Plus Jakarta Sans', sans-serif; margin: 0; color: var(--txt); background: var(--bg); line-height: 1.6; overflow-x: hidden; -webkit-font-smoothing: antialiased; scroll-behavior: smooth; }
        .bg-grid { position: fixed; top: 0; left: 0; width: 100vw; height: 100vh; background-image: linear-gradient(var(--brd) 1px, transparent 1px), linear-gradient(90deg, var(--brd) 1px, transparent 1px); background-size: 40px 40px; opacity: 0.2; z-index: -1; }
        .nav { padding: 25px 5%; display: flex; justify-content: space-between; align-items: center; max-width: 1400px; margin: 0 auto; position: sticky; top: 0; background: rgba(5,5,5,0.8); backdrop-filter: blur(15px); z-index: 100; border-bottom: 1px solid var(--brd); }
        .logo { font-weight: 800; font-size: 2.2rem; letter-spacing: -2px; color: var(--txt); text-decoration: none; }
        .logo span { color: var(--prim); }
        .nav-links { display: flex; gap: 25px; align-items: center; }
        .nav-links a { color: var(--txt-m); text-decoration: none; font-size: 0.85rem; font-weight: 600; transition: 0.3s; }
        .btn-core { background: var(--prim); color: white !important; padding: 10px 20px; border-radius: 100px; font-size: 0.8rem; font-weight: 800; text-decoration: none; }
        .hero { text-align: center; padding: 120px 5% 80px; max-width: 1100px; margin: 0 auto; }
        .hero h1 { font-size: clamp(3rem, 8vw, 6.5rem); font-weight: 800; margin: 0; letter-spacing: -4px; line-height: 1.15; background: linear-gradient(135deg, #FFF 0%, #94A3B8 100%); -webkit-background-clip: text; -webkit-text-fill-color: transparent; padding-bottom: 10px; }
        .hero p { font-size: clamp(1.1rem, 3vw, 1.4rem); color: var(--txt-m); margin: 25px auto 45px; max-width: 750px; }
        .btn-prim { padding: 20px 45px; border-radius: 100px; text-decoration: none; font-weight: 800; font-size: 1.05rem; transition: 0.3s; display: inline-block; }
        .btn-solid { background: var(--txt); color: var(--bg); }
        .btn-outline { background: transparent; border: 1px solid var(--brd); color: white; }
        .roadmap { padding: 80px 5%; max-width: 1200px; margin: 0 auto; }
        .rm-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; }
        .rm-item { border-left: 2px solid var(--prim); padding-left: 25px; position: relative; margin-bottom: 30px; }
        .rm-item::before { content: ''; position: absolute; left: -6px; top: 0; width: 10px; height: 10px; background: var(--prim); border-radius: 50%; box-shadow: 0 0 10px var(--prim); }
        .pre-sale { background: linear-gradient(180deg, var(--bg-c) 0%, #000 100%); border: 1px solid var(--brd); padding: 80px 5%; text-align: center; border-radius: 40px; max-width: 850px; margin: 0 auto 100px; width: 90%; position: relative; }
        .price { font-size: clamp(4rem, 12vw, 6rem); font-weight: 800; margin: 10px 0 20px; letter-spacing: -4px; }
        .w-addr { font-family: 'JetBrains Mono', monospace; font-size: clamp(0.75rem, 2.5vw, 0.95rem); color: var(--txt); word-break: break-all; background: rgba(255,255,255,0.05); padding: 15px; border-radius: 12px; margin-top: 15px; }
        .btn-verify { background: var(--acc); color: #000; padding: 20px 50px; border-radius: 100px; text-decoration: none; font-weight: 800; font-size: 1.1rem; display: inline-flex; align-items: center; gap: 10px; transition: 0.3s; }
        footer { padding: 80px 5% 40px; border-top: 1px solid var(--brd); background: #020202; }
        .footer-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 40px; max-width: 1200px; margin: 0 auto; text-align: left; }
        .social-icons a { display: flex; align-items: center; gap: 10px; color: var(--txt); font-weight: 800; text-decoration: none; margin-top: 15px; }
        @media (max-width: 850px) { .nav-links { display: none; } }
    </style>
</head>
<body>
    <div class="bg-grid"></div>
    <nav class="nav">
        <a href="/" class="logo"><span>A</span>strali<span>X</span></a>
        <div class="nav-links">
            <a href="#roadmap">Roadmap</a>
            <a href="/whitepaper">Whitepaper</a>
            <a href="/dashboard" class="btn-core">LAUNCH APP</a>
        </div>
    </nav>
    <header class="hero">
        <div class="badge" style="background: rgba(59,130,246,0.1); color: var(--prim); padding: 8px 20px; border-radius: 100px; font-weight: 800; font-size: 0.75rem; margin-bottom: 25px; display: inline-block;">Decentralized Future</div>
        <h1>Securing the Era of<br>Quantum Intelligence.</h1>
        <p>A mission-critical Layer 1 protocol doubling cryptographic security standards to 512-bit. Built for the next 100 years of digital asset sovereignty.</p>
        <div style="display:flex; gap:15px; justify-content:center; flex-wrap:wrap;">
            <a href="#buy" class="btn-prim btn-solid">Get Founder Node</a>
            <a href="/whitepaper" class="btn-prim btn-outline">Read Whitepaper</a>
        </div>
    </header>
    <section class="roadmap" id="roadmap">
        <div style="text-align:center; margin-bottom:50px;"><h2 style="font-size:2.5rem; font-weight:800;">Strategic Roadmap</h2></div>
        <div class="rm-grid">
            <div class="rm-item"><div style="color:var(--prim); font-weight:800; font-size:0.8rem; margin-bottom:10px;">Q1 2026</div><h4>Genesis Alpha</h4><p>Launch of the 512-bit core engine and founder node allocation program.</p></div>
            <div class="rm-item"><div style="color:var(--prim); font-weight:800; font-size:0.8rem; margin-bottom:10px;">Q3 2026</div><h4>Ecosystem Expansion</h4><p>Public testnet deployment and developer grants for quantum-safe DApps.</p></div>
            <div class="rm-item"><div style="color:var(--prim); font-weight:800; font-size:0.8rem; margin-bottom:10px;">Q1 2027</div><h4>Mainnet Launch</h4><p>Transition to full decentralized consensus and token migration 1:1.</p></div>
        </div>
    </section>
    <section id="buy" class="pre-sale">
        <div style="color:var(--prim); font-weight:800; letter-spacing:2px; margin-bottom:10px;">FOUNDER NODE ACCESS</div>
        <div class="price">21 USDT</div>
        <div class="addr-box">
            <div style="color:#F3BA2F; font-size:0.8rem; font-weight:800; margin-bottom:10px;">BSC NETWORK (BEP-20)</div>
            <div class="w-addr">0x948a663b1bd1292ded76a8412af2092bf0462d7c</div>
        </div>
        <a href="https://tally.so/r/jaxlL1" target="_blank" class="btn-verify">VERIFY DEPOSIT <i class="fas fa-arrow-right"></i></a>
    </section>
    <footer>
        <div class="footer-grid">
            <div class="f-col">
                <a href="/" class="logo" style="font-size:1.5rem;"><span>A</span>strali<span>X</span></a>
                <p style="font-size:0.8rem; color:var(--txt-m); margin-top:15px; line-height:1.8;">Next-generation Layer 1 infrastructure for the post-quantum world.</p>
                <div class="social-icons">
                    <a href="https://x.com/XAstraliX" target="_blank"><i class="fab fa-x-twitter" style="font-size:1.5rem;"></i> @XAstraliX</a>
                </div>
            </div>
            <div class="f-col"><h4>Protocol</h4><a href="#roadmap">Roadmap</a><a href="/whitepaper">Whitepaper</a><a href="https://github.com" target="_blank">Github</a></div>
            <div class="f-col"><h4>Governance</h4><a href="/dashboard">Core Dashboard</a><a href="#">Network Status</a><a href="https://tally.so/r/jaxlL1" target="_blank">Verify Node</a></div>
        </div>
        <div style="text-align:center; margin-top:60px; color:var(--txt-m); font-size:0.8rem; opacity:0.5;">© 2026 AstraliX Core Engine. Built for Sovereign Security.</div>
    </footer>
</body>
</html>
`

// ==========================================
// 📄 WHITEPAPER WEB PAGE
// ==========================================

const whitepaperHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AstraliX Whitepaper | 512-bit Protocol</title>
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;600;800&family=JetBrains+Mono:wght@400&display=swap');
        :root { --bg: #050505; --prim: #3B82F6; --txt: #FFFFFF; --txt-m: #94A3B8; --brd: #1F2937; }
        body { font-family: 'Plus Jakarta Sans', sans-serif; background: var(--bg); color: var(--txt); margin: 0; line-height: 1.8; }
        .container { max-width: 800px; margin: 0 auto; padding: 60px 20px; }
        .nav-back { margin-bottom: 40px; }
        .nav-back a { color: var(--prim); text-decoration: none; font-weight: 700; display: flex; align-items: center; gap: 8px; }
        h1 { font-size: clamp(2.5rem, 5vw, 3.5rem); font-weight: 800; margin-bottom: 10px; letter-spacing: -2px; }
        .meta { color: var(--txt-m); font-size: 0.9rem; margin-bottom: 60px; border-bottom: 1px solid var(--brd); padding-bottom: 20px; }
        h2 { font-size: 1.8rem; font-weight: 800; margin-top: 50px; color: var(--prim); }
        p { font-size: 1.1rem; color: #cbd5e1; margin-bottom: 25px; }
        .code-box { background: #0A0A0A; border: 1px solid var(--brd); padding: 25px; border-radius: 16px; font-family: 'JetBrains Mono', monospace; font-size: 0.9rem; margin: 30px 0; color: var(--txt-m); }
        .highlight-box { border-left: 4px solid var(--prim); background: rgba(59,130,246,0.05); padding: 25px; border-radius: 0 16px 16px 0; margin: 40px 0; }
        footer { text-align: center; margin-top: 100px; padding: 40px 0; border-top: 1px solid var(--brd); font-size: 0.9rem; color: var(--txt-m); }
    </style>
</head>
<body>
    <div class="container">
        <div class="nav-back"><a href="/"><i class="fas fa-arrow-left"></i> Back to Home</a></div>
        <h1>AstraliX Protocol</h1>
        <div class="meta">Version 1.0 (Alpha Genesis) • Lead Architect: Denis Waldemar • March 2026</div>

        <h2>1. Executive Summary</h2>
        <p>AstraliX is a next-generation Layer 1 blockchain infrastructure engineered to redefine digital security standards. While current industry leaders rely on 256-bit protocols, AstraliX doubles the entropy to <strong>512-bit (SHA-512)</strong>, creating a quantum-resistant fortress for digital assets.</p>

        <h2>2. The Quantum Vulnerability</h2>
        <p>Current blockchains use 256-bit encryption. Quantum computers, using algorithms like Shor’s, will eventually be able to derive private keys from public addresses. AstraliX eliminates this threat by expanding the cryptographic search space to 2^512 combinations, a number so large it surpasses the estimated atoms in the known universe.</p>

        <div class="code-box">
            // SHA-256 (Legacy): 2^256 combinations<br>
            // SHA-512 (AstraliX): 2^512 combinations<br>
            // Difference: Exponentially more secure against brute-force.
        </div>

        <h2>3. Technical Foundation</h2>
        <p>Built using <strong>Golang</strong>, the AstraliX core engine handles massive transaction throughput with zero memory collisions. Our client-side "Zero-Trust Vault" ensures that 512-bit private keys are generated locally and never touch our servers.</p>

        <div class="highlight-box">
            <strong>The founder Node Program:</strong> We are allocating 100 Founder Nodes during the Alpha phase. Each node secures the early network and receives 10,000 AX tokens in the Genesis block.
        </div>

        <h2>4. Vision</h2>
        <p>Our mission is to provide an unbreakable foundation for the next 100 years of digital asset sovereignty. AstraliX isn't just a network; it's an insurance policy for value in the post-quantum era.</p>

        <footer>© 2026 AstraliX Foundation. Built for the Future.</footer>
    </div>
</body>
</html>
`

// ==========================================
// 📱 DASHBOARD OPTIMIZADO
// ==========================================

const dashboardHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>AstraliX Core | OS</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;600;800&family=JetBrains+Mono:wght@400&display=swap');
        :root { --bg: #050505; --card: #0F172A; --primary: #3B82F6; --text: #FFFFFF; --text-m: #94A3B8; --border: #1E293B; }
        body { background: var(--bg); font-family: 'Plus Jakarta Sans', sans-serif; margin: 0; padding-bottom: 100px; color: var(--text); overflow-x: hidden; }
        .container { max-width: 500px; margin: 0 auto; padding: 0 20px; width: 100%; }
        .header-ax { padding: 40px 0 20px; text-align: center; }
        .header-ax h5 { font-weight: 800; letter-spacing: 1px; margin: 0; font-size: 1.4rem; color: var(--text); }
        .status-box { display: inline-flex; align-items: center; background: rgba(16, 185, 129, 0.1); padding: 8px 16px; border-radius: 100px; margin-top: 12px; border: 1px solid rgba(16, 185, 129, 0.2); }
        .status-dot { height: 8px; width: 8px; background: #10B981; border-radius: 50%; margin-right: 10px; box-shadow: 0 0 12px #10B981; }
        .status-text { font-size: 0.7rem; font-weight: 800; color: #10B981; letter-spacing: 1px; text-transform: uppercase; }
        .view-ax { display: none; flex-direction: column; width: 100%; gap: 20px; margin-top: 10px; }
        .card-ax { background: var(--card); border-radius: 28px; padding: 30px; width: 100%; border: 1px solid var(--border); position: relative; }
        .card-dark { background: linear-gradient(145deg, #0F172A 0%, #020617 100%); border-color: var(--primary); }
        .balance-label { font-size: 0.7rem; text-transform: uppercase; letter-spacing: 2px; color: var(--text-m); font-weight: 700; margin-bottom: 8px; display: block; }
        .balance-amount { font-size: 2.4rem; font-weight: 800; margin-bottom: 20px; letter-spacing: -1px; }
        .pill-address { background: rgba(255,255,255,0.03); padding: 16px; border-radius: 18px; font-family: 'JetBrains Mono', monospace; font-size: 0.65rem; word-break: break-all; color: var(--text-m); border: 1px solid var(--border); line-height: 1.4; }
        .input-ax { width: 100%; padding: 18px; border-radius: 18px; border: 1px solid var(--border); background: rgba(255,255,255,0.03); color: white; font-family: 'Plus Jakarta Sans', sans-serif; font-size: 0.95rem; outline: none; transition: 0.3s; }
        .btn-ax { background: var(--primary); color: white; border-radius: 18px; padding: 18px; font-weight: 800; border: none; width: 100%; font-size: 1rem; cursor: pointer; display: flex; align-items: center; justify-content: center; gap: 10px; }
        .block-card { background: rgba(255,255,255,0.02); border-radius: 24px; padding: 20px; width: 100%; border: 1px solid var(--border); margin-bottom: 15px; }
        .block-hash { font-family: 'JetBrains Mono', monospace; font-size: 0.6rem; color: var(--text-m); word-break: break-all; background: #000; padding: 12px; border-radius: 12px; margin-top: 10px; border: 1px solid var(--border); }
        .bottom-bar { background: rgba(10, 10, 10, 0.8); backdrop-filter: blur(20px); position: fixed; bottom: 0; width: 100%; height: 85px; display: flex; justify-content: space-around; align-items: center; border-top: 1px solid var(--border); z-index: 1000; }
        .nav-link-ax { color: #475569; text-decoration: none; font-size: 0.6rem; font-weight: 800; display: flex; flex-direction: column; align-items: center; gap: 6px; cursor: pointer; text-transform: uppercase; letter-spacing: 0.5px; }
        .nav-link-ax.active { color: var(--primary); }
        .nav-link-ax i { font-size: 1.3rem; }
        .seed-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; margin-top: 15px; }
        .seed-word { background: rgba(255,255,255,0.03); padding: 12px; border-radius: 12px; font-size: 0.8rem; border: 1px solid var(--border); color: var(--text-m); display: flex; gap: 8px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header-ax"><h5>Astrali<span>X</span> Core</h5><div class="status-box"><span class="status-dot"></span><span class="status-text">NETWORK ACTIVE</span></div></div>
        <div id="v-dash" class="view-ax" style="display:flex;">
            <div class="card-ax card-dark"><span class="balance-label">Total Supply Balance</span><div id="bal-txt" class="balance-amount">0.00 AX</div><div id="addr-txt" class="pill-address" style="text-align:center;">Vault Locked</div></div>
            <div class="card-ax"><span class="balance-label">Treasury Reward Pool</span><div id="pool-txt" class="balance-amount" style="color:#10B981; font-size:1.8rem;">0.00 AX</div><div class="pill-address">AXf7ca3d5889ed99de642913af6c5630d6c491732b44180771cba042a4eb5a7109cc3ccde9e1a24d5315947415d5e592123ab90edcc4ea85415c1747fbe1684158</div></div>
            <button class="btn-ax" onclick="mine()"><i class="fas fa-hammer"></i> VALIDATE NETWORK (+50 AX)</button>
        </div>
        <div id="v-wallet" class="view-ax">
            <div class="card-ax"><span class="balance-label">Asset Transfer Protocol</span><div style="display:flex; flex-direction:column; gap:12px;"><input type="text" id="tx-to" class="input-ax" placeholder="Recipient AX Address"><input type="number" id="tx-amt" class="input-ax" placeholder="Amount (AX Tokens)"><button class="btn-ax" onclick="send()"><i class="fas fa-paper-plane"></i> CONFIRM TRANSFER</button></div></div>
        </div>
        <div id="v-explorer" class="view-ax"><span class="balance-label">On-Chain Block Explorer</span><div id="block-list" style="width:100%;"></div></div>
        <div id="v-sec" class="view-ax">
            <div class="card-ax">
                <span class="balance-label">Identity Access Point</span>
                <textarea id="i-seed" class="input-ax" style="min-height:100px; resize:none;" placeholder="Enter your 24-word recovery phrase..."></textarea>
                <button class="btn-ax" style="margin-top:12px;" onclick="login()">RESTORE WALLET</button>
                <div style="margin:25px 0; display:flex; align-items:center; gap:15px; opacity:0.2;"><hr style="flex:1; border:0; border-top:1px solid white;"><span style="font-size:0.6rem; font-weight:800;">OR</span><hr style="flex:1; border:0; border-top:1px solid white;"></div>
                <button class="btn-ax" style="background:transparent; border:1px solid var(--border); color:var(--text-m);" onclick="gen()">NEW 512-BIT IDENTITY</button>
                <div id="g-res" style="display:none; margin-top:30px; text-align:left;">
                    <span class="balance-label" style="color:#EF4444;">Generated Recovery Seed</span>
                    <div id="g-seed" class="seed-grid"></div>
                    <span class="balance-label" style="margin-top:25px;">Public Identity</span>
                    <div class="pill-address" id="g-pub"></div>
                </div>
            </div>
        </div>
    </div>
    <div class="bottom-bar">
        <a class="nav-link-ax active" id="n-dash" onclick="nav('dash')"><i class="fas fa-chart-pie"></i>DASH</a>
        <a class="nav-link-ax" id="n-wallet" onclick="nav('wallet')"><i class="fas fa-exchange-alt"></i>SEND</a>
        <a class="nav-link-ax" id="n-explorer" onclick="nav('explorer')"><i class="fas fa-cubes"></i>CHAIN</a>
        <a class="nav-link-ax" id="n-sec" onclick="nav('sec')"><i class="fas fa-shield-halved"></i>VAULT</a>
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
            document.querySelectorAll(".nav-link-ax").forEach(n => n.classList.remove("active"));
            document.getElementById("n-" + id).classList.add("active");
            if(id === 'explorer') renderExplorer();
            window.scrollTo(0,0);
        }
        async function renderExplorer() {
            const r = await fetch("/api/chain");
            const chain = await r.json();
            const list = document.getElementById("block-list");
            let html = ""; const revChain = chain.reverse();
            for(let i=0; i<revChain.length; i++) {
                let b = revChain[i]; let idx = b.Index !== undefined ? b.Index : b.index;
                let hash = b.Hash || b.hash;
                let txs = b.Transactions || b.transactions || [];
                html += '<div class="block-card"><div style="display:flex; justify-content:space-between; align-items:center;"><span style="background:var(--primary); padding:4px 8px; border-radius:6px; font-size:0.6rem; font-weight:800;">BLOCK #' + idx + '</span></div><div class="block-hash">' + hash + '</div><div style="font-size:0.6rem; color:var(--text-m); margin-top:10px;">TX COUNT: ' + txs.length + ' • SHA-512 SECURED</div></div>';
            }
            list.innerHTML = html;
        }
        async function login() {
            const s = document.getElementById("i-seed").value.trim().toLowerCase();
            if(!s) return; const keys = await derive(s);
            session = { pub: keys.pub, priv: keys.priv, seed: s };
            localStorage.setItem("ax_v18_session", JSON.stringify(session)); location.reload();
        }
        async function gen() {
            let seed = []; for(let i=0; i<24; i++) seed.push(words[Math.floor(Math.random()*words.length)]);
            const keys = await derive(seed.join(" "));
            document.getElementById("g-res").style.display = "block";
            let seedHtml = ""; for(let i=0; i<seed.length; i++) {
                seedHtml += '<div class="seed-word"><span style="color:var(--primary); font-weight:800; opacity:0.5; margin-right:8px;">'+(i+1)+'</span>'+seed[i]+'</div>';
            }
            document.getElementById("g-seed").innerHTML = seedHtml; document.getElementById("g-pub").innerText = keys.pub;
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
            if(!session) return alert("Vault Identity Required");
            const r = await fetch("/api/mine?address=" + session.pub);
            if(r.ok) { alert("¡Network Validated!"); load(); } else { alert("Mempool empty. Send a transaction first!"); }
        }
        async function send() {
            const tx = { sender: session.pub, recipient: document.getElementById("tx-to").value, amount: parseFloat(document.getElementById("tx-amt").value) };
            const r = await fetch("/api/transactions/new", { method: "POST", body: JSON.stringify(tx) });
            if(r.ok) { alert("Transaction Propagated!"); nav('dash'); load(); } else { alert("Validation failed."); }
        }
        load(); setInterval(load, 15000);
    </script>
</body>
</html>
`