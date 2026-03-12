import os
import time
import requests
from dotenv import load_dotenv
from sqlalchemy import create_engine, Column, String, Float, Integer, Boolean
from sqlalchemy.orm import declarative_base, sessionmaker

# Cargar variables de entorno
load_dotenv()

# Configuración de Base de Datos
DATABASE_URL = os.getenv("DATABASE_URL")
if DATABASE_URL and DATABASE_URL.startswith("postgres://"):
    DATABASE_URL = DATABASE_URL.replace("postgres://", "postgresql://", 1)

engine = create_engine(DATABASE_URL)
Base = declarative_base()
Session = sessionmaker(bind=engine)
session = Session()

# API Key de BscScan
BSCSCAN_API_KEY = os.getenv("BSCSCAN_API_KEY")

# Modelo de la tabla en PostgreSQL
class Transaccion(Base):
    __tablename__ = 'transacciones_smart_wallets'
    
    hash_tx = Column(String, primary_key=True)
    wallet = Column(String)
    token_symbol = Column(String)
    token_address = Column(String)
    cantidad = Column(Float)
    tipo = Column(String) # COMPRA o VENTA
    timestamp = Column(Integer)

# Crear la tabla si no existe
Base.metadata.create_all(engine)

# Wallets que queremos rastrear (Tus "Smart Wallets" de prueba)
TARGET_WALLETS = [
    "0x...PegaAcaLaWallet1...", 
    "0x...PegaAcaLaWallet2..."
]

def obtener_transacciones(wallet):
    url = f"https://api.bscscan.com/api?module=account&action=tokentx&address={wallet}&page=1&offset=50&sort=desc&apikey={BSCSCAN_API_KEY}"
    try:
        res = requests.get(url).json()
        if res['status'] == '1':
            return res['result']
    except Exception as e:
        print(f"Error conectando a BscScan: {e}")
    return []

def procesar_y_guardar(wallet):
    print(f"Buscando datos para {wallet}...")
    txs = obtener_transacciones(wallet)
    
    nuevas_txs = 0
    for tx in txs:
        # Verificar si la transacción ya existe en la base de datos
        existe = session.query(Transaccion).filter_by(hash_tx=tx['hash']).first()
        
        if not existe:
            # Calcular cantidad real usando los decimales del token
            cantidad_real = float(tx['value']) / (10 ** int(tx['tokenDecimal']))
            
            # Determinar si es compra o venta
            tipo_tx = "COMPRA" if tx['to'].lower() == wallet.lower() else "VENTA"
            
            nueva_tx = Transaccion(
                hash_tx=tx['hash'],
                wallet=wallet,
                token_symbol=tx['tokenSymbol'],
                token_address=tx['contractAddress'],
                cantidad=cantidad_real,
                tipo=tipo_tx,
                timestamp=int(tx['timeStamp'])
            )
            session.add(nueva_tx)
            nuevas_txs += 1
            print(f"NUEVA {tipo_tx}: {cantidad_real:.2f} {tx['tokenSymbol']}")
            
    if nuevas_txs > 0:
        session.commit()
        print(f"✅ Se guardaron {nuevas_txs} transacciones nuevas en la base de datos.")
    else:
        print("Sin movimientos nuevos.")

def iniciar_radar():
    print("🚀 Iniciando Radar de Smart Wallets en BSC...")
    while True:
        for wallet in TARGET_WALLETS:
            # Evitar analizar wallets vacías o mal configuradas
            if len(wallet) == 42 and wallet.startswith("0x"):
                procesar_y_guardar(wallet)
                # Pausa de 5 segundos entre wallets para no saturar la API gratuita
                time.sleep(5) 
            
        print("⏳ Esperando 60 segundos para el próximo escaneo...")
        time.sleep(60)

if __name__ == "__main__":
    if not BSCSCAN_API_KEY or not DATABASE_URL:
        print("❌ Faltan configurar las variables de entorno (DATABASE_URL o BSCSCAN_API_KEY).")
    else:
        iniciar_radar()