import time
import requests
from web3 import Web3
from web3.middleware import ExtraDataToPOAMiddleware 

# --- 🛰️ CONEXIÓN Y RED ---
RPC_NODES = ["https://bsc-dataseed.binance.org/", "https://bsc-dataseed1.defibit.io/"]

def conectar_nodo():
    for rpc in RPC_NODES:
        try:
            w3 = Web3(Web3.HTTPProvider(rpc, request_kwargs={'timeout': 15}))
            w3.middleware_onion.inject(ExtraDataToPOAMiddleware, layer=0)
            if w3.is_connected(): return w3
        except: pass
    return None

w3 = conectar_nodo()

# --- 🧨 CONFIGURACIÓN DE ESTRATEGIA ---
CAPITAL_SNIPER = 0.005 # BNB a invertir por token
GAS_MULTIPLIER = 10.0  # Multiplicador de prioridad
RETRASO_COMPRA = 5     # Segundos de espera para evadir el Anti-Bot
RETRASO_VENTA = 15     # Segundos en hold antes de la venta total

FOUR_MEME_ROUTER = w3.to_checksum_address("0x5c952063c7fc8610ffdb798152d69f0b9550762b")
FIRMA_CREACION = "0x519ebb10" 
FIRMA_COMPRA = "0x87f27655"   # Método: buyTokenAMAP
FIRMA_VENTA = "0x06e7b98f"    # Método: sellTokenAMAP

# --- 🔑 IDENTIDAD ---
PRIV_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"
MI_BILLETERA = w3.eth.account.from_key(PRIV_KEY).address 
TG_TOKEN = '8783847744:AAHdwwlEqP7HCgSXoFxRdD8snr5FRhT1OUo'
TG_ID = '6580309816'

ABI_ERC20 = '[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"},{"inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"name":"approve","outputs":[{"name":"bool","name":"","type":"bool"}],"type":"function"}]'

TOKENS_COMPRADOS = set()

def notify(msg):
    try: requests.post(f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage", json={"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}, timeout=5)
    except: pass

def chequear_fondos():
    """Verifica que haya BNB suficiente para la inversión y las comisiones de red."""
    balance_wei = w3.eth.get_balance(MI_BILLETERA)
    # Se exige el capital + un margen seguro de ~0.002 BNB para pagar el GAS
    minimo_necesario = w3.to_wei(CAPITAL_SNIPER + 0.002, 'ether')
    
    if balance_wei < minimo_necesario:
        print(f"⚠️ BNB INSUFICIENTE. Tenés {w3.from_wei(balance_wei, 'ether'):.4f} BNB. Necesitás recargar.", flush=True)
        return False
    return True

