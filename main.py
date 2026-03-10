import os
import time
import requests
from web3 import Web3

# --- 🛰️ CONEXIÓN PRIVADA ---
RPC_URL = "https://solemn-orbital-thunder.bsc.quiknode.pro/70d0d80f07303278accd2349e2fc01c95018d18c/"
w3 = Web3(Web3.HTTPProvider(RPC_URL))

# --- ⚙️ CONFIGURACIÓN ---
CAPITAL_SNIPER = 0.005 
TARGET_PROFIT = 1.15   # Bajamos a 15% para asegurar el gas
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
    except: print("⚠️ Error enviando a Telegram")

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
            notify("💰 *VENTA EJECUTADA POR PROFIT!*")
            return True
    except: pass
    return False

def monitor_profit(token_addr, monto_invertido):
    print(f"👀 Monitoreando profit para {token_addr}...")
    meme_c = w3.eth.contract(address=token_addr, abi=ABI_ERC20)
    start_t = time.time()
    while time.time() - start_t < 600:
        try:
            bal = meme_c.functions.balanceOf(CONTRATO_ADDR).call()
            if bal > 0:
                current_val = get_current_value(token_addr, bal)
                if current_val >= int(monto_invertido * TARGET_PROFIT):
                    if execute_sell(token_addr): break
        except: pass
        time.sleep(3)

def scan(last_b):
    try:
        now_b = w3.eth.block_number
        if now_b <= last_b: return last_b
        # Simplificamos la llamada al contrato Factory
        factory = w3.eth.contract(address=PANCAKE_FACTORY, abi=ABI_FACTORY)
        logs = factory.events.PairCreated.get_logs(from_block=last_b+1, to_block=now_b)
        for ev in logs:
            t = ev.args.token1 if ev.args.token0 == WBNB_ADDR else (ev.args.token0 if ev.args.token1 == WBNB_ADDR else None)
            if t and t.lower().endswith(('444', '777')):
                # Compra (simplificada para el ejemplo)
                print(f"🎯 OBJETIVO DETECTADO: {t}")
                notify(f"🚀 *OBJETIVO:* `{t}`")
        return now_b
    except Exception as e:
        print(f"⚠️ Error scan: {e}")
        return w3.eth.block_number

# --- 🔥 INICIO CON PUNTOS DE CONTROL ---
print("🚀 AstraliX V7.1 Iniciando...")

print("Paso 1: Verificando conexión al nodo...")
if w3.is_connected():
    print("✅ Nodo conectado correctamente.")
    
    print("Paso 2: Obteniendo bloque actual...")
    last_block = w3.eth.block_number
    print(f"✅ Bloque inicial: {last_block}")
    
    print("Paso 3: Notificando a Telegram...")
    notify("💰 *ASTRALIX V7.1 ONLINE*")
    
    print("✅ TODO LISTO. Empezando patrullaje...")
    timer_hb = time.time()
    while True:
        try:
            last_block = scan(last_block)
            if time.time() - timer_hb > 20:
                print(f"🔎 Patrullando bloque {last_block}...")
                timer_hb = time.time()
            time.sleep(ESPERA_ENTRE_BLOQUES)
        except Exception as e:
            print(f"🔄 Reintentando por error: {e}")
            time.sleep(5)
else:
    print("❌ ERROR: No se pudo conectar al RPC. Revisá tu URL de QuickNode.")