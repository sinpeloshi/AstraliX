import time
import requests
from web3 import Web3
from web3.middleware import ExtraDataToPOAMiddleware 

# --- 🛰️ CONEXIÓN Y NODOS ---
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

# --- 🧨 LA FÓRMULA GUERRILLA: MODO MICRO-SNIPER ---
CAPITAL_SNIPER = 0.0005 # Reducido a 0.0005 BNB para adaptarse a tu saldo de 0.0023 BNB
RETRASO_COMPRA = 3      
RETRASO_VENTA = 3       

# 🛡️ GESTIÓN DE GAS (Ajustada al milímetro)
GAS_COMPRA_GWEI = w3.to_wei(1.5, 'gwei') 
GAS_VENTA_GWEI = w3.to_wei(2.0, 'gwei')  

FOUR_MEME_ROUTER = w3.to_checksum_address("0x5c952063c7fc8610ffdb798152d69f0b9550762b")
FIRMA_CREACION = "0x519ebb10" 
FIRMA_COMPRA = "0x87f27655"   
FIRMA_VENTA = "0x0da74935"    

# --- 🔑 CREDENCIALES ---
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
    balance = w3.eth.get_balance(MI_BILLETERA)
    # Ajustamos el margen de seguridad para que permita operar con tu saldo bajo
    # Necesita al menos CAPITAL_SNIPER (0.0005) + Gas Estimado (0.001) = 0.0015 BNB
    return balance > w3.to_wei(CAPITAL_SNIPER + 0.001, 'ether')

def fire_strike_full_cycle(token_addr):
    if token_addr in TOKENS_COMPRADOS: return
    TOKENS_COMPRADOS.add(token_addr)
    
    if not chequear_fondos():
        notify("⚠️ *SALDO BAJO*: El radar detectó un token pero no llegamos al mínimo operativo (0.0015 BNB).")
        return

    print(f"\n🎯 OBJETIVO FIJADO: {token_addr}", flush=True)
    notify(f"🎯 *TARGET DETECTADO*\n`{token_addr}`\nIniciando micro-compra en {RETRASO_COMPRA}s...")
    
    time.sleep(RETRASO_COMPRA)
    
    # --- 🛒 FASE 1: COMPRA ---
    try:
        monto_wei = w3.to_wei(CAPITAL_SNIPER, 'ether')
        tx_data = FIRMA_COMPRA + token_addr.lower().replace("0x","").zfill(64) + \
                  hex(monto_wei).replace("0x","").zfill(64) + "0"*64
        
        tx_buy = {
            'chainId': 56, 'from': MI_BILLETERA, 'to': FOUR_MEME_ROUTER, 'value': monto_wei,
            'nonce': w3.eth.get_transaction_count(MI_BILLETERA), 
            'gas': 350000, # Bajado de 600k a 350k para no saturar tu saldo retenido
            'gasPrice': GAS_COMPRA_GWEI, 
            'data': tx_data
        }
        
        tx_h_buy = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_buy, PRIV_KEY).raw_transaction)
        print(f"✅ COMPRA ENVIADA: {w3.to_hex(tx_h_buy)}", flush=True)

        # --- ⏳ FASE 2: HOLD RELÁMPAGO ---
        time.sleep(RETRASO_VENTA) 

        # --- 💰 FASE 3: VENTA (EL PROTOCOLO DENIS MEJORADO) ---
        token_c = w3.eth.contract(address=token_addr, abi=ABI_ERC20)
        
        balance = 0
        for i in range(3):
            balance = token_c.functions.balanceOf(MI_BILLETERA).call()
            if balance > 0: break
            time.sleep(1)

        if balance > 0:
            print(f"🔄 Preparando liquidación de {balance} tokens...", flush=True)
            
            # 🔥 MAGIA DEL NONCE
            current_nonce = w3.eth.get_transaction_count(MI_BILLETERA)
            
            # 1. Approve
            tx_app = token_c.functions.approve(FOUR_MEME_ROUTER, balance).build_transaction({
                'from': MI_BILLETERA, 
                'nonce': current_nonce,
                'gas': 60000, 
                'gasPrice': GAS_VENTA_GWEI
            })
            tx_h_app = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_app, PRIV_KEY).raw_transaction)
            print(f"🔓 Approve enviado: {w3.to_hex(tx_h_app)}", flush=True)
            
            # 2. Venta Inmediata
            origin_param = "0" * 64
            token_param = token_addr.lower().replace("0x","").zfill(64)
            amount_param = hex(balance).replace("0x","").zfill(64)
            minFunds_param = "0" * 64
            
            sell_data = FIRMA_VENTA + origin_param + token_param + amount_param + minFunds_param
            
            tx_sell = {
                'chainId': 56, 'from': MI_BILLETERA, 'to': FOUR_MEME_ROUTER, 'value': 0,
                'nonce': current_nonce + 1, 
                'gas': 400000, # Bajado de 600k a 400k
                'gasPrice': GAS_VENTA_GWEI, 
                'data': sell_data
            }
            
            tx_h_sell = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_sell, PRIV_KEY).raw_transaction)
            print(f"💰 VENTA ENVIADA (Fórmula Denis): {w3.to_hex(tx_h_sell)}", flush=True)
            notify(f"💰 *VENTA EJECUTADA*\nTX: `{w3.to_hex(tx_h_sell)}`")
        else:
            print("❌ Saldo 0. No se pudo vender.", flush=True)

    except Exception as e: 
        print(f"❌ Error: {e}", flush=True)
        notify(f"🚨 *FALLO EN OPERACIÓN*: {str(e)[:50]}")

def scan_blocks():
    global w3
    print(f"☢️ AstraliX V33: PROTOCOLO DENIS. Escaneando Matrix...", flush=True)
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
    notify("🧨 *ASTRALIX V33 ONLINE*\nMotor Micro-Guerrilla activado (Tiempos y Gas Maximizados).")
    scan_blocks()