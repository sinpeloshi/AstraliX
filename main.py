import os
import time
import requests
from web3 import Web3

# --- 🛰️ CONEXIÓN Y CONFIGURACIÓN ---
RPC_URL = "https://bsc-dataseed.binance.org/"
w3 = Web3(Web3.HTTPProvider(RPC_URL))

# --- ⚙️ CONFIGURACIÓN DE COMBATE (AJUSTADO POR SOCIO) ---
CAPITAL_WBNB = 0.039588494902596519 
CAPITAL_SNIPER = 0.015             # Inversión por cada gema nueva
PROFIT_MIN_USD = 0.01              # Gatillo fácil
GAS_LIMIT = 450000                 
FILTRO_RADAR = -0.15               

# Identidad (Clave directa para máxima velocidad en Repo Privado)
CONTRATO_ADDR = w3.to_checksum_address(os.environ.get('DIRECCION_CONTRATO', '').strip())
MI_BILLETERA = w3.to_checksum_address(os.environ.get('MI_BILLETERA', '').strip())
PRIV_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"

TG_TOKEN = os.environ.get('TELEGRAM_TOKEN', '').strip()
TG_ID = os.environ.get('TELEGRAM_CHAT_ID', '').strip()

# Direcciones Base
WBNB_ADDR = w3.to_checksum_address("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")
USDT_ADDR = w3.to_checksum_address("0x55d398326f99059fF775485246999027B3197955")
PANCAKE_FACTORY = "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73"

# --- 📜 ABIs (Mantenemos retirarTokens con 's') ---
ABI_ROUTER = '[{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"}],"name":"getAmountsOut","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"view","type":"function"}]'
ABI_ASTRALIX = '[{"inputs":[{"internalType":"address","name":"routerCompra","type":"address"},{"internalType":"address","name":"routerVenta","type":"address"},{"internalType":"address","name":"tokenBase","type":"address"},{"internalType":"address","name":"tokenArbitraje","type":"address"},{"internalType":"uint256","name":"montoInversion","type":"uint256"}],"name":"ejecutarArbitraje","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"}],"name":"retirarTokens","outputs":[],"stateMutability":"nonpayable","type":"function"}]'
ABI_ERC20 = '[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"}]'
ABI_FACTORY = '[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"token0","type":"address"},{"indexed":true,"internalType":"address","name":"token1","type":"address"},{"indexed":false,"internalType":"address","name":"pair","type":"address"},{"indexed":false,"internalType":"uint256","name":"","type":"uint256"}],"name":"PairCreated","type":"event"}]'

# --- 📊 ESTADÍSTICAS ---
total_snipe = 0
total_arbi = 0
profit_acumulado = 0.0

# --- 📡 DICCIONARIOS ---
DEXs = {
    "Pancake":  "0x10ED43C718714eb63d5aA57B78B54704E256024E", 
    "Biswap":   "0x3a6d8cA21D1CF76F653A67577FA0D27453350dD8",
    "ApeSwap":  "0xcF0feBd3f17CEf5b47b0cD257aCf6025c5BFf3b7",
    "BabyDoge": "0xc82819F72A9e77E2c0c3A69B3196478f44303cf4"
}

TOKENS_MEME = {
    "CAT": "0x6894cde390a3f51155ea41ed24a33a4827d3063d", "BDOG": "0x1C45366641014069114c78962bDc371F534Bc81c",
    "BUILD": "0x6bdcce4a559076e37755a78ce0c06214e59e4444", "LUM": "0x4dE1486E27237F170Cd92fF1Efb17eF4c2C74444",
    "4MEME": "0x0a43fc31a73013089df59194872ecae4cae14444", "BABY": "0xc748673057861a797275CD8A068AbB95A902e8de",
    "FLOKI": "0xfb5b838b6cfeedc2873ab27866079ac55363d37e", "PEPE": "0x25d887ce7335150ad2744d010c836a31940e7010"
}

