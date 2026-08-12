package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"Prueba-Go/gateway/internal/middleware"
	cursospb "Prueba-Go/gen/cursos"
	"Prueba-Go/pkg/mailer"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v78"
	"google.golang.org/grpc/metadata"
)

// ─────────────────────────────────────────────────────────────────────────────
// Procesamiento de compras
//
// Una compra pagada puede llegar por dos caminos que compiten entre sí:
//   1. El webhook de Stripe (autoritativo, pero asíncrono y puede tardar).
//   2. /verify-checkout-session, que el frontend llama al volver del pago.
//
// Ambos ejecutan el MISMO procesamiento —las operaciones de BD son idempotentes—
// pero el correo no lo es: enviarlo dos veces se le nota al usuario. Por eso el
// envío pasa por `yaNotificado`, que descarta el duplicado.
// ─────────────────────────────────────────────────────────────────────────────

// itemCompra es un renglón ya procesado de la compra.
type itemCompra struct {
	Tipo       string `json:"tipo"` // "b2c" | "b2b_direct"
	CursoID    string `json:"curso_id"`
	Titulo     string `json:"titulo"`
	CursoType  string `json:"curso_type"`
	Cantidad   int    `json:"cantidad"`
	LicenciaID string `json:"licencia_id,omitempty"`

	// Código de acceso. Deliberadamente sin exportar: viaja por correo, no en
	// la respuesta HTTP de la pantalla de éxito.
	codigoAcceso string
}

// resumenCompra es lo que se devuelve al frontend para pintar la pantalla de
// éxito y decidir a dónde mandar al usuario.
type resumenCompra struct {
	Items    []itemCompra `json:"items"`
	Total    float64      `json:"total"`
	Redirect string       `json:"redirect"`
	Etiqueta string       `json:"etiqueta"`
}

// ── Deduplicación de notificaciones ──────────────────────────────────────────

const ventanaDedupe = 30 * time.Minute

var sesionesNotificadas sync.Map // sessionID -> time.Time (momento del envío)

// yaNotificado marca la sesión y devuelve true si ya se había notificado.
// Es memoria de proceso: si el gateway reinicia entre el webhook y el verify,
// el comprador podría recibir el acuse dos veces. Es un mal menor frente a la
// alternativa (no enviarlo nunca si el webhook no está configurado).
func yaNotificado(sessionID string) bool {
	if sessionID == "" {
		return false
	}
	ahora := time.Now()
	if prev, cargado := sesionesNotificadas.LoadOrStore(sessionID, ahora); cargado {
		if ahora.Sub(prev.(time.Time)) < ventanaDedupe {
			return true
		}
		sesionesNotificadas.Store(sessionID, ahora)
		return false
	}

	// Limpieza perezosa: evita que el mapa crezca sin límite en instancias de
	// larga vida sin necesitar una goroutine dedicada.
	go func() {
		time.Sleep(ventanaDedupe)
		sesionesNotificadas.Delete(sessionID)
	}()
	return false
}

// ── Procesamiento ────────────────────────────────────────────────────────────

// procesarSesion ejecuta las altas correspondientes a una sesión de Stripe
// pagada y devuelve el resumen. Es seguro llamarla varias veces.
func (h *CursosHandler) procesarSesion(ctx context.Context, sess *stripe.CheckoutSession, userID string) resumenCompra {
	var res resumenCompra
	if sess.AmountTotal > 0 {
		res.Total = float64(sess.AmountTotal) / 100
	}

	ref := sess.ClientReferenceID
	parts := strings.Split(ref, "||")
	if len(parts) == 0 {
		return res
	}

	// El userID del contexto tiene prioridad (viene del JWT); el de la referencia
	// es el respaldo para el webhook, donde no hay sesión HTTP.
	if userID == "" && len(parts) >= 2 {
		userID = parts[1]
	}
	if userID == "" {
		slog.Error("procesarSesion: sin user_id", "session", sess.ID, "ref", ref)
		return res
	}

	// Se parte de la metadata que ya traía el contexto (x-user-name/x-user-email
	// desde el JWT) en lugar de reemplazarla: el repositorio de cursos la usa
	// para registrar quién se inscribió.
	mdBase, _ := metadata.FromOutgoingContext(ctx)
	mdFor := func(sufijo string) context.Context {
		md := mdBase.Copy()
		md.Set("x-stripe-session-id", sess.ID+sufijo)
		return metadata.NewOutgoingContext(ctx, md)
	}

	switch {
	case parts[0] == "curso" && len(parts) >= 3:
		res.Items = append(res.Items, h.enrolarB2C(mdFor(""), userID, parts[2]))

	case parts[0] == "licencia" && len(parts) == 3:
		if _, err := h.c.Cursos.WebhookComprarLicencia(mdFor(""), &cursospb.WebhookComprarLicenciaRequest{
			UserId:     userID,
			LicenciaId: parts[2],
		}); err != nil {
			slog.Error("WebhookComprarLicencia falló", "error", err, "session", sess.ID)
		}
		res.Items = append(res.Items, itemCompra{Tipo: "b2b_direct", LicenciaID: parts[2], Cantidad: 1})

	case parts[0] == "b2b_direct" && len(parts) >= 4:
		cantidad, _ := strconv.Atoi(parts[3])
		res.Items = append(res.Items, h.comprarB2B(mdFor(""), userID, parts[2], cantidad))

	case parts[0] == "cart":
		// El orden de las claves de un map de Go es aleatorio; se ordenan para
		// que el resumen y el correo salgan siempre igual.
		claves := make([]string, 0, len(sess.Metadata))
		for k := range sess.Metadata {
			if strings.HasPrefix(k, "item_") {
				claves = append(claves, k)
			}
		}
		ordenarClavesItem(claves)

		for _, k := range claves {
			itemParts := strings.Split(sess.Metadata[k], "||")
			if len(itemParts) < 2 {
				continue
			}
			switch itemParts[0] {
			case "b2c":
				res.Items = append(res.Items, h.enrolarB2C(mdFor("_"+k), userID, itemParts[1]))
			case "b2b_direct":
				if len(itemParts) < 3 {
					continue
				}
				cantidad, _ := strconv.Atoi(itemParts[2])
				res.Items = append(res.Items, h.comprarB2B(mdFor("_"+k), userID, itemParts[1], cantidad))
			}
		}
	}

	res.Redirect, res.Etiqueta = destinoTrasCompra(res.Items)
	return res
}

