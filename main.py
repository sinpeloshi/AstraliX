import time
import requests
from web3 import Web3
from web3.middleware import geth_poa_middleware

# --- 🛰️ CONEXIÓN HTTP (ESTABLE) ---
HTTP_URL = "https://solemn-orbital-thunder.bsc.quiknode.pro/70d0d80f07303278accd2349e2fc01c95018d18c/"
w3 = Web3(Web3.HTTPProvider(HTTP_URL, request_kwargs={'timeout': 20}))

# Inyectamos el parche PoA para que no crashee con bloques de BSC
w3.middleware_onion.inject(geth_poa_middleware, layer=0) 

# --- 🧨 CONFIGURACIÓN DE COMBATE ---
CAPITAL_SNIPER = 0.005 # Tu saldo para comprar
GAS_MULTIPLIER = 3.0   # Gas alto para entrar en el bloque siguiente
TARGET_PROFIT = 1.15   # 15% de ganancia

# --- 🎯 OBJETIVO (Four.meme) ---
FOUR_MEME_MANAGER = w3.to_checksum_address("0x5c952063c7fc8610ffdb798152d69f0b9550762b")
CREATE_METHOD_ID = "0xedf9e251" 

# --- 🔑 IDENTIDAD Y CONTRATOS ---
PRIV_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"
CONTRATO_ADDR = w3.to_checksum_address("0xF44f4D75Efc8d60d9383319D1C69553A1201be28")
MI_BILLETERA = w3.eth.account.from_key(PRIV_KEY).address 

TG_TOKEN = '8783847744:AAHdwwlEqP7HCgSXoFxRdD8snr5FRhT1OUo'
TG_ID = '6580309816'

WBNB_ADDR = w3.to_checksum_address("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")
PANCAKE_ROUTER = w3.to_checksum_address("0x10ED43C718714eb63d5aA57B78B54704E256024E")

# --- 📜 ABIs ---
ABI_ERC20 = '[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"},{"inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"name":"approve","outputs":[{"name":"bool","name":"","type":"bool"}],"type":"function"}]'
ABI_ROUTER = '[{"inputs":[{"name":"amountIn","type":"uint256"},{"name":"amountOutMin","type":"uint256"},{"name":"path","type":"address[]"},{"name":"to","type":"address"},{"name":"deadline","type":"uint256"}],"name":"swapExactTokensForTokensSupportingFeeOnTransferTokens","outputs":[],"type":"function"},{"inputs":[{"name":"amountIn","type":"uint256"},{"name":"path","type":"address[]"}],"name":"getAmountsOut","outputs":[{"name":"amounts","type":"uint256[]"}],"type":"function"}]'
ABI_APEX = '[{"inputs":[{"name":"targets","type":"address[]"},{"name":"payloads","type":"bytes[]"},{"name":"values","type":"uint256[]"},{"name":"minerBribe","type":"uint256"}],"name":"apexStrike","outputs":[],"type":"function"}]'

def notify(msg):
    try: requests.post(f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage", json={"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}, timeout=5)
    except: pass

def get_val(token, amount):
    try:
        r = w3.eth.contract(address=PANCAKE_ROUTER, abi=ABI_ROUTER)
        return r.functions.getAmountsOut(amount, [token, WBNB_ADDR]).call()[1]
    except: return 0

def execute_sell(token):
    print(f"💰 PROFIT ALCANZADO. Vendiendo {token}...", flush=True)
    try:
        meme_c = w3.eth.contract(address=token, abi=ABI_ERC20)
        router_c = w3.eth.contract(address=PANCAKE_ROUTER, abi=ABI_ROUTER)
        apex_c = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_APEX)
        bal = meme_c.functions.balanceOf(CONTRATO_ADDR).call()
        if bal > 0:
            p_app = meme_c.encode_abi("approve", args=[PANCAKE_ROUTER, bal])
            p_swp = router_c.encode_abi("swapExactTokensForTokensSupportingFeeOnTransferTokens", args=[bal, 0, [token, WBNB_ADDR], CONTRATO_ADDR, int(time.time()) + 120])
            tx = apex_c.functions.apexStrike([token, PANCAKE_ROUTER], [w3.to_bytes(hexstr=p_app), w3.to_bytes(hexstr=p_swp)], [0, 0], 0).build_transaction({
                'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA), 'gas': 600000, 'gasPrice': int(w3.eth.gas_price * 2.0)
            })
            w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx, PRIV_KEY).raw_transaction)
            notify("✅ *VENTA EXITOSA*")
            return True
    except Exception as e: print(f"❌ Error Venta: {e}", flush=True)
    return False

