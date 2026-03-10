import os
import time
import requests
from web3 import Web3

# --- 🛰️ NODOS ESTRATÉGICOS (Redundancia Total) ---
NODOS = [
    "https://bsc-dataseed.binance.org/",
    "https://bsc-dataseed1.binance.org/",
    "https://rpc.ankr.com/bsc"
]

# --- ⚙️ CONFIGURACIÓN ---
CAPITAL_SNIPER = 0.015            
TIEMPO_ESPERA_VENTA = 15          
GAS_LIMIT = 500000                

# --- 🔑 IDENTIDAD ---
# Se calcula automáticamente de tu Priv Key
PRIV_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"
CONTRATO_ADDR = Web3.to_checksum_address("0xF44f4D75Efc8d60d9383319D1C69553A1201bE28")

TG_TOKEN = os.environ.get('TELEGRAM_TOKEN', '').strip()
TG_ID = os.environ.get('TELEGRAM_CHAT_ID', '').strip()

WBNB_ADDR = Web3.to_checksum_address("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")
PANCAKE_FACTORY = Web3.to_checksum_address("0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73")
PANCAKE_ROUTER = Web3.to_checksum_address("0x10ED43C718714eb63d5aA57B78B54704E256024E")

# --- 📜 ABIs ---
ABI_ERC20 = '[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"approve","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"}]'
ABI_FACTORY = '[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"token0","type":"address"},{"indexed":true,"internalType":"address","name":"token1","type":"address"},{"indexed":false,"internalType":"address","name":"pair","type":"address"},{"indexed":false,"internalType":"uint256","name":"","type":"uint256"}],"name":"PairCreated","type":"event"}]'
ABI_APEX = '[{"inputs":[{"internalType":"address[]","name":"targets","type":"address[]"},{"internalType":"bytes[]","name":"payloads","type":"bytes[]"},{"internalType":"uint256[]","name":"values","type":"uint256[]"},{"internalType":"uint256","name":"minerBribe","type":"uint256"}],"name":"apexStrike","outputs":[],"stateMutability":"payable","type":"function"}]'
ABI_ROUTER = '[{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForTokensSupportingFeeOnTransferTokens","outputs":[],"stateMutability":"nonpayable","type":"function"}]'

def notify(msg):
    if not TG_TOKEN: return
    try: requests.post(f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage", json={"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}, timeout=5)
    except: pass

def es_honeypot(token_addr):
    try: return requests.get(f"https://api.honeypot.is/v2/IsHoneypot?address={token_addr}", timeout=3).json().get("honeypotResult", {}).get("isHoneypot", True)
    except: return True

def execute_hit_and_run(w3, target_token, mi_billetera):
    print(f"🎯 OBJETIVO ENCONTRADO: {target_token}")
    notify(f"🚀 *OBJETIVO DETECTADO:* `{target_token}`")
    try:
        wbnb_contract = w3.eth.contract(address=WBNB_ADDR, abi=ABI_ERC20)
        target_contract = w3.eth.contract(address=target_token, abi=ABI_ERC20)
        router_contract = w3.eth.contract(address=PANCAKE_ROUTER, abi=ABI_ROUTER)
        apex_contract = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_APEX)
        monto = w3.to_wei(CAPITAL_SNIPER, 'ether')

        # COMPRA
        p_app = wbnb_contract.encodeABI(fn_name="approve", args=[PANCAKE_ROUTER, monto])
        p_swp = router_contract.encodeABI(fn_name="swapExactTokensForTokensSupportingFeeOnTransferTokens", args=[monto, 0, [WBNB_ADDR, target_token], CONTRATO_ADDR, int(time.time()) + 120])
        
        tx = apex_contract.functions.apexStrike([WBNB_ADDR, PANCAKE_ROUTER], [w3.to_bytes(hexstr=p_app), w3.to_bytes(hexstr=p_swp)], [0, 0], 0).build_transaction({
            'from': mi_billetera, 'nonce': w3.eth.get_transaction_count(mi_billetera), 'gas': GAS_LIMIT, 'gasPrice': int(w3.eth.gas_price * 1.3)
        })
        h = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx, PRIV_KEY).raw_transaction)
        w3.eth.wait_for_transaction_receipt(h, timeout=60)
        notify(f"🛒 *COMPRA OK!* Esperando {TIEMPO_ESPERA_VENTA}s...")

        time.sleep(TIEMPO_ESPERA_VENTA)

        # VENTA
        bal = target_contract.functions.balanceOf(CONTRATO_ADDR).call()
        if bal > 0:
            p_app_s = target_contract.encodeABI(fn_name="approve", args=[PANCAKE_ROUTER, bal])
            p_swp_s = router_contract.encodeABI(fn_name="swapExactTokensForTokensSupportingFeeOnTransferTokens", args=[bal, 0, [target_token, WBNB_ADDR], CONTRATO_ADDR, int(time.time()) + 120])
            tx_s = apex_contract.functions.apexStrike([target_token, PANCAKE_ROUTER], [w3.to_bytes(hexstr=p_app_s), w3.to_bytes(hexstr=p_swp_s)], [0, 0], 0).build_transaction({
                'from': mi_billetera, 'nonce': w3.eth.get_transaction_count(mi_billetera), 'gas': GAS_LIMIT, 'gasPrice': int(w3.eth.gas_price * 1.3)
            })
            w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_s, PRIV_KEY).raw_transaction)
            notify("💰 *VENTA COMPLETADA!*")
    except Exception as e:
        print(f"❌ Error en transaccion: {e}")

def scan(w3, last_b, mi_billetera):
    now_b = w3.eth.block_number
    if now_b > last_b:
        try:
            logs = w3.eth.contract(address=PANCAKE_FACTORY, abi=ABI_FACTORY).events.PairCreated.get_logs(from_block=last_b+1, to_block=now_b)
            for ev in logs:
                t = ev.args.token1 if ev.args.token0 == WBNB_ADDR else (ev.args.token0 if ev.args.token1 == WBNB_ADDR else None)
                if t and not es_honeypot(t): execute_hit_and_run(w3, t, mi_billetera)
        except Exception as e:
            print(f"⚠️ Error de red en bloque {now_b}. Saltando...")
    return now_b

# --- 🔥 INICIO ---
print("🚀 Motor TrenchBot V3.4 (Blindado Multi-Nodo) Iniciando...")
current_rpc_idx = 0

while True:
    try:
        rpc = NODOS[current_rpc_idx]
        w3 = Web3(Web3.HTTPProvider(rpc))
        
        if w3.is_connected():
            mi_billetera = w3.eth.account.from_key(PRIV_KEY).address
            last_block = w3.eth.block_number
            notify(f"💰 *TRENCHBOT V3.4 ONLINE*\nNodo: `{rpc.split('/')[2]}`")
            print(f"✅ Conectado a {rpc}. Escaneando...")
            
            timer_hb = time.time()
            while True:
                last_block = scan(w3, last_block, mi_billetera)
                if time.time() - timer_hb > 30:
                    print(f"🔎 Patrullando bloque {last_block}...")
                    timer_hb = time.time()
                time.sleep(1)
                
    except Exception as e:
        current_rpc_idx = (current_rpc_idx + 1) % len(NODOS)
        print(f"🔄 Cambiando a nodo secundario por error: {e}")
        time.sleep(3)