// enrolarB2C inscribe al comprador en un curso individual.
func (h *CursosHandler) enrolarB2C(ctx context.Context, userID, cursoID string) itemCompra {
	item := itemCompra{Tipo: "b2c", CursoID: cursoID, Cantidad: 1}

	resp, err := h.c.Cursos.WebhookEnroll(ctx, &cursospb.WebhookEnrollRequest{
		UserId:         userID,
		CapacitacionId: cursoID,
	})
	if err != nil {
		slog.Error("WebhookEnroll falló", "error", err, "curso_id", cursoID, "user_id", userID)
		return item
	}
	item.Titulo = resp.CapacitacionTitulo
	item.CursoType = resp.CapacitacionType
	return item
}

// comprarB2B crea la licencia corporativa y guarda sus códigos en el ítem.
func (h *CursosHandler) comprarB2B(ctx context.Context, userID, cursoID string, cantidad int) itemCompra {
	item := itemCompra{Tipo: "b2b_direct", CursoID: cursoID, Cantidad: cantidad}

	resp, err := h.c.Cursos.WebhookComprarB2BDirect(ctx, &cursospb.WebhookComprarB2BDirectRequest{
		UserId:   userID,
		CursoId:  cursoID,
		Cantidad: int32(cantidad),
	})
	if err != nil {
		slog.Error("WebhookComprarB2BDirect falló", "error", err, "curso_id", cursoID, "user_id", userID)
		return item
	}
	item.Titulo = resp.CapacitacionTitulo
	item.CursoType = resp.CapacitacionType
	item.LicenciaID = resp.LicenciaId
	item.codigoAcceso = resp.CodigoAcceso
	return item
}

// destinoTrasCompra decide a dónde llevar al comprador al terminar.
//
// Esta era la principal queja del flujo anterior: el carrito siempre mandaba a
// /usuario/capacitaciones, incluso al comprar licencias corporativas, que viven
// en otra pantalla.
func destinoTrasCompra(items []itemCompra) (ruta, etiqueta string) {
	if len(items) == 0 {
		return "/usuario/dashboard", "Ir a mi panel"
	}

	tieneB2B, tieneB2C := false, false
	for _, it := range items {
		if it.Tipo == "b2b_direct" {
			tieneB2B = true
		} else {
			tieneB2C = true
		}
	}

	switch {
	// Compra individual de un solo curso: se entra directo al contenido.
	case !tieneB2B && len(items) == 1 && items[0].CursoID != "":
		return "/usuario/capacitaciones/" + items[0].CursoID, "Comenzar el curso"
	case tieneB2B && !tieneB2C:
		return "/usuario/licencias", "Repartir los accesos"
	case tieneB2B && tieneB2C:
		return "/usuario/licencias", "Ver mis licencias"
	default:
		return "/usuario/capacitaciones", "Ver mis capacitaciones"
	}
}

// ── Correos de compra ────────────────────────────────────────────────────────

