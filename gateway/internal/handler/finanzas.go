package handler

import (
	"context"
	"net/http"

	cursospb "Prueba-Go/gen/cursos"
	usuariospb "Prueba-Go/gen/usuarios"

	"github.com/gin-gonic/gin"
)

// resolverNombres traduce IDs de usuario a nombre y correo.
//
// Hace falta porque cada microservicio tiene su propia base: `users` vive en la
// de auth y `ordenes` en la de cursos, así que un JOIN entre ambas es
// imposible. El gateway es el único sitio que habla con los dos servicios.
//
// La caché evita pedir el mismo perfil dos veces cuando un cliente aparece en
// varias filas. Un usuario que no se puede resolver —cuenta borrada, o el
// servicio caído— no rompe la pantalla: se queda con un texto de respaldo.
func (h *CursosHandler) resolverNombres(ctx context.Context, ids []string) map[string]*usuariospb.PerfilResponse {
	cache := make(map[string]*usuariospb.PerfilResponse, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, visto := cache[id]; visto {
			continue
		}
		u, err := h.c.Usuarios.GetPublicPerfil(ctx, &usuariospb.UserIDRequest{UserId: id})
		if err != nil || u == nil {
			cache[id] = nil
			continue
		}
		cache[id] = u
	}
	return cache
}

// nombreOr saca el nombre de un perfil que puede no haberse podido resolver.
func nombreOr(u *usuariospb.PerfilResponse, respaldo string) string {
	if u != nil && u.Name != "" {
		return u.Name
	}
	return respaldo
}

func correoDe(u *usuariospb.PerfilResponse) string {
	if u == nil {
		return ""
	}
	return u.Email
}

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

	ids := make([]string, 0, len(resp.TransaccionesRecientes))
	for _, t := range resp.TransaccionesRecientes {
		ids = append(ids, t.UserId)
	}
	perfiles := h.resolverNombres(ctx.Request.Context(), ids)

	movs := make([]gin.H, 0, len(resp.TransaccionesRecientes))
	for _, t := range resp.TransaccionesRecientes {
		u := perfiles[t.UserId]
		movs = append(movs, gin.H{
			"id":                t.Id,
			"cliente":           nombreOr(u, "Cuenta eliminada"),
			"email":             correoDe(u),
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

	ids := make([]string, 0, len(resp.Licencias))
	for _, l := range resp.Licencias {
		ids = append(ids, l.CompradorId)
	}
	perfiles := h.resolverNombres(ctx.Request.Context(), ids)

	items := make([]gin.H, 0, len(resp.Licencias))
	for _, l := range resp.Licencias {
		u := perfiles[l.CompradorId]
		items = append(items, gin.H{
			"id":                  l.Id,
			"nombre":              l.Nombre,
			"capacitacion_id":     l.CapacitacionId,
			"capacitacion_titulo": l.CapacitacionTitulo,
			"comprador_id":        l.CompradorId,
			"comprador_nombre":    nombreOr(u, "Cuenta eliminada"),
			"comprador_email":     correoDe(u),
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
