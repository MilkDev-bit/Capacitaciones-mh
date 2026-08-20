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

# La variante Narrow va en un paquete APARTE de fonts-liberation.
#
# Se dio por hecho que fonts-liberation la incluía y no es cierto: ese paquete
# trae Liberation Sans, Serif y Mono, todas de ancho normal. Con solo eso,
# fc-match resolvía "Arial Narrow" a "Liberation Sans" —la comprobación de más
# abajo lo delató— y volvíamos al mismo documento descuadrado.
#
# Se instalan dos candidatos, en orden de preferencia:
#
#   fonts-liberation-sans-narrow → Liberation Sans Narrow, compatible métrica
#                                  directa de Arial Narrow.
#   fonts-urw-base35             → Nimbus Sans Narrow, compatible de Helvetica
#                                  Narrow. Arial es métricamente Helvetica, así
#                                  que sirve igual y está en cualquier Debian.
#
# Se instalan de uno en uno y sin abortar: si un paquete no existe en la versión
# de Debian de la imagen base, interesa seguir y que decida la comprobación
# final, no que reviente el apt por un nombre.
RUN apt-get update -qq \
    && for p in fonts-liberation fonts-liberation-sans-narrow fonts-urw-base35; do \
         DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --no-install-recommends "$p" \
           || echo "AVISO: no se pudo instalar $p"; \
       done \
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
