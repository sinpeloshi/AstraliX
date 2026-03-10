import os
import time
import requests
from web3 import Web3

print("⏳ Iniciando motor TRENCHBOT V2.1 (Hit & Run Blindado)...")

# --- 🛰️ CONEXIÓN ---
RPC_URL = "https://bsc-dataseed.binance.org/"
w3 = Web3(Web3.HTTPProvider(RPC_URL))

if w3.is_connected():
    print("✅ Conectado a la BSC (Nodo Público)")

# --- ⚙️ CONFIGURACIÓN DE ATAQUE ---
CAPITAL_SNIPER = 0.015            
TIEMPO_ESPERA_VENTA = 15          
GAS_LIMIT = 500000                
MINER_BRIBE = 0                   

# --- 🔑 IDENTIDAD ---
CONTRATO_ADDR = w3.to_checksum_address("0xF44f4D75Efc8d60d9383319D1C69553A1201bE28")
MI_BILLETERA = w3.to_checksum_address(os.environ.get('MI_BILLETERA', '0xTuBilleteraAqui').strip())
PRIV_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"

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

# --- 📱 MÓDULO TELEGRAM ---
def notify(msg, buttons=False):
    if not TG_TOKEN: return
    url = f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage"
    data = {"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}
    if buttons:
        data["reply_markup"] = {"keyboard": [[{"text": "/status"}, {"text": "/balance"}], [{"text": "/retiro_wbnb"}]], "resize_keyboard": True}
    try: requests.post(url, json=data, timeout=5)
    except: pass

def check_commands():
    global last_update_id
    if not TG_TOKEN: return
    try:
        r = requests.get(f"https://api.telegram.org/bot{TG_TOKEN}/getUpdates?offset={last_update_id + 1}", timeout=5).json()
        for u in r.get("result", []):
            last_update_id = u["update_id"]
            txt = u.get("message", {}).get("text", "")
            if txt == "/status": 
                notify("🛰️ *TrenchBot V2.1 Activo*\nRadar Inmune a cortes. Auto-Venta (15s).", buttons=True)
            elif txt == "/balance":
                try:
                    c = w3.eth.contract(address=WBNB_ADDR, abi=ABI_ERC20)
                    b = w3.from_wei(c.functions.balanceOf(CONTRATO_ADDR).call(), 'ether')
                    notify(f"🏦 *Munición:* {b:.6f} WBNB", buttons=True)
                except: pass
            elif txt == "/retiro_wbnb":
                ejecutar_rescate(WBNB_ADDR)
    except: pass

def ejecutar_rescate(token_addr):
    notify("🚨 *Iniciando barrido de emergencia...*")
    try:
        c = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_APEX)
        tx = c.functions.emergencySweep(token_addr).build_transaction({
            'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA),
            'gas': 150000, 'gasPrice': w3.eth.gas_price
        })
        s = w3.eth.account.sign_transaction(tx, PRIV_KEY)
        h = w3.eth.send_raw_transaction(s.raw_transaction)
        notify(f"✅ *Barrido Exitoso!*\nHash: {w3.to_hex(h)}")
    except Exception as e: notify(f"❌ *Fallo al retirar:* {str(e)[:50]}")

def es_honeypot(token_addr):
    # API corregida para evitar falsos positivos
    try: return requests.get(f"https://api.honeypot.is/v2/IsHoneypot?address={token_addr}", timeout=3).json().get("honeypotResult", {}).get("isHoneypot", True)
    except: return True