def monitor_profit(token, invertido):
    print(f"👀 Monitoreando {token}...", flush=True)
    meme_c = w3.eth.contract(address=token, abi=ABI_ERC20)
    start = time.time()
    while time.time() - start < 600:
        try:
            bal = meme_c.functions.balanceOf(CONTRATO_ADDR).call()
            if bal > 0:
                actual = get_val(token, bal)
                if actual >= int(invertido * TARGET_PROFIT):
                    if execute_sell(token): break
        except: pass
        time.sleep(2)

def fire_strike(tx_input):
    token_to_buy = "0x" + tx_input[34:74] if len(tx_input) > 74 else None
    if not token_to_buy: return
    token_to_buy = w3.to_checksum_address(token_to_buy)
    
    print(f"🚀 ATACANDO NUEVO TOKEN: {token_to_buy}", flush=True)
    notify(f"🎯 *OBJETIVO DETECTADO:* `{token_to_buy}`\nLanzando compra...")
    
    try:
        wbnb_c = w3.eth.contract(address=WBNB_ADDR, abi=ABI_ERC20)
        router_c = w3.eth.contract(address=PANCAKE_ROUTER, abi=ABI_ROUTER)
        apex_c = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_APEX)
        monto = w3.to_wei(CAPITAL_SNIPER, 'ether')

        p_app = wbnb_c.encode_abi("approve", args=[PANCAKE_ROUTER, monto])
        p_swp = router_c.encode_abi("swapExactTokensForTokensSupportingFeeOnTransferTokens", args=[monto, 0, [WBNB_ADDR, token_to_buy], CONTRATO_ADDR, int(time.time()) + 120])
        
        tx = apex_c.functions.apexStrike([WBNB_ADDR, PANCAKE_ROUTER], [w3.to_bytes(hexstr=p_app), w3.to_bytes(hexstr=p_swp)], [0, 0], 0).build_transaction({
            'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA), 'gas': 700000, 'gasPrice': int(w3.eth.gas_price * GAS_MULTIPLIER)
        })
        
        tx_hash = w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx, PRIV_KEY).raw_transaction)
        print(f"✅ Compra enviada. Hash: {w3.to_hex(tx_hash)}", flush=True)
        monitor_profit(token_to_buy, monto)
        
    except Exception as e:
        print(f"❌ Error en Ataque: {e}", flush=True)
        notify(f"❌ *FALLO EN COMPRA:* {str(e)}")

def main_loop():
    print("☢️ AstraliX V12.4: Escaneando bloques...", flush=True)
    last_block = w3.eth.block_number
    
    while True:
        try:
            current_block = w3.eth.block_number
            if current_block > last_block:
                # Evita que el bot se atrase procesando bloques viejos
                if current_block - last_block > 3: last_block = current_block - 1
                
                print(f"🔎 Bloque {current_block}", flush=True)
                block = w3.eth.get_block(current_block, full_transactions=True)
                
                for tx in block.transactions:
                    if tx.to and tx.to.lower() == FOUR_MEME_MANAGER.lower():
                        if tx.input.startswith(CREATE_METHOD_ID):
                            print("\n🚨 ¡CREACIÓN ENCONTRADA!", flush=True)
                            fire_strike(tx.input)
                
                last_block = current_block
            
            time.sleep(1.5) # Ritmo estable para no saturar QuickNode
            
        except Exception as e:
            print(f"⚠️ Reintentando... ({e})", flush=True)
            time.sleep(3)

if __name__ == "__main__":
    notify("🛰️ *ASTRALIX V12.4 ONLINE*")
    main_loop()
