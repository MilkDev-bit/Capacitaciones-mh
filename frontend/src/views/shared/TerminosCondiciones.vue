<!-- src/views/shared/TerminosCondiciones.vue -->
<script setup lang="ts">
/**
 * Términos y Condiciones de uso.
 *
 * A diferencia del aviso de privacidad, este documento no está atado a un
 * marco regulatorio único con artículos numerados: es un contrato de
 * adhesión regido principalmente por el Código de Comercio (para actos de
 * comercio electrónico) y la Ley Federal de Protección al Consumidor (LFPC),
 * con las obligaciones DC-3 sujetas a la normativa de la STPS.
 *
 * Como con el aviso, esto es una base razonable para operar la plataforma,
 * no un sustituto de la revisión de un abogado mercantilista, sobre todo en
 * las cláusulas de responsabilidad y las de suscripción/renovación
 * automática, que en México están bajo escrutinio activo de PROFECO.
 *
 * Al modificar cualquier cláusula sustantiva (precios, reembolsos,
 * responsabilidad), sube la fecha de "Última actualización" igual que en el
 * aviso de privacidad.
 */
const CORREO = 'soporte@mhsolucionesempresariales.com'
</script>

<template>
    <div class="aviso-page">
        <article class="aviso">
            <h1>Términos y Condiciones</h1>
            <p class="aviso-marco">
                Este documento regula el acceso y uso de la Plataforma de
                Capacitaciones operada por MH Soluciones Empresariales, S.A.S de C.V.
                (en adelante, «el Responsable» o «la Plataforma»). Al crear una cuenta
                o realizar una compra, usted acepta quedar obligado por estos
                Términos y Condiciones.
            </p>
            <p class="aviso-version">Última actualización: agosto de 2026</p>

            <h2>I. Objeto y Aceptación</h2>
            <p>
                Estos Términos y Condiciones (en adelante, «los Términos») regulan la
                relación entre el Responsable y toda persona física o moral (en
                adelante, «el Usuario») que se registre, navegue, adquiera cursos o
                contrate suscripciones y licencias corporativas a través de la
                Plataforma.
            </p>
            <p>
                El uso de la Plataforma implica la aceptación plena de estos Términos
                y del <a href="/privacidad">Aviso de Privacidad</a>. Si no está de
                acuerdo con alguna disposición, debe abstenerse de utilizar el
                servicio.
            </p>

            <h2>II. Definiciones</h2>
            <ul>
                <li><strong>Plataforma:</strong> el sitio web y sus aplicaciones asociadas, propiedad del Responsable.
                </li>
                <li><strong>Curso:</strong> contenido de capacitación (video, documento, texto o evaluación) disponible
                    en la Plataforma.</li>
                <li><strong>Suscripción:</strong> plan de pago recurrente que otorga acceso a todo o parte del catálogo
                    mientras esté vigente.</li>
                <li><strong>Compra individual:</strong> pago único que otorga acceso permanente a un Curso específico.
                </li>
                <li><strong>Licencia corporativa:</strong> conjunto de accesos adquiridos por una empresa para repartir
                    entre sus colaboradores.</li>
                <li><strong>Constancia DC-3:</strong> documento de acreditación de capacitación conforme a la normativa
                    de la Secretaría del Trabajo y Previsión Social (STPS).</li>
            </ul>

            <h2>III. Registro de Cuenta</h2>
            <p>
                Para adquirir Cursos o Suscripciones es necesario crear una cuenta
                proporcionando información veraz, actual y completa. El Usuario es
                responsable de:
            </p>
            <ul>
                <li>Mantener la confidencialidad de su contraseña y credenciales de acceso.</li>
                <li>Toda actividad realizada desde su cuenta, incluida la que provenga de un tercero autorizado por él.
                </li>
                <li>Notificar de inmediato al Responsable ante cualquier uso no autorizado de su cuenta.</li>
            </ul>
            <p>
                El Responsable podrá suspender o cancelar cuentas con información
                falsa, duplicada o utilizada con fines contrarios a estos Términos.
            </p>

            <h2>IV. Descripción del Servicio</h2>
            <p>
                La Plataforma ofrece dos modalidades de acceso a los Cursos, descritas
                también en la sección «Cómo funciona» del sitio:
            </p>
            <ul>
                <li>
                    <strong>Compra individual:</strong> pago único que otorga acceso
                    indefinido al Curso adquirido, sujeto a la continuidad operativa de
                    la Plataforma.
                </li>
                <li>
                    <strong>Suscripción:</strong> pago recurrente (mensual o el
                    intervalo contratado) que otorga acceso a todo el catálogo,
                    incluidos los Cursos publicados durante la vigencia del plan. Al
                    cancelarse o vencer la Suscripción, el Usuario pierde el acceso a
                    los Cursos que no haya adquirido de forma individual.
                </li>
            </ul>
            <p>
                Las Licencias corporativas se rigen por las mismas reglas de la
                modalidad contratada (individual o por suscripción), aplicadas a cada
                colaborador al que la empresa contratante asigne un acceso.
            </p>

            <h2>V. Precios y Medios de Pago</h2>
            <p>
                Los precios se muestran en la moneda indicada en cada Curso o plan y
                pueden modificarse sin previo aviso; el precio aplicable es el
                vigente al momento de confirmarse la compra.
            </p>
            <p>
                El procesamiento de pagos se realiza a través de <strong>Stripe,
                    Inc.</strong>, un proveedor externo especializado. El Responsable
                <strong>no almacena ni tiene acceso a los datos completos de su
                    tarjeta</strong>; dicha información es capturada y resguardada
                directamente por Stripe conforme a sus propias políticas de
                seguridad y privacidad, disponibles en
                <a href="https://stripe.com/mx/privacy" target="_blank" rel="noopener">stripe.com/privacy</a>.
            </p>
            <p>
                Para las condiciones de devolución de cargos, consulte la
                <a href="/reembolsos">Política de Reembolsos</a>.
            </p>

            <h2>VI. Suscripciones y Renovación Automática</h2>
            <p>
                Las Suscripciones se renuevan automáticamente al vencer cada periodo,
                cargando el monto vigente al método de pago registrado, salvo que el
                Usuario la cancele antes de la fecha de renovación desde su panel de
                cuenta o solicitándolo al correo de contacto.
            </p>
            <p>
                La cancelación detiene renovaciones futuras pero no genera un
                reembolso del periodo ya iniciado: el acceso se mantiene hasta el
                último día del periodo pagado. Ver excepciones en la
                <a href="/reembolsos">Política de Reembolsos</a>.
            </p>

            <h2>VII. Propiedad Intelectual</h2>
            <p>
                Todos los Cursos, materiales, marcas, logotipos, textos, videos y
                demás contenidos disponibles en la Plataforma son propiedad del
                Responsable o de sus instructores licenciantes, y están protegidos por
                la Ley Federal del Derecho de Autor y la Ley Federal de Protección a
                la Propiedad Industrial.
            </p>
            <p>
                La compra o suscripción otorga al Usuario una licencia de uso
                personal, no exclusiva e intransferible para consumo del contenido con
                fines de capacitación propia. Queda prohibido:
            </p>
            <ul>
                <li>Reproducir, distribuir o compartir credenciales de acceso con terceros no autorizados.</li>
                <li>Descargar, grabar o redistribuir el contenido fuera de la Plataforma.</li>
                <li>Utilizar el material con fines comerciales distintos a la capacitación propia o de su organización.
                </li>
            </ul>

            <h2>VIII. Constancias DC-3</h2>
            <p>
                La emisión de constancias DC-3 depende de que el Usuario complete el
                Curso conforme a los criterios de evaluación establecidos y de que los
                datos proporcionados (nombre, CURP, ocupación) sean correctos. El
                Responsable no se hace responsable por constancias no emitidas o
                rechazadas ante la autoridad debido a datos erróneos proporcionados
                por el Usuario.
            </p>

            <h2>IX. Uso Aceptable</h2>
            <p>El Usuario se obliga a no utilizar la Plataforma para:</p>
            <ul>
                <li>Actividades ilícitas o que infrinjan derechos de terceros.</li>
                <li>Intentar vulnerar la seguridad, integridad o disponibilidad del sistema.</li>
                <li>Suplantar la identidad de otra persona o proporcionar información falsa.</li>
                <li>Presentar exámenes o evaluaciones por cuenta de otro Usuario.</li>
            </ul>
            <p>
                El incumplimiento de esta sección puede resultar en la suspensión o
                cancelación de la cuenta sin derecho a reembolso, sin perjuicio de las
                acciones legales que correspondan.
            </p>

            <h2>X. Limitación de Responsabilidad</h2>
            <p>
                El Responsable pone a disposición del Usuario la Plataforma en las
                condiciones existentes, procurando su disponibilidad continua, sin
                que ello constituya una garantía absoluta de operación ininterrumpida
                o libre de errores.
            </p>
            <p>
                En la máxima medida permitida por la legislación aplicable, el
                Responsable no será responsable por daños indirectos, incidentales o
                derivados del uso o la imposibilidad de uso de la Plataforma, salvo en
                los casos que la ley mexicana determine como irrenunciables.
            </p>

            <h2>XI. Modificaciones a los Términos</h2>
            <p>
                El Responsable podrá modificar estos Términos en cualquier momento.
                Los cambios sustanciales se notificarán mediante aviso en la
                Plataforma o por correo electrónico, y el uso continuado del servicio
                tras su publicación implica la aceptación de la versión actualizada.
            </p>

            <h2>XII. Legislación Aplicable y Jurisdicción</h2>
            <p>
                Estos Términos se rigen por las leyes de los Estados Unidos Mexicanos.
                Para la interpretación y cumplimiento de los mismos, el Usuario y el
                Responsable se someten a los tribunales competentes de la jurisdicción
                del domicilio del Responsable, renunciando a cualquier otro fuero que
                pudiera corresponderles por razón de su domicilio presente o futuro,
                sin perjuicio de los derechos irrenunciables que la Ley Federal de
                Protección al Consumidor reconozca al Usuario en su carácter de
                consumidor.
            </p>

            <h2>XIII. Contacto</h2>
            <p>
                Para dudas relacionadas con estos Términos, escriba a:
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

.aviso h3 {
    margin: 22px 0 6px;
    font-size: 0.95rem;
    color: var(--text);
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