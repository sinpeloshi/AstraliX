import os
import time
from web3 import Web3
from web3.middleware import ExtraDataToPOAMiddleware 
from dotenv import load_dotenv
from sqlalchemy import create_engine, Column, String, Integer
from sqlalchemy.orm import declarative_base, sessionmaker

load_dotenv()

# ==========================================
# 🛠️ MOTOR DE BASE DE DATOS (AUTO-CORRECTOR)
# ==========================================
DATABASE_URL = os.getenv("DATABASE_URL")

def limpiar_url(url):
    if not url: return "sqlite:///astralix_radar.db"
    # Forzamos minúsculas y corregimos el protocolo para SQLAlchemy
    url = url.replace("Postgresql://", "postgresql://")
    url = url.replace("postgres://", "postgresql://")
    return url

engine = create_engine(limpiar_url(DATABASE_URL))
Base = declarative_base()
Session = sessionmaker(bind=engine)
session = Session()

class WalletDescubierta(Base):
    __tablename__ = 'wallets_detectadas'
    address = Column(String, primary_key=True)
    interacciones_pancakeswap = Column(Integer, default=1)
    ultimo_visto = Column(Integer)

Base.metadata.create_all(engine)

# ==========================================
# 📡 CONEXIÓN A BSC
# ==========================================
BSC_RPC_URL = "https://bsc-dataseed.binance.org/"
w3 = Web3(Web3.HTTPProvider(BSC_RPC_URL))
w3.middleware_onion.inject(ExtraDataToPOAMiddleware, layer=0)

PANCAKESWAP_ROUTERS = [
    "0x10ED43C718714eb63d5aA57B78B54704E256024E".lower(),
    "0x13f4EA83D0bd40E75C8222255bc855a974568Dd4".lower()
]

def procesar_bloque(numero_bloque):
    try:
        bloque = w3.eth.get_block(numero_bloque, full_transactions=True)
        print(f"🔍 Escaneando Bloque {numero_bloque} | Tx: {len(bloque.transactions)}")
        nuevas = 0
        for tx in bloque.transactions:
            if tx['to'] and tx['to'].lower() in PANCAKESWAP_ROUTERS:
                wallet_trader = tx['from'].lower()
                wallet_db = session.query(WalletDescubierta).filter_by(address=wallet_trader).first()
                if wallet_db:
                    wallet_db.interacciones_pancakeswap += 1
                    wallet_db.ultimo_visto = int(time.time())
                else:
                    nueva_wallet = WalletDescubierta(address=wallet_trader, interacciones_pancakeswap=1, ultimo_visto=int(time.time()))
                    session.add(nueva_wallet)
                    nuevas += 1
        session.commit()
        if nuevas > 0: print(f"🎯 AstraliX atrapó {nuevas} nuevas wallets!")
    except Exception as e:
        print(f"❌ Error bloque {numero_bloque}: {e}")
        session.rollback()

def iniciar_radar():
    if not w3.is_connected():
        print("❌ Error de conexión al nodo BSC.")
        return
    print("📡 AstraliX Radar ONLINE. Escuchando mercado...")
    ultimo_bloque = w3.eth.block_number
    while True:
        try:
            actual = w3.eth.block_number
            if actual > ultimo_bloque:
                for n in range(ultimo_bloque + 1, actual + 1):
                    procesar_bloque(n)
                ultimo_bloque = actual
            time.sleep(3)
        except Exception as e:
            print(f"⚠️ Reintentando... {e}")
            time.sleep(5)

if __name__ == "__main__":
    iniciar_radar()