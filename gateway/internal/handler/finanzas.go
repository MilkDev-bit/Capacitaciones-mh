package handler

import (
	"net/http"

	cursospb "Prueba-Go/gen/cursos"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────────────────────────────────────
// Panel financiero (solo admin)
//
// Todo viaja en CENTAVOS y como entero, hasta el navegador. El panel anterior
// mandaba `float` y arrastraba el error de coma flotante hasta la pantalla de
// la directiva; un centavo de redondeo por orden se nota al sumar un año.
// El formateo a "$1,234.56" se hace en el frontend, que es donde se sabe el
// idioma del usuario.
// ─────────────────────────────────────────────────────────────────────────────

// GET /api/admin/finanzas
func (h *CursosHandler) GetFinanzasAdmin(ctx *gin.Context) {
	resp, err := h.c.Cursos.GetFinanzasAdmin(ctx.Request.Context(), &cursospb.EmptyRequest{})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}

	serie := make([]gin.H, 0, len(resp.Serie))
	for _, p := range resp.Serie {
		serie = append(serie, gin.H{
			"mes":               p.Mes,
			"bruto_centavos":    p.BrutoCentavos,
			"comision_centavos": p.ComisionCentavos,
			"neto_centavos":     p.NetoCentavos,
		})
	}

	movs := make([]gin.H, 0, len(resp.TransaccionesRecientes))
	for _, t := range resp.TransaccionesRecientes {
		movs = append(movs, gin.H{
			"id":                t.Id,
			"cliente":           t.Cliente,
			"email":             t.Email,
			"fecha":             t.Fecha,
			"bruto_centavos":    t.BrutoCentavos,
			"comision_centavos": t.ComisionCentavos,
			"neto_centavos":     t.NetoCentavos,
			"comision_conocida": t.ComisionConocida,
			"origen":            t.Origen,
			"concepto":          t.Concepto,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"bruto_centavos":    resp.BrutoCentavos,
		"comision_centavos": resp.ComisionCentavos,
		"neto_centavos":     resp.NetoCentavos,
		"moneda":            resp.Moneda,
		"transacciones":     resp.Transacciones,
		// Cobros liquidados a los que aún les falta la comisión. Mientras no sea
		// cero, el neto es un piso y la pantalla lo advierte.
		"sin_comision":            resp.SinComision,
		"serie":                   serie,
		"transacciones_recientes": movs,
	})
}

// GET /api/admin/licencias
//
// Licencias agrupadas por empresa compradora. El enlace "Empresas / Licencias"
// del menú de administración llevaba a una ruta que no existía y la pantalla
// salía en blanco; esto es lo que debía haber detrás.
func (h *CursosHandler) AdminListLicenciasEmpresas(ctx *gin.Context) {
	resp, err := h.c.Cursos.AdminListLicenciasEmpresas(ctx.Request.Context(), &cursospb.EmptyRequest{})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}

	items := make([]gin.H, 0, len(resp.Licencias))
	for _, l := range resp.Licencias {
		items = append(items, gin.H{
			"id":                  l.Id,
			"nombre":              l.Nombre,
			"capacitacion_id":     l.CapacitacionId,
			"capacitacion_titulo": l.CapacitacionTitulo,
			"comprador_id":        l.CompradorId,
			"comprador_nombre":    l.CompradorNombre,
			"comprador_email":     l.CompradorEmail,
			"precio_centavos":     l.PrecioCentavos,
			"capacidad_maxima":    l.CapacidadMaxima,
			"usadas":              l.Usadas,
			"codigo_acceso":       l.CodigoAcceso,
			"created_at":          l.CreatedAt,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"licencias":       items,
		"total_licencias": resp.TotalLicencias,
		"total_asientos":  resp.TotalAsientos,
		"asientos_usados": resp.AsientosUsados,
	})
}
