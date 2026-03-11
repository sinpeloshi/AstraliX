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
CAPITAL_SNIPER = 0.005 # Inversión en BNB
GAS_MULTIPLIER = 5.0   # MÁXIMA PRIORIDAD
TIEMPO_VENTA = 15      # Segundos a esperar antes de vender

# 🎯 OBJETIVO: FOUR.MEME
FOUR_MEME_ROUTER = w3.to_checksum_address("0x5c952063c7fc8610ffdb798152d69f0b9550762b")
FIRMA_CREACION = "0x519ebb10" 

# --- 🔑 IDENTIDAD ---
PRIV_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"
MI_BILLETERA = w3.eth.account.from_key(PRIV_KEY).address 

TG_TOKEN = '8783847744:AAHdwwlEqP7HCgSXoFxRdD8snr5FRhT1OUo'
TG_ID = '6580309816'

# --- 📜 ABIs (Compra, Venta y ERC20) ---
ABI_ERC20 = '[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"},{"inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"name":"approve","outputs":[{"name":"bool","name":"","type":"bool"}],"type":"function"}]'
ABI_FOUR_MEME_COMPRA = '[{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"minAmountOut","type":"uint256"}],"name":"buy","outputs":[],"stateMutability":"payable","type":"function"}]'
# Asumimos una firma estándar de venta para clones de Pump.fun
ABI_FOUR_MEME_VENTA = '[{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"minAmountOut","type":"uint256"}],"name":"sell","outputs":[],"stateMutability":"nonpayable","type":"function"}]'

TOKENS_COMPRADOS = set()

def notify(msg):
    try: requests.post(f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage", json={"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}, timeout=5)
    except: pass

def auto_sell(token_addr):
    print(f"\n⏳ Iniciando cuenta regresiva de {TIEMPO_VENTA} segundos para VENDER...", flush=True)
    time.sleep(TIEMPO_VENTA)
    
    print(f"🔥 PREPARANDO VENTA DE EXTRACCIÓN...", flush=True)
    try:
        # 1. Chequeamos cuántos tokens compramos realmente
        token_c = w3.eth.contract(address=token_addr, abi=ABI_ERC20)
        balance = token_c.functions.balanceOf(MI_BILLETERA).call()
        
        if balance == 0:
            print("❌ Balance cero. La compra falló o no se procesó a tiempo.", flush=True)
            notify(f"⚠️ *FALLO EN VENTA*\nEl balance de `{token_addr}` es 0.")
            return

        print(f"💼 Balance detectado: {balance} tokens. Aprobando contrato...", flush=True)
        
        # 2. Aprobación (Approve) para que Four.meme pueda retirar tus tokens
        nonce = w3.eth.get_transaction_count(MI_BILLETERA)
        tx_app = token_c.functions.approve(FOUR_MEME_ROUTER, balance).build_transaction({
            'from': MI_BILLETERA, 'nonce': nonce, 'gas': 100000, 'gasPrice': int(w3.eth.gas_price * GAS_MULTIPLIER)
        })
        signed_app = w3.eth.account.sign_transaction(tx_app, PRIV_KEY)
        tx_hash_app = w3.eth.send_raw_transaction(signed_app.raw_transaction)
        
        print("⏳ Esperando confirmación de la red para el Approve...", flush=True)
        w3.eth.wait_for_transaction_receipt(tx_hash_app, timeout=30)
        print("✅ Aprobación confirmada. Lanzando VENTA...", flush=True)
        
        # 3. Venta Total (Sell)
        four_c_sell = w3.eth.contract(address=FOUR_MEME_ROUTER, abi=ABI_FOUR_MEME_VENTA)
        nonce_sell = w3.eth.get_transaction_count(MI_BILLETERA)
        tx_sell = four_c_sell.functions.sell(token_addr, balance, 0).build_transaction({
            'from': MI_BILLETERA, 'nonce': nonce_sell, 'gas': 800000, 'gasPrice': int(w3.eth.gas_price * GAS_MULTIPLIER)
        })
        signed_sell = w3.eth.account.sign_transaction(tx_sell, PRIV_KEY)
        tx_hash_sell = w3.eth.send_raw_transaction(signed_sell.raw_transaction)
        
        print(f"💰 ¡VENTA ENVIADA!: https://bscscan.com/tx/{w3.to_hex(tx_hash_sell)}", flush=True)
        notify(f"💰 *VENTA EJECUTADA EXITOSAMENTE*\nRevisa tu billetera. TX: `{w3.to_hex(tx_hash_sell)}`")
        
    except Exception as e:
        print(f"❌ Error catastrófico en la venta: {e}", flush=True)
        notify(f"🚨 *ERROR AL VENDER*\nTuviste un error: {e}\n¡Vende manual urgente!")


