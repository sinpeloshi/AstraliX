import asyncio
import json
import time
import requests
from web3 import Web3
from websockets import connect

# --- 🛰️ CONEXIÓN WSS (CAMBIADO A ANKR PARA SALTAR EL BLOQUEO) ---
WSS_URL = "wss://rpc.ankr.com/bsc/ws/"
HTTP_URL = "https://rpc.ankr.com/bsc"
w3 = Web3(Web3.HTTPProvider(HTTP_URL))

# --- 🧨 CONFIGURACIÓN KAMIKAZE ---
CAPITAL_SNIPER = 0.005 # BNB a invertir
GAS_MULTIPLIER = 5.0   # 500% de gas para atropellar a los demás bots
TARGET_PROFIT = 1.15   # 15% de ganancia (Hit & Run)

# --- 🎯 OBJETIVO FIJO (Four.meme) ---
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

# --- 📜 ABIs Esenciales ---
ABI_ERC20 = '[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"},{"inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"name":"approve","outputs":[{"name":"bool","name":"","type":"bool"}],"type":"function"}]'
ABI_ROUTER = '[{"inputs":[{"name":"amountIn","type":"uint256"},{"name":"amountOutMin","type":"uint256"},{"name":"path","type":"address[]"},{"name":"to","type":"address"},{"name":"deadline","type":"uint256"}],"name":"swapExactTokensForTokensSupportingFeeOnTransferTokens","outputs":[],"type":"function"},{"inputs":[{"name":"amountIn","type":"uint256"},{"name":"path","type":"address[]"}],"name":"getAmountsOut","outputs":[{"name":"amounts","type":"uint256[]"}],"type":"function"}]'
ABI_APEX = '[{"inputs":[{"name":"targets","type":"address[]"},{"name":"payloads","type":"bytes[]"},{"name":"values","type":"uint256[]"},{"name":"minerBribe","type":"uint256"}],"name":"apexStrike","outputs":[],"type":"function"}]'

def notify(msg):
    try: requests.post(f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage", json={"chat_id": TG_ID, "text": msg, "parse_mode": "Markdown"}, timeout=2)
    except: pass

def get_current_value(token_addr, amount_in):
    try:
        router = w3.eth.contract(address=PANCAKE_ROUTER, abi=ABI_ROUTER)
        return router.functions.getAmountsOut(amount_in, [token_addr, WBNB_ADDR]).call()[1]
    except: return 0

async def execute_sell(token_addr):
    print(f"💥 ¡PROFIT DETECTADO! SALIENDO DEL TOKEN...")
    try:
        meme_c = w3.eth.contract(address=token_addr, abi=ABI_ERC20)
        router_c = w3.eth.contract(address=PANCAKE_ROUTER, abi=ABI_ROUTER)
        apex_c = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_APEX)
        bal = meme_c.functions.balanceOf(CONTRATO_ADDR).call()
        
        if bal > 0:
            p_app = meme_c.encode_abi("approve", args=[PANCAKE_ROUTER, bal])
            p_swp = router_c.encode_abi("swapExactTokensForTokensSupportingFeeOnTransferTokens", args=[bal, 0, [token_addr, WBNB_ADDR], CONTRATO_ADDR, int(time.time()) + 120])
            tx = apex_c.functions.apexStrike([token_addr, PANCAKE_ROUTER], [w3.to_bytes(hexstr=p_app), w3.to_bytes(hexstr=p_swp)], [0, 0], 0).build_transaction({
                'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA), 'gas': 600000, 'gasPrice': int(w3.eth.gas_price * 2.0)
            })
            w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx, PRIV_KEY).raw_transaction)
            notify("💰 *¡BOOM! VENTA EJECUTADA (15% PROFIT)*")
            return True
    except Exception as e: print(f"❌ Error Venta: {e}")
    return False