def fire_strike_full_cycle(token_addr):
    if token_addr in TOKENS_COMPRADOS: return
    TOKENS_COMPRADOS.add(token_addr)
    
    # Check de seguridad antes de actuar
    if not chequear_fondos():
        notify("🚨 *FONDOS INSUFICIENTES*\nEl radar detectó un token, pero no hay suficiente BNB para ejecutar la operación segura. Recargá la wallet.")
        return

    print(f"\n🎯 OBJETIVO DETECTADO: {token_addr}", flush=True)
    print(f"⏳ Esperando {RETRASO_COMPRA}s para evitar bloqueos...", flush=True)
    notify(f"🎯 *TARGET FIJADO*\n`{token_addr}`\nIniciando ciclo automático...")
    
    time.sleep(RETRASO_COMPRA)
    
    # --- 🛒 FASE 1: COMPRA RAW ---
    try:
        monto_wei = w3.to_wei(CAPITAL_SNIPER, 'ether')
        
        # Estructura AMAP: Function ID + Token + Funds + MinOut
        tx_data = FIRMA_COMPRA + token_addr.lower().replace("0x","").zfill(64) + \
                  hex(monto_wei).replace("0x","").zfill(64) + "0"*64
        
        tx_buy = {
            'chainId': 56, 'from': MI_BILLETERA, 'to': FOUR_MEME_ROUTER, 'value': monto_wei,
            'nonce': w3.eth.get_transaction_count(MI_BILLETERA), 'gas': 800000, 
            'gasPrice': int(w3.eth.gas_price * GAS_MULTIPLIER), 'data': tx_data
        }
        
        tx_hash = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_buy, PRIV_KEY).raw_transaction)
        print(f"✅ COMPRA ENVIADA: {w3.to_hex(tx_hash)}", flush=True)

        # --- ⏳ FASE 2: HOLD ---
        print(f"⏳ Posición tomada. Esperando {RETRASO_VENTA}s para vender...", flush=True)
        time.sleep(RETRASO_VENTA)

        # --- 💰 FASE 3: VENTA RAW ---
        token_c = w3.eth.contract(address=token_addr, abi=ABI_ERC20)
        balance = token_c.functions.balanceOf(MI_BILLETERA).call()
        
        if balance > 0:
            print(f"🔄 Liquidando {balance} tokens...", flush=True)
            
            # 1. Autorización de gasto (Approve)
            tx_app = token_c.functions.approve(FOUR_MEME_ROUTER, balance).build_transaction({
                'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA),
                'gas': 120000, 'gasPrice': int(w3.eth.gas_price * GAS_MULTIPLIER)})
            w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_app, PRIV_KEY).raw_transaction)
            
            time.sleep(2) # Pausa técnica para asentamiento en la blockchain
            
            # 2. Venta total AMAP
            sell_data = FIRMA_VENTA + token_addr.lower().replace("0x","").zfill(64) + \
                        hex(balance).replace("0x","").zfill(64) + "0"*64
            
            tx_sell = {
                'chainId': 56, 'from': MI_BILLETERA, 'to': FOUR_MEME_ROUTER, 'value': 0,
                'nonce': w3.eth.get_transaction_count(MI_BILLETERA), 'gas': 800000,
                'gasPrice': int(w3.eth.gas_price * GAS_MULTIPLIER), 'data': sell_data
            }
            
            tx_h_sell = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_sell, PRIV_KEY).raw_transaction)
            print(f"💰 VENTA EJECUTADA: {w3.to_hex(tx_h_sell)}", flush=True)
            notify(f"💰 *CICLO COMPLETADO*\nOperación finalizada. TX de Venta: `{w3.to_hex(tx_h_sell)}`")
        else:
            print("❌ Balance de tokens en cero. La compra falló o se demoró demasiado.", flush=True)

    except Exception as e: 
        print(f"❌ Error operativo: {e}", flush=True)

def scan_blocks():
    global w3
    print(f"☢️ ASTRALIX FINAL EDITION. Escáner RAW activo.", flush=True)
    if not chequear_fondos(): 
        print("⚠️ Advertencia inicial: Fondos críticamente bajos.", flush=True)
    
    last_block = w3.eth.block_number
    while True:
        try:
            current_block = w3.eth.block_number
            if current_block > last_block:
                block = w3.eth.get_block(current_block, full_transactions=True)
                for tx in block.transactions:
                    if tx.to and tx.to.lower() == FOUR_MEME_ROUTER.lower():
                        if w3.to_hex(tx["input"])[:10] == FIRMA_CREACION:
                            receipt = w3.eth.get_transaction_receipt(tx.hash)
                            for log in receipt['logs']:
                                if len(log['topics']) > 0:
                                    t = log['address']
                                    if t.lower() != FOUR_MEME_ROUTER.lower():
                                        fire_strike_full_cycle(t)
                                        break
                last_block = current_block
            time.sleep(1)
        except:
            time.sleep(2)
            w3 = conectar_nodo()

if __name__ == "__main__":
    notify("🧨 *ASTRALIX INICIADO*\nVersión final desplegada y protecciones activas.")
    scan_blocks()