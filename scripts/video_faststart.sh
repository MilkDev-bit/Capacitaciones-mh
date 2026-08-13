#!/usr/bin/env bash
#
# video_faststart.sh — diagnostica y corrige el arranque lento de los MP4.
#
# EL PROBLEMA
# -----------
# Un MP4 guarda su índice en un atom llamado `moov`: duración, códecs y la
# tabla que dice en qué byte empieza cada fotograma. La mayoría de los
# codificadores lo escriben AL FINAL del archivo, porque hasta terminar no
# conocen su tamaño definitivo.
#
# El navegador no puede reproducir nada sin ese índice. Si `moov` está al
# final, tiene que descargar prácticamente el archivo entero antes del primer
# fotograma — de ahí los ~10 s de rectángulo negro, sin importar que haya un
# CDN delante: el CDN sirve rápido, pero sirve mucho.
#
# `-movflags +faststart` reescribe el contenedor moviendo `moov` al principio.
# No recodifica: `-c copy` deja intactos los flujos de video y audio, así que
# la operación es de segundos y no hay pérdida de calidad.
#
# USO
#   ./scripts/video_faststart.sh check  <archivo.mp4|url>   # diagnostica
#   ./scripts/video_faststart.sh fix    <entrada.mp4> [salida.mp4]
#   ./scripts/video_faststart.sh fixdir <carpeta>           # lote, in-place
#
# Requiere ffmpeg (incluye ffprobe). En Debian/Ubuntu: apt install ffmpeg

set -euo pipefail

requiere() {
	command -v "$1" >/dev/null 2>&1 || {
		echo "Falta '$1'. Instala ffmpeg: apt install ffmpeg" >&2
		exit 1
	}
}

# check informa dónde está el atom moov. Se leen los primeros ~5 MB: si moov
# aparece ahí, el archivo ya arranca rápido.
check() {
	local origen="$1"
	requiere ffprobe

	echo "── $origen"

	# ffprobe con -v trace expone el orden en que encuentra los atoms; el
	# primero de los dos que aparezca (moov o mdat) indica el layout real.
	local orden
	orden=$(ffprobe -v trace -i "$origen" 2>&1 |
		grep -oE "type:'(moov|mdat)'" |
		head -2 |
		grep -oE 'moov|mdat' |
		tr '\n' ' ')

	case "$orden" in
	"moov mdat "*|"moov "*)
		echo "   OK — moov al principio, el video arranca de inmediato."
		;;
	"mdat"*)
		echo "   LENTO — moov al final. El navegador debe descargar casi todo"
		echo "   el archivo antes del primer fotograma. Corrige con:"
		echo "     $0 fix \"$origen\""
		;;
	*)
		echo "   No se pudo determinar el layout (¿no es MP4?)."
		;;
	esac

	ffprobe -v error -show_entries format=duration,size,bit_rate \
		-of default=noprint_wrappers=1 "$origen" 2>/dev/null || true
}

# fix reescribe el contenedor sin recodificar.
fix() {
	local entrada="$1"
	local salida="${2:-}"
	requiere ffmpeg

	if [[ -z "$salida" ]]; then
		salida="${entrada%.*}.faststart.mp4"
	fi

	echo "Reescribiendo contenedor: $entrada -> $salida"
	ffmpeg -hide_banner -loglevel warning -y \
		-i "$entrada" \
		-c copy -map 0 \
		-movflags +faststart \
		"$salida"

	echo "Listo. Verificando:"
	check "$salida"
}

# fixdir procesa una carpeta completa reemplazando los originales.
fixdir() {
	local carpeta="$1"
	requiere ffmpeg

	find "$carpeta" -type f \( -iname '*.mp4' -o -iname '*.m4v' -o -iname '*.mov' \) -print0 |
		while IFS= read -r -d '' archivo; do
			local tmp="${archivo}.tmp.mp4"
			echo "→ $archivo"
			if ffmpeg -hide_banner -loglevel error -y \
				-i "$archivo" -c copy -map 0 -movflags +faststart "$tmp"; then
				mv "$tmp" "$archivo"
			else
				echo "   falló, se conserva el original" >&2
				rm -f "$tmp"
			fi
		done
	echo "Lote terminado."
}

case "${1:-}" in
check) shift; [[ $# -ge 1 ]] || { echo "uso: $0 check <archivo|url>" >&2; exit 1; }; for f in "$@"; do check "$f"; done ;;
fix) shift; [[ $# -ge 1 ]] || { echo "uso: $0 fix <entrada> [salida]" >&2; exit 1; }; fix "$@" ;;
fixdir) shift; [[ $# -eq 1 ]] || { echo "uso: $0 fixdir <carpeta>" >&2; exit 1; }; fixdir "$1" ;;
*)
	sed -n '2,30p' "$0" | sed 's/^# \{0,1\}//'
	exit 1
	;;
esac
