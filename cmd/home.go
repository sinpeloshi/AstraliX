package main

import (
	"fmt"
	"net/http"
)

// La función init se ejecuta sola al arrancar el programa.
// Registra la landing page en la raíz (/) sin tocar main.go.
func init() {
	http.HandleFunc("/", HomeHandler)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Solo mostramos la landing si es exactamente la raíz
	if r.URL.Path != "/" {
		return 
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, landingHTML)
}

const landingHTML = `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AstraliX | Nodos Fundadores</title>
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Outfit:wght@400;700;900&display=swap');
        :root { --primary: #0D6EFD; --dark: #0F172A; --accent: #25D366; }
        body { font-family: 'Outfit', sans-serif; margin: 0; color: var(--dark); background: #F8FAFC; line-height: 1.6; }
        
        /* Navbar */
        .nav { padding: 25px; display: flex; justify-content: space-between; align-items: center; max-width: 1200px; margin: 0 auto; }
        .logo { font-weight: 900; font-size: 1.8rem; letter-spacing: -1.5px; color: var(--dark); text-decoration: none; }
        .nav-link { text-decoration: none; color: var(--primary); font-weight: 700; font-size: 0.9rem; background: #E0E7FF; padding: 10px 20px; border-radius: 12px; }

        /* Hero Section */
        .hero { text-align: center; padding: 100px 20px 60px; background: white; border-bottom: 1px solid #E2E8F0; }
        .badge { background: #E0E7FF; color: var(--primary); padding: 8px 16px; border-radius: 100px; font-weight: 700; font-size: 0.75rem; text-transform: uppercase; letter-spacing: 1px; }
        .hero h1 { font-size: 3.8rem; font-weight: 900; margin: 20px 0; letter-spacing: -3px; line-height: 0.9; }
        .hero p { font-size: 1.2rem; color: #64748B; max-width: 650px; margin: 0 auto 40px; }
        .btn-primary { background: var(--primary); color: white; padding: 20px 40px; border-radius: 20px; text-decoration: none; font-weight: 700; font-size: 1.1rem; box-shadow: 0 10px 25px rgba(13, 110, 253, 0.25); display: inline-block; }

        /* Features */
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 30px; max-width: 1200px; margin: 60px auto; padding: 0 20px; }
        .card { background: white; padding: 40px; border-radius: 35px; border: 1px solid #F1F5F9; text-align: center; transition: 0.3s; }
        .card:hover { transform: translateY(-10px); box-shadow: 0 20px 40px rgba(0,0,0,0.03); }
        .card i { font-size: 2.5rem; color: var(--primary); margin-bottom: 20px; }
        .card h3 { font-weight: 800; margin-bottom: 10px; }
        .card p { color: #64748B; font-size: 0.95rem; }

        /* Pre-Sale Box */
        .pre-sale { background: var(--dark); color: white; padding: 80px 20px; text-align: center; border-radius: 50px; max-width: 900px; margin: 40px auto; box-sizing: border-box; }
        .price { font-size: 4.5rem; font-weight: 900; margin: 15px 0; letter-spacing: -2px; }
        .address-container { background: rgba(255,255,255,0.05); padding: 25px; border-radius: 20px; border: 1px dashed rgba(255,255,255,0.2); margin: 30px 0; }
        .wallet-addr { font-family: 'JetBrains Mono', monospace; font-size: 0.85rem; word-break: break-all; color: #CBD5E1; }
        .btn-wa { background: var(--accent); color: white; padding: 20px 40px; border-radius: 20px; text-decoration: none; font-weight: 700; display: inline-block; margin-top: 20px; }

        footer { text-align: center; padding: 40px; color: #94A3B8; font-size: 0.8rem; font-weight: 600; }
    </style>
</head>
<body>

    <nav class="nav">
        <a href="/" class="logo">AstraliX</a>
        <a href="/dashboard" class="nav-link">ACCESO ALPHA CORE <i class="fas fa-external-link-alt"></i></a>
    </nav>

    <header class="hero">
        <span class="badge">Preventa Exclusiva: Nodos Fundadores</span>
        <h1>Invertí en el futuro<br>de la Capa 1.</h1>
        <p>AstraliX es la red de 512 bits diseñada para la máxima seguridad. Sé parte del nacimiento de la infraestructura desde el Chaco.</p>
        <a href="#comprar" class="btn-primary">Adquirir Nodo de Validación</a>
    </header>

    <main class="grid">
        <div class="card">
            <i class="fas fa-fingerprint"></i>
            <h3>Identidad 512-bit</h3>
            <p>Utilizamos las frases semilla más seguras del mercado para proteger tu capital.</p>
        </div>
        <div class="card">
            <i class="fas fa-gift"></i>
            <h3>Airdrop VIP</h3>
            <p>Por cada nodo adquirido, recibís un bono de 10.000 AX directo en tu billetera.</p>
        </div>
        <div class="card">
            <i class="fas fa-exchange-alt"></i>
            <h3>Reconocimiento 1:1</h3>
            <p>Tus activos minados en la Testnet serán migrados automáticamente a la Mainnet final.</p>
        </div>
    </main>

    <section id="comprar" class="pre-sale">
        <span style="text-transform:uppercase; font-weight:800; opacity:0.5; letter-spacing:2px;">Costo de Inscripción</span>
        <div class="price">21 USDT</div>
        <p style="max-width:500px; margin: 0 auto; opacity:0.8;">Para activar tu nodo, enviá el pago (Red BEP-20) a la dirección oficial:</p>
        
        <div class="address-container">
            <div class="wallet-addr">0xTU_DIRECCION_DE_BINANCE_AQUÍ</div>
        </div>

        <p style="font-size:0.9rem; opacity:0.6;">Una vez enviado, envianos el comprobante por WhatsApp para darte de alta en la red.</p>
        
        <a href="https://wa.me/5493735XXXXXX" class="btn-wa">
            <i class="fab fa-whatsapp"></i> ENVIAR COMPROBANTE
        </a>
    </section>

    <footer>
        &copy; 2026 AstraliX Core Engine • Chaco, Argentina
    </footer>

</body>
</html>
`