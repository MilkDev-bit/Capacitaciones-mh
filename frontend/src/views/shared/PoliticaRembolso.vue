<!-- src/views/shared/PoliticaReembolsos.vue -->
<script setup lang="ts">
/**
 * Política de reembolsos.
 *
 * Se redactó pensando en el modelo real de la plataforma: Stripe como único
 * procesador, dos vías de compra (individual y suscripción) y contenido
 * digital consumible de inmediato (video/lecturas). Por eso el eje central
 * es "¿se consumió el contenido o no", que es también el criterio que Stripe
 * espera ver documentado si un usuario abre una disputa (chargeback) — un
 * "no reembolsable" sin política escrita pierde casi siempre la disputa.
 *
 * México no tiene un "derecho de retracto" genérico como la UE; lo que hay
 * es la garantía general de la LFPC y lo que la propia empresa ofrezca por
 * contrato. Aquí se ofrece una ventana voluntaria de 7 días como práctica
 * comercial, no como obligación legal — eso se aclara explícitamente para no
 * sobre-prometer un derecho que la ley no otorga en automático.
 *
 * Los plazos de "5 a 10 días hábiles" para ver el reembolso reflejan los
 * tiempos típicos que Stripe documenta para reversas a tarjeta; no son un
 * compromiso del Responsable, que no controla el banco emisor del Usuario.
 */
const CORREO = 'soporte@mhsolucionesempresariales.com'
</script>

