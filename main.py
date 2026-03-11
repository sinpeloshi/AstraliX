import time
import requests
from web3 import Web3
from web3.middleware import ExtraDataToPOAMiddleware 

RPC_NODES = ["https://bsc-dataseed.binance.org/", "https://bsc-dataseed1.defibit.io/", "https://bsc-dataseed1.ninicoin.io/"]

def conectar_nodo():
    for rpc in RPC_NODES:
        try:
            w3 = Web3(Web3.HTTPProvider(rpc, request_kwargs={'timeout': 15}))
            w3.middleware_onion.inject(ExtraDataToPOAMiddleware, layer=0)
            if w3.is_connected():
                print(f"✅ Conectado a: {rpc}", flush=True)
                return w3
        except: pass
    return None

w3 = conectar_nodo()
GRAFUN_ROUTER = w3.to_checksum_address("0x63395669b9213ef3A1343750529d3851538356E2")
CREATE_METHOD_ID = "0x1f748108" 

def scan_blocks():
    global w3
    print("☢️ AstraliX V16.3: MODO RAYOS X ACTIVADO (Diagnóstico Profundo)", flush=True)
    last_block = w3.eth.block_number
    
    while True:
        try:
            current_block = w3.eth.block_number
            if current_block > last_block:
                block = w3.eth.get_block(current_block, full_transactions=True)
                print(f"\n📦 Bloque {current_block} ({len(block.transactions)} TXs)", flush=True)
                
                grafun_hits = 0
                for tx in block.transactions:
                    # Contamos si ALGUIEN toca GraFun (compra, venta, lo que sea)
                    if tx.to and tx.to.lower() == GRAFUN_ROUTER.lower():
                        grafun_hits += 1
                        
                        # Si encima es una creación de token, gritamos fuerte
                        if tx.input.hex().startswith(CREATE_METHOD_ID):
                            print(f"   🚨 ¡BINGO! CREACIÓN DETECTADA EN GRAFUN (TX: {w3.to_hex(tx.hash)})", flush=True)

                if grafun_hits > 0:
                    print(f"   📊 Actividad en GraFun: {grafun_hits} transacciones en este bloque.", flush=True)
                else:
                    print(f"   👻 GraFun está muerto en este bloque.", flush=True)

                last_block = current_block
            time.sleep(2)
        except Exception as e:
            time.sleep(3)
            w3 = conectar_nodo()

if __name__ == "__main__":
    scan_blocks()