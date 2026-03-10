import os
import time
import requests
from web3 import Web3

# --- 🛰️ CONEXIÓN1 ---
RPC_URL = "https://bsc-dataseed.binance.org/"
w3 = Web3(Web3.HTTPProvider(RPC_URL))

# --- ⚙️ CONFIGURACIÓN ---
CAPITAL_SNIPER = 0.015            
TIEMPO_ESPERA_VENTA = 15          
GAS_LIMIT = 500000                

# --- 🔑 IDENTIDAD ---
CONTRATO_ADDR = w3.to_checksum_address("0xF44f4D75Efc8d60d9383319D1C69553A1201bE28")
PRIV_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"
MI_BILLETERA = w3.eth.account.from_key(PRIV_KEY).address 

TG_TOKEN = os.environ.get('TELEGRAM_TOKEN', '').strip()
TG_ID = os.environ.get('TELEGRAM_CHAT_ID', '').strip()

WBNB_ADDR = w3.to_checksum_address("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")
PANCAKE_FACTORY = w3.to_checksum_address("0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73")
PANCAKE_ROUTER = w3.to_checksum_address("0x10ED43C718714eb63d5aA57B78B54704E256024E")

# --- 📜 ABIs ---
ABI_ROUTER = '[{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForTokensSupportingFeeOnTransferTokens","outputs":[],"stateMutability":"nonpayable","type":"function"}]'
ABI_ERC20 = '[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"approve","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"}]'
ABI_FACTORY = '[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"token0","type":"address"},{"indexed":true,"internalType":"address","name":"token1","type":"address"},{"indexed":false,"internalType":"address","name":"pair","type":"address"},{"indexed":false,"internalType":"uint256","name":"","type":"uint256"}],"name":"PairCreated","type":"event"}]'
ABI_APEX = '[{"inputs":[{"internalType":"address[]","name":"targets","type":"address[]"},{"internalType":"bytes[]","name":"payloads","type":"bytes[]"},{"internalType":"uint256[]","name":"values","type":"uint256[]"},{"internalType":"uint256","name":"minerBribe","type":"uint256"}],"name":"apexStrike","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"}],"name":"emergencySweep","outputs":[],"stateMutability":"nonpayable","type":"function"}]'

last_update_id = 0

def notify(msg, buttons=False):
    if not TG_TOKEN: return
    url = f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage"
    data = {"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}
    if buttons: data["reply_markup"] = {"keyboard": [[{"text": "/status"}, {"text": "/balance"}]], "resize_keyboard": True}
    try: requests.post(url, json=data, timeout=5)
    except: pass

def check_commands():
    global last_update_id
    try:
        r = requests.get(f"https://api.telegram.org/bot{TG_TOKEN}/getUpdates?offset={last_update_id + 1}", timeout=5).json()
        for u in r.get("result", []):
            last_update_id = u["update_id"]
            txt = u.get("message", {}).get("text", "")
            if txt == "/status": notify("🛰️ *TrenchBot V3 Online*")
            elif txt == "/balance":
                c = w3.eth.contract(address=WBNB_ADDR, abi=ABI_ERC20)
                b = w3.from_wei(c.functions.balanceOf(CONTRATO_ADDR).call(), 'ether')
                notify(f"🏦 *Munición:* {b:.6f} WBNB")
    except: pass

def es_honeypot(token_addr):
    try: return requests.get(f"https://api.honeypot.is/v2/IsHoneypot?address={token_addr}", timeout=3).json().get("honeypotResult", {}).get("isHoneypot", True)
    except: return True

def execute_hit_and_run(target_token):
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
            'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA), 'gas': GAS_LIMIT, 'gasPrice': int(w3.eth.gas_price * 1.5)
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
                'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA), 'gas': GAS_LIMIT, 'gasPrice': int(w3.eth.gas_price * 1.5)
            })
            w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_s, PRIV_KEY).raw_transaction)
            notify("💰 *VENTA COMPLETADA!*")
    except Exception as e: notify(f"❌ Error: {str(e)[:50]}")

def scan(last_b):
    now_b = w3.eth.block_number
    if now_b > last_b:
        try:
            logs = w3.eth.contract(address=PANCAKE_FACTORY, abi=ABI_FACTORY).events.PairCreated.get_logs(fromBlock=last_b+1, toBlock=now_b)
            for ev in logs:
                t = ev.args.token1 if ev.args.token0 == WBNB_ADDR else (ev.args.token0 if ev.args.token1 == WBNB_ADDR else None)
                if t and not es_honeypot(t): execute_hit_and_run(t)
        except: pass
    return now_b

# --- 🔥 BUCLE INMORTAL ---
print("🚀 Motor TrenchBot V3 Iniciando...")
while True:
    try:
        if w3.is_connected():
            last_block = w3.eth.block_number
            notify("💰 *TRENCHBOT V3 ONLINE*", buttons=True)
            print("✅ Bot operativo y patrullando.")
            while True:
                check_commands()
                last_block = scan(last_block)
                time.sleep(1)
    except Exception as e:
        print(f"⚠️ Micro-corte detectado. Reintentando en 5s... ({e})")
        time.sleep(5)
