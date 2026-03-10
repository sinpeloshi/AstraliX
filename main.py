import os
import time
import requests
from web3 import Web3

# --- 🛰️ CONEXIÓN PRIVADA ---
RPC_URL = "https://solemn-orbital-thunder.bsc.quiknode.pro/70d0d80f07303278accd2349e2fc01c95018d18c/"
w3 = Web3(Web3.HTTPProvider(RPC_URL))

# --- ⚙️ CONFIGURACIÓN ---
CAPITAL_SNIPER = 0.005 # Ajustado para tu saldo actual
TARGET_PROFIT = 1.15   # Buscamos un 15% de ganancia real
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

def get_val(token, amount):
    try:
        r = w3.eth.contract(address=PANCAKE_ROUTER, abi=ABI_ROUTER)
        return r.functions.getAmountsOut(amount, [token, WBNB_ADDR]).call()[1]
    except: return 0

def execute_sell(token):
    try:
        meme_c = w3.eth.contract(address=token, abi=ABI_ERC20)
        router_c = w3.eth.contract(address=PANCAKE_ROUTER, abi=ABI_ROUTER)
        apex_c = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_APEX)
        bal = meme_c.functions.balanceOf(CONTRATO_ADDR).call()
        if bal > 0:
            p_app = meme_c.encode_abi("approve", args=[PANCAKE_ROUTER, bal])
            p_swp = router_c.encode_abi("swapExactTokensForTokensSupportingFeeOnTransferTokens", args=[bal, 0, [token, WBNB_ADDR], CONTRATO_ADDR, int(time.time()) + 120])
            tx = apex_c.functions.apexStrike([token, PANCAKE_ROUTER], [w3.to_bytes(hexstr=p_app), w3.to_bytes(hexstr=p_swp)], [0, 0], 0).build_transaction({
                'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA), 'gas': GAS_LIMIT, 'gasPrice': int(w3.eth.gas_price * 1.6)
            })
            w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx, PRIV_KEY).raw_transaction)
            notify("💰 *VENTA EXITOSA (PROFIT ALCANZADO)*")
            return True
    except: pass
    return False

def monitor(token, invertido):
    print(f"👀 Vigilando profit de {token}...")
    meme_c = w3.eth.contract(address=token, abi=ABI_ERC20)
    start_time = time.time()
    while time.time() - start_time < 300: # 5 minutos de vigilancia
        try:
            bal = meme_c.functions.balanceOf(CONTRATO_ADDR).call()
            if bal > 0:
                actual = get_val(token, bal)
                if actual >= int(invertido * TARGET_PROFIT):
                    if execute_sell(token): break
        except: pass
        time.sleep(3)

def execute_buy(target):
    print(f"🚀 LANZANDO ATAQUE A: {target}")
    try:
        wbnb_c = w3.eth.contract(address=WBNB_ADDR, abi=ABI_ERC20)
        router_c = w3.eth.contract(address=PANCAKE_ROUTER, abi=ABI_ROUTER)
        apex_c = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_APEX)
        monto = w3.to_wei(CAPITAL_SNIPER, 'ether')

        p_app = wbnb_c.encode_abi("approve", args=[PANCAKE_ROUTER, monto])
        p_swp = router_c.encode_abi("swapExactTokensForTokensSupportingFeeOnTransferTokens", args=[monto, 0, [WBNB_ADDR, target], CONTRATO_ADDR, int(time.time()) + 120])
        
        tx = apex_c.functions.apexStrike([WBNB_ADDR, PANCAKE_ROUTER], [w3.to_bytes(hexstr=p_app), w3.to_bytes(hexstr=p_swp)], [0, 0], 0).build_transaction({
            'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA), 'gas': GAS_LIMIT, 'gasPrice': int(w3.eth.gas_price * 1.6)
        })
        
        w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx, PRIV_KEY).raw_transaction)
        notify(f"🛒 *COMPRA EJECUTADA:* `{target}`\nMonitoreando profit...")
        monitor(target, monto)
    except Exception as e:
        print(f"❌ Error en Ataque: {e}")

def scan(last_b):
    try:
        now_b = w3.eth.block_number
        if now_b <= last_b: return last_b
        factory = w3.eth.contract(address=PANCAKE_FACTORY, abi=ABI_FACTORY)
        logs = factory.events.PairCreated.get_logs(from_block=last_b+1, to_block=now_b)
        for ev in logs:
            t = ev.args.token1 if ev.args.token0 == WBNB_ADDR else (ev.args.token0 if ev.args.token1 == WBNB_ADDR else None)
            if t and t.lower().endswith(('444', '777')):
                print(f"🎯 OBJETIVO DETECTADO: {t}")
                execute_buy(t) # <--- AQUÍ ESTÁ EL GATILLO AHORA
        return now_b
    except: return w3.eth.block_number

print("🚀 AstraliX V7.2 (Gatillo Real) Iniciando...")
if w3.is_connected():
    last_block = w3.eth.block_number
    notify("🛡️ *ASTRALIX V7.2 ONLINE*")
    timer_hb = time.time()
    while True:
        try:
            last_block = scan(last_block)
            if time.time() - timer_hb > 30:
                print(f"🔎 Patrullando bloque {last_block}...")
                timer_hb = time.time()
            time.sleep(ESPERA_ENTRE_BLOQUES)
        except: time.sleep(5)