// Documento de terminos y condiciones para Mexico, rubro medicina.
//
// La clave de version codifica pais y rubro: <pais>-<rubro>-v<n>.
// Este documento NO es universal: referencia LFPDPPP, NOM-004, NOM-024,
// COFEPRIS y tribunales mexicanos. No aplica a tenants fuera de Mexico
// ni a rubros distintos de medicina.
//
// DEUDA ARQUITECTONICA: la resolucion de que documento corresponde a cada
// tenant pertenece al shader consent_policy_* (marcado FUTURO en
// 03_VUHMIK_SHADERS_reglas.md seccion 5.3). Hoy existe una sola opcion.
// Requiere ADR antes de agregar una segunda.
//
// Al publicar una version nueva: crear archivo v2, actualizar
// CurrentTermsVersion en internal/delivery/http/api/terms.go. Las
// aceptaciones de versiones previas quedan invalidadas automaticamente.

export const TERMS_VERSION = 'mx-medicine-v1'

export const TERMS_HTML = `
<p class="terms-updated">Última actualización: agosto 2026</p>

          <section>
            <h3>1. Naturaleza del servicio, partes y objeto</h3>
            <p>
              VUHMÍK es una herramienta de <strong>administración y organización de información</strong>.
              Opera como sistema de registro documental: recibe, estructura, conserva y permite
              recuperar la información que el Médico decide asentar.
            </p>
            <p>
              <strong>VUHMÍK no ejerce la medicina.</strong> En particular, el Servicio no:
            </p>
            <ul>
              <li>emite, sugiere ni valida diagnósticos</li>
              <li>calcula, recomienda ni verifica dosis o posologías</li>
              <li>evalúa interacciones medicamentosas, alergias ni contraindicaciones</li>
              <li>interpreta resultados de laboratorio, estudios de imagen ni signos vitales</li>
              <li>genera alertas clínicas ni recordatorios de decisión médica</li>
              <li>sustituye, complementa ni condiciona el criterio del profesional</li>
            </ul>
            <p>
              El Servicio no constituye un dispositivo médico ni software de apoyo a la decisión
              clínica. Su función es equivalente a la de un archivo estructurado: organiza lo que
              el Médico escribe, sin interpretar su contenido.
            </p>
            <p>
              Toda decisión clínica —diagnóstica, terapéutica o de seguimiento— es tomada
              exclusivamente por el Médico, con base en su juicio profesional, su formación y su
              relación directa con el paciente.
            </p>
            <p>
              El Servicio se ofrece bajo modalidad de suscripción mensual a médicos independientes
              en México. Al acceder y utilizar el Servicio, el profesional de la salud (en adelante
              "el Médico") acepta los presentes Términos y Condiciones en su totalidad.
            </p>
          </section>

          <section>
            <h3>2. Cumplimiento normativo</h3>
            <p>
              VUHMÍK está diseñado para apoyar el cumplimiento del Decreto de Digitalización de
              Salud 2026, la <strong>NOM-024-SSA3-2012</strong> y la <strong>NOM-004-SSA3-2012</strong>.
              El Médico es responsable de verificar que su práctica clínica cumple con la normativa
              vigente. El Servicio es una herramienta de apoyo y no sustituye el criterio clínico
              ni las obligaciones legales del profesional.
            </p>
          </section>

          <section>
            <h3>3. Integridad e inalterabilidad del expediente</h3>
            <p>
              Cada nota, receta o registro clínico asentado en VUHMÍK queda <strong>sellado de
              forma permanente</strong> en el momento de su emisión. El sistema conserva la versión
              original de todo documento sin posibilidad de borrado.
            </p>
            <p>
              Las correcciones al expediente se registran como <strong>enmiendas formales</strong>,
              de modo que el historial completo de modificaciones queda siempre visible y verificable,
              en cumplimiento del principio de integridad que exige la NOM-004. Este mecanismo
              convierte al expediente en VUHMÍK en una prueba legal defendible ante COFEPRIS o en
              un procedimiento por responsabilidad médica.
            </p>
            <p>
              El Médico no deberá intentar alterar ni suprimir registros clínicos por ningún medio.
              Cualquier acción en ese sentido podrá dar lugar a la suspensión inmediata de la cuenta.
            </p>
          </section>

          <section>
            <h3>4. Protección de datos personales y confidencialidad clínica</h3>
            <p>
              Los datos de los pacientes almacenados en VUHMÍK son datos personales sensibles de
              salud conforme a la <strong>Ley Federal de Protección de Datos Personales en Posesión
              de los Particulares (LFPDPPP)</strong>. El Médico, como responsable del tratamiento
              de dichos datos, se obliga a:
            </p>
            <ul>
              <li>Obtener el consentimiento informado del paciente conforme a la NOM-004.</li>
              <li>No compartir sus credenciales de acceso (usuario y contraseña) con terceros.</li>
              <li>Notificar a VUHMÍK de inmediato ante cualquier acceso no autorizado sospechado.</li>
              <li>Mantener el secreto profesional conforme al Código de Ética Médica.</li>
            </ul>
            <p>
              VUHMÍK no accede al contenido clínico de los expedientes salvo para la prestación
              del propio servicio. Los datos no se venden ni se utilizan con fines ajenos a los
              declarados en este documento.
            </p>
            <p>
              VUHMÍK puede generar estadísticas agregadas sobre el uso del Servicio con fines de
              mejora del producto, definición de precios y comunicación comercial. Estas
              estadísticas <strong>no contienen información clínica, datos de pacientes ni
              información que permita identificar al Médico o a sus pacientes</strong>.
            </p>
          </section>

          <section>
            <h3>5. Privacidad de la cuenta</h3>
            <p>
              La información de cada cuenta es estrictamente privada e independiente de cualquier
              otro usuario del sistema. Ningún médico puede ver ni acceder a los expedientes,
              pacientes o datos de otra cuenta. Ante cualquier ambigüedad en los permisos de
              acceso, el sistema deniega el acceso por defecto.
            </p>
          </section>

          <section>
            <h3>6. Portabilidad del expediente</h3>
            <p>
              El Médico puede exportar el expediente completo de sus pacientes en el formato
              internacional de Resumen del Paciente (<strong>IPS/FHIR</strong>), compatible con
              el sistema nacional de salud. La portabilidad es un derecho del Médico y de sus
              pacientes, y puede solicitarse en cualquier momento desde el perfil del sistema.
            </p>
          </section>

          <section>
            <h3>7. Suscripción y facturación</h3>
            <p>
              El Servicio opera bajo suscripción mensual por médico con módulos clínicos
              activables de forma independiente. La cuota vigente se informa al Médico antes de
              activar cualquier módulo o renovar la suscripción. El incumplimiento de pago por
              más de <strong>15 días naturales</strong> puede resultar en la suspensión del
              acceso, sin pérdida de información. Los expedientes se conservan por un mínimo de
              <strong>90 días</strong> tras la suspensión para permitir la exportación o la
              regularización.
            </p>
          </section>

          <section>
            <h3>8. Respaldos y disponibilidad</h3>
            <p>
              VUHMÍK realiza respaldos automáticos periódicos de la información clínica. El
              Servicio se presta con el objetivo de mantener alta disponibilidad, pero no garantiza
              operación ininterrumpida. Ante mantenimientos programados, se notificará al Médico
              con anticipación razonable. VUHMÍK no asume responsabilidad por interrupciones fuera
              de su control (fuerza mayor, fallas en servicios de telecomunicaciones).
            </p>
          </section>

          <section>
            <h3>9. Uso permitido</h3>
            <p>
              El Servicio es de uso exclusivo para médicos con cédula profesional vigente en el
              ejercicio de su práctica clínica privada. Queda expresamente prohibido:
            </p>
            <ul>
              <li>Usar el sistema sin título y cédula profesional habilitante para el ejercicio de la medicina.</li>
              <li>Compartir una misma cuenta entre dos o más profesionales.</li>
              <li>Registrar datos de pacientes sin su consentimiento informado.</li>
              <li>Usar el sistema para actividades contrarias a la ética médica o a la legislación mexicana.</li>
            </ul>
          </section>

          <section>
            <h3>10. Responsabilidad clínica</h3>
            <p>
              VUHMÍK es una herramienta de apoyo al registro clínico. El diagnóstico, la
              prescripción y toda decisión clínica son responsabilidad exclusiva del Médico. El
              Servicio no asume responsabilidad por actos u omisiones clínicas. La responsabilidad
              máxima del Servicio ante el Médico no excederá el monto pagado por suscripción en
              los tres meses anteriores al evento que origine la reclamación.
            </p>
          </section>

          <section>
            <h3>11. Modificaciones</h3>
            <p>
              VUHMÍK puede actualizar estos Términos notificando al Médico con al menos
              <strong>30 días naturales de anticipación</strong> mediante el correo registrado y
              un aviso en la plataforma. El uso continuado del Servicio tras esa fecha constituye
              aceptación de los nuevos términos.
            </p>
          </section>

          <section>
            <h3>12. Jurisdicción</h3>
            <p>
              Estos Términos se rigen por las leyes de los <strong>Estados Unidos Mexicanos</strong>.
              Cualquier controversia se someterá a la jurisdicción de los tribunales competentes
              de la Ciudad de México, con renuncia expresa a cualquier otro fuero que pudiera
              corresponder.
            </p>
          </section>

          <p class="terms-contact">
            Contacto: <strong>legal@vuhmik.mx</strong>
          </p>
`
