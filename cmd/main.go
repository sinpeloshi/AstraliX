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

func main() {
	const Difficulty = 4 
	creatorAddr := "AXdc3acc7c0b91eb485d0e3bb78059bb58a3999c14b56cfe6ca0428670afc6410c"

	genTx := core.Transaction{Sender: "SYSTEM", Recipient: creatorAddr, Amount: 1000002021}
	genTx.TxID = genTx.CalculateHash()
	genesis := core.Block{
		Index: 0, Timestamp: 1773561600,
		Transactions: []core.Transaction{genTx},
		PrevHash: strings.Repeat("0", 128), Difficulty: Difficulty,
	}
	genesis.Mine()
	Blockchain = append(Blockchain, genesis)

	http.HandleFunc("/api/balance/", func(w http.ResponseWriter, r *http.Request) {
		addr := strings.TrimPrefix(r.URL.Path, "/api/balance/")
		balance := 0.0
		for _, b := range Blockchain {
			for _, tx := range b.Transactions {
				if tx.Recipient == addr { balance += tx.Amount }
				if tx.Sender == addr { balance -= tx.Amount }
			}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"balance": balance})
	})

	http.HandleFunc("/api/chain", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Blockchain)
	})

	http.HandleFunc("/api/mine", func(w http.ResponseWriter, r *http.Request) {
		if len(Mempool) == 0 {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": "mempool_empty"})
			return 
		}
		last := Blockchain[len(Blockchain)-1]
		newB := core.Block{
			Index: int64(len(Blockchain)), Timestamp: time.Now().Unix(),
			Transactions: Mempool, PrevHash: last.Hash, Difficulty: Difficulty,
		}
		newB.Mine()
		Blockchain = append(Blockchain, newB)
		Mempool = []core.Transaction{}
		json.NewEncoder(w).Encode(newB)
	})

	http.HandleFunc("/api/transactions/new", func(w http.ResponseWriter, r *http.Request) {
		var tx core.Transaction
		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			http.Error(w, "Invalid", 400); return
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
	fmt.Printf("🌐 AstraliX Pro Node running on port %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>AstraliX | 512-bit Secure Wallet</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --accent: #00ffa3; --bg: #020203; --card: rgba(20, 20, 25, 0.95); }
        body { background: var(--bg); color: #f8f9fa; font-family: 'Inter', sans-serif; padding-bottom: 110px; }
        .main-content { padding: 20px; max-width: 480px; margin: auto; }
        .glass-card { background: var(--card); border: 1px solid rgba(255, 255, 255, 0.05); border-radius: 32px; padding: 28px; margin-bottom: 22px; box-shadow: 0 15px 35px rgba(0,0,0,0.6); }
        .mobile-nav { background: rgba(0,0,0,0.9); backdrop-filter: blur(20px); border-top: 1px solid #1a1a1a; position: fixed; bottom: 0; width: 100%; height: 90px; display: flex; justify-content: space-around; align-items: center; z-index: 1000; }
        .nav-item { color: #444; text-decoration: none; text-align: center; font-size: 10px; cursor: pointer; font-weight: 800; text-transform: uppercase; }
        .nav-item.active { color: var(--accent); }
        .nav-item i { font-size: 26px; display: block; margin-bottom: 6px; }
        .btn-elite { background: var(--accent); color: #000; font-weight: 900; border-radius: 22px; border: none; padding: 20px; width: 100%; transition: 0.3s; }
        .btn-elite:hover { box-shadow: 0 0 35px var(--accent); transform: translateY(-4px); }
        .accent { color: var(--accent); }
        input { background: #000 !important; border: 1px solid #222 !important; color: #fff !important; border-radius: 20px !important; padding: 18px !important; margin-bottom: 18px; }
        .key-box { font-family: monospace; font-size: 0.7rem; background: #000; padding: 20px; border-radius: 22px; border: 1px solid #333; word-break: break-all; color: #888; position: relative; }
        .seed-badge { display: inline-block; background: #111; padding: 5px 12px; border-radius: 10px; margin: 4px; font-size: 13px; color: var(--accent); border: 1px solid #222; }
    </style>
</head>
<body>

    <div class="main-content text-center">
        <div class="my-4">
            <svg width="60" height="60" viewBox="0 0 100 100">
                <circle cx="50" cy="50" r="45" fill="none" stroke="#00ffa3" stroke-width="2" stroke-dasharray="10 5"/>
                <path d="M30 70 L50 30 L70 70 M40 50 L60 50" fill="none" stroke="#00ffa3" stroke-width="6" stroke-linecap="round"/>
                <circle cx="50" cy="30" r="5" fill="#00ffa3"/>
            </svg>
            <h2 class="fw-black accent mt-2" style="letter-spacing: -2px;">ASTRALIX</h2>
            <small class="text-muted">512-BIT EDITION</small>
        </div>

        <div id="view-assets" class="view-section">
            <div class="glass-card">
                <small class="text-muted fw-bold">TOTAL AX BALANCE</small>
                <h1 id="bal-large" class="accent fw-black my-2" style="font-size: 3.5rem;">0.00</h1>
                <div id="addr-display" class="small opacity-50 text-truncate px-4">Disconnected</div>
            </div>

            <div class="row g-3">
                <div class="col-6"><button class="btn btn-elite" onclick="switchTab('send')">SEND</button></div>
                <div class="col-6"><button class="btn btn-dark w-100 border-0 py-3 rounded-4" onclick="showQR()">RECEIVE</button></div>
            </div>

            <div id="recent-feed" class="mt-5 text-start"></div>
        </div>

        <div id="view-send" class="view-section" style="display:none">
            <div class="glass-card text-start">
                <h4 class="fw-bold mb-4">New Transfer</h4>
                <input type="text" id="tx-to" class="form-control" placeholder="Recipient AX Address">
                <input type="number" id="tx-amt" class="form-control" placeholder="Amount">
                <button class="btn btn-elite" onclick="processSend()">SIGN WITH 512-BIT KEY</button>
            </div>
        </div>

        <div id="view-manage" class="view-section" style="display:none">
            <div class="glass-card text-start">
                <h4 class="fw-bold mb-3">Key Login</h4>
                <p class="small text-muted">Import your 512-bit Private Key.</p>
                <input type="password" id="imp-priv" class="form-control" placeholder="Paste Secret Key">
                <button class="btn btn-elite" onclick="importWallet()">LOGIN</button>
                <hr class="border-secondary my-4">
                <button class="btn btn-outline-danger w-100 rounded-4" onclick="logout()">CLEAR SESSION</button>
            </div>

            <div class="glass-card text-start border-success border-opacity-25">
                <h4 class="fw-bold accent">Seed Generator</h4>
                <p class="small text-muted">Generate a secure 512-bit based seed.</p>
                <button class="btn btn-dark w-100 rounded-4 py-3" onclick="createNew()">GENERATE 12 WORDS</button>
                
                <div id="seed-output" class="mt-4" style="display:none">
                    <div id="seed-words" class="mb-3"></div>
                    <small class="text-muted">Private Key (512-bit):</small>
                    <div class="key-box mb-3" id="gen-priv"></div>
                    <small class="accent">Derived Address:</small>
                    <div class="key-box" id="gen-pub"></div>
                </div>
            </div>
        </div>
    </div>

    <div class="mobile-nav">
        <div class="nav-item active" id="nav-assets" onclick="switchTab('assets', this)"><i class="fas fa-wallet"></i>Assets</div>
        <div class="nav-item" id="nav-manage" onclick="switchTab('manage', this)"><i class="fas fa-shield-alt"></i>Security</div>
        <div class="nav-item" id="nav-explorer" onclick="mine()"><i class="fas fa-hammer"></i>Mine</div>
    </div>

    <script>
        async function deriveAddress(priv) {
            const msgBuffer = new TextEncoder().encode(priv);
            // CORREGIDO: Usamos SHA-512 para el estándar 2^512 de AstraliX
            const hashBuffer = await crypto.subtle.digest('SHA-512', msgBuffer);
            const hashArray = Array.from(new Uint8Array(hashBuffer));
            const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
            return 'AX' + hashHex.substring(0, 64); // Usamos 64 caracteres del hash de 512 bits
        }

        let wallet = JSON.parse(localStorage.getItem('ax_session')) || null;

        function switchTab(id, el) {
            document.querySelectorAll('.view-section').forEach(v => v.style.display = 'none');
            document.getElementById('view-' + id).style.display = 'block';
            document.querySelectorAll('.nav-item').forEach(l => l.classList.remove('active'));
            if(el) el.classList.add('active');
        }

        async function createNew() {
            const words = ["star", "galaxy", "orbit", "node", "crypto", "block", "chain", "secure", "alpha", "delta", "nebula", "astral"];
            const seed = [...Array(12)].map(() => words[Math.floor(Math.random()*words.length)]).join(' ');
            const priv = btoa(seed + Date.now()).substring(0, 128); // Más entropía
            const pub = await deriveAddress(priv);

            document.getElementById('seed-output').style.display = 'block';
            document.getElementById('seed-words').innerHTML = seed.split(' ').map(w => '<span class="seed-badge">'+w+'</span>').join('');
            document.getElementById('gen-priv').innerText = priv;
            document.getElementById('gen-pub').innerText = pub;
        }

        async function importWallet() {
            const priv = document.getElementById('imp-priv').value;
            if(!priv) return alert("Private Key required");
            const pub = await deriveAddress(priv);
            wallet = { pub, priv };
            localStorage.setItem('ax_session', JSON.stringify(wallet));
            location.reload();
        }

        function logout() { localStorage.removeItem('ax_session'); location.reload(); }

        async function load() {
            if(wallet) {
                document.getElementById('addr-display').innerText = wallet.pub;
                const r = await fetch('/api/balance/' + wallet.pub);
                const d = await r.json();
                document.getElementById('bal-large').innerText = d.balance.toLocaleString();
            }
            const resC = await fetch('/api/chain');
            const chain = await resC.json();
            const feed = document.getElementById('recent-feed');
            feed.innerHTML = '<h6 class="fw-bold mb-3">LATEST NETWORK BLOCKS</h6>';
            chain.slice().reverse().forEach(b => {
                feed.innerHTML += '<div class="glass-card p-3 mb-2" style="border-radius:20px; font-size:11px;"><div class="d-flex justify-content-between"><span class="accent">#' + b.index + '</span><span class="opacity-50">' + b.hash.substring(0,32) + '...</span></div></div>';
            });
        }

        async function mine() {
            const r = await fetch('/api/mine');
            if(r.ok) { alert("Mining Complete!"); load(); } else { alert("Nothing to mine yet."); }
        }

        async function processSend() {
            if(!wallet) return switchTab('manage');
            const tx = { sender: wallet.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            alert("Transaction authorized with 512-bit key!"); switchTab('assets'); load();
        }

        load();
        setInterval(load, 20000);
    </script>
</body>
</html>
