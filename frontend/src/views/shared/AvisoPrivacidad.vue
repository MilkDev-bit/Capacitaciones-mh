<script setup lang="ts">
/**
 * Aviso de privacidad.
 *
 * El texto lo proporcionó el responsable. Sobre esa base se corrigieron las
 * referencias legales, que apuntaban a la LFPDPPP de 2010: esa ley quedó
 * ABROGADA por el Transitorio Segundo, fracción I, del decreto que publicó la
 * ley vigente el 20 de marzo de 2025. Cambiaron el marco, la autoridad y la
 * numeración de todos los artículos citados.
 *
 * Cada artículo que aparece abajo se contrastó con el texto vigente publicado
 * por la Cámara de Diputados (última reforma DOF 14-11-2025). Aun así, esto no
 * sustituye la revisión de un abogado.
 *
 * Al modificar CUALQUIER línea hay que subir AVISO_VERSION en utils/aviso.ts y
 * la fecha de "Última actualización". Si no se sube, quedará registrado en la
 * base que las personas aceptaron una versión que ya no es la publicada.
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
        Posesión de los Particulares (en adelante, la «Ley»), publicada en el
        Diario Oficial de la Federación el 20 de marzo de 2025, con su última
        reforma publicada el 14 de noviembre de 2025.
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
        <li>
          Clave Única de Registro de Población (CURP) y ocupación específica,
          exclusivamente cuando se emite la constancia DC-3.
        </li>
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
          Comprobantes de pago. Los datos de su tarjeta los captura y procesa
          directamente el proveedor de pagos; el Responsable no los almacena.
        </li>
      </ul>

      <p class="aviso-nota">
        No se recaban datos personales sensibles en los términos del artículo 3,
        fracción VI de la Ley.
      </p>

      <h2>III. Finalidades del Tratamiento</h2>
      <p>
        El tratamiento de los datos personales se realiza bajo los principios de
        licitud, finalidad, lealtad, consentimiento, calidad, proporcionalidad,
        información y responsabilidad, previstos en el <strong>artículo 5</strong>
        de la Ley.
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
        Sus datos podrán ser utilizados para las siguientes finalidades no
        esenciales, a las cuales puede negarse sin que ello afecte la prestación
        del servicio principal:
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

      <h2>IV. Medios para Limitar el Uso o Divulgación de sus Datos</h2>
      <p>
        Además de la negativa señalada en el apartado anterior, usted dispone de
        los siguientes medios, conforme al <strong>artículo 15, fracción IV</strong>
        de la Ley:
      </p>
      <ul>
        <li>
          <strong>Comunicaciones promocionales:</strong> cada correo incluye un
          enlace para darse de baja, y puede solicitarlo también al correo de
          contacto.
        </li>
        <li>
          <strong>Visibilidad de su perfil:</strong> la fotografía de perfil es
          opcional y puede retirarla en cualquier momento desde su cuenta.
        </li>
        <li>
          <strong>Verificación de constancias:</strong> la consulta pública por
          folio muestra únicamente su nombre, el curso, la empresa y la fecha de
          emisión. No expone su CURP, su RFC ni su correo.
        </li>
      </ul>

      <h2>V. Transferencias de Datos Personales</h2>
      <p>
        Sus datos personales no serán vendidos, cedidos ni transferidos a
        terceros ajenos al Responsable sin su consentimiento previo. De
        conformidad con el <strong>artículo 36</strong> de la Ley, el Responsable
        podrá realizar transferencias sin el consentimiento del titular en los
        siguientes supuestos:
      </p>
      <ol>
        <li>
          <strong>Cumplimiento legal (artículo 36, fracción I):</strong> a
          autoridades competentes, tales como la Secretaría del Trabajo y
          Previsión Social (STPS) para el registro de capacitaciones y emisión
          de constancias normativas, o el Servicio de Administración Tributaria
          (SAT) para la emisión de comprobantes fiscales.
        </li>
        <li>
          <strong>Empresas empleadoras (artículo 36, fracciones IV y VII):</strong>
          a la empresa o institución a la que pertenece el usuario, con la única
          finalidad de reportar el cumplimiento, calificaciones y estatus de
          capacitación cuando el curso sea costeado o requerido por dicho
          empleador.
        </li>
        <li>
          <strong>Defensa de derechos (artículo 36, fracciones V y VI):</strong>
          cuando resulte necesario para la procuración o administración de
          justicia, o para el reconocimiento o defensa de un derecho en un
          proceso judicial.
        </li>
      </ol>
      <p>
        Conforme al <strong>artículo 35</strong> de la Ley, se hace constar la
        siguiente cláusula: <strong>al aceptar este aviso, usted consiente las
        transferencias descritas en este apartado.</strong> Si no desea
        consentirlas, puede manifestarlo al correo de contacto; tenga en cuenta
        que la transferencia a la STPS es indispensable para emitir su
        constancia DC-3, de modo que su negativa impediría expedirla.
      </p>

      <h2>VI. Proveedores que tratan datos por cuenta del Responsable</h2>
      <p>
        Para operar la plataforma intervienen proveedores de alojamiento,
        almacenamiento en la nube, transmisión de video, envío de correo y
        procesamiento de pagos. Estos proveedores actúan como
        <strong>personas encargadas</strong> en términos del
        <strong>artículo 3, fracción XII</strong> de la Ley: tratan los datos
        únicamente por cuenta y bajo instrucciones del Responsable, y no pueden
        usarlos para fines propios.
      </p>
      <p>
        Su intervención <strong>no constituye una transferencia</strong>: el
        artículo 35 de la Ley excluye expresamente a la persona encargada del
        supuesto de transferencia a terceros. Se informa aquí por transparencia.
      </p>

      <h2>VII. Derechos ARCO y Revocación del Consentimiento</h2>
      <p>
        Conforme a los <strong>artículos 21 a 26</strong> de la Ley, usted tiene
        derecho a Acceder, Rectificar, Cancelar u Oponerse (Derechos ARCO) al
        tratamiento de sus datos personales, así como a revocar el consentimiento
        otorgado. Para ejercer estos derechos, envíe una solicitud a:
        <a :href="`mailto:${CORREO}`">{{ CORREO }}</a>
      </p>
      <p>
        Su solicitud deberá contener, conforme al <strong>artículo 28</strong> de
        la Ley:
      </p>
      <ol>
        <li>
          Su nombre y domicilio, o cualquier otro medio para recibir la
          respuesta.
        </li>
        <li>
          Los documentos que acrediten su identidad o, en su caso, la
          personalidad e identidad de su representante.
        </li>
        <li>
          La descripción clara y precisa de los datos personales respecto de los
          que se busca ejercer el derecho, salvo que se trate del derecho de
          acceso.
        </li>
        <li>La descripción del derecho ARCO que pretende ejercer.</li>
        <li>
          Cualquier otro elemento o documento que facilite la localización de
          sus datos personales.
        </li>
      </ol>
      <p>
        Conforme al <strong>artículo 31</strong> de la Ley, recibirá respuesta
        sobre la procedencia de su solicitud en un plazo máximo de
        <strong>veinte días hábiles</strong>, y de resultar procedente se hará
        efectiva dentro de los quince días siguientes. Ambos plazos podrán
        ampliarse por una sola vez y por un periodo igual cuando las
        circunstancias del caso lo justifiquen.
      </p>
      <p>
        El ejercicio de los derechos ARCO es <strong>gratuito</strong>
        (artículo 34 de la Ley); solo podrán cobrarse los costos de
        reproducción, copias o envío.
      </p>
      <p>
        En caso de considerar que su derecho a la protección de datos personales
        ha sido vulnerado, puede acudir ante la
        <strong>Secretaría Anticorrupción y Buen Gobierno</strong>, autoridad
        competente en la materia conforme al artículo 3, fracción XV de la Ley,
        tras la extinción del Instituto Nacional de Transparencia, Acceso a la
        Información y Protección de Datos Personales.
      </p>

      <h2>VIII. Retención y Cancelación de Datos</h2>
      <p>
        Sus datos personales e historial académico serán conservados durante el
        tiempo que mantenga activa su cuenta en la plataforma y durante el
        periodo legalmente requerido para respaldar la validez de los
        certificados emitidos o para auditorías laborales y fiscales.
      </p>
      <p>
        Una vez concluida la relación, cumplidos los plazos legales o solicitada
        la cancelación de la cuenta, sus datos entrarán en un periodo de bloqueo
        y posteriormente serán eliminados de nuestros sistemas de manera segura.
      </p>

      <h2>IX. Medidas de Seguridad</h2>
      <p>
        Conforme al <strong>artículo 18</strong> de la Ley, el Responsable ha
        implementado medidas de seguridad administrativas, técnicas y físicas
        para proteger sus datos personales contra daño, pérdida, alteración,
        destrucción o uso, acceso o tratamiento no autorizado, entre las que se
        incluyen:
      </p>
      <ul>
        <li>Cifrado de comunicaciones mediante protocolos HTTPS/TLS.</li>
        <li>Encabezados de seguridad HTTP (HSTS, CSP, X-Frame-Options).</li>
        <li>Control de acceso basado en roles y credenciales autenticadas.</li>
        <li>
          Verificación en dos pasos mediante código de un solo uso para el
          cambio de contraseña.
        </li>
        <li>Almacenamiento seguro de contenidos y expedientes en la nube.</li>
      </ul>
      <p>
        En caso de una vulneración de seguridad que afecte de forma significativa
        sus derechos patrimoniales o morales, se le informará conforme al
        <strong>artículo 19</strong> de la Ley.
      </p>

      <h2>X. Uso de Cookies y Tecnologías de Rastreo</h2>
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

      <h2>XI. Cambios al Aviso de Privacidad</h2>
      <p>
        El Responsable se reserva el derecho de efectuar modificaciones o
        actualizaciones a este Aviso de Privacidad en cualquier momento,
        derivadas de novedades legislativas, políticas internas o nuevos
        requerimientos para la prestación de nuestros servicios.
      </p>
      <p>
        Los cambios se publicarán en esta misma página con su fecha de última
        actualización. Cuando el cambio sea sustancial, se le mostrará el aviso
        actualizado al iniciar sesión y se le solicitará aceptarlo de nuevo; la
        versión que usted aceptó queda registrada en su cuenta.
      </p>

      <p class="aviso-marco">
        <strong>Marco normativo:</strong> este Aviso de Privacidad se elaboró
        conforme a la Ley Federal de Protección de Datos Personales en Posesión
        de los Particulares publicada en el DOF el 20 de marzo de 2025, con su
        última reforma publicada el 14 de noviembre de 2025, que abrogó la ley
        homónima del 5 de julio de 2010. El Reglamento de la ley abrogada,
        publicado el 21 de diciembre de 2011, continúa aplicándose de forma
        supletoria en lo que no contradiga a la ley vigente, en tanto se expide
        el nuevo Reglamento.
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

.aviso-nota {
  margin: 0 0 12px;
  color: var(--muted);
  font-size: 0.88rem;
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
