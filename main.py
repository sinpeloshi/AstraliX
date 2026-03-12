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

# --- 🧨 LA FÓRMULA GUERRILLA ---
CAPITAL_SNIPER = 0.0004 
RETRASO_COMPRA = 3      
RETRASO_VENTA = 3       

# 🛡️ GESTIÓN DE GAS (Ajustado según capturas de Denis)
GAS_PRICE_GWEI = w3.to_wei(1.05, 'gwei') 
GAS_LIMIT_COMPRA = 250000
GAS_LIMIT_APPROVE = 75000  # Captura mostró ~46,000 usados
GAS_LIMIT_VENTA = 250000   # Captura mostró ~160,000 usados

FOUR_MEME_ROUTER = w3.to_checksum_address("0x5c952063c7fc8610ffdb798152d69f0b9550762b")

# 💥 LAS FIRMAS RAW
FIRMA_CREACION = "0x519ebb10" 
FIRMA_COMPRA = "0x87f27655"   
FIRMA_APPROVE = "0x095ea7b3"  
FIRMA_VENTA = "0x0da74935"    

# --- 🔑 CREDENCIALES ---
PRIV_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"
MI_BILLETERA = w3.eth.account.from_key(PRIV_KEY).address 
TG_TOKEN = '8783847744:AAHdwwlEqP7HCgSXoFxRdD8snr5FRhT1OUo'
TG_ID = '6580309816'

ABI_ERC20 = '[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"}]'

TOKENS_COMPRADOS = set()

def notify(msg):
    try: requests.post(f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage", json={"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}, timeout=5)
    except: pass

def chequear_fondos():
    balance = w3.eth.get_balance(MI_BILLETERA)
    return balance >= w3.to_wei(CAPITAL_SNIPER + 0.0005, 'ether')

def fire_strike_full_cycle(token_addr):
    if token_addr in TOKENS_COMPRADOS: return
    TOKENS_COMPRADOS.add(token_addr)
    
    if not chequear_fondos():
        notify("⚠️ *SALDO BAJO*: Radar activo, pero saldo insuficiente para operar.")
        return

    print(f"\n🎯 OBJETIVO FIJADO: {token_addr}", flush=True)
    notify(f"🎯 *TARGET DETECTADO*\n`{token_addr}`\nSecuencia RAW iniciada...")
    
    time.sleep(RETRASO_COMPRA)
    
    # --- 🛒 FASE 1: COMPRA RAW ---
    try:
        monto_wei = w3.to_wei(CAPITAL_SNIPER, 'ether')
        tx_data_buy = FIRMA_COMPRA + token_addr.lower().replace("0x","").zfill(64) + hex(monto_wei).replace("0x","").zfill(64) + "0"*64
        
        tx_buy = {
            'chainId': 56, 'from': MI_BILLETERA, 'to': FOUR_MEME_ROUTER, 'value': monto_wei,
            'nonce': w3.eth.get_transaction_count(MI_BILLETERA), 
            'gas': GAS_LIMIT_COMPRA, 'gasPrice': GAS_PRICE_GWEI, 'data': tx_data_buy
        }
        
        tx_h_buy = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_buy, PRIV_KEY).raw_transaction)
        print(f"✅ COMPRA ENVIADA: {w3.to_hex(tx_h_buy)}", flush=True)

        # --- ⏳ FASE 2: HOLD RELÁMPAGO ---
        time.sleep(RETRASO_VENTA) 

        # --- 💰 FASE 3: APPROVE + VENTA (Bypass Slippage) ---
        token_c = w3.eth.contract(address=token_addr, abi=ABI_ERC20)
        
        balance = 0
        for i in range(3):
            balance = token_c.functions.balanceOf(MI_BILLETERA).call()
            if balance > 0: break
            time.sleep(1)

        if balance > 0:
            print(f"🔄 Liquidando {balance} tokens...", flush=True)
            current_nonce = w3.eth.get_transaction_count(MI_BILLETERA)
            
            # 1. APPROVE RAW (Al contrato del Token)
            router_param = FOUR_MEME_ROUTER.lower().replace("0x","").zfill(64)
            approve_amount_param = hex(balance).replace("0x","").zfill(64)
            tx_data_app = FIRMA_APPROVE + router_param + approve_amount_param
            
            tx_app = {
                'chainId': 56, 'from': MI_BILLETERA, 'to': token_addr, 'value': 0, 
                'nonce': current_nonce, 
                'gas': GAS_LIMIT_APPROVE, 'gasPrice': GAS_PRICE_GWEI, 'data': tx_data_app
            }
            tx_h_app = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_app, PRIV_KEY).raw_transaction)
            print(f"🔓 APPROVE ENVIADO: {w3.to_hex(tx_h_app)}", flush=True)
            
            # 2. VENTA RAW (Con "minFunds" dinámico)
            origin_param = "0" * 64
            token_param = token_addr.lower().replace("0x","").zfill(64)
            amount_param = hex(balance).replace("0x","").zfill(64)
            
            # EL TRUCO: minFunds = 1 Wei (Hex: 0x1) en lugar de ceros.
            minFunds_param = hex(1).replace("0x","").zfill(64) 
            
            tx_data_sell = FIRMA_VENTA + origin_param + token_param + amount_param + minFunds_param
            
            tx_sell = {
                'chainId': 56, 'from': MI_BILLETERA, 'to': FOUR_MEME_ROUTER, 'value': 0, 
                'nonce': current_nonce + 1, # Disparo doble
                'gas': GAS_LIMIT_VENTA, 'gasPrice': GAS_PRICE_GWEI, 'data': tx_data_sell
            }
            tx_h_sell = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_sell, PRIV_KEY).raw_transaction)
            print(f"💰 VENTA ENVIADA: {w3.to_hex(tx_h_sell)}", flush=True)
            notify(f"💰 *VENTA EJECUTADA*\nTX: `{w3.to_hex(tx_h_sell)}`")
        else:
            print("❌ Saldo 0. No se pudo vender.", flush=True)

    except Exception as e: 
        print(f"❌ Error: {e}", flush=True)
        notify(f"🚨 *FALLO EN OPERACIÓN*: {str(e)[:50]}")

def scan_blocks():
    global w3
    print(f"☢️ AstraliX V35: MODO BYPASS SLIPPAGE. Escaneando...", flush=True)
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
    notify("🧨 *ASTRALIX V35 ONLINE*\nSlippage Dinámico inyectado.")
    scan_blocks()