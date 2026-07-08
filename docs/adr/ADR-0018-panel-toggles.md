# ADR-0018 — Panel de administracion: toggles por cuenta

## Estado
Aceptado

## Fecha
2026-06-24

## Contexto

El registro de capacidades (ADR-0017) define que modulos tiene activos cada
cuenta, pero no existe ningun mecanismo para que un administrador los gestione
sin tocar codigo ni base de datos directamente. Sin un panel, activar o
desactivar un modulo para un medico requeriria una intervencion manual en la
BD, lo que es lento, propenso a errores y no auditable.

El panel de toggles es la interfaz administrativa que escribe en
TENANT_CAPABILITIES (plano de datos, ADR-0017). Es la superficie de escritura
mas privilegiada del sistema despues de las migraciones: cambiar un toggle
cambia que puertas al Core existen para ese medico.

El modelo de referencia es Odoo: el administrador ve la lista de modulos
disponibles por cuenta y los prende o apaga segun las necesidades del medico,
en vez de tener todo activo por defecto.

## Decision

### Alcance estrictamente administrativo-comercial

El panel de toggles opera EXCLUSIVAMENTE sobre el registro de capacidades.
Su universo de accion es:

  - Lista de cuentas (medicos registrados): identificador, estado, plan.
  - Modulos disponibles por cuenta (solo los publicados en MODULES).
  - Toggle activo/inactivo por modulo por cuenta.
  - Conteo de modulos activos por cuenta.
  - Estado de la cuenta: activa / suspendida.
  - Plan y costo asociado a los modulos activos.

El panel NUNCA accede a:
  - Contenido del expediente (notas, recetas, alergias, etc.).
  - Datos de pacientes (nombres, diagnosticos, PHI de cualquier tipo).
  - Metricas de uso clinico (cuantas consultas dio el medico, etc.).

Esta frontera es absoluta e invariante. Cruzarla convierte el panel
administrativo en una ventana al expediente, lo que viola el aislamiento
multi-tenant y expone PHI.

### Bandera is_admin

Se introduce la bandera `is_admin` en la tabla `users`. Un usuario con
`is_admin = true` puede acceder al panel de toggles. Un usuario sin ella
no puede ver ni las rutas del panel.

El admin no es un tenant medico: opera sobre todos los tenants pero sin
pertenecer a ninguno. Sus rutas viven bajo `/admin/` con middleware propio
que verifica `is_admin`.

### Catalogo visible: solo lo publicado

El panel muestra unicamente los modulos con `publication_status = 'publicado'`
en MODULES para el rubro correspondiente. Los modulos en desarrollo o de otro
rubro son invisibles para el admin. Esto garantiza que el admin no puede
activar accidentalmente un modulo incompleto o fuera del rubro contratado.

### Modelo tipo Odoo: apagado por defecto

Una cuenta nueva no tiene ningun modulo activo. El admin activa lo que el
medico necesita. El sistema no asume que el medico quiere todo.

Ventaja operativa: el medico paga solo por lo que usa.
Ventaja de seguridad: superficie minima activa = menor riesgo.
Ventaja comercial: expansion de ingreso por cuenta al activar modulos.

### Suspension de cuenta

El admin puede suspender una cuenta por falta de pago u otro motivo
administrativo. La suspension:

  - Bloquea el login del medico (el middleware rechaza el JWT del tenant
    suspendido antes de llegar a ningun handler).
  - NO borra, NO altera y NO vuelve inaccesibles los datos del expediente.
  - Es reversible: reactivar la cuenta restaura el acceso completo.

El expediente permanece integro, append-only, aunque la cuenta este
suspendida. El medico no pierde su historial clinico.

### Super-admin de rubro: diferido

El super-admin que define que rubros y modulos existen en la plataforma
NO se construye en v1 (ADR-0020). En v1 la publicacion de modulos se
hace por migracion. El panel de toggles opera sobre lo ya publicado;
no tiene ni la capacidad ni la UI para publicar modulos nuevos.

