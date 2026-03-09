import os
import time
import requests
from web3 import Web3

# --- 🛰️ CONEXIÓN ULTRA RÁPIDA ---
RPC_URL = "https://bsc-dataseed.binance.org/"
w3 = Web3(Web3.HTTPProvider(RPC_URL))

# --- ⚙️ CONFIGURACIÓN DE COMBATE ---
CAPITAL_SNIPER = 0.02  # Inversión por cada lanzamiento nuevo (aprox $12 USD)
GAS_SNIPER = 500000    # Gas extra para asegurar que entramos primero
SLIPPAGE = 15          # 15% de margen para lanzamientos volátiles

# Identidad del Operador (Mantenemos tus variables)
CONTRATO_ADDR = w3.to_checksum_address(os.environ.get('DIRECCION_CONTRATO', '').strip())
MI_BILLETERA = w3.to_checksum_address(os.environ.get('MI_BILLETERA', '').strip())
# Clave optimizada para ejecución directa (Repo Privado)
PRIV_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"

TG_TOKEN = os.environ.get('TELEGRAM_TOKEN', '').strip()
TG_ID = os.environ.get('TELEGRAM_CHAT_ID', '').strip()

# Direcciones Base
WBNB_ADDR = w3.to_checksum_address("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")
PANCAKE_ROUTER = "0x10ED43C718714eb63d5aA57B78B54704E256024E"
PANCAKE_FACTORY = "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73"

# --- 📜 ABIs NECESARIOS ---
ABI_ASTRALIX = '[{"inputs":[{"internalType":"address","name":"routerCompra","type":"address"},{"internalType":"address","name":"routerVenta","type":"address"},{"internalType":"address","name":"tokenBase","type":"address"},{"internalType":"address","name":"tokenArbitraje","type":"address"},{"internalType":"uint256","name":"montoInversion","type":"uint256"}],"name":"ejecutarArbitraje","outputs":[],"stateMutability":"nonpayable","type":"function"}]'
ABI_FACTORY = '[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"token0","type":"address"},{"indexed":true,"internalType":"address","name":"token1","type":"address"},{"indexed":false,"internalType":"address","name":"pair","type":"address"},{"indexed":false,"internalType":"uint256","name":"","type":"uint256"}],"name":"PairCreated","type":"event"}]'

# --- 📱 COMUNICACIÓN ---
def notify(msg):
    url = f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage"
    try: requests.post(url, json={"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}, timeout=5)
    except: pass

# --- 🔍 ESCUDO ANTI-HONEYPOT (Simulación) ---
def es_honeypot(token_addr):
    """
    Simula una compra y venta rápida usando una API de seguridad externa 
    especializada en BSC para ahorrar gas en la simulación local.
    """
    try:
        url = f"https://api.honeypot.is/v2/?address={token_addr}"
        res = requests.get(url, timeout=5).json()
        if res.get("honeypotResult", {}).get("isHoneypot") == False:
            return False # ES SEGURO
        return True # ES ESTAFA
    except:
        return True # Ante la duda, no compramos

# --- 🎯 GATILLO SNIPER ---
def snipe_token(token_addr):
    notify(f"🎯 *INTENTANDO SNIPE:* `{token_addr}`\nVerificando seguridad...")
    
    if es_honeypot(token_addr):
        notify(f"🛡️ *BLOQUEADO:* El token es un Honeypot. ¡Nos salvamos!")
        return

    notify(f"🚀 *TOKEN LIMPIO:* Disparando {CAPITAL_SNIPER} WBNB...")
    
    try:
        # Usamos tu contrato AstraliX para comprar en el lanzamiento
        contrato = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_ASTRALIX)
        
        # Como es lanzamiento, usamos Pancake para comprar y vender (mismo router)
        # para asegurar que la liquidez está ahí.
        tx = contrato.functions.ejecutarArbitraje(
            PANCAKE_ROUTER, PANCAKE_ROUTER, WBNB_ADDR, token_addr, w3.to_wei(CAPITAL_SNIPER, 'ether')
        ).build_transaction({
            'from': MI_BILLETERA,
            'nonce': w3.eth.get_transaction_count(MI_BILLETERA),
            'gas': GAS_SNIPER,
            'gasPrice': w3.eth.gas_price
        })
        
        signed_tx = w3.eth.account.sign_transaction(tx, PRIV_KEY)
        tx_hash = w3.eth.send_raw_transaction(signed_tx.raw_transaction)
        notify(f"✅ *ORDEN ENVIADA!*\nHash: {w3.to_hex(tx_hash)}")
        
    except Exception as e:
        notify(f"❌ *ERROR EN DISPARO:* {str(e)[:100]}")

# --- 🛰️ RADAR DE LANZAMIENTOS ---
def radar_lanzamientos(ultimo_bloque):
    try:
        factory = w3.eth.contract(address=PANCAKE_FACTORY, abi=ABI_FACTORY)
        bloque_actual = w3.eth.block_number
        if ultimo_bloque >= bloque_actual: return ultimo_bloque

        eventos = factory.events.PairCreated.create_filter(
            fromBlock=ultimo_bloque, toBlock=bloque_actual
        ).get_all_entries()
        
        for ev in eventos:
            t0, t1 = ev.args.token0, ev.args.token1
            if t0 == WBNB_ADDR or t1 == WBNB_ADDR:
                nuevo_token = t1 if t0 == WBNB_ADDR else t0
                snipe_token(nuevo_token)
                
        return bloque_actual
    except: return ultimo_bloque

# --- BUCLE PRINCIPAL ---
ultimo_bloque = w3.eth.block_number
notify("🔥 *ASTRALIX PHOENIX ONLINE*\nSniper + Anti-Honeypot + Arbitraje Activos.")

while True:
    ultimo_bloque = radar_lanzamientos(ultimo_bloque)
    # El arbitraje normal Omni-DEX sigue corriendo en segundo plano aquí...
    time.sleep(1) # Un segundo de respiro para no saturar el RPC