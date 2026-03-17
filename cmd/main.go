Package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"astralix/core"
	_ "github.com/lib/pq"
)

var Blockchain []core.Block
var Mempool []core.Transaction
var db *sql.DB

const TREASURY_POOL_ADDR = "AXf7ca3d5889ed99de642913af6c5630d6c491732b44180771cba042a4eb5a7109cc3ccde9e1a24d5315947415d5e592123ab90edcc4ea85415c1747fbe1684158"

// 🛡️ LISTA NEGRA DE DIRECCIONES (ACCESO RESTRINGIDO A NIVEL DE PROTOCOLO)
var Blacklist = map[string]bool{
	"AXaadcc656dffb1a2c6a86d7cbb9a3ad04e5c278fe17171ea30021e6b49aae98021b5aa570d829aaf5b5a492238ca30152e5c3ed09dfe42843dd6fe45049486758": true,
}

// ==========================================
// ⚙️ MOTOR BLOCKCHAIN (PRO CORE)
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
	http.HandleFunc("/api/holders", func(w http.ResponseWriter, r *http.Request) {
		balances := make(map[string]float64)
		for _, block := range Blockchain {
			for _, tx := range block.Transactions {
				balances[tx.Recipient] += tx.Amount
				if tx.Sender != "SYSTEM" {
					balances[tx.Sender] -= tx.Amount
				}
			}
		}
		type Holder struct {
			Address string  `json:"address"`
			Balance float64 `json:"balance"`
		}
		var holders []Holder
		for addr, bal := range balances {
			if bal > 0 {
				holders = append(holders, Holder{Address: addr, Balance: bal})
			}
		}
		sort.Slice(holders, func(i, j int) bool {
			return holders[i].Balance > holders[j].Balance
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(holders)
	})
	http.HandleFunc("/api/mine", func(w http.ResponseWriter, r *http.Request) {
		miner := r.URL.Query().Get("address")
		if miner == "" || len(Mempool) == 0 { http.Error(w, "Error", 400); return }
		
		// 🛡️ RESTRICCIÓN DE BLACKLIST
		if Blacklist[miner] {
			http.Error(w, "Forbidden: This address is blacklisted for security reasons.", 403)
			return
		}

		// 🛡️ REGLA ANTI-SPAM: Requiere saldo mínimo (Stake) para validar
		if getBalance(miner) < 500 { 
			http.Error(w, "Unauthorized: Minimum 500 AX required to validate.", 401)
			return 
		}

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

		// 🛡️ RESTRICCIÓN DE BLACKLIST (Emisor o Receptor)
		if Blacklist[tx.Sender] || Blacklist[tx.Recipient] {
			http.Error(w, "Forbidden: One of the addresses is blacklisted.", 403)
			return
		}

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
	// RUTAS DE DOCUMENTOS
	http.HandleFunc("/tech-paper", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, techPaperHTML)
	})

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}

// ==========================================
// 🎨 LANDING PAGE (VALLEY STYLE)
// ==========================================

const landingHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <title>AstraliX | The 512-bit Layer 1 Protocol</title>
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.2/css/all.min.css" rel="stylesheet">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700;800;900&family=JetBrains+Mono:wght@400;700&display=swap');
        
        :root { --bg: #020202; --bg-card: #080808; --prim: #3B82F6; --txt: #FFFFFF; --txt-m: #8899A6; --brd: #1A1A1A; --acc: #10B981; }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: 'Inter', sans-serif; background: var(--bg); color: var(--txt); line-height: 1.5; overflow-x: hidden; -webkit-font-smoothing: antialiased; scroll-behavior: smooth; letter-spacing: -0.02em; }
        
        .bg-p { position: fixed; width: 100vw; height: 100vh; background-image: radial-gradient(rgba(255,255,255,0.08) 1px, transparent 0); background-size: 30px 30px; z-index: -1; }
        
        .nav { padding: 20px 6%; display: flex; justify-content: space-between; align-items: center; position: sticky; top: 0; background: rgba(2,2,2,0.7); backdrop-filter: blur(24px); z-index: 100; border-bottom: 1px solid rgba(255,255,255,0.05); }
        .logo { display: flex; align-items: center; text-decoration: none; }
        .logo img { height: 45px; width: auto; mix-blend-mode: screen; } 
        
        .nav-links { display: flex; align-items: center; }
        .nav-links a { color: var(--txt-m); text-decoration: none; font-size: 0.85rem; font-weight: 600; transition: 0.2s; margin-right: 25px; }
        .nav-links a:hover { color: var(--txt); }
        .btn-core-nav { background: rgba(255,255,255,0.1); color: white !important; padding: 10px 22px; border-radius: 100px; font-size: 0.75rem; font-weight: 700; text-decoration: none; transition: 0.3s; margin-right: 0 !important; border: 1px solid rgba(255,255,255,0.1); }
        .btn-core-nav:hover { background: rgba(255,255,255,1); color: #000 !important; }
        .nav-socials a:hover { color: var(--prim) !important; transform: translateY(-2px); }
        
        .hero { text-align: center; padding: 100px 6% 60px; max-width: 1200px; margin: 0 auto; position: relative; }
        
        .hero-title-container { position: relative; display: inline-block; margin-bottom: 30px; }
        .hero-title-container::before { content: ''; position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); width: 120%; height: 120%; background: radial-gradient(circle, rgba(59, 130, 246, 0.2) 0%, transparent 60%); filter: blur(40px); z-index: -1; pointer-events: none; }

        .hero h1 { font-size: clamp(3.5rem, 10vw, 7rem); font-weight: 900; letter-spacing: -0.05em; line-height: 1.05; margin: 0; background: radial-gradient(100% 100% at 50% 0%, #FFFFFF 0%, #A1A1AA 100%); -webkit-background-clip: text; -webkit-text-fill-color: transparent; text-shadow: 0px 4px 20px rgba(0,0,0,0.5); padding: 10px 0; }
        .hero p { font-size: clamp(1.1rem, 2.5vw, 1.3rem); color: #A1A1AA; max-width: 700px; margin: 0 auto 50px; font-weight: 400; line-height: 1.6; letter-spacing: -0.01em; }
        
        .hero-btns { display: flex; gap: 15px; justify-content: center; flex-wrap: wrap; margin-bottom: 60px; }
        .btn-p { padding: 16px 36px; border-radius: 100px; font-weight: 600; text-decoration: none; font-size: 1rem; transition: 0.3s; display: inline-flex; align-items: center; justify-content: center; gap: 10px; letter-spacing: -0.01em; }
        .btn-blue { background: var(--txt); color: #000; box-shadow: 0 0 20px rgba(255,255,255,0.2); }
        .btn-blue:hover { background: #E4E4E7; transform: translateY(-2px); }
        .btn-white { background: rgba(255,255,255,0.05); color: #FFF; border: 1px solid rgba(255,255,255,0.1); }
        .btn-white:hover { background: rgba(255,255,255,0.1); transform: translateY(-2px); }
        .btn-dark { color: #A1A1AA; border: 1px solid rgba(255,255,255,0.1); }
        .btn-dark:hover { color: #FFF; background: rgba(255,255,255,0.05); }
        
        .mockup-container { max-width: 900px; margin: 0 auto; padding: 0 6%; position: relative; perspective: 1000px; }
        .mockup-glow { position: absolute; top: 20%; left: 50%; transform: translate(-50%, -50%); width: 80%; height: 50%; background: var(--prim); filter: blur(120px); opacity: 0.15; z-index: -1; }
        .mockup-window { background: rgba(5,5,5,0.8); border: 1px solid rgba(255,255,255,0.1); border-radius: 16px; overflow: hidden; backdrop-filter: blur(20px); box-shadow: 0 30px 60px -12px rgba(0,0,0,0.8); transform: rotateX(2deg); transition: 0.5s; text-align: left; }
        .mockup-window:hover { transform: rotateX(0deg) translateY(-5px); border-color: rgba(255,255,255,0.2); }
        .mockup-header { background: rgba(255,255,255,0.02); padding: 12px 20px; display: flex; align-items: center; gap: 8px; border-bottom: 1px solid rgba(255,255,255,0.05); }
        .m-dot { width: 10px; height: 10px; border-radius: 50%; }
        .mockup-body { padding: 30px; font-family: 'JetBrains Mono', monospace; font-size: 0.85rem; color: #A1A1AA; line-height: 1.8; }
        .m-highlight { color: var(--acc); font-weight: 700; }
        .m-address { background: rgba(255,255,255,0.03); padding: 10px; border-radius: 8px; font-size: 0.7rem; margin-top: 15px; word-break: break-all; border: 1px solid rgba(255,255,255,0.05); }
        
        /* TOKENOMICS SECTION */
        .tokenomics { max-width: 1000px; margin: 100px auto; padding: 0 6%; text-align: center; }
        .tok-flex { display: flex; align-items: center; justify-content: center; gap: 50px; flex-wrap: wrap; margin-top: 50px; }
        .tok-chart { position: relative; width: 300px; height: 300px; border-radius: 50%; background: conic-gradient(var(--acc) 0% 12.5%, #4B5563 12.5% 52.5%, #8B5CF6 52.5% 67.5%, var(--prim) 67.5% 82.5%, #F59E0B 82.5% 92.5%, #EC4899 92.5% 100%); }
        .tok-chart::after { content: '40% REWARDS'; position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); width: 180px; height: 180px; background: var(--bg); border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: 800; font-size: 0.85rem; color: var(--acc); letter-spacing: 2px; text-align: center; }
        .tok-legend { text-align: left; display: flex; flex-direction: column; gap: 15px; }
        .leg-item { display: flex; align-items: center; gap: 12px; font-size: 0.95rem; color: var(--txt-m); }
        .leg-color { width: 12px; height: 12px; border-radius: 3px; }

        .sec-q { display: flex; gap: 20px; max-width: 1200px; margin: 80px auto; padding: 0 6%; }
        .q-box { flex: 1; border-radius: 20px; padding: 40px; text-align: left; border: 1px solid; }
        .q-box h3 { font-size: 0.75rem; text-transform: uppercase; letter-spacing: 2px; color: var(--txt-m); margin-bottom: 15px; }
        .q-box .val { font-family: 'JetBrains Mono'; font-size: 2rem; font-weight: 800; margin-bottom: 10px; }
        
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; max-width: 1200px; margin: 0 auto 100px; padding: 0 6%; }
        .card { background: rgba(255,255,255,0.02); border: 1px solid rgba(255,255,255,0.05); padding: 50px 40px; text-align: left; border-radius: 20px; transition: 0.3s; }
        .card:hover { border-color: rgba(255,255,255,0.2); background: rgba(255,255,255,0.04); transform: translateY(-5px); }
        .card i { color: var(--txt); font-size: 2rem; margin-bottom: 25px; display: block; opacity: 0.8; }
        .card h4 { font-size: 1.3rem; font-weight: 700; margin-bottom: 15px; letter-spacing: -0.02em; }
        .card p { color: #A1A1AA; font-size: 0.95rem; line-height: 1.7; }
        
        .roadmap { max-width: 800px; margin: 100px auto; padding: 0 6%; }
        .rm-step { border-left: 2px solid rgba(255,255,255,0.1); padding: 0 0 50px 30px; position: relative; }
        .rm-step::before { content: ''; position: absolute; left: -6px; top: 0; width: 10px; height: 10px; background: var(--prim); border-radius: 50%; }
        .rm-date { font-weight: 800; color: var(--prim); font-size: 0.8rem; margin-bottom: 10px; text-transform: uppercase; }
        
        .pre-sale { background: rgba(255,255,255,0.01); padding: 100px 6%; text-align: center; border-top: 1px solid rgba(255,255,255,0.05); }
        .tier-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 20px; max-width: 800px; margin: 0 auto 40px; }
        .tier-card { background: #000; border: 1px solid rgba(255,255,255,0.1); border-radius: 20px; padding: 40px 30px; transition: 0.3s; position: relative; overflow: hidden; text-align: center; }
        .tier-card:hover { border-color: rgba(255,255,255,0.3); transform: translateY(-5px); }
        .tier-card.premium { border-color: rgba(59, 130, 246, 0.4); background: radial-gradient(100% 100% at 50% 0%, rgba(59, 130, 246, 0.1) 0%, #000 100%); }
        .tier-card.premium::before { content: 'PRIORITY'; position: absolute; top: 15px; right: -35px; background: var(--prim); color: #FFF; font-size: 0.55rem; font-weight: 800; padding: 5px 35px; transform: rotate(45deg); letter-spacing: 1px; }
        .t-price { font-size: 3rem; font-weight: 900; letter-spacing: -0.05em; margin: 15px 0; color: #FFF; }
        .t-name { font-size: 0.9rem; text-transform: uppercase; letter-spacing: 2px; color: #A1A1AA; font-weight: 800; }
        .premium .t-name { color: var(--txt); }
        .t-list { text-align: left; margin-top: 25px; display: flex; flex-direction: column; gap: 15px; font-size: 0.9rem; color: #A1A1AA; }
        .t-list div { display: flex; align-items: center; gap: 12px; }
        .t-list i { color: var(--txt); background: rgba(255,255,255,0.1); padding: 5px; border-radius: 50%; font-size: 0.6rem; min-width: 12px; text-align: center; }
        
        .inst-box { background: rgba(255,255,255,0.02); border: 1px solid rgba(255,255,255,0.05); border-radius: 20px; padding: 30px; margin-bottom: 20px; text-align: left; }
        .btn-buy { background: var(--txt); color: #000; padding: 18px 40px; border-radius: 100px; font-weight: 700; text-decoration: none; font-size: 1.05rem; display: inline-flex; align-items: center; justify-content: center; gap: 10px; transition: 0.3s; width: 100%; box-sizing: border-box; letter-spacing: -0.01em; }
        .btn-buy:hover { transform: scale(1.02); background: #E4E4E7; }
        
        footer { padding: 80px 6% 40px; border-top: 1px solid rgba(255,255,255,0.05); display: grid; grid-template-columns: 2fr 1fr 1fr; gap: 50px; max-width: 1200px; margin: 0 auto; text-align: left; }
        .f-col h5 { margin-bottom: 20px; font-size: 0.85rem; text-transform: uppercase; letter-spacing: 1px; color: var(--txt); }
        .f-col a { display: block; color: #A1A1AA; text-decoration: none; margin-bottom: 12px; font-size: 0.9rem; transition: 0.2s; }
        .f-col a:hover { color: var(--txt); }
        .f-logo img { height: 50px; mix-blend-mode: screen; }
        
        @media (max-width: 850px) { 
            .nav-links { display: none; } 
            .hero { padding-top: 50px; } 
            .sec-q { flex-direction: column; } 
            .tok-flex { flex-direction: column; }
            .hero-btns { flex-direction: column; width: 100%; gap: 12px; } 
            .hero-btns .btn-p { width: 100%; } 
            .mockup-body { font-size: 0.7rem; padding: 20px; }
            footer { grid-template-columns: 1fr; } 
        }
    </style>
</head>
<body>
    <div class="bg-p"></div>
    <nav class="nav">
        <a href="/" class="logo"><img src="https://iili.io/qMGLM57.jpg" alt="AstraliX Protocol"></a>
        
        <div style="display: flex; align-items: center; gap: 20px;">
            <div class="nav-links">
                <a href="/whitepaper">Protocol</a>
                <a href="#roadmap">Mainnet</a>
                <a href="/dashboard" class="btn-core-nav">ENTER DASHBOARD</a>
            </div>
            <div class="nav-socials" style="display: flex; gap: 15px; align-items: center;">
                <a href="https://x.com/XAstraliX" target="_blank" style="color: var(--txt); font-size: 1.3rem; transition: 0.3s; display: inline-block;"><i class="fa-brands fa-x-twitter"></i></a>
                <a href="https://t.me/AstraliXProtocol" target="_blank" style="color: var(--txt); font-size: 1.3rem; transition: 0.3s; display: inline-block;"><i class="fa-brands fa-telegram"></i></a>
                <a href="mailto:info@astralix.network" style="color: var(--txt); font-size: 1.3rem; transition: 0.3s; display: inline-block;"><i class="fas fa-envelope"></i></a>
            </div>
        </div>
    </nav>
    <header class="hero">
        <div style="background: rgba(16,185,129,0.1); color: var(--acc); padding: 8px 24px; border-radius: 100px; font-size: 0.75rem; font-weight: 800; display: inline-block; margin-bottom: 30px; border: 1px solid rgba(16,185,129,0.2);"><span style="display:inline-block; width:8px; height:8px; background:var(--acc); border-radius:50%; margin-right:8px; box-shadow: 0 0 10px var(--acc);"></span>ALPHA TESTNET LIVE</div>
        
        <div class="hero-title-container">
            <h1>The 512-bit<br>Standard is Here.</h1>
        </div>
        
        <p>Redefining cryptographic sovereignty with the world's first ISP-backed DePIN Layer 1. Mathematically immune. Physically anchored.</p>
        
        <div class="hero-btns">
            <a href="/dashboard" class="btn-p btn-blue"><i class="fas fa-terminal"></i> Launch Testnet App</a>
            <a href="#buy" class="btn-p btn-white">Apply for Node</a>
            <a href="/whitepaper" class="btn-p btn-dark"><i class="fas fa-file-alt"></i> Read Whitepaper</a>
            <a href="/tech-paper" class="btn-p btn-dark" style="border-color: rgba(59, 130, 246, 0.4);"><i class="fas fa-code"></i> Technical Paper</a>
        </div>
    </header>

    <div class="mockup-container">
        <div class="mockup-glow"></div>
        <div class="mockup-window">
            <div class="mockup-header">
                <span class="m-dot" style="background:#FF5F56;"></span>
                <span class="m-dot" style="background:#FFBD2E;"></span>
                <span class="m-dot" style="background:#27C93F;"></span>
                <span style="margin-left:10px; font-family:'JetBrains Mono'; font-size:0.75rem; color:#666;">astralix-core-node ~ bash</span>
            </div>
            <div class="mockup-body">
                <div>> initializing 512-bit quantum-resistant protocol... <span class="m-highlight">[OK]</span></div>
                <div>> connecting to global peer network... <span class="m-highlight">[OK]</span></div>
                <div>> synchronizing genesis ledger... <span class="m-highlight">[DONE]</span></div>
                <br>
                <div style="display:flex; justify-content:space-between; border-top:1px dashed rgba(255,255,255,0.1); padding-top:20px; flex-wrap:wrap; gap:15px;">
                    <div>
                        <div style="font-size:0.7rem; text-transform:uppercase; letter-spacing:1px;">Network Status</div>
                        <div style="color:var(--acc); font-weight:800; font-size:1.2rem;">SYNCED (BLOCK #<span id="mock-block">0</span>)</div>
                    </div>
                    <div>
                        <div style="font-size:0.7rem; text-transform:uppercase; letter-spacing:1px;">Genesis Supply</div>
                        <div style="color:#FFF; font-weight:800; font-size:1.2rem;">1,000,002,021 AX</div>
                    </div>
                </div>
                <div class="m-address">
                    <span style="color:#666;">LATEST_HASH:</span> <span id="mock-hash" style="color:var(--txt-m);">AXec99e78875c95208706ae0be9b90ca7774bdbf458ebefc4307b66d5426385aefc91b072a68e6d567cfb371d01892d892e51c82113de5644ba4f6a973b7db345d</span>
                </div>
            </div>
        </div>
    </div>

    <section class="tokenomics">
        <h2 style="font-size:2.5rem; font-weight:800; margin-bottom:15px; letter-spacing: -0.03em;">Strategic Tokenomics</h2>
        <p style="color:#A1A1AA; max-width:600px; margin:0 auto;">A Fair-Launch distribution designed to reward validators, protect liquidity, and decentralize the network.</p>
        <div class="tok-flex">
            <div class="tok-chart"></div>
            <div class="tok-legend">
                <div class="leg-item"><div class="leg-color" style="background:#4B5563;"></div><span><strong>40.0%</strong> Ecosystem Mining Rewards</span></div>
                <div class="leg-item"><div class="leg-color" style="background:var(--acc);"></div><span><strong>12.5%</strong> Founder Nodes (Seed Round)</span></div>
                <div class="leg-item"><div class="leg-color" style="background:#8B5CF6;"></div><span><strong>15.0%</strong> Locked Liquidity Pool</span></div>
                <div class="leg-item"><div class="leg-color" style="background:var(--prim);"></div><span><strong>15.0%</strong> Treasury & Protocol R&D</span></div>
                <div class="leg-item"><div class="leg-color" style="background:#F59E0B;"></div><span><strong>10.0%</strong> Marketing & Community</span></div>
                <div class="leg-item"><div class="leg-color" style="background:#EC4899;"></div><span><strong>7.5%</strong> Team & Contributors (Locked)</span></div>
            </div>
        </div>
    </section>

    <section class="sec-q">
        <div class="q-box" style="border-color: rgba(239,68,68,0.2); background: rgba(239,68,68,0.03);">
            <h3 style="color: #EF4444;">Legacy Standard (BTC, ETH, SOL)</h3>
            <div class="val" style="color: #EF4444;">256-bit</div>
            <p style="color: #A1A1AA; font-size: 0.9rem; line-height:1.7; margin-top:10px;">Vulnerable to Shor's Algorithm and next-gen quantum decryption within this decade.</p>
        </div>
        <div class="q-box" style="border-color: rgba(16,185,129,0.2); background: rgba(16,185,129,0.03);">
            <h3 style="color: #10B981;">AstraliX Core Standard</h3>
            <div class="val" style="color: #10B981;">512-bit</div>
            <p style="color: #A1A1AA; font-size: 0.9rem; line-height:1.7; margin-top:10px;">Mathematically immune to classical and quantum brute-force attacks via ultra-high entropy.</p>
        </div>
    </section>

    <main class="grid">
        <div class="card"><i class="fas fa-microchip"></i><h4>Quantum-Proof</h4><p>SHA-512 architecture provides exponentially more combinations, securing assets for the next century of computing.</p></div>
        <div class="card"><i class="fas fa-bolt"></i><h4>Go-Native Engine</h4><p>Multi-threaded consensus built in Golang for sub-second block finality and massive node concurrency.</p></div>
        <div class="card"><i class="fas fa-fingerprint"></i><h4>Sovereign Vault</h4><p>Local mnemonic derivation. Your private keys never touch a server. Pure, untamperable decentralization.</p></div>
    </main>

    <section class="roadmap" id="roadmap">
        <div style="margin-bottom: 50px;"><h2 style="font-size:2.5rem; font-weight:800; letter-spacing: -0.03em;">Strategic Roadmap</h2></div>
        <div class="rm-step"><div class="rm-date">Q1 2026</div><h4 style="font-size:1.2rem; margin-bottom:10px;">Genesis Alpha</h4><p style="color:#A1A1AA;">Deployment of the core engine and Founder Node allocation program for early backers.</p></div>
        <div class="rm-step" style="border-left-color: var(--prim);"><div class="rm-date" style="background: var(--prim); color: #FFF; display: inline-block; padding: 4px 12px; border-radius: 4px; font-weight:800;">APRIL 2026</div><h4 style="font-size:1.2rem; margin-bottom:10px;">Mainnet Launch</h4><p style="color:#A1A1AA;">Official network transition. Token migration 1:1 and full decentralized validator onboarding.</p></div>
    </section>

    <section id="buy" class="pre-sale">
        <div style="text-transform: uppercase; letter-spacing: 4px; font-weight: 800; color: var(--prim); font-size: 0.8rem; margin-bottom:15px;">Founder Node Whitelist</div>
        <div style="margin-bottom: 50px;"><h2 style="font-size:2.5rem; font-weight:800; margin:0; color:#FFF; letter-spacing: -0.03em;">Strategic Node Application</h2></div>
        
        <div class="tier-grid">
            <div class="tier-card">
                <div class="t-name">Standard Node</div>
                <div class="t-price">10K<span style="font-size:1.2rem; color:#A1A1AA; font-weight:600;"> AX</span></div>
                <div class="t-list">
                    <div><i class="fas fa-check"></i> <span><strong>Seed Allocation</strong> (Testnet)</span></div>
                    <div><i class="fas fa-check"></i> <span><strong>Validator Rights:</strong> Earn AX</span></div>
                    <div><i class="fas fa-check"></i> <span>Mainnet 1:1 Migration</span></div>
                </div>
            </div>
            <div class="tier-card premium">
                <div class="t-name">Master Node</div>
                <div class="t-price">100K<span style="font-size:1.2rem; color:#A1A1AA; font-weight:600;"> AX</span></div>
                <div class="t-list">
                    <div><i class="fas fa-check"></i> <span><strong>Priority Allocation</strong> (Testnet)</span></div>
                    <div><i class="fas fa-check"></i> <span><strong>Premium Validator Rights</strong></span></div>
                    <div><i class="fas fa-check"></i> <span>Mainnet 1:1 Migration</span></div>
                </div>
            </div>
        </div>

        <div style="max-width: 700px; margin: 0 auto; text-align: left;">
            
            <div class="inst-box">
                <div style="color: var(--txt); font-weight: 800; font-size: 0.8rem; letter-spacing: 1px; margin-bottom: 10px;">STEP 1: GENERATE VAULT</div>
                <p style="color: #A1A1AA; font-size: 0.95rem; margin-bottom: 15px;">Open the <a href="/dashboard" target="_blank" style="color:var(--prim); text-decoration:none; font-weight:600;">Testnet App</a>, go to the <strong>VAULT</strong> tab, and generate your 512-bit Identity. Copy your public <strong>AX Address</strong> to include in your application.</p>
            </div>

            <div class="inst-box">
                <div style="color: var(--txt); font-weight: 800; font-size: 0.8rem; letter-spacing: 1px; margin-bottom: 10px;">STEP 2: SUBMIT APPLICATION</div>
                <p style="color: #A1A1AA; font-size: 0.95rem; margin-bottom: 15px;">We are currently selecting strategic infrastructure partners for the Genesis Block. Complete the application form via Tally.</p>
                <div style="text-align: center; margin-top: 20px;">
                    <a href="https://tally.so/r/jaxlL1" target="_blank" class="btn-buy">APPLY FOR NODE <i class="fas fa-arrow-right"></i></a>
                </div>
            </div>

            <div class="inst-box">
                <div style="color: var(--txt); font-weight: 800; font-size: 0.8rem; letter-spacing: 1px; margin-bottom: 10px;">STEP 3: ONBOARDING</div>
                <p style="color: #A1A1AA; font-size: 0.95rem; margin-bottom: 0;">If your profile is selected, you will receive an official email from <strong>info@astralix.network</strong> with the private SAFT and funding instructions to activate your node.</p>
            </div>
            
        </div>
    </section>

    <footer>
        <div class="footer-grid">
            <div class="f-col">
                <a href="/" class="logo f-logo"><img src="https://iili.io/qMGLM57.jpg" alt="AstraliX"></a>
                <p style="color: #A1A1AA; margin-top: 20px; font-size: 0.9rem; line-height:1.8;">Leading the cryptographic revolution through 512-bit security standards.</p>
                <div style="display:flex; gap:15px; margin-top:20px;">
                    <a href="https://x.com/XAstraliX" target="_blank" style="color:#FFF; font-weight:600; display:flex; align-items:center; gap:10px; text-decoration:none;"><i class="fa-brands fa-x-twitter" style="font-size:1.3rem;"></i> X/Twitter</a>
                    <a href="https://t.me/AstraliXProtocol" target="_blank" style="color:#FFF; font-weight:600; display:flex; align-items:center; gap:10px; text-decoration:none;"><i class="fa-brands fa-telegram" style="font-size:1.3rem;"></i> Telegram</a>
                    <a href="mailto:info@astralix.network" style="color:#FFF; font-weight:600; display:flex; align-items:center; gap:10px; text-decoration:none;"><i class="fas fa-envelope" style="font-size:1.3rem;"></i> Email</a>
                </div>
            </div>
            <div class="f-col"><h5>Protocol</h5><a href="/whitepaper">Whitepaper</a><a href="#roadmap">Roadmap</a></div>
            <div class="f-col"><h5>Resources</h5><a href="/dashboard">Testnet Dashboard</a><a href="https://tally.so/r/jaxlL1">Verify Node</a></div>
        </div>
        <div style="text-align:center; margin-top:60px; color:#A1A1AA; font-size:0.8rem; opacity:0.5;">© 2026 AstraliX Foundation. Designed for Sovereign Security.</div>
    </footer>

    <script>
        async function fetchRealData() {
            try {
                const res = await fetch("/api/chain");
                const chain = await res.json();
                if(chain && chain.length > 0) {
                    const latest = chain[chain.length - 1];
                    const idx = latest.Index !== undefined ? latest.Index : latest.index;
                    const hash = latest.Hash || latest.hash || latest.TxID || "AXec99e78875c95208706ae0be9b90ca7774bdbf458ebefc4307b66d5426385aefc91b072a68e6d567cfb371d01892d892e51c82113de5644ba4f6a973b7db345d";
                    document.getElementById("mock-block").innerText = idx;
                    document.getElementById("mock-hash").innerText = hash;
                }
            } catch(e) {}
        }
        fetchRealData();
        setInterval(fetchRealData, 10000);
    </script>
</body>
</html>
`

// ==========================================
// 📄 DASHBOARD CON LOGO Y RANKING
// ==========================================

const dashboardHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>Core | AstraliX OS</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.2/css/all.min.css" rel="stylesheet">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700;800;900&family=JetBrains+Mono:wght@400&display=swap');
        :root { --bg: #020202; --card: rgba(255,255,255,0.02); --prim: #3B82F6; --txt: #FFFFFF; --txt-m: #A1A1AA; --brd: rgba(255,255,255,0.05); }
        body { background: var(--bg); font-family: 'Inter', sans-serif; margin: 0; padding-bottom: 120px; color: var(--txt); overflow-x: hidden; letter-spacing: -0.01em; }
        .container { max-width: 550px; margin: 0 auto; padding: 0 5%; width: 100%; box-sizing: border-box; }
        .header-ax { padding: 40px 0 20px; text-align: center; }
        .status-box { display: inline-flex; align-items: center; background: rgba(16, 185, 129, 0.1); padding: 8px 16px; border-radius: 100px; margin-top: 15px; border: 1px solid rgba(16, 185, 129, 0.2); }
        .status-dot { height: 8px; width: 8px; background: #10B981; border-radius: 50%; margin-right: 10px; box-shadow: 0 0 12px #10B981; }
        .view-ax { display: none; flex-direction: column; width: 100%; gap: 20px; margin-top: 10px; }
        .card-ax { background: var(--card); border-radius: 24px; padding: 30px 25px; width: 100%; border: 1px solid var(--brd); box-sizing: border-box; }
        .bal-lbl { font-size: 0.7rem; text-transform: uppercase; letter-spacing: 2px; color: var(--txt-m); font-weight: 700; margin-bottom: 12px; display: block; }
        .bal-val { font-size: clamp(2rem, 8vw, 2.5rem); font-weight: 800; margin-bottom: 25px; letter-spacing: -1px; word-break: break-word; }
        .pill { background: rgba(0,0,0,0.5); padding: 15px; border-radius: 15px; font-family: 'JetBrains Mono', monospace; font-size: clamp(0.55rem, 2.2vw, 0.75rem); word-break: break-all; color: var(--txt-m); border: 1px solid var(--brd); line-height: 1.5; width: 100%; box-sizing: border-box; text-align: left; }
        .btn-ax { background: var(--txt); color: #000; border-radius: 15px; padding: 20px; font-weight: 800; border: none; width: 100%; font-size: 0.95rem; cursor: pointer; transition: 0.3s; display: flex; align-items: center; justify-content: center; gap: 10px; }
        .btn-ax:hover { background: #E4E4E7; transform: translateY(-2px); }
        .bottom-bar { background: rgba(2,2,2,0.85); backdrop-filter: blur(20px); position: fixed; bottom: 0; left: 0; width: 100%; height: 85px; display: flex; justify-content: space-around; align-items: center; border-top: 1px solid var(--brd); z-index: 1000; }
        .nav-l { color: #555; text-decoration: none; font-size: 0.6rem; font-weight: 800; display: flex; flex-direction: column; align-items: center; gap: 8px; cursor: pointer; text-transform: uppercase; flex: 1; }
        .nav-l.active { color: var(--txt); }
        .nav-l i { font-size: 1.3rem; }
        .input-ax { width: 100%; padding: 20px; border-radius: 15px; border: 1px solid var(--brd); background: rgba(0,0,0,0.5); color: #FFF; margin-bottom: 12px; box-sizing: border-box; font-family: inherit; font-size: 0.9rem; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header-ax">
            <a href="/"><img src="https://iili.io/qMGLM57.jpg" style="height:55px; mix-blend-mode:screen;" alt="AstraliX Core"></a><br>
            <div class="status-box"><span class="status-dot"></span><span style="font-size:0.65rem; font-weight:800; color:#10B981; letter-spacing:1px;">ALPHA TESTNET ACTIVE</span></div>
        </div>
        
        <div id="v-dash" class="view-ax" style="display:flex;">
            <div class="card-ax" style="border-color: rgba(59, 130, 246, 0.3); background: radial-gradient(100% 100% at 50% 0%, rgba(59, 130, 246, 0.05) 0%, rgba(255,255,255,0.02) 100%);">
                <span class="bal-lbl">Personal Ledger</span>
                <div id="bal-txt" class="bal-val">0.00 AX</div>
                <div id="addr-txt" class="pill" style="text-align:center;">VAULT LOCKED</div>
            </div>
            <div class="card-ax">
                <span class="bal-lbl">Reward Reserve</span>
                <div id="pool-txt" class="bal-val" style="color:#10B981;">0.00 AX</div>
                <div class="pill">AXf7ca3d5889ed99de642913af6c5630d6c491732b44180771cba042a4eb5a7109cc3ccde9e1a24d5315947415d5e592123ab90edcc4ea85415c1747fbe1684158</div>
            </div>
            <button class="btn-ax" onclick="mine()"><i class="fas fa-hammer"></i> VALIDATE NETWORK (+50 AX)</button>
        </div>

        <div id="v-wallet" class="view-ax">
            <div class="card-ax">
                <span class="bal-lbl">Secure Asset Transfer</span>
                <input type="text" id="tx-to" class="input-ax" placeholder="Recipient AX Address">
                <input type="number" id="tx-amt" class="input-ax" placeholder="Amount AX Tokens">
                <button class="btn-ax" onclick="send()"><i class="fas fa-paper-plane"></i> CONFIRM DEPOSIT</button>
            </div>
        </div>

        <div id="v-explorer" class="view-ax">
            <span class="bal-lbl" style="margin-top:10px;">Real-Time Explorer</span>
            <div id="block-list"></div>
        </div>

        <div id="v-holders" class="view-ax">
            <span class="bal-lbl" style="margin-top:10px;">Network Rich List</span>
            <div id="holders-list"></div>
        </div>

        <div id="v-sec" class="view-ax">
            <div class="card-ax">
                <span class="bal-lbl">Vault Security</span>
                <textarea id="i-seed" class="input-ax" style="height:120px; resize:none;" placeholder="Enter 24-word seed phrase..."></textarea>
                <button class="btn-ax" onclick="login()">RESTORE WALLET</button>
                <div style="text-align:center; margin: 25px 0; color:#444; font-size:0.7rem; font-weight:800; letter-spacing: 1px;">SECURE ENCRYPTION</div>
                <button class="btn-ax" style="background:transparent; border:1px solid rgba(255,255,255,0.1); color:var(--txt-m);" onclick="gen()">GENERATE IDENTITY</button>
                
                <div id="g-res" style="display:none; margin-top:25px;">
                    <button class="btn-ax" style="margin-bottom: 15px;" onclick="copySeed()"><i class="fas fa-copy"></i> COPY 24-WORD SEED</button>
                    
                    <div id="g-seed" style="display:grid; grid-template-columns:1fr 1fr; gap:8px;"></div>
                    <span class="bal-lbl" style="margin-top:25px;">Public Identity</span>
                    <div class="pill" id="g-pub"></div>
                </div>
            </div>
        </div>
    </div>

    <div class="bottom-bar">
        <a class="nav-l active" id="n-dash" onclick="nav('dash')"><i class="fas fa-chart-pie"></i>DASH</a>
        <a class="nav-l" id="n-wallet" onclick="nav('wallet')"><i class="fas fa-exchange-alt"></i>SEND</a>
        <a class="nav-l" id="n-explorer" onclick="nav('explorer')"><i class="fas fa-cubes"></i>CHAIN</a>
        <a class="nav-l" id="n-holders" onclick="nav('holders')"><i class="fas fa-users"></i>USERS</a>
        <a class="nav-l" id="n-sec" onclick="nav('sec')"><i class="fas fa-shield-halved"></i>VAULT</a>
    </div>

    <script>
        const words = ["alpha","bravo","cipher","delta","echo","falcon","ghost","hazard","iron","joker","knight","lunar","matrix","nexus","omega","phantom","quantum","radar","sigma","titan","ultra","vector","wolf","xray","yield","zenith","astral","block","chain","data","edge","fiber","grid","hash","index","joint","kern","link","mine","node","open","peer","root","seed","tech","unit","vault","web","zone"];
        const treasuryAddr = "AXf7ca3d5889ed99de642913af6c5630d6c491732b44180771cba042a4eb5a7109cc3ccde9e1a24d5315947415d5e592123ab90edcc4ea85415c1747fbe1684158";
        
        let currentGeneratedSeed = []; // VARIABLE GLOBAL PARA GUARDAR LA SEMILLA
        
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
            if(id === 'holders') renderHolders();
            window.scrollTo(0,0);
        }

        async function renderExplorer() {
            const r = await fetch("/api/chain"); const chain = await r.json();
            const list = document.getElementById("block-list"); let html = ""; const revChain = chain.reverse();
            for(let i=0; i<revChain.length; i++) {
                let b = revChain[i]; let idx = b.Index !== undefined ? b.Index : b.index;
                let hash = b.Hash || b.hash;
                html += '<div class="card-ax" style="padding:20px; margin-bottom:15px; border-radius:15px;"><span style="background:var(--txt); color:#000; padding:4px 10px; border-radius:6px; font-size:0.6rem; font-weight:800;">BLOCK #' + idx + '</span><div style="margin-top:15px; font-size:clamp(0.55rem, 2vw, 0.65rem); word-break:break-all; font-family:\'JetBrains Mono\', monospace; color:var(--txt-m); line-height:1.4;">' + hash + '</div></div>';
            }
            list.innerHTML = html;
        }

        async function renderHolders() {
            const r = await fetch("/api/holders"); const holders = await r.json();
            const list = document.getElementById("holders-list"); let html = ""; 
            if(holders && holders.length > 0) {
                for(let i=0; i<holders.length; i++) {
                    let h = holders[i];
                    let isTreasury = (h.address === treasuryAddr) ? '<span style="background:var(--txt); color:#000; padding:2px 6px; border-radius:4px; font-size:0.5rem; margin-left:8px; vertical-align:middle; font-weight:800;">TREASURY</span>' : '';
                    let isYou = (session && h.address === session.pub) ? '<span style="background:#10B981; color:#000; padding:2px 6px; border-radius:4px; font-size:0.5rem; margin-left:8px; vertical-align:middle; font-weight:800;">YOU</span>' : '';
                    html += '<div class="card-ax" style="padding:18px; margin-bottom:15px; border-radius:15px;"><div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:10px;"><div style="font-weight:800; font-size:0.9rem; color:#10B981;">#' + (i+1) + isTreasury + isYou + '</div><div style="font-weight:800; font-size:0.9rem; color:#FFF;">' + h.balance.toLocaleString() + ' AX</div></div><div style="font-size:0.55rem; word-break:break-all; font-family:\'JetBrains Mono\', monospace; color:var(--txt-m); line-height:1.4;">' + h.address + '</div></div>';
                }
            } else {
                html = "<p style='color:var(--txt-m); font-size:0.8rem;'>Syncing ledgers...</p>";
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
            currentGeneratedSeed = seed; // GUARDAMOS LA SEMILLA GENERADA
            const keys = await derive(seed.join(" ")); document.getElementById("g-res").style.display = "block";
            let sH = ""; for(let i=0; i<seed.length; i++) sH += '<div style="background:rgba(0,0,0,0.5); padding:8px 12px; border-radius:8px; border:1px solid var(--brd); font-size:0.7rem; color:var(--txt-m);"><span style="color:var(--txt); margin-right:5px; font-weight:800;">'+(i+1)+'</span> '+seed[i]+'</div>';
            document.getElementById("g-seed").innerHTML = sH; document.getElementById("g-pub").innerText = keys.pub;
        }

        // NUEVA FUNCIÓN PARA COPIAR AL PORTAPAPELES
        function copySeed() {
            if(currentGeneratedSeed.length > 0) {
                const seedString = currentGeneratedSeed.join(" ");
                navigator.clipboard.writeText(seedString).then(() => {
                    alert("24-word seed copied to clipboard! Keep it safe.");
                }).catch(err => {
                    alert("Error copying seed. Please copy manually.");
                });
            }
        }

        async function load() {
            if(session) {
                const r = await fetch("/api/balance/" + session.pub); const d = await r.json();
                document.getElementById("bal-txt").innerText = d.balance.toLocaleString() + " AX";
                document.getElementById("addr-txt").innerText = session.pub;
            }
            const rp = await fetch("/api/balance/" + treasuryAddr);
            const dp = await rp.json(); document.getElementById("pool-txt").innerText = dp.balance.toLocaleString() + " AX";
            
            if(document.getElementById("n-holders").classList.contains("active")) {
                renderHolders();
            }
        }

        async function mine() {
            if(!session) return alert("Vault Required. Please restore your identity first.");
            const r = await fetch("/api/mine?address=" + session.pub);
            
            if (r.status === 403) {
                alert("Forbidden: This address is blacklisted for security reasons.");
                return;
            }
            if (r.status === 401) {
                alert("Unauthorized: Minimum 500 AX required to validate blocks.");
                return;
            }
            
            if(r.ok) { alert("Block Validated! +50 AX"); load(); } else { alert("Mempool empty. Send a transaction first."); }
        }

        async function send() {
            const tx = { sender: session.pub, recipient: document.getElementById("tx-to").value, amount: parseFloat(document.getElementById("tx-amt").value) };
            const r = await fetch("/api/transactions/new", { method: "POST", body: JSON.stringify(tx) });
            if(r.ok) { 
                alert("Transaction Sent to Mempool!"); 
                nav('dash'); 
                load(); 
            } else if (r.status === 403) {
                alert("Forbidden: Transaction blocked by security protocols.");
            } else { 
                alert("Transaction Failed. Check Balance."); 
            }
        }
        
        load(); setInterval(load, 15000);
    </script>
</body>
</html>
`

// ==========================================
// 📄 WHITEPAPER ORIGINAL (NO TOCADO)
// ==========================================
const whitepaperHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Whitepaper | AstraliX Protocol</title>
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.2/css/all.min.css" rel="stylesheet">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;600;800&family=JetBrains+Mono:wght@400&display=swap');
        :root { --bg: #020202; --prim: #3B82F6; --txt: #FFFFFF; --txt-m: #8899A6; --brd: #1A1A1A; }
        body { font-family: 'Plus Jakarta Sans', sans-serif; background: var(--bg); color: var(--txt); line-height: 1.8; overflow-x: hidden; }
        .container { max-width: 850px; margin: 0 auto; padding: 60px 6%; }
        h1 { font-size: clamp(2.5rem, 6vw, 4rem); font-weight: 800; letter-spacing: -2px; margin-bottom: 20px; line-height: 1.1; }
        h2 { font-size: 1.8rem; font-weight: 800; margin: 60px 0 25px; color: var(--prim); border-bottom: 1px solid var(--brd); padding-bottom: 15px; }
        h3 { font-size: 1.3rem; color: #FFF; margin: 30px 0 15px; }
        p { margin-bottom: 20px; color: #CCC; font-size: 1.05rem; }
        ul { margin-bottom: 25px; padding-left: 20px; color: #CCC; font-size: 1.05rem; }
        li { margin-bottom: 10px; }
        .tech-box { background: #080808; border: 1px solid var(--brd); padding: 30px; border-radius: 15px; margin: 30px 0; font-family: 'JetBrains Mono'; font-size: 0.85rem; color: var(--txt-m); overflow-x: auto; }
        .quote { font-style: italic; border-left: 4px solid var(--prim); padding-left: 20px; margin: 40px 0; color: #FFF; font-size: 1.2rem; }
        footer { text-align: center; padding: 60px 0; border-top: 1px solid var(--brd); margin-top: 80px; color: var(--txt-m); font-size: 0.9rem; }
    </style>
</head>
<body>
    <div class="container">
        <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:40px;">
            <a href="/" style="color:var(--prim); text-decoration:none; font-weight:800; font-size:0.9rem;"><i class="fas fa-arrow-left"></i> BACK TO HOME</a>
            <img src="https://iili.io/qMGLM57.jpg" style="height:45px; mix-blend-mode:screen;" alt="AstraliX">
        </div>
        
        <h1>AstraliX Protocol: The 512-Bit Architecture</h1>
        <p style="font-size:0.9rem; color:var(--txt-m); text-transform:uppercase; letter-spacing:1px;">Version 1.0 (Alpha Genesis) • Lead Architect: Denis Waldemar • March 2026</p>
        
        <h2>Abstract</h2>
        <p>The AstraliX Protocol is a next-generation Layer 1 blockchain engineered to solve the impending crisis of cryptographic decay. As quantum computing advances at an exponential rate, legacy networks relying on 256-bit Elliptic Curve Digital Signature Algorithms (ECDSA) and SHA-256 hashing face an existential threat. By doubling the cryptographic bit-length to a 512-bit standard, AstraliX establishes a deterministic security moat that remains theoretically immune to both classical brute-force and quantum heuristic attacks.</p>

        <h2>1. The Quantum Threat & Shor's Algorithm</h2>
        <p>Currently, over $2 Trillion in digital assets are secured by 256-bit cryptography (Bitcoin, Ethereum, Solana). The security of these networks relies on the mathematical difficulty of the discrete logarithm problem.</p>
        <p>However, Shor’s algorithm, executed on a sufficiently powerful quantum computer (estimated at 4000+ stable qubits), can solve these problems in polynomial time. Once this threshold is crossed, any exposed public key on a 256-bit network can be reverse-engineered to reveal its private key, rendering the network entirely compromised.</p>
        
        <div class="quote">
            "AstraliX does not wait for the quantum threat to materialize. It pre-emptively neutralizes it by scaling the entropy of the network beyond the physical limits of computation."
        </div>

        <h2>2. The 512-Bit Cryptographic Leap</h2>
        <p>AstraliX mitigates quantum vulnerabilities through sheer mathematical volume. By utilizing <strong>SHA-512</strong> for state transitions, block hashing, and wallet derivation, the search space for potential collisions is expanded exponentially.</p>
        
        <div class="tech-box">
            // MATHEMATICAL ENTROPY COMPARISON<br><br>
            [Legacy 256-bit Standard]<br>
            Combinations: 2^256 ≈ 1.15 x 10^77<br>
            Vulnerability: High (Est. 2030-2035 with Quantum Tech)<br><br>
            [AstraliX 512-bit Standard]<br>
            Combinations: 2^512 ≈ 1.34 x 10^154<br>
            Status: Post-Quantum Immune
        </div>

        <h2>3. High-Concurrency Core Engine (Golang)</h2>
        <p>Heavy cryptography requires a robust, hyper-optimized execution environment. The AstraliX Core is written entirely in <strong>Go (Golang)</strong>.</p>
        <ul>
            <li><strong>Goroutines for Concurrency:</strong> Instead of heavy OS threads, the network uses Goroutines to handle thousands of simultaneous mempool transactions and peer-to-peer gossip protocols without memory bottlenecks.</li>
            <li><strong>Sub-Second Finality:</strong> The underlying consensus mechanism ensures that blocks are minted and verified rapidly, providing UX parity with centralized financial systems while maintaining pure decentralization.</li>
        </ul>

        <h2>4. The Zero-Trust Vault Protocol</h2>
        <p>A blockchain is only as secure as its weakest endpoint. AstraliX implements a strict "Zero-Trust" policy for node operators and end-users.</p>
        <p>When a user creates an identity, the 24-word mnemonic seed phrase undergoes a local, client-side SHA-512 derivation process. The generated 88-character Base64 private key and its corresponding 128-character hexadecimal public address never leave the local environment.</p>

        <h2>5. Network Tokenomics (Fair-Launch Distribution)</h2>
        <p>To ensure a sustainable and decentralized growth model, the AstraliX supply is capped at **1,000,002,021 AX**, distributed as follows:</p>
        <ul>
            <li><strong>Ecosystem Mining Rewards (40.0%):</strong> Emitted over a 10-year decay curve to reward network validators and maintain security.</li>
            <li><strong>Founder Nodes - Seed (12.5%):</strong> Initial allocation for network bootstrap and infrastructure funding.</li>
            <li><strong>Locked Liquidity Pool (15.0%):</strong> Provision for decentralized and centralized exchange depth, ensuring asset stability.</li>
            <li><strong>Treasury & R&D (15.0%):</strong> Managed by the Foundation for protocol upgrades, audits, and long-term research.</li>
            <li><strong>Marketing & Community (10.0%):</strong> Allocation for ecosystem quests, airdrops, and global adoption programs.</li>
            <li><strong>Team & Advisors (7.5%):</strong> Subject to a 24-month linear vesting schedule starting after Mainnet launch.</li>
        </ul>

        <h2>6. Roadmap to Mainnet (April 2026)</h2>
        <p>The network is currently operating on the Alpha Testnet, allowing Founder Nodes to validate blocks and interact with the Core Dashboard. On <strong>April 2026</strong>, the protocol will undergo a Hard Genesis event, officially launching the Mainnet. All Alpha ledger balances will be migrated 1:1 to the main network.</p>

        <footer>© 2026 AstraliX Foundation. Engineered in Argentina, Built for the World.</footer>
    </div>
</body>
</html>
`

// ==========================================
// 📄 TECHNICAL PAPER ACADÉMICO (ULTRA-TECH)
// ==========================================
const techPaperHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Technical Paper | AstraliX Protocol</title>
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Libre+Baskerville:ital,wght@0,400;0,700;1,400&family=Inter:wght@400;600;700&display=swap');
        body { background: #f9f9f9; color: #111; font-family: 'Libre Baskerville', serif; line-height: 1.8; padding: 80px 10%; max-width: 900px; margin: 0 auto; }
        .header { text-align: center; margin-bottom: 60px; border-bottom: 2px solid #eee; padding-bottom: 40px; }
        h1 { font-size: 2.5rem; margin-bottom: 15px; color: #000; letter-spacing: -0.02em; }
        .meta { font-family: 'Inter', sans-serif; font-size: 0.9rem; color: #555; text-transform: uppercase; letter-spacing: 1px; font-weight: 600; }
        h2 { font-size: 1.6rem; margin-top: 60px; border-left: 4px solid #3B82F6; padding-left: 20px; color: #000; font-family: 'Inter', sans-serif; letter-spacing: -0.02em; }
        h3 { font-family: 'Inter', sans-serif; font-size: 1.2rem; color: #333; margin-top: 30px; }
        p { text-align: justify; margin-bottom: 20px; font-size: 1.05rem; color: #333; }
        .math-block { background: #fff; border: 1px solid #e2e8f0; padding: 25px; border-radius: 8px; font-family: 'Courier New', monospace; overflow-x: auto; margin: 30px 0; box-shadow: 0 4px 6px rgba(0,0,0,0.02); font-size: 0.95rem; color: #1e293b; line-height: 1.6; }
        .code-inline { font-family: 'Courier New', monospace; background: #f1f5f9; padding: 2px 6px; border-radius: 4px; font-size: 0.9em; color: #0f172a; }
        .highlight-box { background: rgba(59, 130, 246, 0.05); border-left: 4px solid #3B82F6; padding: 20px; margin: 30px 0; font-family: 'Inter', sans-serif; font-size: 0.95rem; color: #1e293b; }
        footer { margin-top: 100px; padding-top: 40px; border-top: 1px solid #eee; font-family: 'Inter', sans-serif; font-size: 0.8rem; color: #999; text-align: center; }
        .back-link { position: fixed; top: 20px; left: 20px; font-family: 'Inter', sans-serif; text-decoration: none; color: #3B82F6; font-weight: 700; font-size: 0.8rem; background: #fff; padding: 8px 12px; border-radius: 6px; box-shadow: 0 2px 10px rgba(0,0,0,0.05); transition: 0.2s; }
        .back-link:hover { transform: translateY(-2px); box-shadow: 0 4px 15px rgba(0,0,0,0.1); }
    </style>
</head>
<body>
    <a href="/" class="back-link">← RETURN TO DASHBOARD</a>
    <div class="header">
        <h1>AstraliX Protocol: A 512-Bit Post-Quantum DePIN Architecture for Deterministic Consensus</h1>
        <div class="meta">Technical Specification v1.1.0 • Lead Architect: Denis W. • March 2026</div>
    </div>

    <section>
        <h2>Abstract</h2>
        <p>Current decentralized consensus networks rely heavily on 256-bit cryptographic primitives (e.g., ECDSA, SHA-256) which face an imminent existential threat from high-qubit quantum computation. The AstraliX Protocol introduces a paradigm shift by architecting a native 512-bit state machine written entirely in Golang, utilizing <span class="code-inline">goroutines</span> for lock-free mempool ingestion. Furthermore, AstraliX mitigates OSI Layer 1-3 vulnerabilities associated with virtualized hypervisors by enforcing a Decentralized Physical Infrastructure Network (DePIN) physically anchored in proprietary bare-metal Internet Service Provider (ISP) hardware.</p>
    </section>

    <section>
        <h2>1. Cryptographic Entropy & Post-Quantum Resilience</h2>
        <p>The vulnerability of symmetric search spaces to quantum brute-force methodologies is mathematically bounded by <strong>Grover's Algorithm</strong>. Grover's algorithm allows a quantum computer to search an unsorted database of N entries in O(&radic;N) time, effectively halving the bit-security of any symmetric cryptographic hash.</p>
        
        <div class="math-block">
            // GROVER'S BOUND ON LEGACY 256-BIT (BTC, ETH, SOL)<br>
            T &asymp; (&pi; / 4) * &radic;(N / M)<br>
            Effective Quantum Security = 256 / 2 = 128 bits.<br>
            Status: Vulnerable to near-term quantum decryption.
        </div>

        <p>By enforcing <strong>SHA-512</strong> as the foundational hash function for state root derivation and block consensus, AstraliX maintains an absolute minimum quantum-resistant threshold of 256 bits, preserving cryptographic sovereignty well beyond the next century of hardware advancements.</p>
        
        <div class="math-block">
            // ASTRALIX NATIVE 512-BIT ENTROPY POOL<br>
            Combinations: 2^512 &asymp; 1.34 &times; 10^154 states<br>
            Effective Quantum Security (Grover Mitigation) = 512 / 2 = 256 bits.<br>
            Status: Mathematically Immune.
        </div>
    </section>

    <section>
        <h2>2. Go-Native Deterministic Consensus & AX-BFT</h2>
        <p>The AstraliX virtual machine executes state transitions deterministically. To solve the blockchain trilemma (Scalability, Security, Decentralization) without compromising on the 512-bit computational overhead, the engine bypasses traditional OS-level thread blocking.</p>

        <h3>2.1 Lock-Free Mempool Ingestion</h3>
        <p>Transactions are ingested via unbuffered Go channels. Instead of executing computationally expensive mutex locks on the global state, <span class="code-inline">goroutines</span> handle signature verification asynchronously, appending verified structures to the mempool heap in O(1) time complexity.</p>

        <h3>2.2 Hybrid Stake-to-Validate (HPoS)</h3>
        <p>To mitigate Sybil vectors and ensure Byzantine Fault Tolerance (BFT), node operators must cryptographically prove a minimum <strong>500 AX</strong> balance to authorize state execution. The difficulty algorithm deterministically adjusts block time to maintain sub-second finality (latency &lt; 1.2s).</p>
    </section>

    <section>
        <h2>3. Zero-Knowledge Identity Vault</h2>
        <p>Client-side derivation is non-negotiable for a trustless architecture. The AstraliX Vault utilizes a Cryptographically Secure Pseudo-Random Number Generator (CSPRNG) to formulate a 24-word mnemonic. </p>
        
        <div class="highlight-box">
            The entropy seed undergoes a local PBKDF2 operation using HMAC-SHA512 with 2048 iterations. The resulting output yields an 88-character Base64 private key constraint and its deterministic 128-character hexadecimal public address (<span class="code-inline">AX...</span>). The private key never traverses the network payload.
        </div>
    </section>

    <section>
        <h2>4. The DePIN Bare-Metal Sublayer</h2>
        <p>Software security is irrelevant if the underlying physical hardware is compromised via corporate deplatforming or BGP route hijacking. Modern L1s rely heavily on Amazon Web Services (AWS) or Google Cloud, creating centralized points of failure at OSI Layers 1 through 3.</p>
        <p>AstraliX mandates the provisioning of Founder Nodes on proprietary, ISP-grade bare-metal racks. Direct fiber-optic backbones ensure sub-millisecond propagation of the genesis ledger across the peer network, completely decoupling the protocol from virtualized hypervisor vulnerabilities.</p>
    </section>

    <footer>
        © 2026 AstraliX Foundation. Engineered for the Post-Quantum Era.
    </footer>
</body>
</html>
`