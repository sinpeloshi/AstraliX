import os
import time
import requests
from web3 import Web3

# --- 🛰️ CONEXIÓN PRIVADA ---
RPC_URL = "https://solemn-orbital-thunder.bsc.quiknode.pro/70d0d80f07303278accd2349e2fc01c95018d18c/"
w3 = Web3(Web3.HTTPProvider(RPC_URL))

# --- ⚙️ CONFIGURACIÓN ---
CAPITAL_SNIPER = 0.015            
TIEMPO_ESPERA_VENTA = 15          
GAS_LIMIT = 600000 
ESPERA_ENTRE_BLOQUES = 2 

# --- 🔑 IDENTIDAD ---
PRIV_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"
CONTRATO_ADDR = w3.to_checksum_address("0xF44f4D75Efc8d60d9383319D1C69553A1201bE28")
MI_BILLETERA = w3.eth.account.from_key(PRIV_KEY).address 

TG_TOKEN = '8783847744:AAHdwwlEqP7HCgSXoFxRdD8snr5FRhT1OUo'
TG_ID = '6580309816'

WBNB_ADDR = w3.to_checksum_address("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")
PANCAKE_FACTORY = w3.to_checksum_address("0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73")
PANCAKE_ROUTER = w3.to_checksum_address("0x10ED43C718714eb63d5aA57B78B54704E256024E")

# ABIs
ABI_MIN = '[{"anonymous":false,"inputs":[{"indexed":true,"name":"token0","type":"address"},{"indexed":true,"name":"token1","type":"address"},{"indexed":false,"name":"pair","type":"address"},{"indexed":false,"name":"length","type":"uint256"}],"name":"PairCreated","type":"event"}]'
ABI_ERC20 = '[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"approve","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"}]'
ABI_APEX = '[{"inputs":[{"internalType":"address[]","name":"targets","type":"address[]"},{"internalType":"bytes[]","name":"payloads","type":"bytes[]"},{"internalType":"uint256[]","name":"values","type":"uint256[]"},{"internalType":"uint256","name":"minerBribe","type":"uint256"}],"name":"apexStrike","outputs":[],"stateMutability":"payable","type":"function"}]'
ABI_ROUTER = '[{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForTokensSupportingFeeOnTransferTokens","outputs":[],"stateMutability":"nonpayable","type":"function"}]'

def notify(msg):
    try: requests.post(f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage", json={"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}, timeout=5)
    except: pass

def es_honeypot(token_addr):
    try:
        r = requests.get(f"https://api.honeypot.is/v2/IsHoneypot?address={token_addr}", timeout=4).json()
        return r.get("honeypotResult", {}).get("isHoneypot", True)
    except: return True

def execute_strike(target_token):
    print(f"🎯 OBJETIVO: {target_token}")
    notify(f"🚀 *OBJETIVO DETECTADO:* `{target_token}`")
    try:
        wbnb_c = w3.eth.contract(address=WBNB_ADDR, abi=ABI_ERC20)
        meme_c = w3.eth.contract(address=target_token, abi=ABI_ERC20)
        router_c = w3.eth.contract(address=PANCAKE_ROUTER, abi=ABI_ROUTER)
        apex_c = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_APEX)
        monto = w3.to_wei(CAPITAL_SNIPER, 'ether')

        # --- FIX V4.2: SINTAXIS DIRECTA SIN "fn_name=" ---
        p_app = wbnb_c.encode_abi("approve", args=[PANCAKE_ROUTER, monto])
        p_swp = router_c.encode_abi("swapExactTokensForTokensSupportingFeeOnTransferTokens", args=[monto, 0, [WBNB_ADDR, target_token], CONTRATO_ADDR, int(time.time()) + 120])
        
        tx = apex_c.functions.apexStrike(
            [WBNB_ADDR, PANCAKE_ROUTER], 
            [w3.to_bytes(hexstr=p_app), w3.to_bytes(hexstr=p_swp)], 
            [0, 0], 0
        ).build_transaction({
            'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA),
            'gas': GAS_LIMIT, 'gasPrice': int(w3.eth.gas_price * 1.5)
        })
        
        h_buy = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx, PRIV_KEY).raw_transaction)
        notify(f"🛒 *COMPRA ENVIADA!*")
        
        time.sleep(TIEMPO_ESPERA_VENTA)
        
        # --- VENTA ---
        bal = meme_c.functions.balanceOf(CONTRATO_ADDR).call()
        if bal > 0:
            p_app_s = meme_c.encode_abi("approve", args=[PANCAKE_ROUTER, bal])
            p_swp_s = router_c.encode_abi("swapExactTokensForTokensSupportingFeeOnTransferTokens", args=[bal, 0, [target_token, WBNB_ADDR], CONTRATO_ADDR, int(time.time()) + 120])
            
            tx_s = apex_c.functions.apexStrike(
                [target_token, PANCAKE_ROUTER], 
                [w3.to_bytes(hexstr=p_app_s), w3.to_bytes(hexstr=p_swp_s)], 
                [0, 0], 0
            ).build_transaction({
                'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA),
                'gas': GAS_LIMIT, 'gasPrice': int(w3.eth.gas_price * 1.5)
            })
            w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_s, PRIV_KEY).raw_transaction)
            notify("💰 *VENTA COMPLETADA!*")
    except Exception as e:
        print(f"❌ Error TX: {e}")
        notify(f"❌ Error en operacion: {str(e)[:50]}")

def scan(last_b):
    try:
        now_b = w3.eth.block_number
        if now_b <= last_b: return last_b
        if (now_b - last_b) > 5: last_b = now_b - 1

        logs = w3.eth.contract(address=PANCAKE_FACTORY, abi=ABI_MIN).events.PairCreated.get_logs(from_block=last_b+1, to_block=now_b)
        for ev in logs:
            t = ev.args.token1 if ev.args.token0 == WBNB_ADDR else (ev.args.token0 if ev.args.token1 == WBNB_ADDR else None)
            if t and not es_honeypot(t): execute_strike(t)
        return now_b
    except Exception as e:
        return w3.eth.block_number

print("🚀 Motor TrenchBot V4.2 (Gatillo de Pelo) Iniciando...")
if w3.is_connected():
    last_block = w3.eth.block_number
    notify("💰 *TRENCHBOT V4.2 ONLINE*")
    timer_hb = time.time()
    while True:
        try:
            last_block = scan(last_block)
            if time.time() - timer_hb > 30:
                print(f"🔎 Patrullando bloque {last_block}...")
                timer_hb = time.time()
            time.sleep(ESPERA_ENTRE_BLOQUES)
        except Exception: time.sleep(5)