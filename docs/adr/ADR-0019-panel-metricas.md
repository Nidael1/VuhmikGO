# ADR-0019 — Panel de metricas de negocio (solo lectura)

## Estado
Propuesto

## Fecha
2026-06-24

## Contexto

El panel de toggles (ADR-0018) gestiona que modulos tiene activos cada cuenta.
Pero no existe un lugar donde el dueno de la plataforma pueda ver la salud del
negocio: cuantos medicos estan activos, cuanto ingresa, que modulos se usan,
cuales son las cuentas dormidas.

Esta informacion es critica para tomar decisiones comerciales (precio, roadmap,
churn) y para responder la pregunta de rentabilidad que el mercado exige desde
el dia uno.

El panel de metricas se separa del de toggles de forma deliberada y por una
razon de arquitectura: uno ESCRIBE (toggles, ADR-0018) y el otro solo LEE.
Mezclarlos pondria en el mismo lugar codigo que modifica el acceso al Core y
codigo que reporta — superficies con privilegios distintos no deben compartir
handlers ni middleware.

## Decision

### Solo lectura, sin efecto

El panel de metricas no modifica ningun dato del sistema. Es un informe.
No tiene handlers POST, PUT ni DELETE. No escribe en ninguna tabla.
Si el panel cae o da datos incorrectos, el sistema sigue funcionando igual.

### Frontera absoluta: agregados, nunca PHI

El panel muestra numeros y tendencias. NUNCA el contenido detras del numero.

  Permitido: el Dr. Garcia tiene 142 pacientes registrados.
  Prohibido: quienes son esos 142 pacientes o sus datos clinicos.

  Permitido: se emitieron 1,200 recetas este mes en la plataforma.
  Prohibido: que se receto, a quien, o cualquier dato del expediente.

Toda metrica es un agregado (COUNT, SUM, AVG) calculado sobre metadata
de los registros, sin exponer filas individuales ni contenido clinico.
Esto es coherente con la regla de observabilidad sin PHI/PII que ya
rige los logs y metricas del sistema (ADR-0001).

### Contenido del panel

  Por cuenta (fila por medico):
    - Identificador de cuenta (no nombre completo en v1, solo ID/email).
    - Estado: activa / suspendida.
    - Modulos activos y plan actual.
    - Costo mensual de la cuenta (desde TENANT_CAPABILITIES).
    - Conteo de pacientes registrados.
    - Conteo de registros emitidos por tipo (notas, recetas, etc.)
      en el periodo seleccionado.
    - Fecha de alta y fecha de ultimo registro emitido.

  Metricas agregadas de la plataforma:
    - Cuentas activas / suspendidas / en periodo de prueba.
    - MRR (Monthly Recurring Revenue): suma de costos de cuentas activas.
    - Churn del periodo: cuentas canceladas / total del periodo anterior.
    - Distribucion de modulos: cuantas cuentas tienen activo cada modulo.
    - Cuentas muy activas vs dormidas (sin registros en N dias).

### Precalculo por worker WAR-A

Las metricas NO se calculan en vivo en cada request del panel. Se
precalculan periodicamente por un worker WAR-A y se almacenan en una
tabla de snapshot `metrics_snapshot`.

Razon: calcular MRR, churn y conteos sobre toda la BD en cada carga
del panel escala mal y compite con las consultas del medico. El worker
corre fuera de la ventana de uso pico (de madrugada) y el panel lee
el snapshot, que es casi instantaneo.

El snapshot se marca con su timestamp de calculo para que el admin
sepa que tan reciente es el dato.

### Rutas del panel de metricas

  GET /admin/metrics                  resumen agregado de la plataforma
  GET /admin/metrics/accounts         lista de cuentas con conteos
  GET /admin/metrics/accounts/:id     detalle de una cuenta (conteos)
  GET /admin/metrics/modules          distribucion de uso por modulo

Todas las rutas requieren is_admin = true (mismo middleware que ADR-0018).
Ninguna ruta expone PHI.

### Separacion fisica de handlers

Los handlers de metricas viven en un archivo separado de los de toggles
(metrics_handlers.go vs admin_handlers.go). Comparten el middleware
AdminOnly pero no comparten logica ni repositorios de escritura.

## Dependencias

  - ADR-0017: lee TENANT_CAPABILITIES para modulos activos, plan y costo.
  - ADR-0018: comparte el middleware AdminOnly (is_admin); no comparte
              handlers ni repositorios de escritura.
  - Workers existentes: el MetricsPurgeWorker (issue #119) ya existe;
    el worker de precalculo de metricas es un nuevo worker WAR-A.

## Estado de implementacion

  No implementado.
  Requiere issues de implementacion con:
    - Migracion: tabla metrics_snapshot (id, calculated_at, payload JSON).
    - MetricsWorker: calcula MRR, churn, conteos por cuenta y por modulo;
      inserta snapshot; se ejecuta periodicamente (configurable, default 4h).
    - Handler GET /admin/metrics: lee ultimo snapshot y lo devuelve.
    - Handler GET /admin/metrics/accounts: lista de cuentas con conteos
      desde el snapshot; sin PHI.
    - Handler GET /admin/metrics/accounts/:id: detalle de una cuenta.
    - Handler GET /admin/metrics/modules: distribucion de uso por modulo.
    - Frontend: vista /admin/metrics con dashboard de negocio (MRR, churn,
      lista de cuentas, distribucion de modulos). Sin datos clinicos.

## Consecuencias

  El dueno de la plataforma tiene visibilidad de la salud del negocio
  sin acceder a datos clinicos de ningun medico.
  Las metricas precalculadas no compiten con las consultas del medico.
  La separacion fisica de handlers garantiza que el panel de metricas
  no puede adquirir capacidades de escritura por accidente o por refactor.
  El snapshot tiene timestamp visible: el admin sabe que tan reciente
  es el dato que esta viendo.
  En v1 el detalle de cuenta muestra ID/email, no nombre completo,
  para minimizar la superficie de datos personales en el panel.
