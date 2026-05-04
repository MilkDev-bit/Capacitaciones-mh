import os

FIXES = [
    ("Ã±", "\u00f1"), ("Ã³", "\u00f3"), ("Ã©", "\u00e9"), ("Ã¡", "\u00e1"), ("Ã­", "\u00ed"), ("Ãº", "\u00fa"),
    ("Â¡", "\u00a1"), ("Â¿", "\u00bf"),
    ("â€"", "\u2014"), ("â€"", "\u2013"), ("â€¢", "\u2022"),
    ("â€˜", "\u2018"), ("â€™", "\u2019"), ("â€œ", "\u201c"),
    ("â€¦", "\u2026"),
    ("\u00c3\u00b1", "\u00f1"), ("\u00c3\u00b3", "\u00f3"), ("\u00c3\u00a9", "\u00e9"),
    ("\u00c3\u00a1", "\u00e1"), ("\u00c3\u00ad", "\u00ed"), ("\u00c3\u00ba", "\u00fa"),
    ("\u00c2\u00a1", "\u00a1"), ("\u00c2\u00bf", "\u00bf"),
    ("âœ\"", "\u2713"),
    ("\u00e2\u009c\u0094", "\u2713"),
    ("ðŸŽ\u201c", "\U0001f393"), ("ðŸ«", "\U0001f3eb"), ("ðŸ™ˆ", "\U0001f648"), ("ðŸ\u2019", "\U0001f441"),
    ("\u00e2\u0094\u0080", "\u2500"),
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
