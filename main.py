import os
import time
import requests
from web3 import Web3

# --- 🛰️ CONEXIÓN AL NODO ---
RPC_URL = "https://bsc-dataseed.binance.org/"
w3 = Web3(Web3.HTTPProvider(RPC_URL))

# --- 🧹 LIMPIADOR DE VARIABLES ---
def get_env(label):
    val = os.environ.get(label, "").strip()
    return val

# --- ⚙️ CONFIGURACIÓN DE COMBATE ---
CAPITAL_WBNB = 0.039588494902596519 
PROFIT_MIN_USD = 0.02               
GAS_LIMIT = 600000 
FILTRO_RADAR = -0.30

# Direcciones Base (Checksum forzadas)
WBNB_ADDR = w3.to_checksum_address("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")
USDT_ADDR = w3.to_checksum_address("0x55d398326f99059fF775485246999027B3197955")

# Carga de Variables de Railway
try:
    CONTRATO_ADDR = w3.to_checksum_address(get_env('DIRECCION_CONTRATO'))
    MI_BILLETERA = w3.to_checksum_address(get_env('MI_BILLETERA'))
    PRIV_KEY = get_env('PRIVATE_KEY')
    TG_TOKEN = get_env('TELEGRAM_TOKEN')
    TG_ID = get_env('TELEGRAM_CHAT_ID')
except Exception as e:
    print(f"❌ Error cargando variables: {e}")

# --- 📱 FUNCIÓN NOTIFICACIÓN (MODO DEBUG) ---
def notify(msg):
    print(f"DEBUG: {msg}") # También lo imprime en los logs de Railway
    if TG_TOKEN and TG_ID:
        url = f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage?chat_id={TG_ID}&text={msg}"
        try:
            r = requests.get(url, timeout=5)
            if r.status_code != 200:
                print(f"❌ Error de Telegram: {r.text}")
        except Exception as e:
            print(f"❌ Fallo de conexión Telegram: {e}")

# 📡 LOS 45 OBJETIVOS (LISTA PULIDA)
TOKENS = {
    "USDT":  "0x55d398326f99059fF775485246999027B3197955", "ETH": "0x2170Ed0880ac9A755fd29B2688956BD959F933F8",
    "BABY":  "0xc748673057861a797275CD8A068AbB95A902e8de", "FLOKI": "0xfb5b838b6cfeedc2873ab27866079ac55363d37e",
    "KISHU": "0xa2b4c0af19cc16a6cfacce81f192b024d625817d", "POODL": "0x23396cf899ca06c4472205fc903bdb4de249d6fc",
    "WOOFY": "0xd0660cd418a64a1d44e9214ad8e459324d8157f1", "CAT":   "0x6894cde390a3f51155ea41ed24a33a4827d3063d",
    "SPX":   "0x4f1a9cfdc93fce7b55b7a5a1d9b9add4ef95c0d0", "GOAT":  "0x88e806c1d78c4f5f31f5c9c8a86d9c3f7a3c7f69",
    "SMI":   "0xcd7492db29e2ab436e819b249452ee1bbdf52214", "PIT":   "0xa57ac35ce91ee92caefaa8dc04140c8e232c2e50",
    "CATE":  "0x118f073796821da3e9901061b05c0b36377b877e", "BFLOKI":"0x02a9d7162bd73c2b35c5cf6cdd585e91928c850a",
    "BSHIB": "0xb8ebd245a2a117e4d2de08bb9d0a3b75ea8e784f", "CRAZY": "0x11791e8a593bc1e1fbcB406DEaE5F8c71f85B60F",
    "GDOGE": "0xc9d3924e65913e10ec78c7ca6f19ca3d4d965cc7", "NCAT":  "0x9f9e5fD8bbc25984B178FdCE6117Defa39d2db39",
    "FOX":   "0x00069f5a5fe1f7f9f4edb6f6d5c13b79c7f69a00", "BOSS":  "0x04b2e8227f8b9c8d5b9c6d44f7d409775888b859",
    "FRGST": "0x5e0e7a3d9b9d7a509a68e2db279b9da869651d81", "FLOKN": "0x6e0d9e89fbb0e5f0a89e3297d1e392b74538b0f6",
    "BBALI": "0x09c6230e43acdf4aa89ccab18c5cfd3f780459b7", "REKT":  "0x20482b0b4d9d8f60d3ab432b92f4c4b901a0d10c",
    "SIREN": "0x997a58129890bbda032231a52ed1ddc845fc18e1", "MERL":  "0xa0c56a8c0692bd10b3fa8f8ba79cf5332b7107f9",
    "BSQ":   "0x783c3f003f172c6ac5ac700218a357d2d66ee2a2", "CARD":  "0xdc06717f367e57a16e06cce0c4761604460da8fc",
    "HOME":  "0x4bfaa776991e85e5f8b1255461cbbd216cfc714f", "LBR":   "0x68ebc50a5bbd4a9e9ed6c6b3c2d4a0d2a1a4f3b2",
    "DGRA":  "0x4b86b9d09a9d7e5b45cdaf0c9a2b805d6e7ca3a4", "SHI":   "0x1f6f1a4c5b5c1b5a5e0b1f6c1c5a1f5c6b6a5f4e",
    "KSHIB": "0xc34d8a7c4e8f8e3f5d2c9b0b7c8a3d5f7a9e8c1b", "LFLOK": "0xf5bfa2f3c4d6e1a8b9c7d3f2e1c6b7a9d8f5e1c2",
    "PEPE":  "0x25d887ce7335150ad2744d010c836a31940e7010", "DOGE":  "0xba2ae424d960c26247dd6c32edc70b295c744c43",
    "CAKE":  "0x0E09FaBB73Ade0a17ECC321fD13a19e81cE82", "BSW":   "0x965F527D9159dCe6288a2219DB51fc6Eef120dD1",
    "TWT":   "0x4b0f1812e5df2a09796481ff14017e6005508003", "SOL":   "0x570a5d26f7765ecb712c0924e4de545b89fd1460"
}

