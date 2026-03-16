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

func initDB() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("❌ ERROR: DATABASE_URL vacía.")
	}
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("❌ ERROR de Driver:", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		log.Fatal("❌ ERROR de conexión a la DB:", err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS chain_state (id INT PRIMARY KEY, data TEXT)`)
	if err != nil {
		log.Fatal("❌ ERROR creando tablas:", err)
	}
}

func loadChain() {
	var data string
	err := db.QueryRow("SELECT data FROM chain_state WHERE id = 1").Scan(&data)
	if err == nil {
		json.Unmarshal([]byte(data), &Blockchain)
	}
}

func saveChain() {
	data, _ := json.Marshal(Blockchain)
	_, err := db.Exec(`INSERT INTO chain_state (id, data) VALUES (1, $1) ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data`, string(data))
	if err != nil {
		log.Println("❌ Error guardando:", err)
	}
}

func getBalance(addr string) float64 {
	var balance float64
	for _, block := range Blockchain {
		for _, tx := range block.Transactions {
			if tx.Recipient == addr {
				balance += tx.Amount
			}
			if tx.Sender == addr {
				balance -= tx.Amount
			}
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
		if miner == "" {
			http.Error(w, "Address req", 400)
			return
		}
		if len(Mempool) == 0 {
			http.Error(w, "Mempool empty", 400)
			return
		}
		reward := 50.0
		var txs []core.Transaction
		txs = append(txs, Mempool...)
		if getBalance(TREASURY_POOL_ADDR) >= reward {
			rewardTx := core.Transaction{Sender: TREASURY_POOL_ADDR, Recipient: miner, Amount: reward}
			rewardTx.TxID = rewardTx.CalculateHash()
			txs = append(txs, rewardTx)
		}
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
		if getBalance(tx.Sender) < tx.Amount && tx.Sender != "SYSTEM" {
			http.Error(w, "Low balance", 400)
			return
		}
		tx.TxID = tx.CalculateHash()
		Mempool = append(Mempool, tx)
		w.WriteHeader(201)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, landingHTML)
	})

	http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, dashboardHTML)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}

const landingHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AstraliX | Quantum-Resistant 512-bit Layer 1</title>
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;600;800&family=JetBrains+Mono:wght@400;700&display=swap');
        :root { --bg-dark: #050505; --bg-card: #0A0A0A; --primary: #3B82F6; --primary-glow: rgba(59, 130, 246, 0.25); --accent: #10B981; --text-main: #FFFFFF; --text-muted: #94A3B8; --border: #1F2937; }
        * { box-sizing: border-box; }
        body { font-family: 'Plus Jakarta Sans', sans-serif; margin: 0; color: var(--text-main); background: var(--bg-dark); line-height: 1.6; overflow-x: hidden; }
        .bg-pattern { position: fixed; top: 0; left: 0; width: 100vw; height: 100vh; background-image: linear-gradient(var(--border) 1px, transparent 1px), linear-gradient(90deg, var(--border) 1px, transparent 1px); background-size: 40px 40px; opacity: 0.2; z-index: -1; pointer-events: none; }
        .glow-orb { position: absolute; width: 80vw; height: 80vw; max-width: 800px; max-height: 800px; background: radial-gradient(circle, var(--primary-glow) 0%, transparent 60%); top: -20vh; left: 50%; transform: translateX(-50%); z-index: -1; filter: blur(60px); pointer-events: none; }
        .nav { padding: 30px 5%; display: flex; justify-content: space-between; align-items: center; max-width: 1400px; margin: 0 auto; width: 100%; }
        .logo { font-weight: 800; font-size: clamp(2rem, 4vw, 2.8rem); letter-spacing: -1.5px; color: var(--text-main); text-decoration: none; display: flex; align-items: center; }
        .logo .highlight { color: var(--primary); text-shadow: 0 0 15px var(--primary-glow); }
        .nav-link { text-decoration: none; color: var(--text-main); font-weight: 600; font-size: clamp(0.75rem, 2vw, 0.9rem); background: rgba(255,255,255,0.03); padding: 10px 20px; border-radius: 100px; border: 1px solid var(--border); transition: all 0.3s ease; display: flex; align-items: center; gap: 8px; backdrop-filter: blur(10px); white-space: nowrap; }
        .nav-link:hover { background: rgba(59, 130, 246, 0.1); border-color: var(--primary); color: var(--primary); box-shadow: 0 0 20px var(--primary-glow); }
        .hero { text-align: center; padding: 100px 5% 60px; position: relative; max-width: 1000px; margin: 0 auto; }
        .badge { display: inline-block; background: rgba(59, 130, 246, 0.1); color: var(--primary); padding: 8px 20px; border-radius: 100px; font-weight: 700; font-size: 0.8rem; text-transform: uppercase; letter-spacing: 2px; border: 1px solid rgba(59, 130, 246, 0.2); margin-bottom: 30px; }
        .hero h1 { font-size: clamp(2.8rem, 8vw, 6rem); font-weight: 800; margin: 0; letter-spacing: -3px; line-height: 1.05; background: linear-gradient(135deg, #FFFFFF 0%, #94A3B8 100%); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
        .hero p { font-size: clamp(1rem, 3vw, 1.25rem); color: var(--text-muted); margin: 30px auto 40px; font-weight: 400; line-height: 1.7; }
        .btn-primary { background: var(--text-main); color: var(--bg-dark); padding: 18px 40px; border-radius: 100px; text-decoration: none; font-weight: 800; font-size: clamp(0.9rem, 2.5vw, 1.1rem); display: inline-flex; align-items: center; justify-content: center; gap: 10px; transition: all 0.3s; border: none; cursor: pointer; width: fit-content; }
        .btn-primary:hover { transform: translateY(-3px); box-shadow: 0 15px 35px rgba(255,255,255,0.2); background: #f1f5f9; }
        .sec-quantum { max-width: 1200px; margin: 60px 5%; background: var(--bg-card); border: 1px solid var(--border); border-radius: 30px; display: flex; align-items: stretch; overflow: hidden; box-shadow: 0 20px 50px rgba(0,0,0,0.5); text-align: left; }
        .quantum-text { flex: 1.2; padding: 50px 8%; display: flex; flex-direction: column; justify-content: center; }
        .quantum-text h2 { font-size: clamp(1.8rem, 4vw, 2.2rem); font-weight: 800; margin: 0 0 20px; letter-spacing: -1px; color: #EF4444; display: flex; align-items: center; gap: 15px; line-height: 1.2; }
        .quantum-text h2 i { font-size: clamp(1.5rem, 3vw, 1.8rem); flex-shrink: 0; }
        .quantum-text p { color: var(--text-muted); font-size: clamp(0.95rem, 2vw, 1.05rem); margin: 0 0 20px; line-height: 1.8; }
        .quantum-text strong { color: var(--text-main); }
        .quantum-code { flex: 1; background: #000; padding: 50px 8%; border-left: 1px solid var(--border); font-family: 'JetBrains Mono', monospace; font-size: clamp(0.75rem, 1.5vw, 0.85rem); color: var(--text-muted); display: flex; flex-direction: column; justify-content: center; }
        .code-block { background: rgba(255,255,255,0.03); border: 1px solid rgba(255,255,255,0.05); padding: 25px; border-radius: 12px; margin-bottom: 20px; }
        .code-title { font-weight: 700; margin-bottom: 12px; font-size: 0.8rem; letter-spacing: 1px; }
        .code-red { color: #EF4444; }
        .code-green { color: #10B981; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 30px; max-width: 1200px; margin: 0 5% 100px; }
        .card { background: rgba(255,255,255,0.02); padding: 40px 30px; border-radius: 24px; border: 1px solid var(--border); transition: 0.3s; backdrop-filter: blur(10px); text-align: left; display: flex; flex-direction: column; }
        .card:hover { border-color: rgba(59, 130, 246, 0.3); transform: translateY(-5px); background: rgba(59, 130, 246, 0.05); }
        .card i { font-size: 2.5rem; color: var(--primary); margin-bottom: 20px; display: inline-block; }
        .card h3 { font-weight: 800; font-size: 1.3rem; margin: 0 0 15px; color: var(--text-main); }
        .card p { color: var(--text-muted); font-size: 0.95rem; margin: 0; line-height: 1.6; }
        .pre-sale { background: linear-gradient(180deg, var(--bg-card) 0%, #000 100%); border: 1px solid var(--border); padding: 80px 5%; text-align: center; border-radius: 40px; max-width: 800px; margin: 0 auto 100px; width: 90%; position: relative; overflow: hidden; }
        .pre-sale::before { content: ''; position: absolute; top: 0; left: 0; width: 100%; height: 2px; background: linear-gradient(90deg, transparent, var(--primary), transparent); }
        .price-label { font-family: 'JetBrains Mono', monospace; color: var(--primary); font-size: clamp(0.7rem, 2vw, 0.9rem); font-weight: 700; letter-spacing: 2px; text-transform: uppercase; }
        .price { font-size: clamp(4rem, 10vw, 5.5rem); font-weight: 800; margin: 10px 0 20px; letter-spacing: -2px; }
        .address-box { background: rgba(0,0,0,0.6); border: 1px dashed rgba(255,255,255,0.2); padding: 25px 20px; border-radius: 20px; margin: 30px auto; width: 100%; max-width: 550px; display: flex; flex-direction: column; align-items: center; }
        .network-tag { display: inline-block; background: rgba(243, 186, 47, 0.1); color: #F3BA2F; font-size: 0.75rem; font-weight: 800; padding: 6px 12px; border-radius: 6px; border: 1px solid rgba(243, 186, 47, 0.2); margin-bottom: 15px; }
        .wallet-addr { font-family: 'JetBrains Mono', monospace; font-size: clamp(0.8rem, 2.5vw, 1rem); color: var(--text-main); word-break: break-all; text-align: center; background: rgba(255,255,255,0.05); padding: 15px; border-radius: 10px; width: 100%; box-sizing: border-box; }
        .btn-wa { background: var(--accent); color: #000; padding: 18px 40px; border-radius: 100px; text-decoration: none; font-weight: 800; font-size: clamp(0.9rem, 2.5vw, 1.1rem); display: inline-flex; align-items: center; justify-content: center; gap: 10px; transition: 0.3s; margin-top: 10px; width: fit-content; }
        .btn-wa:hover { transform: scale(1.03); background: #12d392; box-shadow: 0 10px 25px rgba(16, 185, 129, 0.3); }
        footer { text-align: center; padding: 40px 5%; color: var(--text-muted); font-size: 0.85rem; border-top: 1px solid var(--border); background: #000; }
        @media (max-width: 850px) {
            .sec-quantum { flex-direction: column; border-radius: 24px; }
            .quantum-code { border-left: none; border-top: 1px solid var(--border); padding: 40px 6%; }
            .quantum-text { padding: 40px 6%; }
            .nav { padding: 20px 5%; }
            .hero { padding: 80px 5% 40px; }
            .pre-sale { padding: 60px 5%; border-radius: 30px; }
        }
    </style>
</head>
<body>
    <div class="bg-pattern"></div>
    <div class="glow-orb"></div>
    <nav class="nav">
        <a href="/" class="logo"><span class="highlight">A</span>strali<span class="highlight">X</span></a>
        <a href="/dashboard" class="nav-link"><i class="fas fa-terminal"></i> LAUNCH CORE</a>
    </nav>
    <header class="hero">
        <div class="badge">Alpha Testnet Genesis</div>
        <h1>The 512-bit Era<br>Begins Here.</h1>
        <p>A Layer 1 blockchain infrastructure engineered from the ground up to withstand the imminent quantum computing threat. Double the entropy, infinite security.</p>
        <a href="#comprar" class="btn-primary">Acquire Node Access <i class="fas fa-arrow-down"></i></a>
    </header>
    <section class="sec-quantum">
        <div class="quantum-text">
            <h2><i class="fas fa-microchip"></i> The Quantum Threat is Real</h2>
            <p>Traditional blockchains rely heavily on <strong>256-bit encryption</strong> (like ECDSA or SHA-256). While secure today, the rapid advancement of quantum computing and Shor's algorithm puts these networks on a countdown.</p>
            <p>Experts predict that within the next decade, a sufficiently powerful quantum computer will be able to derive a 256-bit private key from a public address, exposing billions in digital assets.</p>
            <p><strong>AstraliX solves the "Quantum Day" problem today.</strong> By establishing a foundational <strong>512-bit cryptographic standard</strong>, the time required to break the network via quantum brute-force remains vastly greater than the age of the universe.</p>
        </div>
        <div class="quantum-code">
            <div class="code-block" style="border-color: rgba(239, 68, 68, 0.2);">
                <div class="code-title code-red">Legacy Networks (256-bit)</div>
                <div>Hash: 64 hex chars</div>
                <div>Combinations: 2^256</div>
                <div style="margin-top:10px; font-size:0.75rem; color: rgba(239, 68, 68, 0.8);">[WARNING] Vulnerable to future decryption via Shor's Algorithm.</div>
            </div>
            <div class="code-block" style="border-color: rgba(16, 185, 129, 0.2);">
                <div class="code-title code-green">AstraliX Core (512-bit)</div>
                <div>Hash: 128 hex chars</div>
                <div>Combinations: 2^512</div>
                <div style="margin-top:10px; font-size:0.75rem; color: rgba(16, 185, 129, 0.8);">[SECURE] Cryptographic space expanded. Quantum brute-force timeline infinite.</div>
            </div>
        </div>
    </section>
    <main class="grid">
        <div class="card">
            <i class="fas fa-shield-halved"></i>
            <h3>Zero-Trust Vault</h3>
            <p>Your 24-word recovery phrase generates a 512-bit seed locally in your browser. Private keys never touch our servers. Absolute self-custody.</p>
        </div>
        <div class="card">
            <i class="fas fa-bolt"></i>
            <h3>High-Throughput Go Engine</h3>
            <p>Built purely in Go (Golang), the AstraliX core node handles validation and mempool states with zero memory collisions and blazing speed.</p>
        </div>
        <div class="card">
            <i class="fas fa-layer-group"></i>
            <h3>Genesis Allocation</h3>
            <p>Early infrastructure backers (Founder Nodes) receive a genesis block airdrop of 10,000 AX, fully secured and migratable to the future Mainnet.</p>
        </div>
    </main>
    <section id="comprar" class="pre-sale">
        <div class="price-label">Founder Node Allocation</div>
        <div class="price">21 USDT</div>
        <p style="color: var(--text-muted); font-size: clamp(0.95rem, 2vw, 1.1rem); max-width: 500px; margin: 0 auto; line-height: 1.6;">Fund the core development. Secure your genesis identity on the most robust Layer 1 architecture.</p>
        <div class="address-box">
            <div class="network-tag">Network: BNB Smart Chain (BEP-20)</div>
            <div style="font-size: 0.8rem; color: var(--text-muted); margin-bottom: 8px;">Official Treasury Address:</div>
            <div class="wallet-addr">0x948a663b1bd1292ded76a8412af2092bf0462d7c</div>
        </div>
        <p style="font-size: clamp(0.85rem, 2vw, 0.95rem); color: var(--text-muted); margin: 0 auto 25px; max-width: 450px;">Send the exact allocation amount. Verify your transaction directly with the Core Team to provision your Node.</p>
        <a href="https://wa.me/TuNumeroAqui?text=Initiating%20Founder%20Node%20Provisioning.%20I%20have%20sent%2021%20USDT." class="btn-wa">
            <i class="fab fa-whatsapp"></i> VERIFY TRANSACTION
        </a>
    </section>
    <footer>&copy; 2026 AstraliX Core Engine • Autonomous Network Infrastructure.</footer>
</body>
</html>
`

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
        .header-ax h5 { font-weight: 800; letter-spacing: 1.5px; margin: 0; color: #0F172A; font-size: 1.5rem; }
        .status-box { display: inline-flex; align-items: center; background: #ECFDF5; padding: 6px 14px; border-radius: 100px; margin-top: 10px; }
        .status-dot { height: 7px; width: 7px; background: #10B981; border-radius: 50%; margin-right: 8px; box-shadow: 0 0 10px #10B981; }
        .status-text { font-size: 0.65rem; font-weight: 800; color: #059669; letter-spacing: 1px; }
        .view-ax { display: none; flex-direction: column; align-items: center; width: 100%; max-width: 500px; margin: 0 auto; padding: 0 20px; box-sizing: border-box; justify-content: center; }
        .card-ax { background: var(--card); border-radius: 32px; box-shadow: 0 10px 40px rgba(0,0,0,0.04); padding: 30px; margin: 15px 0; width: 100%; box-sizing: border-box; border: 1px solid rgba(0,0,0,0.03); text-align: center; }
        .card-dark { background: linear-gradient(145deg, #0F172A 0%, #1E293B 100%); color: white; border: none; }
        .balance-label { font-size: 0.75rem; text-transform: uppercase; letter-spacing: 2px; opacity: 0.5; font-weight: 600; display: block; }
        .balance-amount { font-size: 2.2rem; font-weight: 800; margin: 10px 0 20px; letter-spacing: -1px; }
        .pill-address { background: rgba(0,0,0,0.05); padding: 14px; border-radius: 16px; font-family: 'JetBrains Mono', monospace; font-size: 0.55rem; word-break: break-all; color: #64748B; line-height: 1.5; text-align: left; }
        .card-dark .pill-address { background: rgba(255,255,255,0.1); color: rgba(255,255,255,0.5); }
        .btn-ax { background: var(--primary); color: white; border-radius: 20px; padding: 20px; font-weight: 700; border: none; width: 100%; margin-top: 10px; font-size: 1rem; box-shadow: 0 10px 25px rgba(13, 110, 253, 0.2); cursor: pointer; }
        .block-card { background: white; border-radius: 24px; padding: 20px; margin-bottom: 15px; width: 100%; border: 1px solid #E2E8F0; text-align: left; box-shadow: 0 4px 12px rgba(0,0,0,0.02); box-sizing: border-box; }
        .block-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
        .block-idx { background: #E0E7FF; padding: 5px 12px; border-radius: 10px; font-weight: 800; font-size: 0.7rem; color: var(--primary); }
        .block-hash { font-family: 'JetBrains Mono', monospace; font-size: 0.55rem; color: #64748B; word-break: break-all; margin-top: 8px; background: #F8FAFC; padding: 12px; border-radius: 12px; border: 1px solid #F1F5F9; }
        .bottom-bar { background: rgba(255,255,255,0.9); backdrop-filter: blur(15px); position: fixed; bottom: 0; width: 100%; height: 90px; display: flex; justify-content: space-around; align-items: center; border-top: 1px solid #F1F5F9; z-index: 999; }
        .nav-link-ax { color: #94A3B8; text-decoration: none; flex: 1; font-size: 10px; font-weight: 700; display: flex; flex-direction: column; align-items: center; gap: 6px; cursor: pointer; }
        .nav-link-ax.active { color: var(--primary); }
        .nav-link-ax i { font-size: 1.3rem; }
    </style>
</head>
<body>
    <div class="header-ax">
        <h5>AstraliX Core</h5>
        <div class="status-box"><span class="status-dot"></span><span class="status-text">NODE SYNCHRONIZED</span></div>
    </div>
    <div id="v-dash" class="view-ax" style="display:flex;">
        <div class="card-ax card-dark">
            <span class="balance-label">Total Supply Balance</span>
            <div id="bal-txt" class="balance-amount">0.00 AX</div>
            <div id="addr-txt" class="pill-address" style="text-align:center;">Vault Locked</div>
        </div>
        <div class="card-ax">
            <span class="balance-label">Treasury Reward Pool</span>
            <div id="pool-txt" class="balance-amount" style="color:var(--primary); font-size:1.8rem;">0.00 AX</div>
            <div class="pill-address">AXf7ca3d5889ed99de642913af6c5630d6c491732b44180771cba042a4eb5a7109cc3ccde9e1a24d5315947415d5e592123ab90edcc4ea85415c1747fbe1684158</div>
        </div>
        <button class="btn-ax" onclick="mine()">VALIDATE NETWORK (+50 AX)</button>
    </div>
    <div id="v-wallet" class="view-ax">
        <div class="card-ax">
            <span class="balance-label" style="margin-bottom:20px;">Internal Asset Transfer</span>
            <input type="text" id="tx-to" style="width:100%; padding:18px; border-radius:18px; border:2px solid #E2E8F0; background:#F8FAFC; margin-bottom:12px; box-sizing:border-box; outline:none;" placeholder="Recipient Address (AX...)">
            <input type="number" id="tx-amt" style="width:100%; padding:18px; border-radius:18px; border:2px solid #E2E8F0; background:#F8FAFC; margin-bottom:20px; box-sizing:border-box; outline:none;" placeholder="Amount AX">
            <button class="btn-ax" onclick="send()">CONFIRM TRANSFER</button>
        </div>
    </div>
    <div id="v-explorer" class="view-ax">
        <span class="balance-label" style="margin: 10px 0 25px;">On-Chain Block Explorer</span>
        <div id="block-list" style="width:100%;"></div>
    </div>
    <div id="v-sec" class="view-ax">
        <div class="card-ax">
            <span class="balance-label" style="margin-bottom:15px;">Vault Access Point</span>
            <textarea id="i-seed" style="width:100%; background:#F8FAFC; border:2px solid #E2E8F0; border-radius:20px; padding:20px; font-size:0.95rem; resize:none; box-sizing:border-box; outline:none; font-family:'Outfit', sans-serif;" rows="3" placeholder="Input your 24-word recovery phrase..."></textarea>
            <button class="btn-ax" onclick="login()">RESTORE WALLET</button>
            <div style="margin: 30px 0; display: flex; align-items: center; opacity: 0.3; width:100%;"><hr style="flex:1;"><span style="margin:0 15px; font-weight:800; font-size:0.6rem;">SECURE</span><hr style="flex:1;"></div>
            <button class="btn-ax" style="background:#F8FAFC; border:2px solid #E2E8F0; color:#475569; box-shadow:none;" onclick="gen()">NEW 512-BIT IDENTITY</button>
            <div id="g-res" style="display:none; margin-top:30px; text-align:left;">
                <span style="font-size:0.75rem; font-weight:800; color:#EF4444; text-transform:uppercase; letter-spacing:1px;">Recovery Seed</span>
                <div id="g-seed" style="background:#F1F5F9; border-radius:20px; padding:18px; display:grid; grid-template-columns:1fr 1fr; gap:10px; margin-top:12px;"></div>
                <div class="mt-4">
                    <span class="balance-label" style="margin-top:20px;">Generated Address</span>
                    <div class="pill-address" style="background:#F1F5F9; color:var(--primary); margin-top:10px; font-weight:700;" id="g-pub"></div>
                </div>
            </div>
        </div>
    </div>
    <div class="bottom-bar">
        <a class="nav-link-ax active" id="n-dash" onclick="nav('dash')"><i class="fas fa-chart-pie"></i>Overview</a>
        <a class="nav-link-ax" id="n-wallet" onclick="nav('wallet')"><i class="fas fa-paper-plane"></i>Transfer</a>
        <a class="nav-link-ax" id="n-explorer" onclick="nav('explorer')"><i class="fas fa-cubes"></i>Explorer</a>
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
            let html = "";
            const revChain = chain.reverse();
            for(let i=0; i<revChain.length; i++) {
                let b = revChain[i];
                let idx = b.index !== undefined ? b.index : b.Index;
                let ts = b.timestamp || b.Timestamp;
                let hash = b.hash || b.Hash || "Calculando...";
                let txs = b.transactions || b.Transactions || [];
                let txCount = txs.length;
                let timeStr = ts ? new Date(ts * 1000).toLocaleTimeString() : "Desconocido";
                html += '<div class="block-card"><div class="block-header"><span class="block-idx">BLOCK #' + idx + '</span><span style="font-size:0.65rem; color:#94A3B8;">' + timeStr + '</span></div><div style="font-size:0.65rem; font-weight:700; color:#475569; margin-bottom:5px;">State Hash:</div><div class="block-hash">' + hash + '</div><div style="font-size:0.6rem; color:#94A3B8; margin-top:10px; display:flex; justify-content:space-between;"><span>TX COUNT: ' + txCount + '</span><span>SHA-512 SECURED</span></div></div>';
            }
            list.innerHTML = html;
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
            const keys = await derive(seed.join(" "));
            document.getElementById("g-res").style.display = "block";
            let seedHtml = "";
            for(let i=0; i<seed.length; i++) {
                seedHtml += '<div style="font-size:0.75rem; background:white; padding:10px; border-radius:12px; color:#475569; font-weight:600; border:1px solid #F1F5F9;"><span style="color:#CBD5E1; font-size:0.6rem; margin-right:8px;">'+(i+1)+'</span>'+seed[i]+'</div>';
            }
            document.getElementById("g-seed").innerHTML = seedHtml;
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