### Rutas del panel

  GET  /admin/accounts              lista de cuentas con estado y conteos
  GET  /admin/accounts/:tenant_id   detalle de cuenta: modulos y toggles
  POST /admin/accounts/:tenant_id/modules/:module_id/enable   activa
  POST /admin/accounts/:tenant_id/modules/:module_id/disable  desactiva
  POST /admin/accounts/:tenant_id/suspend    suspende la cuenta
  POST /admin/accounts/:tenant_id/activate   reactiva la cuenta

Todas las rutas requieren `is_admin = true`. Middleware dedicado.
Ninguna ruta del panel lee o escribe PHI.

## Dependencias

  - ADR-0017: el panel escribe en TENANT_CAPABILITIES (plano de datos)
              y lee MODULES para el catalogo visible.
  - ADR-0019: el panel de metricas es una ruta separada de solo lectura;
              no comparte handlers ni middleware con el de toggles.
  - ADR-0020: el super-admin (diferido) sera el que publique modulos;
              este panel solo activa lo ya publicado.

## Estado de implementacion

  Implementado. Rutas y nombres de handler difieren de los propuestos
  originalmente pero cubren la misma funcionalidad; ver detalle.
  Migracion 000012_admin_flags (is_admin, is_suspended en users).
  admin_handlers.go, AdminMiddleware en router.go, AdminView.vue.

    - Migracion 000012: columna is_admin (bool default false) e
      is_suspended (bool default false) en tabla users. El usuario
      dev@vuhmik.com se marca is_admin = true por seed.
    - Middleware AdminMiddleware (router.go): verifica is_admin = true
      en el JWT antes de cualquier handler bajo /api/v1/admin/*;
      rechaza con 403 si no cumple.
    - Handler HandleAdminTenants (GET /api/v1/admin/tenants): lista
      tenants con estado y conteos. Sin PHI. Equivalente a la ruta
      propuesta GET /admin/accounts.
    - Handler HandleAdminCapabilityToggle (POST /api/v1/admin/capabilities):
      un solo endpoint para activar/desactivar modulo por tenant, en vez
      de dos rutas separadas enable/disable — misma funcionalidad,
      forma distinta.
    - Handler HandleAdminSuspend (POST /api/v1/admin/suspend): un solo
      endpoint con flag is_suspended, en vez de dos rutas separadas
      suspend/activate — misma funcionalidad, forma distinta.
    - Middleware de suspension: verificacion de IsSuspended en
      auth_handlers.go (HandleLogin) — bloquea el login del tenant
      suspendido antes de emitir JWT. No borra ni altera datos.
    - Frontend: AdminView.vue — vista de administracion con toggles,
      sin acceso a datos clinicos. Ampliada en issue #233 con secciones
      de metricas y actividad (ADR-0019, ADR-0023).

  Nota: las rutas reales (/api/v1/admin/tenants, /capabilities, /suspend)
  no coinciden literalmente con las propuestas en la decision
  (/admin/accounts/:tenant_id/modules/:module_id/enable, etc.). La
  funcionalidad y las garantias (fail-closed, sin PHI, is_admin
  obligatorio) se cumplen; el nombrado de rutas quedo mas compacto
  de lo planeado.

## Consecuencias

  El admin puede gestionar modulos por cuenta sin tocar codigo ni BD.
  Activar o desactivar un modulo es un cambio de dato auditado, no un
  deploy.
  La suspension bloquea acceso sin perder datos — el medico puede
  retomar donde lo dejo al reactivar.
  El catalogo solo muestra modulos publicados: imposible activar algo
  en desarrollo accidentalmente.
  El panel no expone PHI en ningun camino de ejecucion.
  La bandera is_admin es la unica diferencia entre un medico y un admin
  en el modelo de datos — simple y auditable.