// notificarCompra envía el acuse de pago y, si hubo licencias corporativas, el
// correo con los accesos. Se ejecuta una sola vez por sesión de Stripe.
func (h *CursosHandler) notificarCompra(sess *stripe.CheckoutSession, res resumenCompra, nombre, email string) {
	if !h.mail.Enabled() || email == "" || len(res.Items) == 0 {
		return
	}
	if yaNotificado(sess.ID) {
		return
	}

	base := strings.TrimRight(h.cfg.AppURL, "/")

	lineas := make([]mailer.PurchaseLine, 0, len(res.Items))
	for _, it := range res.Items {
		tipo := "Inscripción individual"
		if it.Tipo == "b2b_direct" {
			tipo = "Licencias corporativas"
		}
		titulo := it.Titulo
		if titulo == "" {
			titulo = "Capacitación"
		}
		lineas = append(lineas, mailer.PurchaseLine{Titulo: titulo, Tipo: tipo, Cantidad: it.Cantidad})
	}

	confirm := h.mail.PurchaseConfirmation(nombre, lineas, res.Total, res.Etiqueta, base+res.Redirect)
	confirm.To = []string{email}
	h.mail.SendAsync(confirm)

	// Segundo correo: los accesos propiamente dichos. Va aparte para que el
	// comprador pueda reenviarlo a su equipo sin exponer el importe pagado.
	for _, it := range res.Items {
		if it.Tipo != "b2b_direct" {
			continue
		}
		msg := h.mail.CorporateLicense(
			nombre, it.Titulo, it.Cantidad, it.codigoAcceso,
			base+"/usuario/licencias",
		)
		msg.To = []string{email}
		h.mail.SendAsync(msg)
	}
}

// ── Reparto de accesos a participantes ───────────────────────────────────────

// POST /api/licencias/:id/enviar-accesos
//
// Recibe los correos de los participantes, reserva un acceso para cada uno y se
// lo envía. Es la alternativa a que el comprador copie códigos a mano desde el
// módulo de licencias.
func (h *CursosHandler) EnviarAccesosLicencia(ctx *gin.Context) {
	licenciaID := ctx.Param("id")
	userID := ctx.GetString(middleware.CtxUserID)

	var body struct {
		Participantes []struct {
			Nombre string `json:"nombre"`
			Email  string `json:"email" binding:"required,email"`
		} `json:"participantes" binding:"required,min=1,max=100,dive"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "revisa los correos capturados: " + err.Error()})
		return
	}

	// Se deduplica en el gateway para no gastar lugares de la licencia por un
	// copy/paste repetido del comprador.
	vistos := map[string]bool{}
	req := &cursospb.AsignarAccesosLicenciaRequest{LicenciaId: licenciaID, CompradorId: userID}
	for _, p := range body.Participantes {
		email := strings.ToLower(strings.TrimSpace(p.Email))
		if email == "" || vistos[email] {
			continue
		}
		vistos[email] = true
		req.Participantes = append(req.Participantes, &cursospb.ParticipanteInput{
			Nombre: strings.TrimSpace(p.Nombre),
			Email:  email,
		})
	}
	if len(req.Participantes) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no hay correos válidos que enviar"})
		return
	}

	resp, err := h.c.Cursos.AsignarAccesosLicencia(genMetadata(ctx), req)
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}

	if !h.mail.Enabled() {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "los accesos se reservaron pero el correo no está configurado en el servidor",
		})
		return
	}

	comprador := ctx.GetString(middleware.CtxUserName)
	base := strings.TrimRight(h.cfg.AppURL, "/")

	// Un solo lote a Resend en vez de N peticiones: con 50 participantes la
	// diferencia entre una llamada y cincuenta es notoria.
	msgs := make([]mailer.Message, 0, len(resp.Accesos))
	for _, a := range resp.Accesos {
		destino := fmt.Sprintf("%s/unirse/%s", base, a.Codigo)
		msg := h.mail.ParticipantAccess(a.Nombre, comprador, a.CapacitacionTitulo, a.Codigo, destino)
		msg.To = []string{a.Email}
		msgs = append(msgs, msg)
	}

	go func() {
		bg, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := h.mail.SendBatch(bg, msgs); err != nil {
			slog.Error("EnviarAccesosLicencia: fallo al enviar el lote",
				"error", err, "licencia_id", licenciaID, "destinatarios", len(msgs))
		}
	}()

	ctx.JSON(http.StatusOK, gin.H{"enviados": len(resp.Accesos), "accesos": resp.Accesos})
}

// GET /api/licencias/:id/invitaciones
func (h *CursosHandler) ListInvitacionesLicencia(ctx *gin.Context) {
	resp, err := h.c.Cursos.ListInvitacionesLicencia(genMetadata(ctx), &cursospb.LicenciaIDRequest{
		Id: ctx.Param("id"),
	})
	if err != nil {
		grpcToHTTP(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp.Invitaciones)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// ordenarClavesItem ordena "item_0", "item_1", ... "item_10" numéricamente.
// Un sort lexicográfico pondría "item_10" antes que "item_2".
func ordenarClavesItem(claves []string) {
	for i := 1; i < len(claves); i++ {
		for j := i; j > 0 && indiceItem(claves[j]) < indiceItem(claves[j-1]); j-- {
			claves[j], claves[j-1] = claves[j-1], claves[j]
		}
	}
}

func indiceItem(clave string) int {
	n, err := strconv.Atoi(strings.TrimPrefix(clave, "item_"))
	if err != nil {
		return 1 << 30
	}
	return n
}
