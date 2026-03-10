import os
import time
import requests
from web3 import Web3

# --- 🛰️ CONEXIÓN PRIVADA ---
RPC_URL = "https://solemn-orbital-thunder.bsc.quiknode.pro/70d0d80f07303278accd2349e2fc01c95018d18c/"
w3 = Web3(Web3.HTTPProvider(RPC_URL))

# --- ⚙️ CONFIGURACIÓN ---
CAPITAL_SNIPER = 0.005 # Capital por trade
TARGET_PROFIT = 1.20   # 1.20 = 20% de ganancia. 1.50 = 50%, etc.
GAS_LIMIT = 600000 
ESPERA_ENTRE_BLOQUES = 2 

# --- 🔑 IDENTIDAD ---
PRIV_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"
CONTRATO_ADDR = w3.to_checksum_address("0xF44f4D75Efc8d60d9383319D1C69553A1201be28")
MI_BILLETERA = w3.eth.account.from_key(PRIV_KEY).address 

TG_TOKEN = '8783847744:AAHdwwlEqP7HCgSXoFxRdD8snr5FRhT1OUo'
TG_ID = '6580309816'

WBNB_ADDR = w3.to_checksum_address("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")
PANCAKE_FACTORY = w3.to_checksum_address("0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73")
PANCAKE_ROUTER = w3.to_checksum_address("0x10ED43C718714eb63d5aA57B78B54704E256024E")

# ABIs
ABI_ERC20 = '[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"},{"inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"name":"approve","outputs":[{"name":"bool","name":"","type":"bool"}],"type":"function"}]'
ABI_FACTORY = '[{"anonymous":false,"inputs":[{"indexed":true,"name":"token0","type":"address"},{"indexed":true,"name":"token1","type":"address"},{"indexed":false,"name":"pair","type":"address"},{"indexed":false,"name":"length","type":"uint256"}],"name":"PairCreated","type":"event"}]'
ABI_APEX = '[{"inputs":[{"name":"targets","type":"address[]"},{"name":"payloads","type":"bytes[]"},{"name":"values","type":"uint256[]"},{"name":"minerBribe","type":"uint256"}],"name":"apexStrike","outputs":[],"type":"function"}]'
ABI_ROUTER = '[{"inputs":[{"name":"amountIn","type":"uint256"},{"name":"amountOutMin","type":"uint256"},{"name":"path","type":"address[]"},{"name":"to","type":"address"},{"name":"deadline","type":"uint256"}],"name":"swapExactTokensForTokensSupportingFeeOnTransferTokens","outputs":[],"type":"function"},{"inputs":[{"name":"amountIn","type":"uint256"},{"name":"path","type":"address[]"}],"name":"getAmountsOut","outputs":[{"name":"amounts","type":"uint256[]"}],"type":"function"}]'

def notify(msg):
    try: requests.post(f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage", json={"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}, timeout=5)
    except: pass

def get_current_value(token_addr, amount_in):
    try:
        router = w3.eth.contract(address=PANCAKE_ROUTER, abi=ABI_ROUTER)
        amounts = router.functions.getAmountsOut(amount_in, [token_addr, WBNB_ADDR]).call()
        return amounts[1]
    except: return 0

def execute_sell(token_addr):
    try:
        meme_c = w3.eth.contract(address=token_addr, abi=ABI_ERC20)
        router_c = w3.eth.contract(address=PANCAKE_ROUTER, abi=ABI_ROUTER)
        apex_c = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_APEX)
        bal = meme_c.functions.balanceOf(CONTRATO_ADDR).call()
        
        if bal > 0:
            p_app_s = meme_c.encode_abi("approve", args=[PANCAKE_ROUTER, bal])
            p_swp_s = router_c.encode_abi("swapExactTokensForTokensSupportingFeeOnTransferTokens", args=[bal, 0, [token_addr, WBNB_ADDR], CONTRATO_ADDR, int(time.time()) + 120])
            tx_s = apex_c.functions.apexStrike([token_addr, PANCAKE_ROUTER], [w3.to_bytes(hexstr=p_app_s), w3.to_bytes(hexstr=p_swp_s)], [0, 0], 0).build_transaction({
                'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA), 'gas': GAS_LIMIT, 'gasPrice': int(w3.eth.gas_price * 1.5)
            })
            w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_s, PRIV_KEY).raw_transaction)
            notify(f"💰 *VENTA EJECUTADA POR PROFIT!*")
            return True
    except Exception as e: print(f"❌ Error Venta: {e}")
    return False

def monitor_profit(token_addr, monto_invertido):
    print(f"👀 Monitoreando profit para {token_addr}...")
    meme_c = w3.eth.contract(address=token_addr, abi=ABI_ERC20)
    start_time = time.time()
    
    while time.time() - start_time < 600: # Monitorea por 10 minutos máximo
        bal = meme_c.functions.balanceOf(CONTRATO_ADDR).call()
        if bal > 0:
            current_wbnb_val = get_current_value(token_addr, bal)
            objetivo = int(monto_invertido * TARGET_PROFIT)
            
            if current_wbnb_val >= objetivo:
                if execute_sell(token_addr): break
        time.sleep(2)

def execute_buy(target):
    print(f"🎯 COMPRANDO: {target}")
    try:
        wbnb_c = w3.eth.contract(address=WBNB_ADDR, abi=ABI_ERC20)
        router_c = w3.eth.contract(address=PANCAKE_ROUTER, abi=ABI_ROUTER)
        apex_c = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_APEX)
        monto = w3.to_wei(CAPITAL_SNIPER, 'ether')

        p_app = wbnb_c.encode_abi("approve", args=[PANCAKE_ROUTER, monto])
        p_swp = router_c.encode_abi("swapExactTokensForTokensSupportingFeeOnTransferTokens", args=[monto, 0, [WBNB_ADDR, target], CONTRATO_ADDR, int(time.time()) + 120])
        
        tx = apex_c.functions.apexStrike([WBNB_ADDR, PANCAKE_ROUTER], [w3.to_bytes(hexstr=p_app), w3.to_bytes(hexstr=p_swp)], [0, 0], 0).build_transaction({
            'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA), 'gas': GAS_LIMIT, 'gasPrice': int(w3.eth.gas_price * 1.5)
        })
        
        w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx, PRIV_KEY).raw_transaction)
        notify(f"🛒 *COMPRA OK:* `{target}`\nIniciando auto-monitoreo de profit...")
        
        # Después de comprar, entra a vigilar el profit
        monitor_profit(target, monto)
        
    except Exception as e: print(f"❌ Error Compra: {e}")

def scan(last_b):
    try:
        now_b = w3.eth.block_number
        if now_b <= last_b: return last_b
        logs = w3.eth.contract(address=PANCAKE_FACTORY, abi=Web3().eth.contract(abi=ABI_FACTORY).abi).events.PairCreated.get_logs(from_block=last_b+1, to_block=now_b)
        for ev in logs:
            t = ev.args.token1 if ev.args.token0 == WBNB_ADDR else (ev.args.token0 if ev.args.token1 == WBNB_ADDR else None)
            if t and t.lower().endswith(('444', '777')):
                execute_buy(t)
        return now_b
    except: return w3.eth.block_number

print("🚀 AstraliX V7.0 (Auto-Profit) Iniciando...")
if w3.is_connected():
    last_block = w3.eth.block_number
    notify("💰 *ASTRALIX V7.0 ONLINE*\nModo: Auto-Buy & Auto-Profit (20%)")
    while True:
        last_block = scan(last_block)
        time.sleep(ESPERA_ENTRE_BLOQUES)