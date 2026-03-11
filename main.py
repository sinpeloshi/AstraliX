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

# --- 🧨 CONFIGURACIÓN ESTRATÉGICA ---
CAPITAL_SNIPER = 0.005 
GAS_MULTIPLIER = 10.0  
RETRASO_COMPRA = 5     
RETRASO_VENTA = 15     

# ROUTERS
FOUR_MEME_ROUTER = w3.to_checksum_address("0x5c952063c7fc8610ffdb798152d69f0b9550762b")
PANCAKE_ROUTER = w3.to_checksum_address("0x10ED43C718714eb63d5aA57B78B54704E256024E")
WBNB = w3.to_checksum_address("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")

# FIRMAS
FIRMA_CREACION = "0x519ebb10" 
FIRMA_COMPRA = "0x87f27655"   # buyTokenAMAP (Four.meme)

# --- 🔑 IDENTIDAD ---
PRIV_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"
MI_BILLETERA = w3.eth.account.from_key(PRIV_KEY).address 
TG_TOKEN = '8783847744:AAHdwwlEqP7HCgSXoFxRdD8snr5FRhT1OUo'
TG_ID = '6580309816'

ABI_ERC20 = '[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"},{"inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"name":"approve","outputs":[{"name":"bool","name":"","type":"bool"}],"type":"function"}]'
ABI_PANCAKE = '[{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForETHSupportingFeeOnTransferTokens","outputs":[],"stateMutability":"nonpayable","type":"function"}]'

TOKENS_COMPRADOS = set()

def notify(msg):
    try: requests.post(f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage", json={"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}, timeout=5)
    except: pass

def chequear_fondos():
    balance = w3.eth.get_balance(MI_BILLETERA)
    return balance > w3.to_wei(CAPITAL_SNIPER + 0.0015, 'ether')

def fire_strike_hybrid(token_addr):
    if token_addr in TOKENS_COMPRADOS: return
    TOKENS_COMPRADOS.add(token_addr)
    
    if not chequear_fondos(): return

    print(f"\n🎯 OBJETIVO: {token_addr}. Esperando {RETRASO_COMPRA}s...", flush=True)
    time.sleep(RETRASO_COMPRA)
    
    # --- 🛒 COMPRA (FOUR.MEME) ---
    try:
        monto_wei = w3.to_wei(CAPITAL_SNIPER, 'ether')
        tx_data = FIRMA_COMPRA + token_addr.lower().replace("0x","").zfill(64) + \
                  hex(monto_wei).replace("0x","").zfill(64) + "0"*64
        
        tx_buy = {'chainId': 56, 'from': MI_BILLETERA, 'to': FOUR_MEME_ROUTER, 'value': monto_wei,
                  'nonce': w3.eth.get_transaction_count(MI_BILLETERA), 'gas': 850000, 
                  'gasPrice': int(w3.eth.gas_price * GAS_MULTIPLIER), 'data': tx_data}
        
        tx_h_buy = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_buy, PRIV_KEY).raw_transaction)
        print(f"✅ COMPRA ENVIADA: {w3.to_hex(tx_h_buy)}", flush=True)

        # --- ⏳ HOLD ---
        print(f"⏳ Esperando {RETRASO_VENTA}s para vender en PancakeSwap...", flush=True)
        time.sleep(RETRASO_VENTA)

        # --- 💰 VENTA (PANCAKESWAP) ---
        token_c = w3.eth.contract(address=token_addr, abi=ABI_ERC20)
        balance = token_c.functions.balanceOf(MI_BILLETERA).call()
        
        if balance > 0:
            print(f"🔄 Liquidando {balance} tokens en PancakeSwap...", flush=True)
            # Approve Pancake
            tx_app = token_c.functions.approve(PANCAKE_ROUTER, 2**256 - 1).build_transaction({
                'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA),
                'gas': 120000, 'gasPrice': int(w3.eth.gas_price * (GAS_MULTIPLIER + 2))})
            w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_app, PRIV_KEY).raw_transaction)
            
            time.sleep(5) # Buffer de confirmación
            
            pancake_c = w3.eth.contract(address=PANCAKE_ROUTER, abi=ABI_PANCAKE)
            tx_sell = pancake_c.functions.swapExactTokensForETHSupportingFeeOnTransferTokens(
                balance, 0, [token_addr, WBNB], MI_BILLETERA, int(time.time()) + 600
            ).build_transaction({
                'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA),
                'gas': 900000, 'gasPrice': int(w3.eth.gas_price * (GAS_MULTIPLIER + 2))
            })
            
            tx_h_sell = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx_sell, PRIV_KEY).raw_transaction)
            print(f"💰 VENTA EXITOSA: {w3.to_hex(tx_h_sell)}", flush=True)
            notify(f"💰 *CICLO COMPLETADO*\n¡Venta en PancakeSwap exitosa!\nTX: `{w3.to_hex(tx_h_sell)}`")
        else:
            print("❌ Saldo 0. El token podría ser un rugpull.", flush=True)

    except Exception as e: 
        print(f"❌ Error: {e}", flush=True)

def scan_blocks():
    global w3
    print(f"☢️ ASTRALIX V28: HYBRID ENGINE (Four.meme + Pancake).", flush=True)
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
                                        fire_strike_hybrid(t)
                                        break
                last_block = current_block
            time.sleep(1)
        except:
            time.sleep(2)
            w3 = conectar_nodo()

if __name__ == "__main__":
    scan_blocks()