# --- 🎯 GATILLO HIT & RUN ---
def execute_hit_and_run(target_token):
    notify(f"🚀 *OBJETIVO DETECTADO:* `{target_token}`\nArmando misil de COMPRA...")
    
    try:
        wbnb_contract = w3.eth.contract(address=WBNB_ADDR, abi=ABI_ERC20)
        target_contract = w3.eth.contract(address=target_token, abi=ABI_ERC20)
        router_contract = w3.eth.contract(address=PANCAKE_ROUTER, abi=ABI_ROUTER)
        apex_contract = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_APEX)
        
        monto_inversion = w3.to_wei(CAPITAL_SNIPER, 'ether')
        
        # --- FASE 1: COMPRA ATÓMICA ---
        p_approve_buy = wbnb_contract.encodeABI(fn_name="approve", args=[PANCAKE_ROUTER, monto_inversion])
        p_swap_buy = router_contract.encodeABI(fn_name="swapExactTokensForTokensSupportingFeeOnTransferTokens", args=[monto_inversion, 0, [WBNB_ADDR, target_token], CONTRATO_ADDR, int(time.time()) + 120])
        
        tx_buy = apex_contract.functions.apexStrike(
            [WBNB_ADDR, PANCAKE_ROUTER],
            [w3.to_bytes(hexstr=p_approve_buy), w3.to_bytes(hexstr=p_swap_buy)],
            [0, 0], MINER_BRIBE
        ).build_transaction({
            'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA),
            'gas': GAS_LIMIT, 'gasPrice': int(w3.eth.gas_price * 1.5)
        })
        
        hash_buy = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_buy, PRIV_KEY).raw_transaction)
        w3.eth.wait_for_transaction_receipt(hash_buy, timeout=60)
        notify(f"🛒 *COMPRA EXITOSA!* Esperando {TIEMPO_ESPERA_VENTA} segundos para asegurar ganancia...")
        
        # --- FASE 2: ESPERA ESTRATÉGICA ---
        time.sleep(TIEMPO_ESPERA_VENTA)
        
        # --- FASE 3: AUTO-VENTA ATÓMICA ---
        notify("💥 *INICIANDO AUTO-VENTA (DUMP)...*")
        balance_meme = target_contract.functions.balanceOf(CONTRATO_ADDR).call()
        
        if balance_meme > 0:
            p_approve_sell = target_contract.encodeABI(fn_name="approve", args=[PANCAKE_ROUTER, balance_meme])
            p_swap_sell = router_contract.encodeABI(fn_name="swapExactTokensForTokensSupportingFeeOnTransferTokens", args=[balance_meme, 0, [target_token, WBNB_ADDR], CONTRATO_ADDR, int(time.time()) + 120])
            
            tx_sell = apex_contract.functions.apexStrike(
                [target_token, PANCAKE_ROUTER],
                [w3.to_bytes(hexstr=p_approve_sell), w3.to_bytes(hexstr=p_swap_sell)],
                [0, 0], MINER_BRIBE
            ).build_transaction({
                'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA),
                'gas': GAS_LIMIT, 'gasPrice': int(w3.eth.gas_price * 1.5)
            })
            
            hash_sell = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_sell, PRIV_KEY).raw_transaction)
            notify(f"💰 *¡VENTA COMPLETADA!* \nWBNB devueltos al contrato.\nHash: {w3.to_hex(hash_sell)}")
            
            time.sleep(3)
            nuevo_balance = w3.from_wei(wbnb_contract.functions.balanceOf(CONTRATO_ADDR).call(), 'ether')
            notify(f"🏦 *Nuevo Capital:* {nuevo_balance:.6f} WBNB")
        else:
            notify("⚠️ Error: El contrato no recibió los tokens de la compra.")

    except Exception as e:
        notify(f"❌ *ERROR EN EL HIT & RUN:* {str(e)[:70]}")

def scan_all(last_block):
    now_block = w3.eth.block_number
    if now_block > last_block:
        try:
            # FIX: get_logs es 100% seguro en nodos públicos
            event_filter = w3.eth.contract(address=PANCAKE_FACTORY, abi=ABI_FACTORY).events.PairCreated.get_logs(fromBlock=last_block+1, toBlock=now_block)
            for ev in event_filter:
                t = ev.args.token1 if ev.args.token0 == WBNB_ADDR else (ev.args.token0 if ev.args.token1 == WBNB_ADDR else None)
                if t:
                    print(f"💎 Posible Graduación detectada: {t}")
                    if not es_honeypot(t):
                        execute_hit_and_run(t)
                    else:
                        print("🛡️ Estafa evadida.")
        except: pass
    return now_block

print("🚀 Motor TrenchBot encendido...")
try:
    last_block = w3.eth.block_number
    notify("💰 *TRENCHBOT V2.1 ONLINE*\nConectado a ApexTrenchBot. Listo para disparar.", buttons=True)

    while True:
        check_commands()
        last_block = scan_all(last_block)
        time.sleep(1)

except Exception as e:
    print(f"❌ CRASH FATAL: {e}")
