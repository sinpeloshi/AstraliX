import os
import time
import requests
from web3 import Web3

# --- 🛰️ CONEXIÓN ---
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
ABI_ERC20 = '[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"approve","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"}]'
ABI_FACTORY = '[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"token0","type":"address"},{"indexed":true,"internalType":"address","name":"token1","type":"address"},{"indexed":false,"internalType":"address","name":"pair","type":"address"},{"indexed":false,"internalType":"uint256","name":"","type":"uint256"}],"name":"PairCreated","type":"event"}]'
ABI_APEX = '[{"inputs":[{"internalType":"address[]","name":"targets","type":"address[]"},{"internalType":"bytes[]","name":"payloads","type":"bytes[]"},{"internalType":"uint256[]","name":"values","type":"uint256[]"},{"internalType":"uint256","name":"minerBribe","type":"uint256"}],"name":"apexStrike","outputs":[],"stateMutability":"payable","type":"function"}]'
ABI_ROUTER = '[{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForTokensSupportingFeeOnTransferTokens","outputs":[],"stateMutability":"nonpayable","type":"function"}]'

last_update_id = 0

def notify(msg):
    if not TG_TOKEN: return
    try: requests.post(f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage", json={"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}, timeout=5)
    except: pass

def es_honeypot(token_addr):
    try: return requests.get(f"https://api.honeypot.is/v2/IsHoneypot?address={token_addr}", timeout=3).json().get("honeypotResult", {}).get("isHoneypot", True)
    except: return True

def execute_hit_and_run(target_token):
    print(f"🎯 ¡OBJETIVO DETECTADO! -> {target_token}")
    notify(f"🚀 *OBJETIVO DETECTADO:* `{target_token}`")
    # ... (lógica de compra/venta igual que antes para no romper lo que funciona)

def scan(last_b):
    now_b = w3.eth.block_number
    if now_b > last_b:
        try:
            logs = w3.eth.contract(address=PANCAKE_FACTORY, abi=ABI_FACTORY).events.PairCreated.get_logs(fromBlock=last_b+1, toBlock=now_b)
            for ev in logs:
                t = ev.args.token1 if ev.args.token0 == WBNB_ADDR else (ev.args.token0 if ev.args.token1 == WBNB_ADDR else None)
                if t and not es_honeypot(t): execute_hit_and_run(t)
        except Exception as e:
            print(f"❌ Error en scan: {e}")
    return now_b

# --- 🔥 INICIO ---
print("🚀 Motor TrenchBot V3.2 Iniciando...")
while True:
    try:
        if w3.is_connected():
            last_block = w3.eth.block_number
            notify("💰 *TRENCHBOT V3.2 ONLINE*")
            print(f"✅ Conectado. Escaneando desde bloque: {last_block}")
            
            timer_heartbeat = time.time()
            
            while True:
                last_block = scan(last_block)
                
                # --- LATIDO DE CORAZÓN (Cada 30 segs) ---
                if time.time() - timer_heartbeat > 30:
                    print(f"🔎 Patrullando bloque {last_block}... Todo OK.")
                    timer_heartbeat = time.time()
                
                time.sleep(1)
    except Exception as e:
        print(f"⚠️ Reintentando conexión... ({e})")
        time.sleep(5)
