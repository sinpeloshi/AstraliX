import time
import requests
from web3 import Web3
from web3.middleware import geth_poa_middleware

# --- 🛰️ CONEXIÓN ---
HTTP_URL = "https://solemn-orbital-thunder.bsc.quiknode.pro/70d0d80f07303278accd2349e2fc01c95018d18c/"
w3 = Web3(Web3.HTTPProvider(HTTP_URL, request_kwargs={'timeout': 20}))
w3.middleware_onion.inject(geth_poa_middleware, layer=0) 

# --- 🧨 CONFIGURACIÓN ---
CAPITAL_SNIPER = 0.005 
GAS_MULTIPLIER = 3.0   
TARGET_PROFIT = 1.15   

# 🎯 EL CAMBIO CLAVE: Usamos el Manager como Router de entrada
FOUR_MEME_ROUTER = w3.to_checksum_address("0x5c952063c7fc8610ffdb798152d69f0b9550762b")
CREATE_METHOD_ID = "0xedf9e251" 

# --- 🔑 IDENTIDAD ---
PRIV_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"
CONTRATO_ADDR = w3.to_checksum_address("0xF44f4D75Efc8d60d9383319D1C69553A1201be28")
MI_BILLETERA = w3.eth.account.from_key(PRIV_KEY).address 

TG_TOKEN = '8783847744:AAHdwwlEqP7HCgSXoFxRdD8snr5FRhT1OUo'
TG_ID = '6580309816'
WBNB_ADDR = w3.to_checksum_address("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")

# --- 📜 ABIs ---
ABI_ERC20 = '[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"},{"inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"name":"approve","outputs":[{"name":"bool","name":"","type":"bool"}],"type":"function"}]'
# ABI simplificada para la función de compra de Four.meme
ABI_FOUR_MEME = '[{"inputs":[{"name":"token","type":"address"},{"name":"amountIn","type":"uint256"},{"name":"minAmountOut","type":"uint256"}],"name":"buy","outputs":[],"type":"function"}]'
ABI_APEX = '[{"inputs":[{"name":"targets","type":"address[]"},{"name":"payloads","type":"bytes[]"},{"name":"values","type":"uint256[]"},{"name":"minerBribe","type":"uint256"}],"name":"apexStrike","outputs":[],"type":"function"}]'

def notify(msg):
    try: requests.post(f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage", json={"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}, timeout=5)
    except: pass

def fire_strike(tx_input):
    # Extraemos el address del token del input de la creación
    token_to_buy = "0x" + tx_input[34:74] if len(tx_input) > 74 else None
    if not token_to_buy: return
    token_to_buy = w3.to_checksum_address(token_to_buy)
    
    print(f"🚀 ATACANDO EN FOUR.MEME: {token_to_buy}", flush=True)
    notify(f"🎯 *ATAQUE DIRECTO:* `{token_to_buy}`")
    
    try:
        wbnb_c = w3.eth.contract(address=WBNB_ADDR, abi=ABI_ERC20)
        four_c = w3.eth.contract(address=FOUR_MEME_ROUTER, abi=ABI_FOUR_MEME)
        apex_c = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_APEX)
        monto = w3.to_wei(CAPITAL_SNIPER, 'ether')

        # 1. Aprobamos WBNB para el Manager de Four.meme
        p_app = wbnb_c.encode_abi("approve", args=[FOUR_MEME_ROUTER, monto])
        # 2. Llamamos a la función de compra del Manager
        p_buy = four_c.encode_abi("buy", args=[token_to_buy, monto, 0])
        
        tx = apex_c.functions.apexStrike(
            [WBNB_ADDR, FOUR_MEME_ROUTER], 
            [w3.to_bytes(hexstr=p_app), w3.to_bytes(hexstr=p_buy)], 
            [0, 0], 0
        ).build_transaction({
            'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA),
            'gas': 800000, 'gasPrice': int(w3.eth.gas_price * GAS_MULTIPLIER)
        })
        
        tx_hash = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx, PRIV_KEY).raw_transaction)
        print(f"✅ Compra enviada! Hash: {w3.to_hex(tx_hash)}", flush=True)
        
    except Exception as e:
        print(f"❌ Error en Ataque: {e}", flush=True)

def scan_blocks():
    print("☢️ AstraliX V12.5: Escaneando bloques de Four.meme...", flush=True)
    last_block = w3.eth.block_number
    while True:
        try:
            current_block = w3.eth.block_number
            if current_block > last_block:
                block = w3.eth.get_block(current_block, full_transactions=True)
                print(f"🔎 Bloque {current_block}", flush=True)
                for tx in block.transactions:
                    if tx.to and tx.to.lower() == FOUR_MEME_ROUTER.lower():
                        if tx.input.startswith(CREATE_METHOD_ID):
                            print("\n🚨 ¡CREACIÓN DETECTADA!", flush=True)
                            fire_strike(tx.input)
                last_block = current_block
            time.sleep(1.5)
        except Exception as e:
            time.sleep(3)

if __name__ == "__main__":
    notify("🛰️ *ASTRALIX V12.5 ONLINE*\nObjetivo: Four.meme Manager Directo.")
    scan_blocks()