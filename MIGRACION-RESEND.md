# Migración: Resend + verificación de correo + flujo de pago

Pasos para dejar el cambio corriendo. Los tres primeros son obligatorios.

## 1. Regenerar protobuf

Se añadieron RPCs y mensajes en `proto/auth` y `proto/cursos`. **El backend no compila
sin esto.**

```bash
make generate
make tidy
make build
```

RPCs nuevos:

| Servicio | RPC | Motivo |
|---|---|---|
| auth | `VerifyEmail`, `ResendVerificationCode` | Código de 6 dígitos |
| cursos | `AsignarAccesosLicencia`, `ListInvitacionesLicencia` | Reparto de accesos |

Cambios de firma (solo los llama el gateway):

- `WebhookEnroll` → devuelve `EnrollResponse` en vez de `EmptyResponse`
- `WebhookComprarB2BDirect` → devuelve `ComprarB2BDirectResponse`

Ambos ahora regresan el detalle de lo comprado para que el gateway arme el correo y
la pantalla de éxito sin una segunda ronda de consultas. `make breaking` marcará
estos dos como incompatibles; es intencional.

## 2. Variables de entorno

Añadir al `.env` (ver `.env.example`):

```env
RESEND_API_KEY=re_...
RESEND_FROM=Capacitaciones MH <no-reply@tudominio.com>
RESEND_REPLY_TO=
EMAIL_VERIFICATION_TTL_MINUTES=15
EMAIL_VERIFICATION_COOLDOWN_SECONDS=60
```

El dominio de `RESEND_FROM` debe estar verificado en Resend o la API responde 403.
Para probar sin dominio propio: `Capacitaciones MH <onboarding@resend.dev>`.

`APP_URL` debe apuntar a la URL pública real: se usa para el logo y los enlaces de
los correos.

Las variables `SMTP_*` quedan deprecadas. `SMTP_FROM` se sigue leyendo como respaldo
de `RESEND_FROM` para no romper despliegues existentes.

## 3. Migraciones de base de datos

Se aplican solas al arrancar (`internal/db/schema.go` y `services/auth/cmd/server/main.go`):

- `users`: `email_verified`, `email_verification_hash`, `email_verification_expires`,
  `email_verification_attempts`, `email_verification_sent_at`
- Tabla nueva `licencia_invitaciones`

**Las cuentas existentes quedan verificadas automáticamente.** La columna se agrega
con `DEFAULT true` y acto seguido el default pasa a `false`, así que solo las cuentas
nuevas nacen sin verificar.

## 4. Retiro de videollamadas

El flujo de capacitaciones por videollamada se desmontó por completo: 5 vistas,
2 componentes, 11 RPCs, 9 rutas del gateway y 2 tablas.

**No se borra nada automáticamente.** El código dejó de leer y escribir
`instructor_schedules` y `videocall_tickets`, pero las tablas siguen ahí. Para
eliminarlas, respalda y corre:

```bash
pg_dump -t instructor_schedules -t videocall_tickets -d "$DATABASE_URL" > respaldo_videollamadas.sql
psql "$DATABASE_URL" -f migrations/001_retirar_videollamadas.sql
```

Los cursos con `type = 'videocall'` se despublican solos al arrancar
(`is_public = false`). Conservan contenido, inscripciones e historial de compras;
solo dejan de ofrecerse. Si quieres recuperarlos como material autoformativo,
el `.sql` incluye al final el `UPDATE` para reclasificarlos a `'video'`.

La videollamada del chat (Jitsi entre dos usuarios en Mensajes) **no se tocó**:
es un flujo independiente que no usa horarios ni tickets.

## 5. Aviso DC-3

El correo de constancias se disparaba **solo** al cerrar una videollamada. Ahora
sale cuando un participante termina todas las lecciones del curso.

La deduplicación vive en la tabla nueva `dc3_avisos`, con clave primaria
`(licencia_id, capacitacion_id)`: el representante recibe un correo por licencia
y curso aunque terminen los 50 participantes. Solo aplica a licencias
corporativas — en compras individuales el propio usuario tramita su constancia
desde *Mis Capacitaciones*, que ya tenía su botón.

## 6. Stripe

Actualizar la URL de éxito si está configurada fuera del código:
`/checkout/exito?session_id={CHECKOUT_SESSION_ID}`.

El webhook (`/api/webhooks/stripe`) sigue siendo el camino autoritativo. La pantalla
de éxito llama a `/verify-checkout-session` como respaldo; ambos ejecutan el mismo
procesamiento idempotente y el correo se envía una sola vez por sesión.

---

## Qué cambió funcionalmente

**Registro.** `POST /api/register` ya no devuelve cookie de sesión: responde
`requires_verification: true` y manda un código de 6 dígitos. `POST /api/login` con
un correo sin verificar responde `403 {code: "email_not_verified"}` — 403 y no 401
a propósito, porque el interceptor de axios trata cualquier 401 como sesión expirada
y expulsaría al usuario. El código se guarda hasheado (SHA-256 + userID) y se compara
en tiempo constante; 5 intentos fallidos lo invalidan.

**Licencias corporativas.** Al confirmarse el pago el comprador recibe dos correos:
el acuse con el resumen y otro con los accesos. Desde *Mis Licencias* puede capturar
los correos de su equipo (o pegar una lista) y cada participante recibe su código
individual. El reparto de tickets es transaccional con `FOR UPDATE SKIP LOCKED`, así
que dos envíos simultáneos no pueden entregar el mismo código a dos personas.

**Pagos.** Nueva pantalla `/checkout/exito` con el resumen de la compra y un CTA que
depende de lo comprado: al curso si fue individual, a licencias si fue corporativa.
Antes el carrito siempre redirigía a `/usuario/capacitaciones`, incluso al comprar
licencias, que viven en otra pantalla. El carrito ahora permite cambiar cantidad y
tipo de compra sin borrar la línea.
