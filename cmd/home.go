package main

import (
	"fmt"
	"net/http"
)

func init() {
	// Registramos la Landing Page en la raíz
	http.HandleFunc("/", HomeHandler)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// IMPORTANTE: Evitamos que otras rutas caigan acá
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
    <title>AstraliX | Layer 1 Infrastructure</title>
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Outfit:wght@400;700;900&display=swap');
        :root { --primary: #0D6EFD; --dark: #0F172A; --accent: #25D366; }
        body { font-family: 'Outfit', sans-serif; margin: 0; color: var(--dark); background: #F8FAFC; line-height: 1.6; }
        
        .nav { padding: 25px; display: flex; justify-content: space-between; align-items: center; max-width: 1200px; margin: 0 auto; }
        .logo { font-weight: 900; font-size: 1.8rem; letter-spacing: -1.5px; color: var(--dark); text-decoration: none; }
        
        /* BOTÓN PARA IR A LA TESTNET */
        .btn-app { text-decoration: none; color: white; background: var(--dark); padding: 12px 24px; border-radius: 14px; font-weight: 700; font-size: 0.85rem; transition: 0.3s; box-shadow: 0 4px 12px rgba(0,0,0,0.1); }
        .btn-app:hover { background: var(--primary); transform: translateY(-2px); }

        .hero { text-align: center; padding: 80px 20px; background: white; border-bottom: 1px solid #E2E8F0; }
        .hero h1 { font-size: 3.5rem; font-weight: 900; margin: 20px 0; letter-spacing: -2px; line-height: 1; }
        .hero p { font-size: 1.2rem; color: #64748B; max-width: 600px; margin: 0 auto 40px; }
        
        .btn-buy { background: var(--primary); color: white; padding: 20px 40px; border-radius: 20px; text-decoration: none; font-weight: 700; font-size: 1.1rem; box-shadow: 0 10px 25px rgba(13, 110, 253, 0.25); display: inline-block; }

        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 30px; max-width: 1200px; margin: 60px auto; padding: 0 20px; }
        .card { background: white; padding: 40px; border-radius: 35px; border: 1px solid #F1F5F9; text-align: center; }
        .card i { font-size: 2.5rem; color: var(--primary); margin-bottom: 20px; }

        .pre-sale { background: var(--dark); color: white; padding: 80px 20px; text-align: center; border-radius: 50px; max-width: 900px; margin: 40px auto; box-sizing: border-box; }
        .price { font-size: 4.5rem; font-weight: 900; margin: 15px 0; }
        .btn-wa { background: var(--accent); color: white; padding: 20px 40px; border-radius: 20px; text-decoration: none; font-weight: 700; display: inline-block; margin-top: 20px; }
        
        footer { text-align: center; padding: 40px; color: #94A3B8; font-size: 0.8rem; }
    </style>
</head>
<body>

    <nav class="nav">
        <a href="/" class="logo">AstraliX</a>
        <a href="/dashboard" class="btn-app">INGRESAR AL CORE <i class="fas fa-rocket" style="margin-left:8px;"></i></a>
    </nav>

    <header class="hero">
        <h1 id="title">Infraestructura Blockchain<br>de 512 bits.</h1>
        <p>Desde el Chaco para el mundo. AstraliX es la red de nueva generación para aplicaciones descentralizadas de alta seguridad.</p>
        <a href="#comprar" class="btn-buy">Adquirir Nodo Fundador</a>
    </header>

    <main class="grid">
        <div class="card">
            <i class="fas fa-server"></i>
            <h3>Nodo de Validación</h3>
            <p>Participá activamente en la seguridad de la red y recibí recompensas por cada bloque verificado.</p>
        </div>
        <div class="card">
            <i class="fas fa-shield-check"></i>
            <h3>Seguridad Máxima</h3>
            <p>Protocolo blindado con derivación de llaves de 512 bits, duplicando el estándar actual.</p>
        </div>
        <div class="card">
            <i class="fas fa-globe-americas"></i>
            <h3>Origen Local</h3>
            <p>Desarrollado íntegramente en Argentina, con visión global y escalabilidad total.</p>
        </div>
    </main>

    <section id="comprar" class="pre-sale">
        <span style="font-weight:800; opacity:0.5; letter-spacing:2px;">OFERTA DE LANZAMIENTO</span>
        <div class="price">21 USDT</div>
        <p>Asegurá tu lugar como Nodo Fundador y recibí 10.000 AX de bienvenida.</p>
        
        <div style="background: rgba(255,255,255,0.05); padding: 20px; border-radius: 20px; border: 1px dashed rgba(255,255,255,0.2); margin: 30px 0;">
            <div style="font-family:monospace; font-size:0.85rem; color:#CBD5E1;">TU_DIRECCION_USDT_BEP20_AQUI</div>
        </div>

        <a href="https://wa.me/TU_WHATSAPP" class="btn-wa">
            <i class="fab fa-whatsapp"></i> ENVIAR COMPROBANTE
        </a>
    </section>

    <footer>
        AstraliX Core Engine • La Tigra, Chaco
    </footer>

</body>
</html>
`