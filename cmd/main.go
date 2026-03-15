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
	// Mantenemos la ley de los 512 bits
	const Difficulty = 4 
	creatorAddr := "AXdc3acc7c0b91eb485d0e3bb78059bb58a3999c14b56cfe6ca0428670afc6410c"

	genTx := core.Transaction{Sender: "SYSTEM", Recipient: creatorAddr, Amount: 1000002021}
	genTx.TxID = genTx.CalculateHash()
	
	// PrevHash de 128 caracteres (512 bits en hex)
	genesis := core.Block{
		Index: 0, Timestamp: 1773561600,
		Transactions: []core.Transaction{genTx},
		PrevHash: strings.Repeat("0", 128), 
		Difficulty: Difficulty,
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
	fmt.Printf("🌐 AstraliX 512-bit Node running on port %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>AstraliX | 512-bit Elite</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root { --accent: #00ffa3; --bg: #010102; --card: rgba(25, 25, 30, 0.95); }
        body { background: var(--bg); color: #f8f9fa; font-family: 'Inter', sans-serif; padding-bottom: 110px; }
        .main-content { padding: 20px; max-width: 480px; margin: auto; }
        .glass-card { background: var(--card); border: 1px solid rgba(0, 255, 163, 0.1); border-radius: 35px; padding: 30px; margin-bottom: 25px; box-shadow: 0 20px 40px rgba(0,0,0,0.7); }
        .mobile-nav { background: rgba(0,0,0,0.9); backdrop-filter: blur(25px); border-top: 1px solid #111; position: fixed; bottom: 0; width: 100%; height: 95px; display: flex; justify-content: space-around; align-items: center; z-index: 1000; }
        .nav-item { color: #333; text-decoration: none; text-align: center; font-size: 10px; cursor: pointer; font-weight: 900; letter-spacing: 1px; }
        .nav-item.active { color: var(--accent); }
        .nav-item i { font-size: 28px; display: block; margin-bottom: 8px; }
        .btn-elite { background: var(--accent); color: #000; font-weight: 900; border-radius: 25px; border: none; padding: 22px; width: 100%; transition: 0.4s; }
        .btn-elite:hover { box-shadow: 0 0 40px var(--accent); transform: translateY(-5px); }
        .accent { color: var(--accent); }
        input { background: #000 !important; border: 1px solid #222 !important; color: #fff !important; border-radius: 22px !important; padding: 20px !important; margin-bottom: 20px; }
        .key-box { font-family: monospace; font-size: 0.7rem; background: #000; padding: 20px; border-radius: 25px; border: 1px solid #333; word-break: break-all; color: #888; }
        .seed-badge { display: inline-block; background: #080808; padding: 6px 14px; border-radius: 12px; margin: 5px; font-size: 13px; color: var(--accent); border: 1px solid #1a1a1a; }
    </style>
</head>
<body>
    <div class="main-content text-center">
        <div class="my-5">
            <h1 class="fw-black accent mb-0" style="letter-spacing: -3px; font-size: 3rem;">AX</h1>
            <small class="text-muted fw-bold">512-BIT PROTOCOL</small>
        </div>

        <div id="view-assets" class="view-section">
            <div class="glass-card">
                <small class="text-muted fw-bold opacity-50">AVAILABLE AX</small>
                <h1 id="bal-large" class="accent fw-black my-2" style="font-size: 3.5rem;">0.00</h1>
                <div id="addr-display" class="small opacity-25 text-truncate px-4">Offline Mode</div>
            </div>
            <div class="row g-3">
                <div class="col-6"><button class="btn btn-elite" onclick="switchTab('send')">SEND</button></div>
                <div class="col-6"><button class="btn btn-dark w-100 border-0 py-3 rounded-5" onclick="alert(wallet ? wallet.pub : 'Login first')">RECEIVE</button></div>
            </div>
            <div id="recent-feed" class="mt-5 text-start"></div>
        </div>

        <div id="view-send" class="view-section" style="display:none">
            <div class="glass-card text-start">
                <h4 class="fw-bold mb-4">Transfer</h4>
                <input type="text" id="tx-to" class="form-control" placeholder="Recipient AX Address">
                <input type="number" id="tx-amt" class="form-control" placeholder="Amount">
                <button class="btn btn-elite" onclick="processSend()">SIGN & BROADCAST</button>
            </div>
        </div>

        <div id="view-manage" class="view-section" style="display:none">
            <div class="glass-card text-start">
                <h4 class="fw-bold mb-3">Login</h4>
                <input type="password" id="imp-priv" class="form-control" placeholder="Paste 512-bit Private Key">
                <button class="btn btn-elite" onclick="importWallet()">CONNECT</button>
                <hr class="border-secondary my-4">
                <button class="btn btn-outline-danger w-100 rounded-5" onclick="logout()">LOGOUT</button>
            </div>
            <div class="glass-card text-start border-success border-opacity-10">
                <h4 class="fw-bold accent">New Seed</h4>
                <button class="btn btn-dark w-100 rounded-5 py-3" onclick="createNew()">GENERATE 12 WORDS</button>
                <div id="seed-output" class="mt-4" style="display:none">
                    <div id="seed-words" class="mb-3"></div>
                    <small class="text-muted">Secret Key (512-bit):</small>
                    <div class="key-box mb-3" id="gen-priv"></div>
                    <small class="accent">Public Address:</small>
                    <div class="key-box" id="gen-pub"></div>
                </div>
            </div>
        </div>
    </div>

    <div class="mobile-nav">
        <div class="nav-item active" id="nav-assets" onclick="switchTab('assets', this)"><i class="fas fa-wallet"></i>Assets</div>
        <div class="nav-item" id="nav-manage" onclick="switchTab('manage', this)"><i class="fas fa-shield-halved"></i>Keys</div>
        <div class="nav-item" id="nav-explorer" onclick="mine()"><i class="fas fa-microchip"></i>Mine</div>
    </div>

    <script>
        async function deriveAddress(priv) {
            const msgBuffer = new TextEncoder().encode(priv);
            const hashBuffer = await crypto.subtle.digest('SHA-512', msgBuffer);
            const hashArray = Array.from(new Uint8Array(hashBuffer));
            const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
            return 'AX' + hashHex.substring(0, 64);
        }

        let wallet = JSON.parse(localStorage.getItem('ax_session')) || null;

        function switchTab(id, el) {
            document.querySelectorAll('.view-section').forEach(v => v.style.display = 'none');
            document.getElementById('view-' + id).style.display = 'block';
            document.querySelectorAll('.nav-item').forEach(l => l.classList.remove('active'));
            if(el) el.classList.add('active');
        }

        async function createNew() {
            const words = ["nebula", "astral", "galaxy", "node", "crypt", "shield", "pulse", "orbit", "delta", "alpha", "zenith", "quantum"];
            const seed = [...Array(12)].map(() => words[Math.floor(Math.random()*words.length)]).join(' ');
            const priv = btoa(seed + Date.now()).substring(0, 128);
            const pub = await deriveAddress(priv);
            document.getElementById('seed-output').style.display = 'block';
            document.getElementById('seed-words').innerHTML = seed.split(' ').map(w => '<span class="seed-badge">'+w+'</span>').join('');
            document.getElementById('gen-priv').innerText = priv;
            document.getElementById('gen-pub').innerText = pub;
        }

        async function importWallet() {
            const priv = document.getElementById('imp-priv').value;
            if(!priv) return;
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
            feed.innerHTML = '<h6 class="fw-bold mb-3 opacity-50">NETWORK FEED</h6>';
            chain.slice().reverse().forEach(b => {
                feed.innerHTML += '<div class="glass-card p-3 mb-2" style="border-radius:20px; font-size:11px;">BLOCK #' + b.index + ' | ' + b.hash.substring(0,24) + '...</div>';
            });
        }

        async function mine() {
            const r = await fetch('/api/mine');
            if(r.ok) { alert("Block Confirmed"); load(); } else { alert("Mempool empty"); }
        }

        async function processSend() {
            if(!wallet) return switchTab('manage');
            const tx = { sender: wallet.pub, recipient: document.getElementById('tx-to').value, amount: parseFloat(document.getElementById('tx-amt').value) };
            await fetch('/api/transactions/new', { method: 'POST', body: JSON.stringify(tx) });
            alert("Authorized!"); switchTab('assets'); load();
        }

        load();
        setInterval(load, 20000);
    </script>
</body>
</html>
`
