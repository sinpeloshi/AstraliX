import os
import time
import requests
from web3 import Web3

print("⏳ Iniciando motor TRENCHBOT V1 (Apex Executor)...")

# --- 🛰️ CONEXIÓN ---
RPC_URL = "https://bsc-dataseed.binance.org/"
w3 = Web3(Web3.HTTPProvider(RPC_URL))

if w3.is_connected():
    print("✅ Conectado a la BSC")

# --- ⚙️ CONFIGURACIÓN DE ATAQUE ---
CAPITAL_SNIPER = 0.015            # WBNB a usar por cada disparo
GAS_LIMIT = 500000                
MINER_BRIBE = 0                   # Propina al minero en Wei (0 por ahora)

# --- 🔑 IDENTIDAD ---
# Tu nueva máquina de guerra ApexTrenchBot
CONTRATO_ADDR = w3.to_checksum_address("0xF44f4D75Efc8d60d9383319D1C69553A1201bE28")
MI_BILLETERA = w3.to_checksum_address(os.environ.get('MI_BILLETERA', '0xTuBilleteraAqui').strip())
# Tu clave privada autorizada
PRIV_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"

TG_TOKEN = os.environ.get('TELEGRAM_TOKEN', '').strip()
TG_ID = os.environ.get('TELEGRAM_CHAT_ID', '').strip()

WBNB_ADDR = w3.to_checksum_address("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")
PANCAKE_FACTORY = w3.to_checksum_address("0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73")
PANCAKE_ROUTER = w3.to_checksum_address("0x10ED43C718714eb63d5aA57B78B54704E256024E")

# --- 📜 ABIs NECESARIOS ---
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
                notify("🛰️ *TrenchBot V1 Activo*\nEsperando graduaciones de GraFun/Four.meme a PancakeSwap.", buttons=True)
            elif txt == "/balance":
                try:
                    c = w3.eth.contract(address=WBNB_ADDR, abi=ABI_ERC20)
                    b = w3.from_wei(c.functions.balanceOf(CONTRATO_ADDR).call(), 'ether')
                    notify(f"🏦 *Munición en el Cañón:* {b:.6f} WBNB", buttons=True)
                except Exception as e:
                    notify(f"❌ Error leyendo saldo: {e}")
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
    except Exception as e: notify(f"❌ *Fallo al retirar:* {str(e)[:100]}")

# --- 🔍 MÓDULO DE SEGURIDAD ---
def es_honeypot(token_addr):
    try: return requests.get(f"https://api.honeypot.is/v2/?address={token_addr}", timeout=3).json().get("honeypotResult", {}).get("isHoneypot", True)
    except: return True

# --- 🎯 GATILLO ATÓMICO (APEX STRIKE) ---
def execute_atomic_snipe(target_token):
    notify(f"🚀 *OBJETIVO DETECTADO:* `{target_token}`\nArmando misil atómico...")
    
    try:
        # Preparamos las herramientas
        wbnb_contract = w3.eth.contract(address=WBNB_ADDR, abi=ABI_ERC20)
        router_contract = w3.eth.contract(address=PANCAKE_ROUTER, abi=ABI_ROUTER)
        apex_contract = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_APEX)
        
        monto_inversion = w3.to_wei(CAPITAL_SNIPER, 'ether')
        
        # 1. Empaquetamos el "Approve" (Permitimos que PancakeSwap gaste nuestro WBNB)
        payload_approve = wbnb_contract.encodeABI(fn_name="approve", args=[PANCAKE_ROUTER, monto_inversion])
        
        # 2. Empaquetamos el "Swap" (Compramos el token meme con WBNB)
        deadline = int(time.time()) + 120
        path = [WBNB_ADDR, target_token]
        # to=CONTRATO_ADDR porque el token lo recibe tu contrato de asalto
        payload_swap = router_contract.encodeABI(fn_name="swapExactTokensForTokensSupportingFeeOnTransferTokens", args=[monto_inversion, 0, path, CONTRATO_ADDR, deadline])
        
        # 3. Armamos la matriz del Multicall Atómico
        targets = [WBNB_ADDR, PANCAKE_ROUTER]
        payloads = [w3.to_bytes(hexstr=payload_approve), w3.to_bytes(hexstr=payload_swap)]
        values = [0, 0] # 0 porque estamos usando WBNB, no BNB nativo
        
        # Disparamos
        tx = apex_contract.functions.apexStrike(targets, payloads, values, MINER_BRIBE).build_transaction({
            'from': MI_BILLETERA,
            'nonce': w3.eth.get_transaction_count(MI_BILLETERA),
            'gas': GAS_LIMIT,
            'gasPrice': int(w3.eth.gas_price * 1.5) # Subimos un poco el gas para pasar al frente
        })
        
        s_tx = w3.eth.account.sign_transaction(tx, PRIV_KEY)
        h = w3.eth.send_raw_transaction(s_tx.raw_transaction)
        notify(f"🎯 *¡STRIKE ATÓMICO ENVIADO!*\nHash: {w3.to_hex(h)}\nEsperando confirmación...")
        
    except Exception as e:
        notify(f"❌ *ERROR EN EL DISPARO:* {str(e)[:50]}")

# --- 🏗️ MOTOR DE BÚSQUEDA ---
def scan_all(last_block):
    now_block = w3.eth.block_number
    if now_block > last_block:
        try:
            for ev in w3.eth.contract(address=PANCAKE_FACTORY, abi=ABI_FACTORY).events.PairCreated.create_filter(fromBlock=last_block+1, toBlock=now_block).get_all_entries():
                t = ev.args.token1 if ev.args.token0 == WBNB_ADDR else (ev.args.token0 if ev.args.token1 == WBNB_ADDR else None)
                if t:
                    print(f"💎 Posible Graduación: {t}")
                    if not es_honeypot(t):
                        execute_atomic_snipe(t)
                        return now_block # Pausamos el escaneo tras disparar
                    else:
                        print("🛡️ Estafa detectada y evadida.")
        except: pass
    return now_block

# --- 🔥 INICIO ---
print("🚀 Motor TrenchBot encendido...")
try:
    last_block = w3.eth.block_number
    notify("💰 *TRENCHBOT V1 ONLINE*\nConectado a ApexTrenchBot.", buttons=True)

    while True:
        check_commands()
        last_block = scan_all(last_block)
        time.sleep(1)

except Exception as e:
    print(f"❌ CRASH FATAL: {e}")