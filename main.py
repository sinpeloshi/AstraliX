import os
import time
import requests
from web3 import Web3

print("⏳ Iniciando motor AstraliX V10 (Titanium Edition - Alto Volumen)...")

# --- 🛰️ CONEXIÓN ---
RPC_URL = "https://bsc-dataseed.binance.org/"
w3 = Web3(Web3.HTTPProvider(RPC_URL))

if w3.is_connected():
    print("✅ Conectado a la BSC")

# --- ⚙️ CONFIGURACIÓN DE ATAQUE ---
CAPITAL_WBNB = 0.061              
CAPITAL_SNIPER = 0.02             
PROFIT_MIN_USD = 0.005            # Gatillo ultra-sensible (medio centavo) para forzar acción
GAS_LIMIT = 350000                
FILTRO_RADAR = -0.15              # Verás más movimiento en el radar

# --- 🔑 IDENTIDAD ---
CONTRATO_ADDR = w3.to_checksum_address("0x2093cd0b3F75A1E6ff750E1F871C234C1abF3d3c")
MI_BILLETERA = w3.to_checksum_address(os.environ.get('MI_BILLETERA', '0xTuBilleteraAqui').strip())
PRIV_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"

TG_TOKEN = os.environ.get('TELEGRAM_TOKEN', '').strip()
TG_ID = os.environ.get('TELEGRAM_CHAT_ID', '').strip()

WBNB_ADDR = w3.to_checksum_address("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")
USDT_ADDR = w3.to_checksum_address("0x55d398326f99059fF775485246999027B3197955")
PANCAKE_FACTORY = "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73"

# --- 📜 ABIs ---
ABI_ROUTER = '[{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"}],"name":"getAmountsOut","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"view","type":"function"}]'
ABI_ASTRALIX = '[{"inputs":[{"internalType":"address","name":"routerCompra","type":"address"},{"internalType":"address","name":"routerVenta","type":"address"},{"internalType":"address","name":"tokenBase","type":"address"},{"internalType":"address","name":"tokenArbitraje","type":"address"},{"internalType":"uint256","name":"montoInversion","type":"uint256"}],"name":"ejecutarArbitraje","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"}],"name":"retirarTokens","outputs":[],"stateMutability":"nonpayable","type":"function"}]'
ABI_ERC20 = '[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"}]'
ABI_FACTORY = '[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"token0","type":"address"},{"indexed":true,"internalType":"address","name":"token1","type":"address"},{"indexed":false,"internalType":"address","name":"pair","type":"address"},{"indexed":false,"internalType":"uint256","name":"","type":"uint256"}],"name":"PairCreated","type":"event"}]'

# --- 📡 ESCUADRÓN MIXTO (Tus 6 + 4 Reyes del Volumen) ---
TOKENS_MEME = {
    # Los 4 Reyes para asegurar liquidez y acción:
    "PEPE":    "0x25d887ce7335150ad2744d010c836a31940e7010",
    "FLOKI":   "0xfb5b838b6cfeedc2873ab27866079ac55363d37e",
    "BABY":    "0xc748673057861a797275CD8A068AbB95A902e8de",
    "CAT":     "0x6894cde390a3f51155ea41ed24a33a4827d3063d",
    # Tus 6 tokens exclusivos:
    "MILADY":  "0xc20E45E49e0E79f0fC81E71F05fD2772d6587777",
    "ION":     "0xE1ab61f7b093435204dF32F5b3A405de55445Ea8",
    "LOBSTER": "0xeCCBb861c0dda7eFd964010085488B69317e4444",
    "FOM":     "0x3e17ee3B1895dD1A7CF993A89769C5e029584444",
    "TKN5":    "0x9570Ff1d7eC2992Fb7728d632738A229A69a7400",
    "ARK":     "0xCae117ca6Bc8A341D2E7207F30E180f0e5618B9D"
}

DEXs = {
    "Pancake": "0x10ED43C718714eb63d5aA57B78B54704E256024E", 
    "Biswap":  "0x3a6d8cA21D1CF76F653A67577FA0D27453350dD8",
    "ApeSwap": "0xcF0feBd3f17CEf5b47b0cD257aCf6025c5BFf3b7"
}

