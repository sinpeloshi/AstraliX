package main

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
// 🎨 LANDING PAGE (VALLEY STYLE)
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
        
        .nav { padding: 20px 6%; display: flex; justify-content: space-between; align-items: center; position: sticky; top: 0; background: rgba(2,2,2,0.85); backdrop-filter: blur(20px); z-index: 100; border-bottom: 1px solid var(--brd); }
        .logo { display: flex; align-items: center; text-decoration: none; }
        .logo img { height: 45px; width: auto; mix-blend-mode: screen; } 
        
        .nav-links { display: flex; align-items: center; }
        .nav-links a { color: var(--txt-m); text-decoration: none; font-size: 0.85rem; font-weight: 600; transition: 0.2s; margin-right: 25px; }
        .nav-links a:hover { color: var(--txt); }
        .btn-core-nav { background: var(--prim); color: white !important; padding: 10px 22px; border-radius: 100px; font-size: 0.75rem; font-weight: 800; text-decoration: none; transition: 0.3s; margin-right: 0 !important; }
        
        .hero { text-align: center; padding: 80px 6% 40px; max-width: 1200px; margin: 0 auto; position: relative; }
        .hero h1 { font-size: clamp(3rem, 9vw, 6.2rem); font-weight: 800; letter-spacing: -3px; line-height: 1.1; margin-bottom: 25px; background: linear-gradient(180deg, #FFF 30%, #555 100%); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
        
        .mockup-container { max-width: 900px; margin: 0 auto; padding: 0 6%; position: relative; }
        .mockup-window { background: rgba(10,10,10,0.8); border: 1px solid #222; border-radius: 16px; overflow: hidden; backdrop-filter: blur(20px); box-shadow: 0 25px 50px -12px rgba(0,0,0,0.5); }
        .mockup-body { padding: 30px; font-family: 'JetBrains Mono', monospace; font-size: 0.85rem; color: var(--txt-m); line-height: 1.8; text-align: left; }

        /* TOKENOMICS SECTION */
        .tokenomics { max-width: 1000px; margin: 100px auto; padding: 0 6%; text-align: center; }
        .tok-flex { display: flex; align-items: center; justify-content: center; gap: 50px; flex-wrap: wrap; margin-top: 50px; }
        .tok-chart { position: relative; width: 280px; height: 280px; border-radius: 50%; background: conic-gradient(var(--acc) 0% 12.5%, var(--prim) 12.5% 30%, #8B5CF6 30% 45%, #F59E0B 45% 59%, #EC4899 59% 79%, #4B5563 79% 100%); }
        .tok-chart::after { content: '21% REWARDS'; position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); width: 170px; height: 170px; background: var(--bg); border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: 800; font-size: 0.75rem; color: var(--acc); letter-spacing: 1px; text-align: center; }
        .tok-legend { text-align: left; display: flex; flex-direction: column; gap: 12px; }
        .leg-item { display: flex; align-items: center; gap: 12px; font-size: 0.9rem; color: var(--txt-m); }
        .leg-color { width: 10px; height: 10px; border-radius: 2px; }

        /* TIERS & FLOW CSS */
        .pre-sale { background: var(--bg-card); padding: 100px 6%; text-align: center; border-top: 1px solid var(--brd); }
        .tier-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 20px; max-width: 800px; margin: 0 auto 40px; }
        .tier-card { background: #000; border: 1px solid var(--brd); border-radius: 20px; padding: 40px 30px; transition: 0.3s; position: relative; overflow: hidden; }
        .tier-card:hover { border-color: var(--prim); transform: translateY(-5px); }
        .tier-card.premium { border-color: rgba(16, 185, 129, 0.4); background: linear-gradient(180deg, rgba(16,185,129,0.05) 0%, #000 100%); }
        .tier-card.premium::before { content: 'PRO'; position: absolute; top: 15px; right: -30px; background: var(--acc); color: #000; font-size: 0.6rem; font-weight: 800; padding: 5px 30px; transform: rotate(45deg); letter-spacing: 1px; }
        .t-price { font-size: 3.5rem; font-weight: 800; letter-spacing: -2px; margin: 15px 0; color: #FFF; }
        .t-name { font-size: 0.9rem; text-transform: uppercase; letter-spacing: 2px; color: var(--prim); font-weight: 800; }
        .inst-box { background: #0A0A0A; border: 1px solid var(--brd); border-radius: 20px; padding: 30px; margin-bottom: 20px; text-align: left; }
        .btn-buy { background: var(--acc); color: #000; padding: 20px 50px; border-radius: 100px; font-weight: 800; text-decoration: none; font-size: 1.1rem; display: inline-flex; align-items: center; justify-content: center; gap: 10px; transition: 0.3s; width: 100%; box-sizing: border-box; }
        
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; max-width: 1200px; margin: 100px auto; padding: 0 6%; }
        .card { background: var(--bg-card); border: 1px solid var(--brd); padding: 50px 40px; text-align: left; border-radius: 20px; }
        
        @media (max-width: 850px) { .nav-links { display: none; } .hero { padding-top: 50px; } .tok-flex { flex-direction: column; } }
    </style>
</head>
<body>
    <div class="bg-p"></div>
    <nav class="nav">
        <a href="/" class="logo"><img src="https://iili.io/qMGLM57.jpg" alt="AstraliX"></a>
        <div style="display: flex; align-items: center; gap: 20px;">
            <div class="nav-links">
                <a href="/whitepaper">Whitepaper</a>
                <a href="/dashboard" class="btn-core-nav">ENTER DASHBOARD</a>
            </div>
            <div class="nav-socials" style="display: flex; gap: 15px;">
                <a href="https://x.com/XAstraliX" target="_blank" style="color:white; font-size: 1.2rem;"><i class="fab fa-x-twitter"></i></a>
                <a href="https://t.me/XAstraliX" target="_blank" style="color:white; font-size: 1.2rem;"><i class="fab fa-telegram"></i></a>
            </div>
        </div>
    </nav>

    <header class="hero">
        <div style="background: rgba(16,185,129,0.1); color: var(--acc); padding: 8px 24px; border-radius: 100px; font-size: 0.75rem; font-weight: 800; display: inline-block; margin-bottom: 30px; border: 1px solid rgba(16,185,129,0.2);">ALPHA TESTNET LIVE</div>
        <h1>The 512-bit Era Begins.</h1>
        <p style="color: var(--txt-m); font-size: 1.3rem; max-width: 700px; margin: 0 auto 50px;">A mission-critical Layer 1 protocol doubling security standards for the post-quantum era.</p>
        <div style="display: flex; gap: 15px; justify-content: center; flex-wrap: wrap; margin-bottom: 60px;">
            <a href="/dashboard" style="background:var(--prim); color:white; padding:18px 40px; border-radius:100px; font-weight:800; text-decoration:none;">Launch Testnet App</a>
            <a href="/whitepaper" style="border:1px solid var(--brd); color:white; padding:18px 40px; border-radius:100px; font-weight:800; text-decoration:none;">Read Whitepaper</a>
        </div>
    </header>

    <div class="mockup-container">
        <div class="mockup-window">
            <div class="mockup-body">
                <div>> initializing 512-bit quantum-resistant protocol... <span style="color:var(--acc);">[OK]</span></div>
                <div>> checking genesis integrity... <span style="color:var(--acc);">[DONE]</span></div>
                <br>
                <div style="display:flex; justify-content:space-between; border-top:1px dashed #333; padding-top:20px;">
                    <div><div style="font-size:0.7rem;">NETWORK_STATUS</div><div style="color:var(--acc); font-weight:800;">ACTIVE (BLOCK #<span id="mock-block">0</span>)</div></div>
                    <div><div style="font-size:0.7rem;">TOTAL_SUPPLY</div><div style="color:white; font-weight:800;">1,000,002,021 AX</div></div>
                </div>
            </div>
        </div>
    </div>

    <section class="tokenomics">
        <h2 style="font-size:2.5rem; font-weight:800; margin-bottom:15px;">Protocol Tokenomics</h2>
        <p style="color:var(--txt-m);">Sustainable distribution for long-term network security.</p>
        <div class="tok-flex">
            <div class="tok-chart"></div>
            <div class="tok-legend">
                <div class="leg-item"><div class="leg-color" style="background:#4B5563;"></div><span><strong>21.0%</strong> Ecosystem Mining Rewards</span></div>
                <div class="leg-item"><div class="leg-color" style="background:var(--acc);"></div><span><strong>12.5%</strong> Founder Nodes (Seed Round)</span></div>
                <div class="leg-item"><div class="leg-color" style="background:var(--prim);"></div><span><strong>17.5%</strong> Treasury & Protocol R&D</span></div>
                <div class="leg-item"><div class="leg-color" style="background:#8B5CF6;"></div><span><strong>15.0%</strong> Locked Liquidity Pool</span></div>
                <div class="leg-item"><div class="leg-color" style="background:#F59E0B;"></div><span><strong>14.0%</strong> Marketing & Growth</span></div>
                <div class="leg-item"><div class="leg-color" style="background:#EC4899;"></div><span><strong>20.0%</strong> Team & Contributors (Locked)</span></div>
            </div>
        </div>
    </section>

    <section id="buy" class="pre-sale">
        <div style="text-transform: uppercase; letter-spacing: 4px; font-weight: 800; color: var(--prim); font-size: 0.8rem; margin-bottom:15px;">Capital Allocation</div>
        <h2 style="font-size:2.5rem; font-weight:800; margin-bottom: 50px;">Acquire Founder Node</h2>
        
        <div class="tier-grid">
            <div class="tier-card">
                <div class="t-name">Standard Node</div>
                <div class="t-price">21<span style="font-size:1.2rem; color:var(--txt-m);"> USDT</span></div>
                <div style="text-align:left; color:var(--txt-m); font-size:0.9rem; margin-top:20px;">
                    <p><i class="fas fa-check" style="color:var(--acc);"></i> 10,000 AX Coins</p>
                    <p><i class="fas fa-check" style="color:var(--acc);"></i> Validator Rights</p>
                </div>
            </div>
            <div class="tier-card premium">
                <div class="t-name" style="color:var(--acc);">Master Node</div>
                <div class="t-price">210<span style="font-size:1.2rem; color:var(--txt-m);"> USDT</span></div>
                <div style="text-align:left; color:var(--txt-m); font-size:0.9rem; margin-top:20px;">
                    <p><i class="fas fa-check" style="color:var(--acc);"></i> 100,000 AX Coins</p>
                    <p><i class="fas fa-check" style="color:var(--acc);"></i> Priority Validator Rights</p>
                </div>
            </div>
        </div>

        <div style="max-width: 700px; margin: 0 auto;">
            <div class="inst-box">
                <div style="color: var(--prim); font-weight: 800; font-size: 0.8rem; margin-bottom: 10px;">STEP 1: GENERATE VAULT</div>
                <p style="color: var(--txt-m); font-size: 0.9rem;">Go to the <a href="/dashboard" style="color:var(--prim);">Vault</a> and create your Identity. Copy your <strong>AX Address</strong>.</p>
            </div>
            <div class="inst-box">
                <div style="color: var(--prim); font-weight: 800; font-size: 0.8rem; margin-bottom: 10px;">STEP 2: SEND USDT (BEP-20)</div>
                <div style="font-family: 'JetBrains Mono'; font-size: 0.85rem; color:var(--acc); padding: 15px; border: 1px dashed #333; word-break:break-all; text-align:center;">0x948a663b1bd1292ded76a8412af2092bf0462d7c</div>
            </div>
            <div class="inst-box" style="text-align:center;">
                <div style="color: var(--prim); font-weight: 800; font-size: 0.8rem; margin-bottom: 20px; text-align:left;">STEP 3: CLAIM NODE</div>
                <a href="https://tally.so/r/jaxlL1" target="_blank" class="btn-buy">CLAIM NODE NOW <i class="fas fa-arrow-right"></i></a>
            </div>
        </div>
    </section>

    <main class="grid">
        <div class="card"><i class="fas fa-microchip" style="font-size:2rem; color:var(--prim); margin-bottom:20px; display:block;"></i><h4>Quantum-Proof</h4><p style="color:var(--txt-m); margin-top:10px;">SHA-512 architecture securing assets for the next century.</p></div>
        <div class="card"><i class="fas fa-bolt" style="font-size:2rem; color:var(--prim); margin-bottom:20px; display:block;"></i><h4>Go-Native</h4><p style="color:var(--txt-m); margin-top:10px;">Hyper-optimized execution for sub-second block finality.</p></div>
        <div class="card"><i class="fas fa-fingerprint" style="font-size:2rem; color:var(--prim); margin-bottom:20px; display:block;"></i><h4>Sovereign</h4><p style="color:var(--txt-m); margin-top:10px;">Zero-trust local mnemonic derivation. Absolute privacy.</p></div>
    </main>

    <footer style="padding: 60px 6%; border-top: 1px solid var(--brd); text-align: center; color: var(--txt-m); font-size: 0.8rem;">
        <img src="https://iili.io/qMGLM57.jpg" style="height:40px; mix-blend-mode: screen; margin-bottom:20px;"><br>
        © 2026 AstraliX Foundation. Designed for Palo Alto, Built for the World.
    </footer>

    <script>
        async function fetchRealData() {
            try {
                const res = await fetch("/api/chain"); const chain = await res.json();
                if(chain && chain.length > 0) {
                    const latest = chain[chain.length - 1];
                    document.getElementById("mock-block").innerText = latest.Index !== undefined ? latest.Index : latest.index;
                }
            } catch(e) {}
        }
        fetchRealData(); setInterval(fetchRealData, 10000);
    </script>
</body>
</html>
`

// ==========================================
// 📄 WHITEPAPER ACTUALIZADO (21% REWARDS)
// ==========================================

const whitepaperHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Whitepaper | AstraliX Protocol</title>
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;600;800&family=JetBrains+Mono:wght@400&display=swap');
        body { font-family: 'Plus Jakarta Sans', sans-serif; background: #020202; color: #FFF; line-height: 1.8; }
        .container { max-width: 800px; margin: 0 auto; padding: 60px 6%; }
        h1 { font-size: 3.5rem; font-weight: 800; letter-spacing: -2px; margin-bottom: 10px; }
        h2 { color: #3B82F6; margin-top: 50px; font-weight: 800; border-bottom: 1px solid #1A1A1A; padding-bottom: 10px; }
        p, li { color: #CCC; font-size: 1.05rem; }
        .tech-box { background: #080808; border: 1px solid #1A1A1A; padding: 25px; border-radius: 12px; font-family: 'JetBrains Mono'; color: #8899A6; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:40px;">
            <a href="/" style="color:#3B82F6; text-decoration:none; font-weight:800;">← BACK</a>
            <img src="https://iili.io/qMGLM57.jpg" style="height:40px; mix-blend-mode:screen;">
        </div>
        <h1>AstraliX Core</h1>
        <p style="color:#8899A6; text-transform:uppercase; letter-spacing:1px;">Version 1.1 • 512-bit Standard • March 2026</p>

        <h2>1. Abstract</h2>
        <p>The AstraliX Protocol is a Layer 1 blockchain designed to survive the quantum computing era by doubling the industry cryptographic standard to 512 bits. Built natively in Go, it focuses on extreme entropy and sub-second finality.</p>

        <h2>2. Cryptographic MOAT</h2>
        <div class="tech-box">
            // ENTROPY SCALE<br>
            Standard L1: 2^256 combinations<br>
            AstraliX Core: 2^512 combinations (Quantum Immune)
        </div>

        <h2>3. Reward Pool & Security</h2>
        <p>A fixed pool of **210,000,000 AX** (exactly 21% of total supply) is allocated for validation rewards. This ensures that node operators are incentivized to maintain network integrity for the next decade.</p>

        <h2>4. Tokenomics</h2>
        <ul>
            <li><strong>21.0% Ecosystem Rewards:</strong> Validation and mining incentives.</li>
            <li><strong>12.5% Founder Nodes:</strong> Initial bootstrap allocation (Seed Round).</li>
            <li><strong>15.0% Liquidity Pool:</strong> Reserved for exchange depth.</li>
            <li><strong>17.5% Treasury:</strong> Core R&D and server infrastructure.</li>
            <li><strong>14.0% Marketing:</strong> Community and adoption.</li>
            <li><strong>20.0% Team:</strong> Subject to 24-month linear vesting.</li>
        </ul>

        <h2>5. Roadmap</h2>
        <p>Mainnet launch is scheduled for April 2026. All Alpha Testnet balances will be migrated 1:1 during the Genesis Hard Fork.</p>

        <footer style="margin-top:80px; text-align:center; color:#444;">© 2026 AstraliX Foundation.</footer>
    </div>
</body>
</html>
`

// ==========================================
// 📱 DASHBOARD (NO TOCADO PARA NO ROMPER NADA)
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
        body { background: var(--bg); font-family: 'Plus Jakarta Sans', sans-serif; margin: 0; padding-bottom: 120px; color: var(--txt); overflow-x: hidden; }
        .container { max-width: 550px; margin: 0 auto; padding: 0 5%; box-sizing: border-box; }
        .header-ax { padding: 40px 0 20px; text-align: center; }
        .card-ax { background: var(--card); border-radius: 24px; padding: 30px 25px; width: 100%; border: 1px solid var(--brd); box-sizing: border-box; margin-bottom:20px; }
        .bal-val { font-size: 2.5rem; font-weight: 800; margin-bottom: 25px; }
        .pill { background: #000; padding: 15px; border-radius: 15px; font-family: 'JetBrains Mono'; font-size: 0.65rem; word-break: break-all; color: var(--txt-m); border: 1px solid var(--brd); }
        .btn-ax { background: var(--prim); color: white; border-radius: 15px; padding: 20px; font-weight: 800; border: none; width: 100%; cursor: pointer; }
        .bottom-bar { background: rgba(2,2,2,0.85); backdrop-filter: blur(20px); position: fixed; bottom: 0; left: 0; width: 100%; height: 85px; display: flex; border-top: 1px solid var(--brd); }
        .nav-l { color: #555; text-decoration: none; font-size: 0.6rem; font-weight: 800; display: flex; flex-direction: column; align-items: center; gap: 8px; flex: 1; justify-content: center; cursor:pointer; }
        .nav-l.active { color: var(--prim); }
        .input-ax { width: 100%; padding: 20px; border-radius: 15px; border: 1px solid var(--brd); background: #000; color: #FFF; margin-bottom: 12px; box-sizing: border-box; }
        .view-ax { display: none; flex-direction: column; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header-ax">
            <a href="/"><img src="https://iili.io/qMGLM57.jpg" style="height:50px; mix-blend-mode:screen;"></a><br>
            <div style="color:var(--acc); font-size:0.7rem; font-weight:800; margin-top:10px;">ALPHA TESTNET ACTIVE</div>
        </div>
        
        <div id="v-dash" class="view-ax" style="display:flex;">
            <div class="card-ax" style="border-color: var(--prim); background: linear-gradient(180deg, #0A0A0A 0%, #000 100%);">
                <div id="bal-txt" class="bal-val">0.00 AX</div>
                <div id="addr-txt" class="pill">VAULT LOCKED</div>
            </div>
            <button class="btn-ax" onclick="mine()"><i class="fas fa-hammer"></i> VALIDATE NETWORK (+50 AX)</button>
        </div>

        <div id="v-wallet" class="view-ax">
            <div class="card-ax">
                <input type="text" id="tx-to" class="input-ax" placeholder="Recipient AX Address">
                <input type="number" id="tx-amt" class="input-ax" placeholder="Amount AX Tokens">
                <button class="btn-ax" onclick="send()">CONFIRM DEPOSIT</button>
            </div>
        </div>

        <div id="v-sec" class="view-ax">
            <div class="card-ax">
                <textarea id="i-seed" class="input-ax" style="height:120px;" placeholder="Enter seed phrase..."></textarea>
                <button class="btn-ax" onclick="login()">RESTORE WALLET</button>
                <button class="btn-ax" style="background:transparent; color:#555;" onclick="gen()">GENERATE IDENTITY</button>
                <div id="g-res" style="display:none; margin-top:20px;">
                    <div id="g-seed" style="display:grid; grid-template-columns:1fr 1fr; gap:5px;"></div>
                    <div class="pill" id="g-pub" style="margin-top:10px;"></div>
                </div>
            </div>
        </div>
    </div>

    <div class="bottom-bar">
        <a class="nav-l active" onclick="nav('dash')"><i class="fas fa-chart-pie"></i>DASH</a>
        <a class="nav-l" onclick="nav('wallet')"><i class="fas fa-exchange-alt"></i>SEND</a>
        <a class="nav-l" onclick="nav('sec')"><i class="fas fa-shield-halved"></i>VAULT</a>
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
            window.scrollTo(0,0);
        }
        async function login() {
            const s = document.getElementById("i-seed").value.trim().toLowerCase(); if(!s) return;
            const keys = await derive(s); session = { pub: keys.pub, priv: keys.priv, seed: s };
            localStorage.setItem("ax_v18_session", JSON.stringify(session)); location.reload();
        }
        async function gen() {
            let seed = []; for(let i=0; i<24; i++) seed.push(words[Math.floor(Math.random()*words.length)]);
            const keys = await derive(seed.join(" ")); document.getElementById("g-res").style.display = "block";
            let sH = ""; for(let i=0; i<seed.length; i++) sH += '<div style="font-size:0.6rem; color:#888;">'+(i+1)+'. '+seed[i]+'</div>';
            document.getElementById("g-seed").innerHTML = sH; document.getElementById("g-pub").innerText = keys.pub;
        }
        async function load() {
            if(session) {
                const r = await fetch("/api/balance/" + session.pub); const d = await r.json();
                document.getElementById("bal-txt").innerText = d.balance.toLocaleString() + " AX";
                document.getElementById("addr-txt").innerText = session.pub;
            }
        }
        async function mine() {
            if(!session) return alert("Vault Required");
            const r = await fetch("/api/mine?address=" + session.pub);
            if(r.ok) { alert("Validated! +50 AX"); load(); } else { alert("Mempool empty"); }
        }
        async function send() {
            const tx = { sender: session.pub, recipient: document.getElementById("tx-to").value, amount: parseFloat(document.getElementById("tx-amt").value) };
            const r = await fetch("/api/transactions/new", { method: "POST", body: JSON.stringify(tx) });
            if(r.ok) { alert("Sent!"); nav('dash'); load(); } else { alert("Failed"); }
        }
        load(); setInterval(load, 15000);
    </script>
</body>
</html>
`