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
CAPITAL_SNIPER = 0.005 # Inversión en BNB Nativo
GAS_MULTIPLIER = 5.0   

# 🎯 OBJETIVO: FOUR.MEME
FOUR_MEME_ROUTER = w3.to_checksum_address("0x5c952063c7fc8610ffdb798152d69f0b9550762b")
FIRMA_CREACION = "0x519ebb10" # Firma que crea el token
FIRMA_COMPRA = "0xcce7ec13"   # LLAVE MAESTRA RAW DE COMPRA

# --- 🔑 IDENTIDAD ---
PRIV_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"
MI_BILLETERA = w3.eth.account.from_key(PRIV_KEY).address 

TG_TOKEN = '8783847744:AAHdwwlEqP7HCgSXoFxRdD8snr5FRhT1OUo'
TG_ID = '6580309816'

# 🧠 MEMORIA DEL BOT 
TOKENS_COMPRADOS = set()

def notify(msg):
    try: requests.post(f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage", json={"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}, timeout=5)
    except: pass

def fire_strike_raw(token_to_buy):
    token_addr = w3.to_checksum_address(token_to_buy)
    
    if token_addr in TOKENS_COMPRADOS:
        return False
    TOKENS_COMPRADOS.add(token_addr)
    
    print(f"\n🚨 ¡CREACIÓN DETECTADA! -> {token_addr}", flush=True)
    print(f"🔥 INYECTANDO CÓDIGO RAW (GAS 5.0x)...", flush=True)
    notify(f"💥 *FRANCOTIRADOR RAW ACTIVADO*\nObjetivo: `{token_addr}`\nEntrando con BNB Nativo.")
    
    try:
        method_id = FIRMA_COMPRA
        clean_token = token_addr.lower().replace("0x", "")
        padded_token = clean_token.zfill(64)
        padded_amount = "0" * 64
        
        tx_data = method_id + padded_token + padded_amount

        tx = {
            'chainId': 56, 
            'from': MI_BILLETERA,
            'to': FOUR_MEME_ROUTER,
            'value': w3.to_wei(CAPITAL_SNIPER, 'ether'),
            'nonce': w3.eth.get_transaction_count(MI_BILLETERA),
            'gas': 800000,
            'gasPrice': int(w3.eth.gas_price * GAS_MULTIPLIER),
            'data': tx_data 
        }
        
        signed_tx = w3.eth.account.sign_transaction(tx, PRIV_KEY)
        tx_hash = w3.eth.send_raw_transaction(signed_tx.raw_transaction)
        
        print(f"✅ ¡MISIL ENVIADO!: https://bscscan.com/tx/{w3.to_hex(tx_hash)}", flush=True)
        notify(f"✅ *¡COMPRA CONFIRMADA!*\nVe a Four.meme y preparate para vender.\nTX: `{w3.to_hex(tx_hash)}`")
        return True
        
    except Exception as e: 
        print(f"❌ Error en disparo RAW: {e}", flush=True)
        return False

def scan_blocks():
    global w3
    if not w3: return
    print("☢️ AstraliX V22.2: RAW HEX SNIPER. (Radar Visual Encendido).", flush=True)
    last_block = w3.eth.block_number
    
    while True:
        try:
            current_block = w3.eth.block_number
            if current_block > last_block:
                print(f"📦 Bloque {current_block} - Escaneando...", flush=True)
                block = w3.eth.get_block(current_block, full_transactions=True)
                
                for tx in block.transactions:
                    if tx.to and tx.to.lower() == FOUR_MEME_ROUTER.lower():
                        
                        raw_input = w3.to_hex(tx["input"])
                        method_id = raw_input[:10] if len(raw_input) >= 10 else "0x00000000"
                        
                        if method_id == FIRMA_CREACION:
                            print(f"   ⚠️ FIRMA DE CREACIÓN {FIRMA_CREACION} DETECTADA. Analizando...", flush=True)
                            receipt = w3.eth.get_transaction_receipt(tx.hash)
                            
                            for log in receipt['logs']:
                                if len(log['topics']) > 0:
                                    potential_token = log['address']
                                    if potential_token.lower() != FOUR_MEME_ROUTER.lower():
                                        if fire_strike_raw(potential_token):
                                            break 
                        else:
                            # ESTA ES LA LÍNEA QUE TE DEVUELVE LA TRANQUILIDAD
                            print(f"   📡 Pulso detectado en Four.meme (Firma: {method_id})", flush=True)
                            
                last_block = current_block
            time.sleep(2)
        except Exception:
            time.sleep(3)
            w3 = conectar_nodo()

if __name__ == "__main__":
    notify("🧨 *ASTRALIX V22.2 ONLINE*\nRadar visual encendido.")
    scan_blocks()