# --- 📱 COMUNICACIÓN ---
def notify(msg):
    url = f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage"
    try: requests.post(url, json={"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}, timeout=5)
    except: pass

# --- 🔍 FILTRO DE SEGURIDAD ---
def es_honeypot(token_addr):
    try:
        r = requests.get(f"https://api.honeypot.is/v2/?address={token_addr}", timeout=3).json()
        return r.get("honeypotResult", {}).get("isHoneypot", True)
    except: return True # Ante duda, bloqueamos

# --- 🎯 GATILLO DE EJECUCIÓN ---
def execute_trade(r1, r2, t_addr, t_name, n1, n2, capital, mode="ARBI"):
    global total_arbi, total_snipe, profit_acumulado
    notify(f"🎯 *STRIKE {mode}:* {t_name}\n🔄 {n1} ➡️ {n2}")
    try:
        c_wbnb = w3.eth.contract(address=WBNB_ADDR, abi=ABI_ERC20)
        s_antes = c_wbnb.functions.balanceOf(CONTRATO_ADDR).call()

        contrato = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_ASTRALIX)
        tx = contrato.functions.ejecutarArbitraje(r1, r2, WBNB_ADDR, t_addr, w3.to_wei(capital, 'ether')).build_transaction({
            'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA),
            'gas': GAS_LIMIT, 'gasPrice': int(w3.eth.gas_price * 1.1) # Prioridad
        })
        s_tx = w3.eth.account.sign_transaction(tx, PRIV_KEY)
        h = w3.eth.send_raw_transaction(s_tx.raw_transaction)
        
        rec = w3.eth.wait_for_transaction_receipt(h, timeout=60)
        if rec.status == 1:
            s_despues = c_wbnb.functions.balanceOf(CONTRATO_ADDR).call()
            ganancia = float(w3.from_wei(s_despues - s_antes, 'ether'))
            if mode == "ARBI": total_arbi += 1
            else: total_snipe += 1
            profit_acumulado += ganancia
            notify(f"✅ *ÉXITO:* +{ganancia:.5f} WBNB\nHash: {w3.to_hex(h)}")
        else: notify(f"🛡️ *REVERTIDO:* El contrato evitó pérdidas.")
    except Exception as e: notify(f"❌ *ERROR:* {str(e)[:50]}")

# --- 🛡️ MÓDULO SNIPER ---
def scan_new_tokens(last_block):
    try:
        factory = w3.eth.contract(address=PANCAKE_FACTORY, abi=ABI_FACTORY)
        now_block = w3.eth.block_number
        if last_block >= now_block: return last_block

        events = factory.events.PairCreated.create_filter(fromBlock=last_block+1, toBlock=now_block).get_all_entries()
        for ev in events:
            t0, t1 = ev.args.token0, ev.args.token1
            target = t1 if t0 == WBNB_ADDR else (t0 if t1 == WBNB_ADDR else None)
            
            if target:
                print(f"💎 Nuevo Token Detectado: {target}")
                if not es_honeypot(target):
                    execute_trade(DEXs["Pancake"], DEXs["Pancake"], target, "GEMA_NUEVA", "Pancake", "Sniper", CAPITAL_SNIPER, "SNIPER")
                else: print(f"🛡️ Honeypot bloqueado: {target}")
        return now_block
    except: return last_block

# --- 🏗️ MÓDULO ARBITRAJE OMNI ---
def scan_arbitrage():
    try:
        p_bnb = float(w3.from_wei(w3.eth.contract(address=w3.to_checksum_address(DEXs["Pancake"]), abi=ABI_ROUTER).functions.getAmountsOut(w3.to_wei(1, 'ether'), [WBNB_ADDR, USDT_ADDR]).call()[-1], 'ether'))
        gas_usd = float(w3.from_wei(w3.eth.gas_price * GAS_LIMIT, 'ether')) * p_bnb

        for name, addr in TOKENS_MEME.items():
            for n1, a1 in DEXs.items():
                for n2, a2 in DEXs.items():
                    if n1 == n2: continue
                    router1 = w3.eth.contract(address=w3.to_checksum_address(a1), abi=ABI_ROUTER)
                    router2 = w3.eth.contract(address=w3.to_checksum_address(a2), abi=ABI_ROUTER)
                    
                    try:
                        p1 = w3.from_wei(router1.functions.getAmountsOut(w3.to_wei(CAPITAL_WBNB, 'ether'), [WBNB_ADDR, addr]).call()[-1], 'ether')
                        p2 = w3.from_wei(router2.functions.getAmountsOut(w3.to_wei(p1, 'ether'), [addr, WBNB_ADDR]).call()[-1], 'ether')
                        neto = (float(p2) - CAPITAL_WBNB) * p_bnb - gas_usd
                        
                        if neto > FILTRO_RADAR:
                            print(f"📡 Radar: {n1}->{n2} | {name} | Profit: ${neto:.3f}")
                        if neto > PROFIT_MIN_USD:
                            execute_trade(a1, a2, addr, name, n1, n2, CAPITAL_WBNB, "ARBI")
                            return # Salir para refrescar bloque
                    except: continue
    except: pass

# --- 🔥 INICIO ---
last_block = w3.eth.block_number
notify("🛡️ *ASTRALIX GOD-MODE V5 ONLINE*\nSniper & Arbi listos.")

while True:
    prev_block = last_block
    last_block = scan_new_tokens(last_block)
    
    # Heartbeat: Si el bloque avanzó, imprimimos en consola
    if last_block > prev_block:
        print(f"✅ Bloque {last_block} patrullado. Escaneando rutas...")
    
    scan_arbitrage()
    time.sleep(1)