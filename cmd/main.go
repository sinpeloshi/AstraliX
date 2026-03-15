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
	rootAddr := "AX5eaba583bf646e0e39f41da6f9d8fa6db929c2e858bd32dffe6ac0cee2e3e97400000000000000000000000000000000000000000000000000000000000000000"

	loadChain()

	if len(Blockchain) == 0 {
		genesisTx := core.Transaction{Sender: "SYSTEM", Recipient: rootAddr, Amount: 1000002021}
		genesisTx.TxID = genesisTx.CalculateHash()
		
		genesisBlock := core.Block{
			Index: 0, 
			Timestamp: 1773561600,
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
			http.Error(w, "Nothing to validate", 400); return 
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
    <title>AX Core | Global Enterprise</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --ax-dark: #0F172A; --ax-blue: #2563EB; --ax-bg: #F8FAFC; }
        body { background: var(--ax-bg); font-family: "Inter", sans-serif; margin: 0; padding-bottom: 100px; overflow-x: hidden; }
        
        .nav-header { background: var(--ax-dark); color: white; padding: 20px; text-align: center; box-shadow: 0 4px 12px rgba(0,0,0,0.1); }
        
        .card-ax { background: white; border-radius: 24px; box-shadow: 0 4px 20px rgba(0,0,0,0.02); padding: 25px; margin: 20px; border: none; }
        .hero { background: linear-gradient(135deg, var(--ax-dark) 0%, var(--ax-blue) 100%); color: white; border-radius: 30px; }
        
        .pill-512 { background: #F1F5F9; padding: 15px; border-radius: 16px; font-family: "JetBrains Mono", monospace; font-size: 0.7rem; word-break: break-all; margin-top: 15px; line-height: 1.5; color: #475569; border: 1px solid #E2E8F0; }
        .hero .pill-512 { background: rgba(255,255,255,0.1); color: white; border: none; }
        
        .btn-mine { background: var(--ax-blue); color: white; border-radius: 18px; padding: 18px; font-weight: 800; border: none; width: 100%; font-size: 1.1rem; box-shadow: 0 8px 25px rgba(37, 99, 235, 0.3); }
        
        .bottom-nav { background: white; position: fixed; bottom: 0; width: 100%; height: 85px; display: flex; justify-content: space-around; align-items: center; border-top: 1px solid #E2E8F0; z-index: 9999; box-shadow: 0 -5px 25px rgba(0,0,0,0.05); }
        .nav-item { color: #94A3B8; text-align: center; text-decoration: none; flex: 1; font-size: 11px; font-weight: 700; transition: 0.3s; }
        .nav-item.active { color: var(--ax-blue); }
        .nav-item i { font-size: 24px; display: block; margin-bottom: 5px; }

        .sidebar { display: none; } /* Eliminamos la sidebar en móvil para evitar bugs */
        @media (min-width: 992px) {
            .bottom-nav { display: none; }
            .sidebar { display: block; background: var(--ax-dark); height: 100vh; position: fixed; width: 260px; color: white; z-index: 1000; }
            .main-content { margin-left: 260px; }
            .nav-link-side { color: #94A3B8; padding: 15px 25px; display: block; text-decoration: none; font-weight: 600; border-radius: 12px; margin: 10px; }
            .nav-link-side.active { background: var(--ax-blue); color: white; }
        }
    </style>
</head>
<body>

    <div class="sidebar">
        <div class="p-5 text-center"><h3 class="fw-bold m-0">AX CORE</h3></div>
        <a class="nav-link-side active" onclick="nav('dash')"><i class="fas fa-th-large me-2"></i> Dashboard</a>
        <a class="nav-link-side" onclick="nav('wallet')"><i class="fas fa-wallet me-2"></i> Wallet</a>
        <a class="nav-link-side" onclick="nav('sec')"><i class="fas fa-key me-2"></i> Security</a>
    </div>

    <div class="main-content">
        <div class="nav-header d-lg-none">
            <h4 class="fw-bold m-0">AX CORE</h4>
        </div>

        <div id="v-dash" class="view-ax">
            <div class="card-ax hero text-center">
                <small class="text-uppercase fw-bold opacity-75">Network Balance</small>
                <h1 id="bal-txt" class="display-4 fw-bold my-2">0.00 AX</h1>
                <div id="addr-txt" class="pill-512">Connect in Security Tab</div>
            </div>

            <div class="card-ax text-center" style="border: 2px dashed #CBD5E1;">
                <small class="fw-bold text-muted">TREASURY REWARDS (2^512)</small>
                <h3 id="pool-txt" class="fw-bold m-0 text-primary">0.00 AX</h3>
                <div class="pill-512">AX5def33f67eda5560561837935709169eb17955ffe13c1f112b3a329321bef5400000000000000000000000000000000000000000000000000000000000000000</div>
            </div>

            <div class="px-3">
                <button class="btn-mine" onclick="mine()">MINE BLOCK (+50.00 AX)</button>
            </div>
        </div>

        <div id="v-wallet" class="view-ax" style="display:none">
            <div class="card-ax">
                <h4 class="fw-bold mb-4">Transfer Assets</h4>
                <input type="text" id="tx-to" class="form-control p-3 mb-3 border-0 bg-light rounded-4" placeholder="Recipient 128-char address">
                <input type="number" id="tx-amt" class="form-control p-3 mb-4 border-0 bg-light rounded-4" placeholder="0.00">
                <button class="btn-mine" onclick="send()">CONFIRM TRANSFER</button>
            </div>
        </div>

        <div id="v-sec" class="view-ax" style="display:none">
            <div class="card-ax">
                <h4 class="fw-bold mb-4">Vault Identity</h4>
                <input type="password" id="i-priv" class="form-control p-3 mb-3 border-0 bg-light rounded-4" placeholder="Private Key">
                <button class="btn-mine mb-3" onclick="login()">CONNECT WALLET</button>
                <hr>
                <button class="btn btn-outline-dark w-100 py-3 rounded-4" onclick="gen()">GENERATE 512-BIT KEY</button>
                <div id="g-res" class="mt-4" style="display:none">
                    <small class="fw-bold">Your Secure Key (512 bits):</small>
                    <div class="pill-512" id="g-priv"></div>
                    <small class="fw-bold text-primary">Public Address:</small>
                    <div class="pill-512 text-primary fw-bold" id="g-pub"></div>
                </div>
            </div>
        </div>
    </div>

    <div class="bottom-nav">
        <a class="nav-item active" id="n-dash" onclick="nav('dash')"><i class="fas fa-th-large"></i>Dash</a>
        <a class="nav-item" id="n-wallet" onclick="nav('wallet')"><i class="fas fa-paper-plane"></i>Send</a>
        <a class="nav-item" id="n-sec" onclick="nav('sec')"><i class="fas fa-key"></i>Keys</a>
    </div>

    <script>
        async function derive(priv) {
            const buf = new TextEncoder().encode(priv);
            const hash = await crypto.subtle.digest("SHA-512", buf);
            const hex = Array.from(new Uint8Array(hash)).map(b => b.toString(16).padStart(2,"0")).join("");
            return "AX" + hex; 
        }

        let session = JSON.parse(localStorage.getItem("ax_v17_session")) || null;

        function nav(id) {
            document.querySelectorAll(".view-ax").forEach(v => v.style.display = "none");
            document.getElementById("v-" + id).style.display = "block";
            document.querySelectorAll(".nav-item").forEach(n => n.classList.remove("active"));
            document.getElementById("n-" + id).classList.add("active");
            window.scrollTo(0,0);
        }

        async function login() {
            const p = document.getElementById("i-priv").value;
            const pb = await derive(p);
            session = { pub: pb, priv: p };
            localStorage.setItem("ax_v17_session", JSON.stringify(session));
            location.reload();
        }

        async function load() {
            if(session) {
                const r = await fetch("/api/balance/" + session.pub);
                const d = await r.json();
                document.getElementById("bal-txt").innerText = d.balance.toLocaleString() + " AX";
                document.getElementById("addr-txt").innerText = session.pub; // FULL 128 CHARS
            }
            const rp = await fetch("/api/balance/AX5def33f67eda5560561837935709169eb17955ffe13c1f112b3a329321bef5400000000000000000000000000000000000000000000000000000000000000000");
            const dp = await rp.json();
            document.getElementById("pool-txt").innerText = dp.balance.toLocaleString() + " AX";
        }

        async function mine() {
            if(!session) return alert("Sync first");
            const r = await fetch("/api/mine?address=" + session.pub);
            if(r.ok) { alert("Mined!"); load(); } else { alert("Nothing to mine/Treasury issue."); }
        }

        async function send() {
            const tx = { sender: session.pub, recipient: document.getElementById("tx-to").value, amount: parseFloat(document.getElementById("tx-amt").value) };
            const r = await fetch("/api/transactions/new", { method: "POST", body: JSON.stringify(tx) });
            if(r.ok) { alert("Sent!"); nav("dash"); load(); } else { alert("Error."); }
        }

        async function gen() {
            const p = btoa(Math.random().toString() + Date.now()).substring(0,64);
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
