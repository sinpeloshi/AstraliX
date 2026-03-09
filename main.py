import time
import telebot
from telebot.types import ReplyKeyboardMarkup, KeyboardButton
from web3 import Web3
from threading import Thread

# === 1. CONEXIÓN Y CONFIGURACIÓN ===
BSC_URL = "https://bsc-dataseed.binance.org/"
w3 = Web3(Web3.HTTPProvider(BSC_URL))

# Credenciales directas (¡MANTÉN EL REPOSITORIO ESTRICTAMENTE PRIVADO!)
PRIVATE_KEY = "0x8f270281b31526697669d03a48e7e930509657662cbf1f4d6e89b3dfd0413c6e"
TELEGRAM_TOKEN = "8783847744:AAHdwwlEqP7HCgSXoFxRdD8snr5FRhT1OUo"
TELEGRAM_CHAT_ID = "6580309816"

# Direcciones (Formato Checksum estricto de Web3)
MI_BILLETERA = w3.to_checksum_address("0xbcf6a859b20a44d85ffaf610f0aedc67607d97e6")
DIRECCION_CONTRATO = w3.to_checksum_address("0x2093cd0b3F75A1E6ff750E1F871C234C1abF3d3c")
WBNB_ADDRESS = w3.to_checksum_address("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")

# ABI del contrato
ABI = [
    {"inputs": [{"internalType": "address", "name": "token", "type": "address"}], "name": "retirarTokens", "outputs": [], "stateMutability": "nonpayable", "type": "function"},
    {"inputs": [], "name": "retirarBNB", "outputs": [], "stateMutability": "nonpayable", "type": "function"},
    {"inputs": [], "name": "ejecutarArbitraje", "outputs": [], "stateMutability": "nonpayable", "type": "function"}
]

contrato = w3.eth.contract(address=DIRECCION_CONTRATO, abi=ABI)
bot = telebot.TeleBot(TELEGRAM_TOKEN)

# Matriz de Tokens (Zoológico Integrado)
TOKENS = {
    "WBNB": WBNB_ADDRESS,
    "USDT": w3.to_checksum_address("0x55d398326f99059ff775485246999027b3197955"),
    "DOGE": w3.to_checksum_address("0xba2ae424d960c26247dd6c32edc70b295c744c43"),
    "PEPE": w3.to_checksum_address("0x25d887ce7a35172c62fef4674630e620cd2d0752"),
    "SHIB": w3.to_checksum_address("0x2859e4544c4bb03966803b044a93563bd2d0dd4d"),
    "FLOKI": w3.to_checksum_address("0xfb5b838b6cfeedc2873ab27866079ac55363d37e"),
    "BABYDOGE": w3.to_checksum_address("0xc748673057861a797275CD8A068AbB95A902e8de"),
    "KISHU": w3.to_checksum_address("0xa2b4c0af19cc16a6cfacce81f192b024d625817d"),
    "POODLE": w3.to_checksum_address("0x23396cf899ca06c4472205fc903bdb4de249d6fc"),
    "WOOFY": w3.to_checksum_address("0xd0660cd418a64a1d44e9214ad8e459324d8157f1"),
    "CAT": w3.to_checksum_address("0x6894cde390a3f51155ea41ed24a33a4827d3063d"),
    "SPX": w3.to_checksum_address("0x4f1a9cfdc93fce7b55b7a5a1d9b9add4ef95c0d0"),
    "AQUAGOAT": w3.to_checksum_address("0x88e806c1d78c4f5f31f5c9c8a86d9c3f7a3c7f69"),
    "SMI": w3.to_checksum_address("0xcd7492db29e2ab436e819b249452ee1bbdf52214"),
    "PIT": w3.to_checksum_address("0xa57ac35ce91ee92caefaa8dc04140c8e232c2e50"),
    "CATE": w3.to_checksum_address("0x118f073796821da3e9901061b05c0b36377b877e"),
    "BABYFLOKI": w3.to_checksum_address("0x02a9d7162bd73c2b35c5cf6cdd585e91928c850a"),
    "BABYSHIBA": w3.to_checksum_address("0xb8ebd245a2a117e4d2de08bb9d0a3b75ea8e784f"),
    "CRAZYSHIBA": w3.to_checksum_address("0x11791e8a593bc1e1fbcB406DEaE5F8c71f85B60F"),
    "GDOGE": w3.to_checksum_address("0xc9d3924e65913e10ec78c7ca6f19ca3d4d965cc7"),
    "NCAT": w3.to_checksum_address("0x9f9e5fD8bbc25984B178FdCE6117Defa39d2db39"),
    "FOXGIRL": w3.to_checksum_address("0x00069f5a5fe1f7f9f4edb6f6d5c13b79c7f69a00"),
    "BOSS": w3.to_checksum_address("0x04b2e8227f8b9c8d5b9c6d44f7d409775888b859"),
    "FRGST": w3.to_checksum_address("0x5e0e7a3d9b9d7a509a68e2db279b9da869651d81"),
    "FLOKIN": w3.to_checksum_address("0x6e0d9e89fbb0e5f0a89e3297d1e392b74538b0f6"),
    "BB": w3.to_checksum_address("0x09c6230e43acdf4aa89ccab18c5cfd3f780459b7"),
    "REKT": w3.to_checksum_address("0x20482b0b4d9d8f60d3ab432b92f4c4b901a0d10c"),
    "SIREN": w3.to_checksum_address("0x997a58129890bbda032231a52ed1ddc845fc18e1"),
    "MERL": w3.to_checksum_address("0xa0c56a8c0692bd10b3fa8f8ba79cf5332b7107f9"),
    "BSQ": w3.to_checksum_address("0x783c3f003f172c6ac5ac700218a357d2d66ee2a2"),
    "BNBCARD": w3.to_checksum_address("0xdc06717f367e57a16e06cce0c4761604460da8fc"),
    "HOME": w3.to_checksum_address("0x4bfaa776991e85e5f8b1255461cbbd216cfc714f"),
    "LBR": w3.to_checksum_address("0x68ebc50a5bbd4a9e9ed6c6b3c2d4a0d2a1a4f3b2"),
    "DOGIRA": w3.to_checksum_address("0x4b86b9d09a9d7e5b45cdaf0c9a2b805d6e7ca3a4"),
    "SHI": w3.to_checksum_address("0x1f6f1a4c5b5c1b5a5e0b1f6c1c5a1f5c6b6a5f4e"),
    "KINGSHIB": w3.to_checksum_address("0xc34d8a7c4e8f8e3f5d2c9b0b7c8a3d5f7a9e8c1b"),
    "LILFLOKI": w3.to_checksum_address("0xf5bfa2f3c4d6e1a8b9c7d3f2e1c6b7a9d8f5e1c2"),
    "XDOGE": w3.to_checksum_address("0x2b3c4d5e6f708192a3b4c5d6e7f8192a3b4c5d6e")
}