# --- 📊 ESTADÍSTICAS GLOBALES ---
ultimos_datos = []
last_update_id = 0
total_disparos = 0
disparos_exitosos = 0
profit_total_wbnb = 0.0

# --- 📱 MÓDULO TELEGRAM ---
def notify(msg, buttons=False):
    if not TG_TOKEN: return
    url = f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage"
    data = {"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}
    if buttons:
        data["reply_markup"] = {"keyboard": [[{"text": "/status"}, {"text": "/radar"}], [{"text": "/balance"}, {"text": "/retiro"}]], "resize_keyboard": True}
    try: requests.post(url, json=data, timeout=5)
    except: pass

def check_commands():
    global last_update_id, total_disparos, disparos_exitosos, profit_total_wbnb
    if not TG_TOKEN: return
    url = f"https://api.telegram.org/bot{TG_TOKEN}/getUpdates?offset={last_update_id + 1}"
    try:
        r = requests.get(url, timeout=5).json()
        for update in r.get("result", []):
            last_update_id = update["update_id"]
            m = update.get("message", {})
            txt = m.get("text", "")
            if str(m.get("from", {}).get("id", "")) != TG_ID: continue

            if txt == "/status":
                msg = (f"🛰️ *AstraliX:* V10 Titanium Activo\n"
                       f"⚡ Modo: 0.061 WBNB (10 Objetivos)\n"
                       f"🎯 Intentos: {total_disparos}\n"
                       f"✅ Exitosos: {disparos_exitosos}\n"
                       f"💰 *Profit:* +{profit_total_wbnb:.6f} WBNB")
                notify(msg, buttons=True)
            elif txt == "/balance":
                try:
                    c = w3.eth.contract(address=WBNB_ADDR, abi=ABI_ERC20)
                    b = w3.from_wei(c.functions.balanceOf(CONTRATO_ADDR).call(), 'ether')
                    notify(f"🏦 *Capital:* {b:.6f} WBNB", buttons=True)
                except Exception as e:
                    notify(f"❌ Error leyendo saldo: {e}")
            elif txt == "/radar":
                msg = "📡 *Mejores Rutas Actuales*\n" + "═"*20 + "\n"
                for item in sorted(ultimos_datos, key=lambda x: x[1], reverse=True)[:5]:
                    msg += f"🔸 *{item[0]}:* ${item[1]:.3f}\n"
                notify(msg if ultimos_datos else "Acelerando motores... (intentá en 5 segs)", buttons=True)
            elif txt == "/retiro":
                ejecutar_retiro()
    except: pass

def ejecutar_retiro():
    notify("💰 *Ejecutando retiro...*")
    try:
        c = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_ASTRALIX)
        tx = c.functions.retirarTokens(WBNB_ADDR).build_transaction({
            'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA),
            'gas': 120000, 'gasPrice': w3.eth.gas_price
        })
        s = w3.eth.account.sign_transaction(tx, PRIV_KEY)
        h = w3.eth.send_raw_transaction(s.raw_transaction)
        notify(f"✅ *Retiro Exitoso!*\nHash: {w3.to_hex(h)}")
    except Exception as e: notify(f"❌ *Fallo al retirar:* {str(e)[:100]}")

# --- 🔍 MÓDULO DE SEGURIDAD ---
def es_honeypot(token_addr):
    try:
        r = requests.get(f"https://api.honeypot.is/v2/?address={token_addr}", timeout=3).json()
        return r.get("honeypotResult", {}).get("isHoneypot", True)
    except: return True