async def monitor_and_sell(token_addr, monto_invertido):
    notify(f"☢️ *MODO CAZADOR:* Vigilando profit de `{token_addr}`")
    meme_c = w3.eth.contract(address=token_addr, abi=ABI_ERC20)
    start_t = time.time()
    
    while time.time() - start_t < 300: # 5 minutos de asedio
        try:
            bal = meme_c.functions.balanceOf(CONTRATO_ADDR).call()
            if bal > 0:
                val = get_current_value(token_addr, bal)
                if val >= int(monto_invertido * TARGET_PROFIT):
                    if await execute_sell(token_addr): break
        except: pass
        await asyncio.sleep(0.5)

async def fire_strike_and_monitor(tx_input):
    token_to_buy = "0x" + tx_input[34:74] if len(tx_input) > 74 else "0xTOKEN_NUEVO"
    print(f"🧨 ¡OBJETIVO FIJADO! LANZANDO ATAQUE SIN FRENOS A: {token_to_buy}")
    notify(f"🚀 *¡GATILLO ACCIONADO!* Entrando al nuevo token de Four.meme...")

    try:
        apex_c = w3.eth.contract(address=CONTRATO_ADDR, abi=ABI_APEX)
        monto = w3.to_wei(CAPITAL_SNIPER, 'ether')
        
        tx = apex_c.functions.apexStrike([WBNB_ADDR], [b''], [0], 0).build_transaction({
            'from': MI_BILLETERA, 'nonce': w3.eth.get_transaction_count(MI_BILLETERA),
            'gas': 700000, 'gasPrice': int(w3.eth.gas_price * GAS_MULTIPLIER)
        })
        
        w3.eth.send_raw_transaction(w3.eth.account.sign_transaction(tx, PRIV_KEY).raw_transaction)
        print("✅ Transacción inyectada en la red.")
        
        await monitor_and_sell(token_to_buy, monto)
        
    except Exception as e: 
        print(f"❌ Impacto fallido: {e}")
        notify("❌ *FALLO EN EL IMPACTO.* Gas insuficiente o red saturada.")

async def listen_mempool():
    async with connect(WSS_URL) as ws:
        await ws.send(json.dumps({"jsonrpc": "2.0", "id": 1, "method": "eth_subscribe", "params": ["pendingTransactions"]}))
        print("☢️ AstraliX V11.1 Kamikaze: Conexión enviada. Esperando al nodo...")
        
        # 🚨 LECTURA DEL PRIMER MENSAJE PARA VER SI NOS BLOQUEAN
        primer_mensaje = await ws.recv()
        print(f"📡 RESPUESTA DEL NODO: {primer_mensaje}")
        
        if "error" in primer_mensaje.lower():
            print("❌ EL NODO ESTÁ BLOQUEANDO EL MEMPOOL. Deteniendo bot.")
            return
            
        print("✅ ¡Suscripción aceptada! Arranca el escaneo...")
        contador_tx = 0
        tiempo_inicio = time.time()
        
        while True:
            try:
                msg = await ws.recv()
                
                # Latido
                contador_tx += 1
                if time.time() - tiempo_inicio >= 5:
                    print(f"⏱️ Pulso de red: {contador_tx} transacciones escaneadas en 5 segundos.")
                    contador_tx = 0
                    tiempo_inicio = time.time()

                tx_hash = json.loads(msg)['params']['result']
                
                try:
                    tx = w3.eth.get_transaction(tx_hash)
                    if tx and tx.get('to') and tx['to'].lower() == FOUR_MEME_MANAGER.lower():
                        if tx.get('input', '').startswith(CREATE_METHOD_ID):
                            print("\n🚨 ¡ALERTA ROJA! ¡CREACIÓN EN MEMPOOL DETECTADA!")
                            await fire_strike_and_monitor(tx['input'])
                except:
                    continue
                    
            except Exception:
                continue

if __name__ == "__main__":
    print("🧨 INICIANDO SECUENCIA KAMIKAZE V11.1...")
    if w3.is_connected():
        notify("☢️ *ASTRALIX V11.1 KAMIKAZE ONLINE*\nConectado al Mempool. Reporte de pulso activado.")
        asyncio.run(listen_mempool())