<script setup lang="ts">
/**
 * Aviso de privacidad (texto definitivo, proporcionado por el responsable).
 *
 * Al modificar CUALQUIER línea del texto hay que subir AVISO_VERSION en
 * utils/aviso.ts y la fecha de "Última actualización" de esta página. Si no se
 * sube, quedará registrado en la base que las personas aceptaron una versión
 * que ya no es la publicada, que es justo lo que el registro de versión existe
 * para evitar.
 */
import { AVISO_VERSION } from '../../utils/aviso'

const CORREO = 'soporte@mhsolucionesempresariales.com'
</script>

<template>
  <div class="aviso-page">
    <article class="aviso">
      <h1>Aviso de Privacidad</h1>
      <p class="aviso-marco">
        Con fundamento en la Ley Federal de Protección de Datos Personales en
        Posesión de los Particulares (en adelante, la «Ley»), con su última
        reforma publicada en el Diario Oficial de la Federación (DOF) el 14 de
        noviembre de 2025, y su Reglamento vigente publicado en el DOF el 21 de
        diciembre de 2011.
      </p>
      <p class="aviso-version">
        Última actualización: agosto de 2026
        <span class="aviso-version-id">(versión {{ AVISO_VERSION }})</span>
      </p>

      <h2>I. Identidad y Domicilio del Responsable</h2>
      <p>
        MH Soluciones Empresariales, S.A.S de C.V. (en adelante, «el
        Responsable»), con domicilio fiscal en la República Mexicana, es el
        responsable del tratamiento y protección de los datos personales que
        usted proporciona a través de la presente Plataforma de Capacitaciones y
        Formación Continua.
      </p>
      <p>
        Para cualquier asunto relacionado con el tratamiento y protección de sus
        datos personales, puede contactarnos en:
      </p>
      <ul>
        <li>
          Correo electrónico de atención (Derechos ARCO):
          <a :href="`mailto:${CORREO}`">{{ CORREO }}</a>
        </li>
      </ul>

      <h2>II. Datos Personales que se Recaban</h2>
      <p>
        El Responsable recaba las siguientes categorías de datos personales con
        motivo del registro, acceso, gestión académica y certificación en la
        Plataforma de Capacitaciones:
      </p>

      <h3>Datos de identificación y contacto</h3>
      <ul>
        <li>Nombre completo.</li>
        <li>Correo electrónico institucional y/o personal.</li>
        <li>Número de teléfono.</li>
        <li>Puesto, área o departamento (en caso de usuarios corporativos).</li>
        <li>Fotografía de perfil (opcional).</li>
      </ul>

      <h3>Datos académicos y de desempeño</h3>
      <ul>
        <li>Cursos inscritos, avance, calificaciones y evaluaciones.</li>
        <li>
          Registro de asistencias, conexiones y horas de capacitación
          acumuladas.
        </li>
        <li>Certificados, constancias e insignias obtenidas.</li>
      </ul>

      <h3>Datos de facturación y pago (cuando apliquen capacitaciones con costo)</h3>
      <ul>
        <li>Registro Federal de Contribuyentes (RFC).</li>
        <li>Domicilio fiscal.</li>
        <li>
          Comprobantes de pago o datos bancarios necesarios para la transacción.
        </li>
      </ul>

      <h2>III. Finalidades del Tratamiento</h2>
      <p>
        El tratamiento de los datos personales se realiza bajo los principios de
        licitud, consentimiento, información, calidad, finalidad, lealtad,
        proporcionalidad y responsabilidad, previstos en el Artículo 6 de la
        Ley.
      </p>

      <h3>Finalidades primarias (necesarias para el servicio)</h3>
      <ul>
        <li>Registrar y autenticar su perfil de usuario en la plataforma.</li>
        <li>
          Gestionar la inscripción, acceso a contenidos, evaluaciones y progreso
          en los programas de capacitación.
        </li>
        <li>
          Emitir constancias, certificados de participación o acreditaciones
          laborales (ej. Formatos DC-3 ante la STPS, según corresponda).
        </li>
        <li>
          Reportar los avances académicos a la empresa u organización
          empleadora, en caso de capacitaciones corporativas contratadas por su
          centro de trabajo.
        </li>
        <li>
          Brindar soporte técnico y resolver dudas operativas sobre la
          plataforma.
        </li>
      </ul>

      <h3>Finalidades secundarias (opcionales)</h3>
      <p>
        De acuerdo con el Artículo 14 del Reglamento, sus datos podrán ser
        utilizados para las siguientes finalidades no esenciales, a las cuales
        puede negarse sin que ello afecte la prestación del servicio principal:
      </p>
      <ul>
        <li>
          Enviarle notificaciones, promociones e información sobre nuevos
          cursos, webinars o programas formativos.
        </li>
        <li>
          Realizar encuestas de satisfacción para mejorar la calidad del
          catálogo de cursos.
        </li>
      </ul>
      <p>
        Si no desea que sus datos sean utilizados para las finalidades
        secundarias, puede manifestar su negativa desde este momento enviando un
        correo a <a :href="`mailto:${CORREO}`">{{ CORREO }}</a>.
      </p>

      <h2>IV. Transferencias de Datos Personales</h2>
      <p>
        Sus datos personales no serán vendidos, cedidos ni transferidos a
        terceros ajenos al Responsable sin su consentimiento previo. Sin
        embargo, de conformidad con lo establecido en el Artículo 37 de la Ley y
        el Artículo 10 de su Reglamento, el Responsable podrá realizar
        transferencias de datos sin el consentimiento del titular en los
        siguientes supuestos:
      </p>
      <ol>
        <li>
          <strong>Cumplimiento Legal y Obligaciones Laborales:</strong> a
          autoridades competentes (tales como la Secretaría del Trabajo y
          Previsión Social - STPS) para el registro de capacitaciones y emisión
          de constancias normativas, o a la autoridad fiscal (SAT) para la
          emisión de comprobantes.
        </li>
        <li>
          <strong>Empresas Patrocinadoras/Empleadores:</strong> a la empresa o
          institución a la que pertenece el usuario, con la única finalidad de
          reportar el cumplimiento, calificaciones y estatus de capacitación
          cuando el curso sea costeado o requerido por dicho empleador.
        </li>
        <li>
          <strong>Relación Contractual:</strong> a terceros proveedores que
          actúen como Encargados en términos del Artículo 49 del Reglamento,
          específicamente para servicios de hosting, plataformas de streaming de
          video o infraestructura en la nube.
        </li>
      </ol>

      <h2>V. Derechos ARCO y Revocación del Consentimiento</h2>
      <p>
        Usted tiene derecho a Acceder, Rectificar, Cancelar u Oponerse (Derechos
        ARCO) al tratamiento de sus datos personales, así como a revocar el
        consentimiento otorgado. Para ejercer estos derechos, envíe una
        solicitud a: <a :href="`mailto:${CORREO}`">{{ CORREO }}</a>
      </p>
      <p>Su solicitud deberá contener:</p>
      <ol>
        <li>Nombre completo y correo electrónico con el que se registró.</li>
        <li>
          Copia de identificación oficial para acreditar su identidad (o la de
          su representante legal).
        </li>
        <li>
          Descripción clara y precisa de los datos personales respecto de los
          que se busca ejercer alguno de los derechos mencionados.
        </li>
      </ol>
      <p>
        El procedimiento de atención se rige por los Artículos 28 al 35 de la
        Ley y los Artículos 86 al 102 de su Reglamento. Recibirá una respuesta
        sobre la procedencia de su solicitud en un plazo máximo de 20 días
        hábiles, el cual podrá ser ampliado por una sola vez por un periodo
        igual, siempre que así lo justifiquen las circunstancias (Artículo 32 de
        la Ley).
      </p>
      <p>
        En caso de considerar que su derecho a la protección de datos personales
        ha sido vulnerado, le informamos que puede acudir ante el Instituto
        Nacional de Transparencia, Acceso a la Información y Protección de Datos
        Personales (INAI).
      </p>

      <h2>VI. Retención y Cancelación de Datos</h2>
      <p>
        Sus datos personales e historial académico serán conservados durante el
        tiempo que mantenga activa su cuenta en la plataforma y durante el
        periodo legalmente requerido para respaldar la validez de los
        certificados emitidos o para auditorías laborales/fiscales.
      </p>
      <p>
        Una vez concluida la relación, cumplidos los plazos legales o solicitada
        la cancelación de la cuenta, sus datos entrarán en un periodo de bloqueo
        y posteriormente serán eliminados de nuestros sistemas de manera segura.
      </p>

      <h2>VII. Medidas de Seguridad</h2>
      <p>
        El Responsable ha implementado medidas de seguridad técnicas,
        administrativas y físicas para proteger sus datos personales contra
        daño, pérdida, alteración, destrucción o uso, acceso o tratamiento no
        autorizado, entre las que se incluyen:
      </p>
      <ul>
        <li>Cifrado de comunicaciones mediante protocolos HTTPS/TLS.</li>
        <li>Encabezados de seguridad HTTP (HSTS, CSP, X-Frame-Options).</li>
        <li>Control de acceso basado en roles y credenciales autenticadas.</li>
        <li>Almacenamiento seguro de contenidos y expedientes en la nube.</li>
      </ul>

      <h2>VIII. Uso de Cookies y Tecnologías de Rastreo</h2>
      <p>
        Esta plataforma utiliza archivos temporales y almacenamiento local
        (localStorage y cookies esenciales) exclusivamente para:
      </p>
      <ul>
        <li>Mantener la sesión activa del alumno durante su aprendizaje.</li>
        <li>Guardar el progreso en tiempo real de los videos o lecciones.</li>
        <li>
          Almacenar preferencias del sistema (como idioma o tema claro/oscuro).
        </li>
      </ul>
      <p>
        No se utilizan cookies de rastreo publicitario ni se comparten datos de
        navegación con redes de anuncios de terceros.
      </p>

      <h2>IX. Cambios al Aviso de Privacidad</h2>
      <p>
        El Responsable se reserva el derecho de efectuar modificaciones o
        actualizaciones a este Aviso de Privacidad en cualquier momento,
        derivadas de novedades legislativas, políticas internas o nuevos
        requerimientos para la prestación de nuestros servicios. Dichos cambios
        serán publicados en este mismo portal con la fecha de última
        actualización correspondiente. Se recomienda consultarlo periódicamente.
      </p>

      <p class="aviso-marco">
        <strong>Marco Normativo:</strong> este Aviso de Privacidad fue elaborado
        en estricto cumplimiento de la Ley Federal de Protección de Datos
        Personales en Posesión de los Particulares, publicada originalmente en
        el DOF el 5 de julio de 2010 (con su última reforma publicada en el DOF
        el 14 de noviembre de 2025), y su Reglamento vigente publicado el 21 de
        diciembre de 2011.
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

/* Encabezado normativo y nota de cierre: mismo tratamiento visual porque
 * cumplen la misma función, enmarcar el texto sin competir con él. */
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

/* La versión técnica es la que se guarda en la cuenta al aceptar. Se muestra
 * para poder cotejarla con el registro si alguien reclama. */
.aviso-version-id { opacity: 0.7; }

.aviso h2 {
  margin: 34px 0 8px;
  font-size: 1.05rem;
}

.aviso h3 {
  margin: 22px 0 6px;
  font-size: 0.95rem;
  color: var(--text);
}

.aviso p, .aviso li { font-size: 0.95rem; }
.aviso ul, .aviso ol { padding-left: 1.4em; margin: 0 0 14px; }
.aviso li { margin-bottom: 6px; }
.aviso a { color: var(--brand); text-decoration: underline; overflow-wrap: anywhere; }
</style>
