package main

import (
	"fmt"
	"net/http"
)

// La función init se ejecuta automáticamente al iniciar la aplicación.
// Registra el manejador de la Landing Page en la raíz del sitio.
func init() {
	http.HandleFunc("/", HomeHandler)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// IMPORTANTE: Solo mostramos la landing si la ruta es exactamente "/"
	// Esto evita que "pise" las rutas de la API o el dashboard.
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
    <title>AstraliX | Nodos Fundadores 512-bit</title>
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Outfit:wght@400;700;900&display=swap');
        :root { --primary: #0D6EFD; --dark: #0F172A; --accent: #25D366; --text-light: #64748B; }
        
        * { box-sizing: border-box; }
        body { font-family: 'Outfit', sans-serif; margin: 0; color: var(--dark); background: #F8FAFC; line-height: 1.6; scroll-behavior: smooth; }
        
        /* Navbar */
        .nav { padding: 25px; display: flex; justify-content: space-between; align-items: center; max-width: 1200px; margin: 0 auto; }
        .logo { font-weight: 900; font-size: 2rem; letter-spacing: -1.5px; color: var(--dark); text-decoration: none; }
        .nav-link { text-decoration: none; color: var(--primary); font-weight: 700; font-size: 0.9rem; background: #E0E7FF; padding: 12px 24px; border-radius: 14px; transition: 0.3s; }
        .nav-link:hover { background: var(--primary); color: white; }

        /* Hero Section */
        .hero { text-align: center; padding: 120px 20px 80px; background: white; border-bottom: 1px solid #E2E8F0; }
        .badge { background: #E0E7FF; color: var(--primary); padding: 8px 18px; border-radius: 100px; font-weight: 800; font-size: 0.75rem; text-transform: uppercase; letter-spacing: 1.5px; }
        .hero h1 { font-size: 4.5rem; font-weight: 900; margin: 25px 0; letter-spacing: -4px; line-height: 0.85; color: var(--dark); }
        .hero p { font-size: 1.3rem; color: var(--text-light); max-width: 700px; margin: 0 auto 45px; }
        .btn-primary { background: var(--primary); color: white; padding: 22px 50px; border-radius: 24px; text-decoration: none; font-weight: 700; font-size: 1.2rem; box-shadow: 0 15px 35px rgba(13, 110, 253, 0.3); display: inline-block; transition: 0.3s; }
        .btn-primary:hover { transform: translateY(-3px); box-shadow: 0 20px 40px rgba(13, 110, 253, 0.4); }

        /* Features */
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 40px; max-width: 1200px; margin: 80px auto; padding: 0 20px; }
        .card { background: white; padding: 50px 40px; border-radius: 40px; border: 1px solid #F1F5F9; text-align: center; transition: 0.4s; }
        .card:hover { transform: translateY(-15px); box-shadow: 0 30px 60px rgba(0,0,0,0.04); }
        .card i { font-size: 3rem; color: var(--primary); margin-bottom: 30px; }
        .card h3 { font-weight: 800; font-size: 1.7rem; margin-bottom: 15px; }
        .card p { color: var(--text-light); font-size: 1.05rem; }

        /* Pre-Sale Box */
        .pre-sale { background: var(--dark); color: white; padding: 100px 30px; text-align: center; border-radius: 60px; max-width: 1000px; margin: 60px auto; box-sizing: border-box; }
        .price-label { text-transform: uppercase; font-weight: 800; opacity: 0.5; letter-spacing: 3px; font-size: 0.9rem; }
        .price { font-size: 6rem; font-weight: 900; margin: 10px 0; letter-spacing: -4px; }
        .address-container { background: rgba(255,255,255,0.03); padding: 35px; border-radius: 25px; border: 2px dashed rgba(255,255,255,0.15); margin: 40px 0; }
        .wallet-addr { font-family: 'JetBrains Mono', monospace; font-size: 1rem; word-break: break-all; color: #E2E8F0; line-height: 1.4; }
        .btn-wa { background: var(--accent); color: white; padding: 22px 50px; border-radius: 24px; text-decoration: none; font-weight: 700; font-size: 1.1rem; display: inline-block; transition: 0.3s; box-shadow: 0 15px 35px rgba(37, 211, 102, 0.2); }
        .btn-wa:hover { transform: scale(1.05); background: #20bd5a; }

        footer { text-align: center; padding: 60px 20px; color: #94A3B8; font-size: 0.9rem; font-weight: 600; border-top: 1px solid #E2E8F0; margin-top: 100px; }
        
        @media (max-width: 768px) {
            .hero h1 { font-size: 3rem; }
            .price { font-size: 4rem; }
            .pre-sale { border-radius: 30px; margin: 20px; }
        }
    </style>
</head>
<body>

    <nav class="nav">
        <a href="/" class="logo">AstraliX</a>
        <a href="/dashboard" class="nav-link">DASHBOARD CORE <i class="fas fa-arrow-right" style="margin-left:8px;"></i></a>
    </nav>

    <header class="hero">
        <span class="badge">Nodos Fundadores • 100 Cupos Alpha</span>
        <h1>Seguridad de 512 bits<br>hecha en el Chaco.</h1>
        <p>AstraliX es la infraestructura blockchain de nueva generación. Un motor de Capa 1 diseñado para ser inviolable, veloz y fundacional.</p>
        <a href="#comprar" class="btn-primary">Asegurar mi Nodo Fundador</a>
    </header>

    <main class="grid">
        <div class="card">
            <i class="fas fa-fingerprint"></i>
            <h3>Identidad Mnemónica</h3>
            <p>Utilizamos protocolos de 512 bits para la generación de llaves, estableciendo un nuevo estándar de protección.</p>
        </div>
        <div class="card">
            <i class="fas fa-gem"></i>
            <h3>Beneficios VIP</h3>
            <p>Cada nodo fundador recibe un airdrop de 10.000 AX y acceso prioritario a la minería de red.</p>
        </div>
        <div class="card">
            <i class="fas fa-sync"></i>
            <h3>Respaldo 1:1</h3>
            <p>Todo el progreso y los activos generados en la fase Alpha serán migrados íntegramente a la Mainnet oficial.</p>
        </div>
    </main>

    <section id="comprar" class="pre-sale">
        <span class="price-label">Inscripción Nodo Fundador</span>
        <div class="price">21 USDT</div>
        <p style="max-width:550px; margin: 10px auto 40px; opacity:0.8; font-size: 1.1rem;">Para activar tu identidad en la red, envía el pago (Red Binance Smart Chain BEP-20) a la siguiente dirección oficial:</p>
        
        <div class="address-container">
            <div class="wallet-addr">0x948a663b1bd1292ded76a8412af2092bf0462d7c</div>
        </div>

        <p style="font-size:0.95rem; opacity:0.6; margin-bottom: 30px;">Una vez realizado el envío, presiona el botón de abajo para enviarnos el comprobante por WhatsApp y activar tu cuenta.</p>
        
        <a href="https://wa.me/TuNumeroDeTelefono?text=Hola!%20Envio%20comprobante%20para%20mi%20Nodo%20Fundador%20AstraliX" class="btn-wa">
            <i class="fab fa-whatsapp" style="margin-right:10px;"></i> ENVIAR COMPROBANTE
        </a>
    </section>

    <footer>
        &copy; 2026 AstraliX Core Engine • La Tigra, Chaco, Argentina.
    </footer>

</body>
</html>
`