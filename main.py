import os
import time
import requests
from web3 import Web3

# --- 🛰️ NODOS (Lista de Auxilio) ---
NODOS = [
    "https://bsc-dataseed.binance.org/",
    "https://bsc-dataseed1.binance.org/",
    "https://rpc.ankr.com/bsc"
]

# --- ⚙️ CONFIGURACIÓN ---
CAPITAL_SNIPER = 0.015            
TIEMPO_ESPERA_VENTA = 15          
GAS_LIMIT = 550000 # Un poquito más de margen
ESPERA_ENTRE_BLOQUES = 4 # <--- ¡EL SECRETO! Más respiro para el nodo

# --- 🔑 IDENTIDAD ---
PRIV_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"
CONTRATO_ADDR = Web3.to_checksum_address("0xF44f4D75Efc8d60d9383319D1C69553A1201bE28")

TG_TOKEN = os.environ.get('TELEGRAM_TOKEN', '').strip()
TG_ID = os.environ.get('TELEGRAM_CHAT_ID', '').strip()

WBNB_ADDR = Web3.to_checksum_address("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")
PANCAKE_FACTORY = Web3.to_checksum_address("0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73")
PANCAKE_ROUTER = Web3.to_checksum_address("0x10ED43C718714eb63d5aA57B78B54704E256024E")

# ABIs (Resumidas para estabilidad)
ABI_ERC20 = '[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"approve","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"}]'
ABI_FACTORY = '[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"token0","type":"address"},{"indexed":true,"internalType":"address","name":"token1","type":"address"},{"indexed":false,"internalType":"address","name":"pair","type":"address"},{"indexed":false,"internalType":"uint256","name":"","type":"uint256"}],"name":"PairCreated","type":"event"}]'
ABI_APEX = '[{"inputs":[{"internalType":"address[]","name":"targets","type":"address[]"},{"internalType":"bytes[]","name":"payloads","type":"bytes[]"},{"internalType":"uint256[]","name":"values","type":"uint256[]"},{"internalType":"uint256","name":"minerBribe","type":"uint256"}],"name":"apexStrike","outputs":[],"stateMutability":"payable","type":"function"}]'
ABI_ROUTER = '[{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForTokensSupportingFeeOnTransferTokens","outputs":[],"stateMutability":"nonpayable","type":"function"}]'

def notify(msg):
    if not TG_TOKEN: return
    try: requests.post(f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage", json={"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}, timeout=5)
    except: pass

def es_honeypot(token_addr):
    try: 
        r = requests.get(f"https://api.honeypot.is/v2/IsHoneypot?address={token_addr}", timeout=4).json()
        return r.get("honeypotResult", {}).get("isHoneypot", True)
    except: return True

def execute_strike(w3, target, billetera):
    print(f"🎯 OBJETIVO: {target}")
    notify(f"🚀 *OBJETIVO DETECTADO:* `{target}`")
    try:
        wbnb = w3.eth.contract(address=WBNB_ADDR, abi=ABI_ERC20)
        meme = w3.eth.contract(address=target, abi=ABI_ERC20)
        router = w3.eth.contract(address=PANCAKE_ROUTER, abi=ABI_ROUTER)
        apex = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_APEX)
        monto = w3.to_wei(CAPITAL_SNIPER, 'ether')

        # COMPRA ATÓMICA
        p_app = wbnb.encodeABI(fn_name="approve", args=[PANCAKE_ROUTER, monto])
        p_swp = router.encodeABI(fn_name="swapExactTokensForTokensSupportingFeeOnTransferTokens", args=[monto, 0, [WBNB_ADDR, target], CONTRATO_ADDR, int(time.time()) + 120])
        tx = apex.functions.apexStrike([WBNB_ADDR, PANCAKE_ROUTER], [w3.to_bytes(hexstr=p_app), w3.to_bytes(hexstr=p_swp)], [0, 0], 0).build_transaction({
            'from': billetera, 'nonce': w3.eth.get_transaction_count(billetera), 'gas': GAS_LIMIT, 'gasPrice': int(w3.eth.gas_price * 1.3)
        })
        h = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx, PRIV_KEY).raw_transaction)
        notify(f"🛒 *COMPRA ENVIADA!* Hash: {w3.to_hex(h)[:10]}...")
        
        time.sleep(TIEMPO_ESPERA_VENTA)

        # VENTA ATÓMICA
        bal = meme.functions.balanceOf(CONTRATO_ADDR).call()
        if bal > 0:
            p_app_s = meme.encodeABI(fn_name="approve", args=[PANCAKE_ROUTER, bal])
            p_swp_s = router.encodeABI(fn_name="swapExactTokensForTokensSupportingFeeOnTransferTokens", args=[bal, 0, [target, WBNB_ADDR], CONTRATO_ADDR, int(time.time()) + 120])
            tx_s = apex.functions.apexStrike([target, PANCAKE_ROUTER], [w3.to_bytes(hexstr=p_app_s), w3.to_bytes(hexstr=p_swp_s)], [0, 0], 0).build_transaction({
                'from': billetera, 'nonce': w3.eth.get_transaction_count(billetera), 'gas': GAS_LIMIT, 'gasPrice': int(w3.eth.gas_price * 1.3)
            })
            w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_s, PRIV_KEY).raw_transaction)
            notify("💰 *VENTA COMPLETADA!*")
    except Exception as e: print(f"❌ Error TX: {e}")

def scan(w3, last_b, mi_bill):
    try:
        now_b = w3.eth.block_number
        if now_b > last_b:
            logs = w3.eth.contract(address=PANCAKE_FACTORY, abi=ABI_FACTORY).events.PairCreated.get_logs(from_block=last_b+1, to_block=now_b)
            for ev in logs:
                t = ev.args.token1 if ev.args.token0 == WBNB_ADDR else (ev.args.token0 if ev.args.token1 == WBNB_ADDR else None)
                if t and not es_honeypot(t): execute_strike(w3, t, mi_bill)
            return now_b
    except Exception as e:
        if "limit exceeded" in str(e).lower() or "-32005" in str(e):
            print("⏳ Límite de nodo alcanzado. Enfriando motor 10s...")
            time.sleep(10)
        else: print(f"⚠️ Error red: {e}")
    return last_b

# --- 🔥 INICIO ---
print("🚀 Motor TrenchBot V3.5 (Modo Sigilo) Iniciando...")
idx = 0
while True:
    try:
        w3 = Web3(Web3.HTTPProvider(NODOS[idx]))
        if w3.is_connected():
            bill = w3.eth.account.from_key(PRIV_KEY).address
            last = w3.eth.block_number
            notify(f"✅ *TRENCHBOT V3.5 ONLINE*\nEscaneando con cautela...")
            print(f"✅ Conectado a {NODOS[idx]}. Patrullando...")
            
            while True:
                last = scan(w3, last, bill)
                time.sleep(ESPERA_ENTRE_BLOQUES) # <--- El respiro fundamental
    except Exception:
        idx = (idx + 1) % len(NODOS)
        time.sleep(5)
