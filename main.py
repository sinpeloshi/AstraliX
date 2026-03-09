import os
import time
import requests
from web3 import Web3

# --- 🛰️ CONEXIÓN A BSC ---
RPC_URL = "https://bsc-dataseed.binance.org/"
w3 = Web3(Web3.HTTPProvider(RPC_URL))

def get_env(label):
    return os.environ.get(label, "").strip()

# --- ⚙️ CONFIGURACIÓN SNIPER (MODO AGRESIVO) ---
CAPITAL_WBNB = 0.039588494902596519 
PROFIT_MIN_USD = 0.01      # Dispara a partir de 1 centavo limpio
GAS_LIMIT = 400000         # Gas ajustado a la realidad de BSC (Sin miedo)
FILTRO_RADAR = -0.15       # Radar más sensible

CONTRATO_ADDR = w3.to_checksum_address(get_env('DIRECCION_CONTRATO'))
MI_BILLETERA = w3.to_checksum_address(get_env('MI_BILLETERA'))
PRIV_KEY = get_env('PRIVATE_KEY')
TG_TOKEN = get_env('TELEGRAM_TOKEN')
TG_ID = get_env('TELEGRAM_CHAT_ID')

WBNB_ADDR = w3.to_checksum_address("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")
USDT_ADDR = w3.to_checksum_address("0x55d398326f99059fF775485246999027B3197955")

# --- 📜 ABIs ---
ABI_ROUTER = '[{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"}],"name":"getAmountsOut","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"view","type":"function"}]'
ABI_ASTRALIX = '[{"inputs":[{"internalType":"address","name":"routerCompra","type":"address"},{"internalType":"address","name":"routerVenta","type":"address"},{"internalType":"address","name":"tokenBase","type":"address"},{"internalType":"address","name":"tokenArbitraje","type":"address"},{"internalType":"uint256","name":"montoInversion","type":"uint256"}],"name":"ejecutarArbitraje","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"}],"name":"retirarTokens","outputs":[],"stateMutability":"nonpayable","type":"function"}]'
ABI_ERC20 = '[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"}]'

# --- 📡 ESCUADRÓN DE 40 OBJETIVOS ---
TOKENS = {
    "CAT":   "0x6894cde390a3f51155ea41ed24a33a4827d3063d", "BDOG":  "0x1C45366641014069114c78962bDc371F534Bc81c",
    "BUILD": "0x6bdcce4a559076e37755a78ce0c06214e59e4444", "LUM":   "0x4dE1486E27237F170Cd92fF1Efb17eF4c2C74444",
    "4MEME": "0x0a43fc31a73013089df59194872ecae4cae14444", "BABY":  "0xc748673057861a797275CD8A068AbB95A902e8de",
    "FLOKI": "0xfb5b838b6cfeedc2873ab27866079ac55363d37e", "PEPE":  "0x25d887ce7335150ad2744d010c836a31940e7010",
    "BOB":   "0x51363f073b1e4920fda7aa9e9d84ba97ede1560e", "BOME":  "0x43a85b67a2f5f24f74d758f50c00000000000000",
    "SLERF": "0x6f6f... (BSC version)", "REKT":  "0x20482b0b4d9d8f60d3ab432b92f4c4b901a0d10c",
    "CATE":  "0x118f073796821da3e9901061b05c0b36377b877e", "PIT":   "0xa57ac35ce91ee92caefaa8dc04140c8e232c2e50",
    "KISHU": "0xa2b4c0af19cc16a6cfacce81f192b024d625817d", "POODL": "0x23396cf899ca06c4472205fc903bdb4de249d6fc",
    "SMI":   "0xcd7492db29e2ab436e819b249452ee1bbdf52214", "BOSS":  "0x04b2e8227f8b9c8d5b9c6d44f7d409775888b859",
    "FOX":   "0x00069f5a5fe1f7f9f4edb6f6d5c13b79c7f69a00", "FRGST": "0x5e0e7a3d9b9d7a509a68e2db279b9da869651d81",
    "USDT":  "0x55d398326f99059fF775485246999027B3197955", "ETH":   "0x2170Ed0880ac9A755fd29B2688956BD959F933F8",
    "BTCB":  "0x7130d2A12B9BCbFAe4f2634d864A1Ee1Ce3Ead9c", "SOL":   "0x570a5d26f7765ecb712c0924e4de545b89fd1460",
    "CAKE":  "0x0E09FaBB73Ade0a17ECC321fD13a19e81cE82062", "BSW":   "0x965F527D9159dCe6288a2219DB51fc6Eef120dD1",
    "TWT":   "0x4b0f1812e5df2a09796481ff14017e6005508003", "SAFEM": "0x42981d035AF57553470472a9194BA21FBBe42296",
    "DOGE":  "0xba2ae424d960c26247dd6c32edc70b295c744c43", "ADA":   "0x3EE2200Efb3400fAbB9AacF31297cBdD1d435D47",
    "XRP":   "0x1d2f0da169ceb247c7b4442785b00c6d3714b6d3", "DOT":   "0x7083609fce4d1d8dc0c979aab8c869ea2c873402",
    "LINK":  "0xf8a0bf9cf54bb92f17374d9e9a321e6a111a51bd", "UNI":   "0xbf513a9366e61f592750e82c5053066444444444",
    "MATIC": "0xcc42724c6683b7e57334c4e856f4c9965ed682bd", "SHIB":  "0x2859e4544c4bb03966803b044a72d188207f223c",
    "AIDOG": "0x309a... (AIDoge BSC)", "CRAZY": "0x11791e8a593bc1e1fbcB406DEaE5F8c71f85B60F",
    "BSHIB": "0xb8ebd245a2a117e4d2de08bb9d0a3b75ea8e784f", "KSHIB": "0xc34d8a7c4e8f8e3f5d2c9b0b7c8a3d5f7a9e8c1b"
}

