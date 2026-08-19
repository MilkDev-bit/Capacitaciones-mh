# Gotenberg con las fuentes que exige la plantilla DC-3.
#
# La imagen oficial trae sustitutos métricos para Arial, Times New Roman y
# Courier New (Liberation), pero NO para Arial Narrow, que es la fuente
# dominante del formato oficial de la STPS: 916 de sus referencias.
#
# Sin ella LibreOffice cae a una tipografía más ancha y pasan dos cosas a la vez:
# el texto se sale de unas celdas que tienen altura fija —la CURP, el puesto, la
# ocupación salen cortados a media letra— y el documento entero deja de caber,
# pasando de dos páginas a tres. La documentación de Gotenberg lo señala como la
# causa más común de exactamente estos dos síntomas.
#
# Se usa la variante -libreoffice: solo se llama a /forms/libreoffice/convert, y
# la imagen completa arrastra además Chromium, que aquí no convierte nada.
FROM gotenberg/gotenberg:8-libreoffice

USER root

# fonts-liberation (1.x) incluye Liberation Sans Narrow, métricamente compatible
# con Arial Narrow. Si la imagen base ya lo trae, esto no hace nada; instalarlo
# explícitamente evita que una actualización de la base lo retire sin avisar y
# nos devuelva el problema en silencio.
RUN apt-get update -qq \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --no-install-recommends \
        fonts-liberation \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# La regla de sustitución es lo que realmente cierra el problema.
#
# Tener la fuente instalada no basta: fontconfig tiene que saber que "Arial
# Narrow" debe resolverse a "Liberation Sans Narrow". Sin la regla puede elegir
# cualquier otra familia por proximidad de nombre y volvemos a lo mismo.
COPY gotenberg/fonts-dc3.conf /etc/fonts/conf.d/99-dc3.conf

RUN fc-cache -f \
    && echo "=== Resolución de fuentes de la plantilla DC-3 ===" \
    && echo -n "Arial Narrow -> " && fc-match "Arial Narrow" \
    && echo -n "Arial        -> " && fc-match "Arial" \
    # El build falla si Arial Narrow no resuelve a una Liberation Narrow.
    #
    # No es celo excesivo: la primera constancia que se generó en producción
    # incrustó NotoSans en lugar de una condensada, y ese es exactamente el
    # fallo silencioso que se quiere impedir. Una imagen sin esta fuente produce
    # documentos con los campos cortados, y eso solo se descubre cuando alguien
    # abre el PDF y se queja.
    && fc-match "Arial Narrow" | grep -qi "LiberationSansNarrow" \
    && echo "OK: Arial Narrow resuelve a Liberation Sans Narrow"

USER gotenberg
