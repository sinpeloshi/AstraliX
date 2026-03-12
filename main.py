import os
import time
from web3 import Web3
# 👇 PARCHE VITAL: En las versiones nuevas de Web3 se llama así
from web3.middleware import ExtraDataToPOAMiddleware 
from dotenv import load_dotenv
from sqlalchemy import create_engine, Column, String, Integer
from sqlalchemy.orm import declarative_base, sessionmaker

# Cargamos variables de entorno
load_dotenv()

# ==========================================
# 🛠️ CONFIGURACIÓN DE BASE DE DATOS
# ==========================================
DATABASE_URL = os.getenv("DATABASE_URL")

# Ajuste de protocolo para SQLAlchemy y fallback a SQLite
if DATABASE_URL is None:
    print("⚠️ DATABASE_URL no detectada. Usando SQLite local de emergencia...")
    DATABASE_URL = "sqlite:///astralix_radar.db"
elif DATABASE_URL.startswith("postgres://"):
    DATABASE_URL = DATABASE_URL.replace("postgres://", "postgresql://", 1)

engine = create_engine(DATABASE_URL)
Base = declarative_base()
Session = sessionmaker(bind=engine)
session = Session()

# Definición de la tabla (Se crea sola al arrancar)
class WalletDescubierta(Base):
    __tablename__ = 'wallets_detectadas'
    address = Column(String, primary_key=True)
    interacciones_pancakeswap = Column(Integer, default=1)
    ultimo_visto = Column(Integer)

# Crear la tabla si no existe
Base.metadata.create_all(engine)

# ==========================================
# 📡 CONEXIÓN A BINANCE SMART CHAIN (BSC)
# ==========================================
BSC_RPC_URL = "https://bsc-dataseed.binance.org/"
w3 = Web3(Web3.HTTPProvider(BSC_RPC_URL))

# 👇 INYECCIÓN DEL MIDDLEWARE PARA REDES POA (Como BSC)
w3.middleware_onion.inject(ExtraDataToPOAMiddleware, layer=0)

# Direcciones oficiales de PancakeSwap Routers
PANCAKESWAP_ROUTERS = [
    "0x10ED43C718714eb63d5aA57B78B54704E256024E".lower(), # V2 Router
    "0x13f4EA83D0bd40E75C8222255bc855a974568Dd4".lower()  # V3 Router
]

def procesar_bloque(numero_bloque):
    try:
        # Obtenemos el bloque con todas sus transacciones
        bloque = w3.eth.get_block(numero_bloque, full_transactions=True)
        print(f"🔍 Escaneando Bloque {numero_bloque} | Transacciones: {len(bloque.transactions)}")
        
        nuevas_wallets = 0
        
        for tx in bloque.transactions:
            # Filtramos transacciones dirigidas a PancakeSwap
            if tx['to'] and tx['to'].lower() in PANCAKESWAP_ROUTERS:
                wallet_trader = tx['from'].lower()
                
                # Buscamos si ya la conocemos
                wallet_db = session.query(WalletDescubierta).filter_by(address=wallet_trader).first()
                
                if wallet_db:
                    wallet_db.interacciones_pancakeswap += 1
                    wallet_db.ultimo_visto = int(time.time())
                else:
                    # Si es nueva, al sobre para la base de datos
                    nueva_wallet = WalletDescubierta(
                        address=wallet_trader,
                        interacciones_pancakeswap=1,
                        ultimo_visto=int(time.time())
                    )
                    session.add(nueva_wallet)
                    nuevas_wallets += 1
        
        session.commit()
        if nuevas_wallets > 0:
            print(f"🎯 AstraliX atrapó {nuevas_wallets} nuevas wallets operando.")

    except Exception as e:
        print(f"❌ Error procesando bloque {numero_bloque}: {e}")

def iniciar_radar():
    if not w3.is_connected():
        print("❌ Falló la conexión al nodo de BSC. Revisá el RPC.")
        return

    print("📡 AstraliX Radar ONLINE. Escuchando PancakeSwap...")
    
    ultimo_bloque_procesado = w3.eth.block_number

    while True:
        try:
            bloque_actual = w3.eth.block_number
            
            if bloque_actual > ultimo_bloque_procesado:
                for n_bloque in range(ultimo_bloque_procesado + 1, bloque_actual + 1):
                    procesar_bloque(n_bloque)
                ultimo_bloque_procesado = bloque_actual
            
            # Pausa de 3 segundos para no saturar el nodo gratuito
            time.sleep(3)
        except Exception as e:
            print(f"⚠️ Error en el bucle principal: {e}")
            time.sleep(5)

if __name__ == "__main__":
    iniciar_radar()