DEXs = {"Pancake": "0x10ED43C718714eb63d5aA57B78B54704E256024E", "Biswap": "0x3a6d8cA21D1CF76F653A67577FA0D27453350dD8"}

ultimos_datos = []
last_update_id = 0

# --- 📱 COMUNICACIÓN TELEGRAM ---
def notify(msg, buttons=False):
    if not TG_TOKEN: return
    url = f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage"
    data = {"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}
    if buttons:
        data["reply_markup"] = {"keyboard": [[{"text": "/status"}, {"text": "/radar"}], [{"text": "/balance"}, {"text": "/retiro"}]], "resize_keyboard": True}
    try: requests.post(url, json=data, timeout=5)
    except: pass

def check_commands():
    global last_update_id
    url = f"https://api.telegram.org/bot{TG_TOKEN}/getUpdates?offset={last_update_id + 1}"
    try:
        r = requests.get(url, timeout=5).json()
        for update in r.get("result", []):
            last_update_id = update["update_id"]
            m = update.get("message", {})
            txt = m.get("text", "")
            if str(m.get("from", {}).get("id", "")) != TG_ID: continue

            if txt == "/status":
                notify("🛰️ *AstraliX:* Modo Sniper Activado.\n⚡ Listo para gatillar.", buttons=True)
            elif txt == "/balance":
                c = w3.eth.contract(address=WBNB_ADDR, abi=ABI_ERC20)
                b = w3.from_wei(c.functions.balanceOf(CONTRATO_ADDR).call(), 'ether')
                notify(f"🏦 *Capital:* {b:.6f} WBNB", buttons=True)
            elif txt == "/radar":
                msg = "📡 *Radar de Oportunidades*\n" + "═"*20 + "\n"
                for item in sorted(ultimos_datos, key=lambda x: x[1], reverse=True)[:5]:
                    msg += f"🔸 *{item[0]}:* ${item[1]:.3f}\n"
                notify(msg if ultimos_datos else "Buscando sangre...", buttons=True)
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
        notify(f"✅ *Éxito!*\nHash: {w3.to_hex(h)}")
    except Exception as e: notify(f"❌ *Fallo:* {str(e)[:100]}")

# --- 🛠️ MOTOR DE DISPARO ---
def get_price(router, amount, path):
    c = w3.eth.contract(address=w3.to_checksum_address(router), abi=ABI_ROUTER)
    try: return w3.from_wei(c.functions.getAmountsOut(w3.to_wei(amount, 'ether'), path).call()[-1], 'ether')
    except: return 0

def execute_strike(r1, r2, t_addr, t_name, profit):
    notify(f"🎯 *DISPARANDO:* {t_name} | Profit Estimado: ${profit:.2f}")
    try:
        c = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_ASTRALIX)
        tx = c.functions.ejecutarArbitraje(r1, r2, WBNB_ADDR, t_addr, w3.to_wei(CAPITAL_WBNB, 'ether')).build_transaction({
            'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA),
            'gas': GAS_LIMIT, 'gasPrice': w3.eth.gas_price
        })
        s = w3.eth.account.sign_transaction(tx, PRIV_KEY)
        h = w3.eth.send_raw_transaction(s.raw_transaction)
        notify(f"🚀 *GATILLADO:* {w3.to_hex(h)}")
    except Exception as e: notify(f"🛡️ *Fallo de Red:* {e}")

# --- INICIO ---
timer_10m = time.time()
notify("🚀 *AstraliX SNIPER Online* - A ganar, socio.", buttons=True)

while True:
    check_commands()
    if time.time() - timer_10m > 600:
        try:
            p_bnb = float(get_price(DEXs["Pancake"], 1, [WBNB_ADDR, USDT_ADDR]))
            notify(f"⏱️ *Reporte:* Todo OK.\nBNB: ${p_bnb:.2f}")
        except: pass
        timer_10m = time.time()

    temp_radar = []
    try:
        p_bnb = float(get_price(DEXs["Pancake"], 1, [WBNB_ADDR, USDT_ADDR]))
        
        # EL PRECIO DEL GAS SE LEE AQUÍ EN TIEMPO REAL DESDE LA RED BSC
        gas_usd = float(w3.from_wei(w3.eth.gas_price * GAS_LIMIT, 'ether')) * p_bnb

        for name, addr in TOKENS.items():
            for n1, a1 in DEXs.items():
                for n2, a2 in DEXs.items():
                    if n1 == n2: continue
                    p1 = get_price(a1, CAPITAL_WBNB, [WBNB_ADDR, addr])
                    if p1 == 0: continue
                    p2 = get_price(a2, p1, [addr, WBNB_ADDR])
                    
                    # CÁLCULO NETO EXACTO CON EL GAS DE ESE MILISEGUNDO
                    neto = (float(p2) - CAPITAL_WBNB) * p_bnb - gas_usd
                    
                    if neto > FILTRO_RADAR: 
                        temp_radar.append((name, neto))
                        print(f"📡 Objetivo: {name:<5} | Profit Neto: ${neto:.3f}")
                    
                    if neto > PROFIT_MIN_USD:
                        execute_strike(a1, a2, addr, name, neto)
                        time.sleep(30) # Pausa de recarga después de disparar
        ultimos_datos = temp_radar
    except: time.sleep(5)