def fire_strike_direct(token_to_buy):
    token_addr = w3.to_checksum_address(token_to_buy)
    
    if token_addr in TOKENS_COMPRADOS: return False
    TOKENS_COMPRADOS.add(token_addr)
    
    print(f"\n🚨 ¡CREACIÓN DETECTADA! -> {token_addr}", flush=True)
    print(f"🔥 DISPARANDO BNB DIRECTO (GAS 5.0x)...", flush=True)
    notify(f"💥 *ENTRADA KAMIKAZE*\nObjetivo: `{token_addr}`")
    
    try:
        four_c = w3.eth.contract(address=FOUR_MEME_ROUTER, abi=ABI_FOUR_MEME_COMPRA)
        monto_bnb = w3.to_wei(CAPITAL_SNIPER, 'ether')
        
        tx = four_c.functions.buy(token_addr, 0).build_transaction({
            'from': MI_BILLETERA, 'value': monto_bnb, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA), 
            'gas': 800000, 'gasPrice': int(w3.eth.gas_price * GAS_MULTIPLIER)
        })
        signed_tx = w3.eth.account.sign_transaction(tx, PRIV_KEY)
        tx_hash = w3.eth.send_raw_transaction(signed_tx.raw_transaction)
        
        print(f"✅ ¡COMPRA ENVIADA!: https://bscscan.com/tx/{w3.to_hex(tx_hash)}", flush=True)
        
        # ⚡ INICIA SECUENCIA DE SALIDA AUTOMÁTICA
        auto_sell(token_addr)
        return True
        
    except Exception as e: 
        print(f"❌ Error en compra: {e}", flush=True)
        return False


def scan_blocks():
    global w3
    if not w3: return
    print("☢️ AstraliX V21: AUTO-SELL (15 Segundos). BNB Nativo.", flush=True)
    last_block = w3.eth.block_number
    
    while True:
        try:
            current_block = w3.eth.block_number
            if current_block > last_block:
                print(f"📦 Bloque {current_block} - Escaneando...", flush=True)
                block = w3.eth.get_block(current_block, full_transactions=True)
                
                for tx in block.transactions:
                    if tx.to and tx.to.lower() == FOUR_MEME_ROUTER.lower():
                        input_data = tx.input.hex()
                        method_id = "0x" + input_data[:8] if len(input_data) >= 8 else "0x00000000"
                        
                        if method_id == FIRMA_CREACION:
                            print(f"   ⚠️ FIRMA {FIRMA_CREACION} DETECTADA.", flush=True)
                            receipt = w3.eth.get_transaction_receipt(tx.hash)
                            for log in receipt['logs']:
                                if len(log['topics']) > 0:
                                    potential_token = log['address']
                                    if potential_token.lower() != FOUR_MEME_ROUTER.lower():
                                        if fire_strike_direct(potential_token):
                                            break 
                last_block = current_block
            time.sleep(2)
        except Exception:
            time.sleep(3)
            w3 = conectar_nodo()

if __name__ == "__main__":
    notify("🧨 *ASTRALIX V21 ONLINE*\nModo Auto-Sell en 15s Activado.")
    scan_blocks()