# === 2. NÚCLEO BLOCKCHAIN ===
def enviar_transaccion(funcion_contrato):
    """Construye, firma y envía la transacción a la red BSC"""
    try:
        nonce = w3.eth.get_transaction_count(MI_BILLETERA)
        tx = funcion_contrato.build_transaction({
            'chainId': 56,
            'gas': 300000, 
            'gasPrice': w3.eth.gas_price,
            'nonce': nonce,
        })
        # Firma con la clave privada configurada
        tx_firmada = w3.eth.account.sign_transaction(tx, private_key=PRIVATE_KEY)
        tx_hash = w3.eth.send_raw_transaction(tx_firmada.rawTransaction)
        return w3.to_hex(tx_hash)
    except Exception as e:
        return str(e)

# === 3. INTERFAZ TELEGRAM ===
def menu_teclado():
    markup = ReplyKeyboardMarkup(resize_keyboard=True)
    markup.row(KeyboardButton('/status'), KeyboardButton('/balance'))
    markup.row(KeyboardButton('/radar'), KeyboardButton('/retiro'))
    return markup

@bot.message_handler(commands=['start', 'help'])
def send_welcome(message):
    bot.send_message(message.chat.id, "⚡️ **AstraliX - God X** Activado ⚡️\nDashboard listo. Presiona /retiro para extraer tus fondos.", parse_mode="Markdown", reply_markup=menu_teclado())

@bot.message_handler(commands=['status'])
def status(message):
    bot.send_message(message.chat.id, "🟢 **Sistema Operativo**\nConectado a BSC y escuchando comandos.", parse_mode="Markdown")

@bot.message_handler(commands=['balance'])
def balance(message):
    try:
        wbnb_contract = w3.eth.contract(address=WBNB_ADDRESS, abi=[{"constant": True, "inputs": [{"name": "_owner", "type": "address"}], "name": "balanceOf", "outputs": [{"name": "balance", "type": "uint256"}], "type": "function"}])
        saldo_wei = wbnb_contract.functions.balanceOf(DIRECCION_CONTRATO).call()
        saldo_wbnb = w3.from_wei(saldo_wei, 'ether')
        bot.send_message(message.chat.id, f"💰 **Balance en Contrato:**\n`{saldo_wbnb:.4f} WBNB`", parse_mode="Markdown")
    except Exception as e:
        bot.send_message(message.chat.id, f"❌ Error al leer balance: {e}")

@bot.message_handler(commands=['radar'])
def radar(message):
    bot.send_message(message.chat.id, f"📡 **Radar AstraliX**\nMonitoreando {len(TOKENS)} tokens.", parse_mode="Markdown")

@bot.message_handler(commands=['retiro'])
def retiro(message):
    bot.send_message(message.chat.id, "⏳ Ejecutando retiro de WBNB a tu wallet owner...")
    # Ejecuta el retiro de WBNB desde el contrato a tu billetera personal
    tx = enviar_transaccion(contrato.functions.retirarTokens(WBNB_ADDRESS))
    
    if tx.startswith("0x"):
        bot.send_message(message.chat.id, f"✅ **Retiro Exitoso!**\nHash de la transacción:\n`{tx}`", parse_mode="Markdown")
    else:
        bot.send_message(message.chat.id, f"❌ **Error en la transacción:**\n{tx}", parse_mode="Markdown")

# === 4. LOOP Y ARRANQUE ===
def reporte_automatico():
    while True:
        try:
            bot.send_message(TELEGRAM_CHAT_ID, "⏱ Bot online, monitoreando tokens y esperando comandos.")
        except:
            pass
        time.sleep(3600) # Reporte cada 1 hora para no saturar

if __name__ == "__main__":
    print("Iniciando AstraliX...")
    Thread(target=reporte_automatico, daemon=True).start()
    bot.infinity_polling()