# --- 🎯 GATILLO DE EJECUCIÓN ---
def execute_trade(r1, r2, t_addr, t_name, n1, n2, capital, mode="ARBI"):
    global total_disparos, disparos_exitosos, profit_total_wbnb
    total_disparos += 1
    notify(f"🎯 *STRIKE {mode}:* {t_name}\n🔄 {n1} ➡️ {n2}")
    
    try:
        c_wbnb = w3.eth.contract(address=WBNB_ADDR, abi=ABI_ERC20)
        s_antes = c_wbnb.functions.balanceOf(CONTRATO_ADDR).call()

        contrato = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_ASTRALIX)
        tx = contrato.functions.ejecutarArbitraje(r1, r2, WBNB_ADDR, t_addr, w3.to_wei(capital, 'ether')).build_transaction({
            'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA),
            'gas': GAS_LIMIT, 'gasPrice': int(w3.eth.gas_price * 1.05) 
        })
        s_tx = w3.eth.account.sign_transaction(tx, PRIV_KEY)
        h = w3.eth.send_raw_transaction(s_tx.raw_transaction)
        
        rec = w3.eth.wait_for_transaction_receipt(h, timeout=60)
        if rec.status == 1:
            s_despues = c_wbnb.functions.balanceOf(CONTRATO_ADDR).call()
            ganancia = float(w3.from_wei(s_despues - s_antes, 'ether'))
            profit_total_wbnb += ganancia
            disparos_exitosos += 1
            notify(f"✅ *PROFIT:* +{ganancia:.5f} WBNB\nHash: {w3.to_hex(h)}")
        else:
            notify(f"🛡️ *REVERTIDO:* El escudo evitó pérdidas.")
    except Exception as e:
        notify(f"❌ *ERROR:* {str(e)[:50]}")

# --- 🏗️ MOTORES DE BÚSQUEDA ---
def scan_all(last_block):
    global ultimos_datos
    now_block = w3.eth.block_number
    
    # 1. SNIPER DE LANZAMIENTOS
    if now_block > last_block:
        try:
            events = w3.eth.contract(address=PANCAKE_FACTORY, abi=ABI_FACTORY).events.PairCreated.create_filter(fromBlock=last_block+1, toBlock=now_block).get_all_entries()
            for ev in events:
                t0, t1 = ev.args.token0, ev.args.token1
                target = t1 if t0 == WBNB_ADDR else (t0 if t1 == WBNB_ADDR else None)
                if target:
                    if not es_honeypot(target):
                        execute_trade(DEXs["Pancake"], DEXs["Pancake"], target, "GEMA_NUEVA", "Pancake", "Sniper", CAPITAL_SNIPER, "SNIPER")
        except: pass
    
    # 2. ARBITRAJE SEGURO
    temp_radar = []
    try:
        p_bnb = float(w3.from_wei(w3.eth.contract(address=w3.to_checksum_address(DEXs["Pancake"]), abi=ABI_ROUTER).functions.getAmountsOut(w3.to_wei(1, 'ether'), [WBNB_ADDR, USDT_ADDR]).call()[-1], 'ether'))
        gas_usd = float(w3.from_wei(w3.eth.gas_price * GAS_LIMIT, 'ether')) * p_bnb

        for name, raw_addr in TOKENS_MEME.items():
            addr = w3.to_checksum_address(raw_addr)
            for n1, a1 in DEXs.items():
                for n2, a2 in DEXs.items():
                    if n1 == n2: continue
                    try:
                        p1 = w3.from_wei(w3.eth.contract(address=a1, abi=ABI_ROUTER).functions.getAmountsOut(w3.to_wei(CAPITAL_WBNB, 'ether'), [WBNB_ADDR, addr]).call()[-1], 'ether')
                        p2 = w3.from_wei(w3.eth.contract(address=a2, abi=ABI_ROUTER).functions.getAmountsOut(w3.to_wei(p1, 'ether'), [addr, WBNB_ADDR]).call()[-1], 'ether')
                        neto = (float(p2) - CAPITAL_WBNB) * p_bnb - gas_usd
                        
                        ruta = f"{n1}->{n2}"
                        temp_radar.append((f"{name} ({ruta})", neto))
                        
                        if neto > PROFIT_MIN_USD: 
                            execute_trade(a1, a2, addr, name, n1, n2, CAPITAL_WBNB, "ARBI")
                            return now_block
                    except: continue # Ignora las rutas sin liquidez sin romper el ciclo
        ultimos_datos = temp_radar
    except: pass
    
    return now_block

# --- 🔥 INICIO ---
print("🚀 Motor encendido...")
try:
    last_block = w3.eth.block_number
    notify("💰 *ASTRALIX V10 ONLINE*\nMunición Mixta (10 Objetivos) cargada.", buttons=True)

    while True:
        check_commands()
        last_block = scan_all(last_block)
        time.sleep(1)

except Exception as e:
    print(f"❌ CRASH FATAL: {e}")