DEXs = {"Pancake": "0x10ED43C718714eb63d5aA57B78B54704E256024E", "Biswap": "0x3a6d8cA21D1CF76F653A67577FA0D27453350dD8"}

ABI_ASTRALIX = '[{"inputs":[{"internalType":"address","name":"routerCompra","type":"address"},{"internalType":"address","name":"routerVenta","type":"address"},{"internalType":"address","name":"tokenBase","type":"address"},{"internalType":"address","name":"tokenArbitraje","type":"address"},{"internalType":"uint256","name":"montoInversion","type":"uint256"}],"name":"ejecutarArbitraje","outputs":[],"stateMutability":"nonpayable","type":"function"}]'
ABI_ROUTER = '[{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"}],"name":"getAmountsOut","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"view","type":"function"}]'

def get_price(router, amount_in, path):
    contract = w3.eth.contract(address=w3.to_checksum_address(router), abi=ABI_ROUTER)
    try:
        amounts = contract.functions.getAmountsOut(w3.to_wei(amount_in, 'ether'), path).call()
        return w3.from_wei(amounts[-1], 'ether')
    except:
        return 0

def execute_strike(r_compra, r_venta, t_addr, t_nombre, profit_usd):
    notify(f"🎯 EJECUTANDO: {t_nombre} | Profit Est: ${profit_usd:.2f}")
    try:
        contrato = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_ASTRALIX)
        tx = contrato.functions.ejecutarArbitraje(
            w3.to_checksum_address(r_compra), w3.to_checksum_address(r_venta),
            WBNB_ADDR, w3.to_checksum_address(t_addr), w3.to_wei(CAPITAL_WBNB, 'ether')
        ).build_transaction({
            'from': MI_BILLETERA,
            'nonce': w3.eth.get_transaction_count(MI_BILLETERA),
            'gas': GAS_LIMIT,
            'gasPrice': w3.eth.gas_price
        })
        signed = w3.eth.account.sign_transaction(tx, PRIV_KEY)
        tx_hash = w3.eth.send_raw_transaction(signed.raw_transaction)
        notify(f"✅ ¡DISPARO ENVIADO! Hash: {w3.to_hex(tx_hash)}")
    except Exception as e:
        notify(f"🛡️ BLOQUEO / ERROR: {e}")

# --- 🚀 ARRANQUE ---
notify("🚀 AstraliX Kaioken Online 24/7 en Railway")

while True:
    try:
        precio_bnb = float(get_price(DEXs["Pancake"], 1, [WBNB_ADDR, USDT_ADDR]))
        gas_usd = float(w3.from_wei(w3.eth.gas_price * GAS_LIMIT, 'ether')) * precio_bnb

        for t_name, t_addr in TOKENS.items():
            for n1, a1 in DEXs.items():
                for n2, a2 in DEXs.items():
                    if n1 == n2: continue
                    p1 = get_price(a1, CAPITAL_WBNB, [WBNB_ADDR, t_addr])
                    if p1 == 0: continue
                    p2 = get_price(a2, p1, [t_addr, WBNB_ADDR])
                    
                    neto = (float(p2) - CAPITAL_WBNB) * precio_bnb - gas_usd
                    
                    if neto > PROFIT_MIN_USD:
                        execute_strike(a1, a2, t_addr, t_name, neto)
                        time.sleep(30)
                    elif neto > FILTRO_RADAR: 
                        print(f"🎯 {t_name:<5} | Neto: ${neto:.3f}")
        time.sleep(1)
    except Exception as e:
        print(f"Error Loop: {e}")
        time.sleep(10)