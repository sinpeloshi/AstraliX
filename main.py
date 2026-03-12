import os
import time
from web3 import Web3
from web3.middleware import ExtraDataToPOAMiddleware # <-- MAGIA ACTUALIZADA PARA WEB3 V6+
from dotenv import load_dotenv
from sqlalchemy import create_engine, Column, String, Integer
from sqlalchemy.orm import declarative_base, sessionmaker

load_dotenv()

# ==========================================
# 🛠️ MOTOR DE BASE DE DATOS
# ==========================================
DATABASE_URL = os.getenv("DATABASE_URL")

if DATABASE_URL is None:
    print("⚠️ DATABASE_URL no encontrada. Usando base de datos SQLite de emergencia...")
    DATABASE_URL = "sqlite:///astralix_radar.db"
elif DATABASE_URL.startswith("postgres://"):
    DATABASE_URL = DATABASE_URL.replace("postgres://", "postgresql://", 1)

engine = create_engine(DATABASE_URL)
Base = declarative_base()
Session = sessionmaker(bind=engine)
session = Session()

# Tabla para guardar las Wallets descubiertas
class WalletDescubierta(Base):
    __tablename__ = 'wallets_detectadas'
    
    address = Column(String, primary_key=True)
    interacciones_pancakeswap = Column(Integer, default=1)
    ultimo_visto = Column(Integer)

Base.metadata.create_all(engine)

# ==========================================
# 📡 CONEXIÓN A LA BLOCKCHAIN (BSC)
# ==========================================
BSC_RPC_URL = "https://bsc-dataseed.binance.org/"
w3 = Web3(Web3.HTTPProvider(BSC_RPC_URL))

# ¡INYECCIÓN DEL PARCHE PARA BSC (POA) VERSIÓN 6!
w3.middleware_onion.inject(ExtraDataToPOAMiddleware, layer=0)

# Contratos de PancakeSwap (V2 y V3 Routers)
PANCAKESWAP_ROUTERS = [
    "0x10ED43C718714eb63d5aA57B78B54704E256024E".lower(), # V2
    "0x13f4EA83D0bd40E75C8222255bc855a974568Dd4".lower()  # V3
]

def procesar_bloque(numero_bloque):
    try:
        bloque = w3.eth.get_block(numero_bloque, full_transactions=True)
        print(f"🔍 Escaneando Bloque {numero_bloque} | Transacciones: {len(bloque.transactions)}")
        
        nuevas_wallets = 0
        
        for tx in bloque.transactions:
            if tx['to'] and tx['to'].lower() in PANCAKESWAP_ROUTERS:
                wallet_trader = tx['from'].lower()
                
                wallet_db = session.query(WalletDescubierta).filter_by(address=wallet_trader).first()
                
                if wallet_db:
                    wallet_db.interacciones_pancakeswap += 1
                    wallet_db.ultimo_visto = int(time.time())
                else:
                    nueva_wallet = WalletDescubierta(
                        address=wallet_trader,
                        interacciones_pancakeswap=1,
                        ultimo_visto=int(time.time())
                    )
                    session.add(nueva_wallet)
                    nuevas_wallets += 1
        
        session.commit()
        if nuevas_wallets > 0:
            print(f"🎯 ¡Se atraparon {nuevas_wallets} nuevas wallets operando en PancakeSwap!")

    except Exception as e:
        print(f"Error procesando el bloque: {e}")

def iniciar_escucha():
    if not w3.is_connected():
        print("❌ Error de conexión al nodo BSC.")
        return

    print("📡 Conectado a BSC. Escuchando operaciones en PancakeSwap...")
    
    ultimo_bloque_procesado = w3.eth.block_number

    while True:
        bloque_actual = w3.eth.block_number
        
        if bloque_actual > ultimo_bloque_procesado:
            for n_bloque in range(ultimo_bloque_procesado + 1, bloque_actual + 1):
                procesar_bloque(n_bloque)
            ultimo_bloque_procesado = bloque_actual
        
        # Pausa para no saturar el nodo público gratuito
        time.sleep(3)

if __name__ == "__main__":
    iniciar_escucha()