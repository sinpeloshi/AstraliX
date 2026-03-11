import time
import requests
from web3 import Web3
from web3.middleware import ExtraDataToPOAMiddleware 

# --- 🛰️ CONEXIÓN ---
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

# --- 🧨 CONFIGURACIÓN ---
CAPITAL_SNIPER = 0.005 
GAS_MULTIPLIER = 10.0  
RETRASO_COMPRA = 3     # 3s para saltar Deadblocks
RETRASO_VENTA = 15     # 15s para sacar profit

FOUR_MEME_ROUTER = w3.to_checksum_address("0x5c952063c7fc8610ffdb798152d69f0b9550762b")
FIRMA_CREACION = "0x519ebb10" 
FIRMA_COMPRA = "0x87f27655"   # LA NUEVA QUE ENCONTRASTE
FIRMA_VENTA = "0x06e7b98f"    # Venta AMAP

PRIV_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"
MI_BILLETERA = w3.eth.account.from_key(PRIV_KEY).address 
TG_TOKEN = '8783847744:AAHdwwlEqP7HCgSXoFxRdD8snr5FRhT1OUo'
TG_ID = '6580309816'

ABI_ERC20 = '[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"},{"inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"name":"approve","outputs":[{"name":"bool","name":"","type":"bool"}],"type":"function"}]'

TOKENS_COMPRADOS = set()

def notify(msg):
    try: requests.post(f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage", json={"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}, timeout=5)
    except: pass

def fire_strike_full_cycle(token_addr):
    if token_addr in TOKENS_COMPRADOS: return
    TOKENS_COMPRADOS.add(token_addr)
    
    print(f"\n🚨 OBJETIVO: {token_addr}. Esperando {RETRASO_COMPRA}s...", flush=True)
    time.sleep(RETRASO_COMPRA)
    
    # --- 🛒 FASE DE COMPRA ---
    try:
        monto_wei = w3.to_wei(CAPITAL_SNIPER, 'ether')
        # Payload: ID + token(32b) + funds(32b) + minAmount(32b)
        tx_data = FIRMA_COMPRA + token_addr.lower().replace("0x","").zfill(64) + \
                  hex(monto_wei).replace("0x","").zfill(64) + "0"*64
        
        tx_buy = {'chainId': 56, 'from': MI_BILLETERA, 'to': FOUR_MEME_ROUTER, 'value': monto_wei,
                  'nonce': w3.eth.get_transaction_count(MI_BILLETERA), 'gas': 800000, 
                  'gasPrice': int(w3.eth.gas_price * GAS_MULTIPLIER), 'data': tx_data}
        
        tx_hash = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_buy, PRIV_KEY).raw_transaction)
        print(f"✅ COMPRA ENVIADA: {w3.to_hex(tx_hash)}", flush=True)
        notify(f"🛒 *COMPRA ENVIADA*\nToken: `{token_addr}`\nEsperando {RETRASO_VENTA}s para vender...")

        # --- ⏳ ESPERA PARA PROFIT ---
        time.sleep(RETRASO_VENTA)

        # --- 💰 FASE DE VENTA ---
        token_c = w3.eth.contract(address=token_addr, abi=ABI_ERC20)
        balance = token_c.functions.balanceOf(MI_BILLETERA).call()
        
        if balance > 0:
            print(f"🔄 Vendiendo {balance} tokens...", flush=True)
            # Aprobar
            tx_app = token_c.functions.approve(FOUR_MEME_ROUTER, balance).build_transaction({
                'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA),
                'gas': 100000, 'gasPrice': int(w3.eth.gas_price * GAS_MULTIPLIER)})
            w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_app, PRIV_KEY).raw_transaction)
            time.sleep(2) # Pausa para que el approve asiente
            
            # Vender (Payload RAW para sellTokenAMAP)
            # Payload: ID + token(32b) + amountIn(32b) + minAmountOut(32b)
            sell_data = FIRMA_VENTA + token_addr.lower().replace("0x","").zfill(64) + \
                        hex(balance).replace("0x","").zfill(64) + "0"*64
            
            tx_sell = {'chainId': 56, 'from': MI_BILLETERA, 'to': FOUR_MEME_ROUTER, 'value': 0,
                       'nonce': w3.eth.get_transaction_count(MI_BILLETERA), 'gas': 800000,
                       'gasPrice': int(w3.eth.gas_price * GAS_MULTIPLIER), 'data': sell_data}
            
            tx_h_sell = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_sell, PRIV_KEY).raw_transaction)
            print(f"💰 VENTA ENVIADA: {w3.to_hex(tx_h_sell)}", flush=True)
            notify(f"💰 *VENTA ENVIADA*\nTX: `{w3.to_hex(tx_h_sell)}`")
        else:
            print("❌ No hay tokens para vender.", flush=True)

    except Exception as e: print(f"❌ Error: {e}", flush=True)

def scan_blocks():
    global w3
    print(f"☢️ AstraliX V25: FULL CYCLE (3s/15s). Firma {FIRMA_COMPRA}.", flush=True)
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
    scan_blocks()