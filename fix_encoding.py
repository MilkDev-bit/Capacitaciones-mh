import os, re

# Tabla de sustituciones: mojibake -> correcto
FIXES = [
    # doble-encoding comun
    ("Ã±", "ñ"), ("Ã³", "ó"), ("Ã©", "é"), ("Ã¡", "á"), ("Ã­", "í"), ("Ãº", "ú"),
    ("Ã¼", "ü"), ("Ã'", "Ñ"), ("Ã"", "Ó"), ("Ã‰", "É"), ("Ã", "Á"), ("Ã", "Í"),
    # prefijos comunes
    ("Â¡", "¡"), ("Â¿", "¿"),
    # simbolos
    ("â€"", "—"), ("â€"", "–"), ("â€¢", "•"), ("â€˜", "'"), ("â€™", "'"),
    ("â€œ", "\u201c"), ("â€\u009d", "\u201d"),
    ("â€¦", "…"), ("â€", "-"),
    # checkmark
    ("âœ"", "✓"), ("âœ…", "✅"),
    # emojis doble-encoding
    ("ðŸŽ"", "🎓"), ("ðŸ«", "🏫"), ("ðŸ™ˆ", "🙈"), ("ðŸ'", "👁"),
    ("ðŸ"–", "📖"), ("ðŸ"", "📋"), ("ðŸŽ¬", "🎬"),
    # guiones decorativos en comentarios CSS
    ("â"€", "─"),
]

base = r"C:\Proyectos\capacitaciones'mh\frontend\src"
files = [
    os.path.join(base, "views", "LoginView.vue"),
    os.path.join(base, "views", "user", "MisCapacitaciones.vue"),
]

for f in files:
    with open(f, "r", encoding="utf-8", errors="replace") as fh:
        text = fh.read()
    for bad, good in FIXES:
        text = text.replace(bad, good)
    with open(f, "w", encoding="utf-8") as fh:
        fh.write(text)
    print("OK:", f)
