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
// ⚙️ MOTOR BLOCKCHAIN (INTACTO)
// ==========================================

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

	// ==========================================
	// 🔌 RUTAS DE LA API (INTACTAS)
	// ==========================================

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

	// ==========================================
	// 🌐 RUTAS WEB (FRONTEND)
	// ==========================================

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

// ==========================================
// 🎨 CÓDIGO HTML/CSS CONSTANTES
// ==========================================

const landingHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AstraliX | Next-Gen 512-bit Security</title>
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;600;800&family=JetBrains+Mono:wght@400;700&display=swap');
        
        :root { 
            --bg-dark: #050505; 
            --bg-card: #111111; 
            --primary: #3B82F6; 
            --primary-glow: rgba(59, 130, 246, 0.4);
            --accent: #10B981; 
            --text-main: #FFFFFF; 
            --text-muted: #94A3B8;
            --border: #222222;
        }
        
        * { box-sizing: border-box; }
        body { font-family: 'Plus Jakarta Sans', sans-serif; margin: 0; color: var(--text-main); background: var(--bg-dark); line-height: 1.6; scroll-behavior: smooth; overflow-x: hidden; }
        
        /* Grid Background Pattern */
        .bg-pattern { position: fixed; top: 0; left: 0; width: 100vw; height: 100vh; background-image: linear-gradient(var(--border) 1px, transparent 1px), linear-gradient(90deg, var(--border) 1px, transparent 1px); background-size: 50px 50px; opacity: 0.15; z-index: -1; pointer-events: none; }
        .glow-orb { position: absolute; width: 600px; height: 600px; background: radial-gradient(circle, var(--primary-glow) 0%, transparent 70%); top: -200px; left: 50%; transform: translateX(-50%); z-index: -1; filter: blur(50px); pointer-events: none; }

        /* Navbar */
        .nav { padding: 30px 40px; display: flex; justify-content: space-between; align-items: center; max-width: 1400px; margin: 0 auto; }
        .logo { font-weight: 800; font-size: 2rem; letter-spacing: -1px; color: var(--text-main); text-decoration: none; display: flex; align-items: center; gap: 10px; }
        .logo span { color: var(--primary); }
        .nav-link { text-decoration: none; color: var(--text-main); font-weight: 600; font-size: 0.95rem; background: rgba(255,255,255,0.05); padding: 12px 28px; border-radius: 100px; border: 1px solid var(--border); transition: all 0.3s ease; display: flex; align-items: center; gap: 8px; }
        .nav-link:hover { background: var(--primary); border-color: var(--primary); box-shadow: 0 0 20px var(--primary-glow); }

        /* Hero Section */
        .hero { text-align: center; padding: 120px 20px 80px; position: relative; }
        .badge { display: inline-block; background: rgba(59, 130, 246, 0.1); color: var(--primary); padding: 8px 20px; border-radius: 100px; font-weight: 700; font-size: 0.8rem; text-transform: uppercase; letter-spacing: 2px; border: 1px solid rgba(59, 130, 246, 0.2); margin-bottom: 30px; }
        .hero h1 { font-size: clamp(3rem, 6vw, 5.5rem); font-weight: 800; margin: 0; letter-spacing: -2px; line-height: 1.1; background: linear-gradient(to right, #fff, #94A3B8); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
        .hero p { font-size: 1.25rem; color: var(--text-muted); max-width: 650px; margin: 30px auto 50px; font-weight: 400; }
        .btn-primary { background: var(--primary); color: white; padding: 20px 45px; border-radius: 100px; text-decoration: none; font-weight: 700; font-size: 1.1rem; display: inline-flex; align-items: center; gap: 10px; transition: all 0.3s; border: none; cursor: pointer; }
        .btn-primary:hover { transform: translateY(-2px); box-shadow: 0 10px 30px var(--primary-glow); }

        /* Security Feature Section */
        .sec-feature { max-width: 1200px; margin: 60px auto; background: var(--bg-card); border: 1px solid var(--border); border-radius: 30px; padding: 60px; display: flex; align-items: center; gap: 50px; position: relative; overflow: hidden; }
        .sec-content { flex: 1; }
        .sec-content h2 { font-size: 2.5rem; font-weight: 800; margin-top: 0; margin-bottom: 20px; letter-spacing: -1px; }
        .sec-content p { color: var(--text-muted); font-size: 1.1rem; margin-bottom: 20px; line-height: 1.7; }
        .sec-visual { flex: 1; background: #000; border-radius: 20px; padding: 30px; border: 1px solid var(--border); font-family: 'JetBrains Mono', monospace; font-size: 0.85rem; color: var(--primary); position: relative; }
        .sec-visual::before { content: 'SHA-512 HASH GENERATION'; position: absolute; top: -12px; left: 20px; background: var(--bg-dark); padding: 0 10px; font-size: 0.7rem; font-weight: 700; color: var(--text-muted); }
        .hash-line { margin: 8px 0; opacity: 0.7; word-break: break-all; }
        .hash-line:nth-child(1) { color: #fff; opacity: 1; }

        /* Grid Cards */
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 30px; max-width: 1200px; margin: 0 auto 100px; padding: 0 20px; }
        .card { background: var(--bg-card); padding: 45px; border-radius: 24px; border: 1px solid var(--border); transition: 0.3s; position: relative; overflow: hidden; }
        .card:hover { border-color: rgba(59, 130, 246, 0.4); transform: translateY(-5px); }
        .card i { font-size: 2.5rem; color: var(--primary); margin-bottom: 25px; }
        .card h3 { font-weight: 800; font-size: 1.5rem; margin: 0 0 15px; }
        .card p { color: var(--text-muted); font-size: 1rem; margin: 0; line-height: 1.6; }

        /* Pre-Sale Box */
        .pre-sale { background: linear-gradient(180deg, var(--bg-card) 0%, #000 100%); border: 1px solid var(--border); padding: 80px 40px; text-align: center; border-radius: 40px; max-width: 800px; margin: 0 auto 100px; position: relative; }
        .price-label { font-family: 'JetBrains Mono', monospace; color: var(--primary); font-size: 0.9rem; font-weight: 700; letter-spacing: 2px; }
        .price { font-size: 5rem; font-weight: 800; margin: 10px 0 20px; letter-spacing: -2px; }
        .address-box { background: rgba(0,0,0,0.5); border: 1px solid rgba(255,255,255,0.1); padding: 25px; border-radius: 16px; margin: 30px 0; display: inline-block; width: 100%; max-width: 500px; }
        .wallet-addr { font-family: 'JetBrains Mono', monospace; font-size: 0.95rem; color: var(--text-main); word-break: break-all; }
        .network-tag { display: inline-block; background: #F3BA2F; color: #000; font-size: 0.7rem; font-weight: 800; padding: 4px 10px; border-radius: 6px; margin-bottom: 15px; }
        
        .btn-wa { background: var(--accent); color: #000; padding: 20px 40px; border-radius: 100px; text-decoration: none; font-weight: 800; font-size: 1.1rem; display: inline-flex; align-items: center; gap: 10px; transition: 0.3s; margin-top: 10px; }
        .btn-wa:hover { transform: scale(1.05); background: #12d392; box-shadow: 0 0 20px rgba(16, 185, 129, 0.3); }

        footer { text-align: center; padding: 40px 20px; color: var(--text-muted); font-size: 0.9rem; border-top: 1px solid var(--border); }
        
        @media (max-width: 900px) {
            .sec-feature { flex-direction: column; padding: 40px 20px; }
            .sec-visual { width: 100%; }
        }
    </style>
</head>
<body>
    <div class="bg-pattern"></div>
    <div class="glow-orb"></div>

    <nav class="nav">
        <a href="/" class="logo">AstraliX<span></span></a>
        <a href="/dashboard" class="nav-link"><i class="fas fa-terminal"></i> LAUNCH CORE</a>
    </nav>

    <header class="hero">
        <div class="badge">Alpha Testnet Live</div>
        <h1>Unbreakable Protocol.<br>Layer 1 Innovation.</h1>
        <p>AstraliX isn't just a network; it's a cryptographic fortress. Powered by 512-bit security architecture, designed for speed, and built for the future of digital assets.</p>
        <a href="#comprar" class="btn-primary">Become a Founder Node <i class="fas fa-arrow-right"></i></a>
    </header>

    <section class="sec-feature">
        <div class="sec-content">
            <h2>The Power of 512-bits</h2>
            <p>While industry standards rely on 256-bit encryption, AstraliX doubles the cryptographic complexity. Every wallet generation and transaction is secured using the ultra-robust <strong>SHA-512</strong> algorithm.</p>
            <p>This means your private keys are mathematically derived from a space so massive, it is practically immune to modern brute-force and quantum computing threats. We don't just secure your assets; we future-proof them.</p>
        </div>
        <div class="sec-visual">
            <div class="hash-line">Input: "alpha bravo cipher delta..."</div>
            <div class="hash-line" style="margin-top:20px; color:#555;">Processing SHA-512...</div>
            <div class="hash-line" style="color:#10B981; margin-top:10px;">Output Hash:</div>
            <div class="hash-line">e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855</div>
            <div class="hash-line">121a9c80d85ab2c7d9a16f2c00c5c4...</div>
        </div>
    </section>

    <main class="grid">
        <div class="card">
            <i class="fas fa-shield-halved"></i>
            <h3>Mnemonic Vault</h3>
            <p>Your identity is your seed. Our client-side vault generates 24-word recovery phrases that never touch the server, giving you absolute self-custody.</p>
        </div>
        <div class="card">
            <i class="fas fa-bolt"></i>
            <h3>High-Speed Mempool</h3>
            <p>Transactions are validated and added to the blockchain instantly. The Go-powered backend ensures zero collision and maximum throughput.</p>
        </div>
        <div class="card">
            <i class="fas fa-coins"></i>
            <h3>Alpha Airdrop</h3>
            <p>Early adopters securing a Founder Node receive an immediate genesis allocation of 10,000 AX, fully migratable to the Mainnet.</p>
        </div>
    </main>

    <section id="comprar" class="pre-sale">
        <div class="price-label">Node Allocation Fee</div>
        <div class="price">21 USDT</div>
        <p style="color: var(--text-muted); margin-bottom: 20px;">Secure your position in the network infrastructure.</p>
        
        <div class="address-box">
            <span class="network-tag">BNB Smart Chain (BEP-20)</span>
            <div class="wallet-addr">0x948a663b1bd1292ded76a8412af2092bf0462d7c</div>
        </div>

        <p style="font-size: 0.9rem; color: var(--text-muted); margin: 10px 0 30px;">Send the exact amount and verify your transaction via WhatsApp to activate your 512-bit Identity.</p>
        
        <a href="https://wa.me/TuNumeroAqui?text=Hello,%20I%20have%20sent%2021%20USDT%20for%20my%20AstraliX%20Founder%20Node" class="btn-wa">
            <i class="fab fa-whatsapp"></i> SEND TX RECEIPT
        </a>
    </section>

    <footer>
        &copy; 2026 AstraliX Core Development. All rights reserved.
    </footer>

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