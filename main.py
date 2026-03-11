import time
import requests
from web3 import Web3
from web3.middleware import ExtraDataToPOAMiddleware 

# --- 🛰️ CONEXIÓN BLINDADA ---
RPC_NODES = ["https://bsc-dataseed.binance.org/", "https://bsc-dataseed1.defibit.io/", "https://bsc-dataseed1.ninicoin.io/"]

def conectar_nodo():
    for rpc in RPC_NODES:
        try:
            w3 = Web3(Web3.HTTPProvider(rpc, request_kwargs={'timeout': 15}))
            w3.middleware_onion.inject(ExtraDataToPOAMiddleware, layer=0)
            if w3.is_connected(): return w3
        except: pass
    return None

w3 = conectar_nodo()

# --- 🧨 CONFIGURACIÓN KAMIKAZE ---
CAPITAL_SNIPER = 0.005 
GAS_MULTIPLIER = 5.0   

# 🎯 OBJETIVO: FOUR.MEME
FOUR_MEME_ROUTER = w3.to_checksum_address("0x5c952063c7fc8610ffdb798152d69f0b9550762b")

# 💣 LA ESCOPETA (Dejamos SOLO la que confirmaste que es de Creación)
FIRMACIONES_SOSPECHOSAS = ["0x519ebb10"]

# --- 🔑 IDENTIDAD ---
PRIV_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"
CONTRATO_ADDR = w3.to_checksum_address("0xF44f4D75Efc8d60d9383319D1C69553A1201be28")
MI_BILLETERA = w3.eth.account.from_key(PRIV_KEY).address 
TG_TOKEN = '8783847744:AAHdwwlEqP7HCgSXoFxRdD8snr5FRhT1OUo'
TG_ID = '6580309816'
WBNB_ADDR = w3.to_checksum_address("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")

# --- 📜 ABIs ---
ABI_ERC20 = '[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"},{"inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"name":"approve","outputs":[{"name":"bool","name":"","type":"bool"}],"type":"function"}]'
ABI_FOUR_MEME = '[{"inputs":[{"name":"token","type":"address"},{"name":"amountIn","type":"uint256"},{"name":"minAmountOut","type":"uint256"}],"name":"buy","outputs":[],"type":"function"}]'
ABI_APEX = '[{"inputs":[{"name":"targets","type":"address[]"},{"name":"payloads","type":"bytes[]"},{"name":"values","type":"uint256[]"},{"name":"minerBribe","type":"uint256"}],"name":"apexStrike","outputs":[],"type":"function"}]'

# 🧠 MEMORIA DEL BOT (Para no comprar dos veces lo mismo)
TOKENS_COMPRADOS = set()

def notify(msg):
    try: requests.post(f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage", json={"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}, timeout=5)
    except: pass

def fire_strike(token_to_buy, firma):
    token_addr = w3.to_checksum_address(token_to_buy)
    
    # Seguro anti-ráfagas
    if token_addr in TOKENS_COMPRADOS:
        return False
        
    TOKENS_COMPRADOS.add(token_addr) # Lo guardamos en la memoria
    
    print(f"\n🚨 ¡BLANCO DETECTADO! (Firma: {firma}) -> {token_addr}", flush=True)
    notify(f"💥 *DISPARO ASTRALIX*\nObjetivo: `{token_addr}`")
    try:
        wbnb_c = w3.eth.contract(address=WBNB_ADDR, abi=ABI_ERC20)
        four_c = w3.eth.contract(address=FOUR_MEME_ROUTER, abi=ABI_FOUR_MEME)
        apex_c = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_APEX)
        monto = w3.to_wei(CAPITAL_SNIPER, 'ether')
        
        p_app = wbnb_c.encode_abi("approve", args=[FOUR_MEME_ROUTER, monto])
        p_buy = four_c.encode_abi("buy", args=[token_addr, monto, 0])
        
        tx = apex_c.functions.apexStrike(
            [WBNB_ADDR, FOUR_MEME_ROUTER], [w3.to_bytes(hexstr=p_app), w3.to_bytes(hexstr=p_buy)], [0, 0], 0
        ).build_transaction({'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA), 'gas': 900000, 'gasPrice': int(w3.eth.gas_price * GAS_MULTIPLIER)})
        
        tx_hash = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx, PRIV_KEY).raw_transaction)
        print(f"✅ ¡COMPRA ENVIADA!: {w3.to_hex(tx_hash)}", flush=True)
        return True
    except Exception as e: 
        print(f"❌ Disparo fallido: {e}", flush=True)
        return False

def scan_blocks():
    global w3
    if not w3: return
    print("☢️ AstraliX V19: MIRA CALIBRADA A 0x519ebb10. Memoria Anti-Ráfaga ON.", flush=True)
    last_block = w3.eth.block_number
    
    while True:
        try:
            current_block = w3.eth.block_number
            if current_block > last_block:
                print(f"📦 Bloque {current_block}", flush=True)
                block = w3.eth.get_block(current_block, full_transactions=True)
                for tx in block.transactions:
                    if tx.to and tx.to.lower() == FOUR_MEME_ROUTER.lower():
                        
                        input_data = tx.input.hex()
                        method_id = "0x" + input_data[:8] if len(input_data) >= 8 else "0x00000000"
                        
                        # AHORA SOLO BUSCAMOS LA FIRMA CONFIRMADA DE CREACIÓN
                        if method_id in FIRMACIONES_SOSPECHOSAS:
                            print(f"   ⚠️ FIRMA {method_id} DETECTADA. Analizando recibo...", flush=True)
                            receipt = w3.eth.get_transaction_receipt(tx.hash)
                            for log in receipt['logs']:
                                if len(log['topics']) > 0:
                                    potential_token = log['address']
                                    if potential_token.lower() != FOUR_MEME_ROUTER.lower():
                                        if fire_strike(potential_token, method_id):
                                            break # Rompemos el loop para no disparar a los demás logs de esta misma TX
                        
                last_block = current_block
            time.sleep(2)
        except Exception:
            time.sleep(3)
            w3 = conectar_nodo()

if __name__ == "__main__":
    notify("🧨 *ASTRALIX V19 ONLINE*\nSeguro anti-ráfagas activado.")
    scan_blocks()