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

	// 1. Genesis Setup
	genTx := core.Transaction{Sender: "SYSTEM", Recipient: creatorAddr, Amount: 1000002021}
	genTx.TxID = genTx.CalculateHash()
	genesis := core.Block{
		Index: 0, Timestamp: 1773561600,
		Transactions: []core.Transaction{genTx},
		PrevHash: strings.Repeat("0", 128), Difficulty: Difficulty,
	}
	genesis.Mine()
	Blockchain = append(Blockchain, genesis)

	// --- API ROUTES ---
	
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
		if len(Mempool) == 0 { http.Error(w, "Mempool empty", 400); return }
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

	// SERVIMOS EL DASHBOARD
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, dashboardHTML)
	})

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	fmt.Printf("🌐 AstraliX Live on port %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

// Interfaz Visual (HTML/CSS/JS)
const dashboardHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>AstraliX Network Dashboard</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <style>
        body { background: #0b0b0d; color: #f0f0f0; }
        .card { background: #15151a; border: 1px solid #333; border-radius: 12px; }
        .text-accent { color: #00ffa3; }
        .addr-text { font-family: monospace; font-size: 0.8rem; color: #aaa; word-break: break-all; }
    </style>
</head>
<body class="container py-4">
    <h1 class="text-center mb-4">🌐 AstraliX <span class="text-accent">Network</span></h1>
    <div class="row">
        <div class="col-md-4">
            <div class="card p-3 mb-3">
                <h3>Wallet</h3>
                <input type="text" id="w-addr" class="form-control bg-dark text-white mb-2" placeholder="AX address...">
                <button class="btn btn-success w-100" onclick="checkB()">Check Balance</button>
                <h2 id="bal" class="text-center text-accent my-3">0.00 AX</h2>
                <button class="btn btn-outline-light btn-sm w-100" onclick="genW()">New Wallet</button>
                <div id="new-w" class="mt-2 small" style="display:none">
                    <div class="addr-text" id="p-key"></div>
                    <div class="addr-text text-accent" id="a-key"></div>
                </div>
            </div>
        </div>
        <div class="col-md-8">
            <div class="card p-3">
                <div class="d-flex justify-content-between">
                    <h3>Explorer</h3>
                    <button class="btn btn-sm btn-info" onclick="mine()">Mine TXs</button>
                </div>
                <div id="chain" class="mt-3"></div>
            </div>
        </div>
    </div>
    <script>
        async function checkB() {
            const a = document.getElementById('w-addr').value;
            const r = await fetch('/api/balance/' + a);
            const d = await r.json();
            document.getElementById('bal').innerText = d.balance.toLocaleString() + ' AX';
        }
        async function load() {
            const r = await fetch('/api/chain');
            const c = await r.json();
            let h = '';
            c.reverse().forEach(b => {
                h += '<div class="border-bottom border-secondary mb-2"><b>Block #' + b.index + '</b><br><span class="addr-text">' + b.hash + '</span></div>';
            });
            document.getElementById('chain').innerHTML = h;
        }
        async function mine() {
            const r = await fetch('/api/mine');
            if(r.ok) { alert('Mined!'); load(); } else { alert('Mempool empty'); }
        }
        function genW() {
            const h = l => [...Array(l)].map(()=>Math.floor(Math.random()*16).toString(16)).join('');
            document.getElementById('new-w').style.display='block';
            document.getElementById('p-key').innerText='Priv: ' + h(64);
            document.getElementById('a-key').innerText='Addr: AX' + h(64);
        }
        load();
    </script>
</body>
</html>
`