<template>
    <div class="aviso-page">
        <article class="aviso">
            <h1>Política de Reembolsos</h1>
            <p class="aviso-marco">
                Esta política describe en qué casos MH Soluciones Empresariales,
                S.A.S de C.V. (en adelante, «el Responsable») reembolsa pagos
                realizados en la Plataforma. Todos los cobros y reembolsos se
                procesan a través de <strong>Stripe</strong>, y se abonan siempre al
                mismo método de pago con el que se realizó la compra original.
            </p>
            <p class="aviso-version">Última actualización: agosto de 2026</p>

            <h2>I. Principio General</h2>
            <p>
                Al tratarse de contenido digital de acceso inmediato, el criterio
                principal para evaluar una solicitud de reembolso es el
                <strong>grado de consumo del Curso</strong>: cuanto menos avance
                registre el Usuario, mayor es la probabilidad de que proceda. Esta
                política es una condición contractual ofrecida por el Responsable y
                no sustituye las garantías que, en su caso, reconozca la Ley Federal
                de Protección al Consumidor.
            </p>

            <h2>II. Compras Individuales de Cursos</h2>
            <p>
                Para la compra de un Curso individual, el Responsable ofrece una
                ventana de <strong>7 días naturales</strong> a partir de la fecha de
                compra, sujeta a que se cumplan ambas condiciones:
            </p>
            <ul>
                <li>No se haya avanzado más del <strong>20% del contenido</strong> del Curso.</li>
                <li>No se haya descargado ni emitido la constancia o certificado correspondiente.</li>
            </ul>
            <p>
                Pasado este plazo, o si se supera el porcentaje de avance indicado,
                el Curso se considera consumido y <strong>no procede reembolso</strong>,
                salvo que se trate de un caso de las excepciones señaladas en el
                apartado VI.
            </p>

            <h2>III. Suscripciones</h2>
            <p>
                La cancelación de una Suscripción detiene los cargos futuros pero
                <strong>no genera un reembolso del periodo en curso</strong>: el
                Usuario conserva el acceso hasta el último día del periodo ya pagado,
                conforme a lo señalado en los Términos y Condiciones.
            </p>
            <p>Existen dos excepciones a esta regla:</p>
            <ul>
                <li>
                    <strong>Primera suscripción, primeras 48 horas:</strong> si el
                    Usuario cancela dentro de las 48 horas posteriores a su primer
                    cargo y no ha abierto ningún Curso incluido en el plan, se
                    reembolsa el cargo completo.
                </li>
                <li>
                    <strong>Cobro duplicado o error de facturación:</strong> ver
                    apartado VI.
                </li>
            </ul>

            <h2>IV. Licencias Corporativas</h2>
            <p>
                Las licencias adquiridas por lote (compra corporativa) son
                reembolsables de forma proporcional únicamente por los
                <strong>lugares que no hayan sido asignados ni activados</strong> por
                ningún colaborador al momento de solicitar el reembolso. Los lugares
                ya asignados y con contenido iniciado siguen la regla del apartado II.
            </p>

            <h2>V. Cómo Solicitar un Reembolso</h2>
            <p>
                Actualmente, la funcionalidad de reembolso <strong>no se encuentra automatizada en la
                    plataforma</strong>.
                Si cumple con los criterios descritos en esta política, deberá solicitar la gestión de manera manual
                enviando un correo electrónico a <a :href="`mailto:${CORREO}`">{{ CORREO }}</a> indicando:
            </p>
            <ol>
                <li>Nombre completo y correo registrado en la Plataforma.</li>
                <li>Curso, plan o licencia por el que solicita la devolución.</li>
                <li>Fecha aproximada del cargo y referencia de pago de Stripe.</li>
                <li>Motivo de la solicitud.</li>
            </ol>
            <p>
                El Responsable revisará la solicitud y responderá en un plazo máximo de <strong>5 días hábiles</strong>.
                De ser procedente, el reembolso se procesará de forma manual a través del panel de Stripe
                hacia el método de pago original.
            </p>

            <h2>VI. Excepciones que Siempre Proceden</h2>
            <p>
                Independientemente del avance o del plazo transcurrido, se reembolsa
                de inmediato cuando se acredite:
            </p>
            <ul>
                <li><strong>Cobro duplicado</strong> por un mismo Curso o periodo de Suscripción.</li>
                <li><strong>Error de precio</strong> atribuible a una falla técnica de la Plataforma.</li>
                <li><strong>Curso no entregado o inaccesible</strong> por causas imputables al Responsable, sin que el
                    Usuario haya podido consumir el contenido.</li>
                <li><strong>Cargo no reconocido</strong> derivado de un uso fraudulento comprobado de un método de pago
                    ajeno al titular.</li>
            </ul>

            <h2>VII. Tiempos de Reembolso</h2>
            <p>
                Una vez emitido el reembolso desde Stripe, el banco o emisor de la
                tarjeta del Usuario suele reflejarlo en un plazo de
                <strong>5 a 10 días hábiles</strong>, dependiendo de sus propios
                tiempos de procesamiento internos. Este plazo no depende del
                Responsable, que no tiene control sobre los tiempos del banco
                emisor.
            </p>

            <h2>VIII. Contracargos (Chargebacks)</h2>
            <p>
                Antes de iniciar un contracargo con su banco o tarjeta, se recomienda
                contactar primero al correo de soporte: la mayoría de los casos se
                resuelven de forma más rápida por esta vía. Un contracargo abierto
                directamente con el banco puede resultar, conforme a las reglas de
                Stripe y de las redes de tarjetas (Visa, Mastercard), en la
                <strong>suspensión temporal de la cuenta</strong> del Usuario mientras
                se resuelve la disputa.
            </p>

            <h2>IX. Cambios a esta Política</h2>
            <p>
                El Responsable podrá actualizar esta Política de Reembolsos en
                cualquier momento; los cambios aplican a las compras realizadas
                después de su publicación. Las compras ya realizadas se rigen por la
                versión vigente al momento del pago.
            </p>

            <h2>X. Contacto</h2>
            <p>
                Para cualquier duda sobre reembolsos, escriba a:
                <a :href="`mailto:${CORREO}`">{{ CORREO }}</a>
            </p>
        </article>
    </div>
</template>

<style scoped>
.aviso-page {
    padding: clamp(24px, 5vw, 56px) var(--gutter, 16px);
    background: var(--bg);
    min-height: 100dvh;
}

.aviso {
    max-width: 76ch;
    margin: 0 auto;
    color: var(--text);
    line-height: 1.7;
}

.aviso h1 {
    margin: 0 0 12px;
    font-size: clamp(1.5rem, 1.2rem + 1.4vw, 2rem);
}

.aviso-marco {
    margin: 0 0 12px;
    padding: 12px 14px;
    border-left: 3px solid var(--border);
    color: var(--muted);
    font-size: 0.85rem;
    line-height: 1.6;
}

.aviso-version {
    margin: 0 0 28px;
    color: var(--muted);
    font-size: 0.85rem;
}

.aviso h2 {
    margin: 34px 0 8px;
    font-size: 1.05rem;
}

.aviso p,
.aviso li {
    font-size: 0.95rem;
}

.aviso ul,
.aviso ol {
    padding-left: 1.4em;
    margin: 0 0 14px;
}

.aviso li {
    margin-bottom: 6px;
}

.aviso a {
    color: var(--brand);
    text-decoration: underline;
    overflow-wrap: anywhere;
}
</style>