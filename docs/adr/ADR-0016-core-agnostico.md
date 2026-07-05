# ADR-0016 — Core agnostico: contenido opaco y discriminador de tipo en Shader

## Estado
Aceptado

## Fecha
2026-06-24

## Contexto

El Core de VUHMIK fue construido como motor agnostico de dominio dentro de la
familia de sistemas que requieren registro probatorio e inmutable. La misma base
puede ser un ECE medico hoy, un ERP manana, un sistema notarial o un CRM, sin
reescribir el nucleo — solo cambiando que Shaders y Asteroides se activan.

Sin embargo, en la implementacion actual el Core tiene dos campos que filtran
semantica de dominio medico hacia adentro:

  `notes`      Campo de texto con contenido clinico. "Notas" es vocabulario
               medico. Un CRM de restaurantes no tiene "notas clinicas". Este
               campo rompe el agnosticismo del Core.

  `subject_id` Apunta logicamente a un paciente. El nombre "subject" es mas
               neutral, pero la forma en que se usa (con logica de busqueda
               por paciente dentro de los handlers) filtra conocimiento de
               dominio hacia la capa de entrega.

Adicionalmente, el ADR-0006 (versionado silencioso) y los ADR-0011 a 0015
(modulos clinicos) asumen que los registros clinicos tipados (receta, alergia,
diagnostico, vacuna, laboratorio) viviran como tipos de contenido sobre el
registro generico del Core. Pero las secciones "Estado de implementacion" de
esos ADR mencionan tablas tipadas separadas, lo que contradice el modelo
agnostico y obligaria a re-implementar las garantias de append-only, lifecycle
y hash en cada modulo.

Esta ADR resuelve la contradiccion de raiz y establece la regla definitiva.

## Decision

### El Core es agnostico de dominio, no de todo

El Core es agnostico del RUBRO (medicina, ERP, CRM, notarial), pero NO es
agnostico de la FORMA de tratar los datos. Solo sirve para dominios donde
la regla "se registra, se emite y no se altera; las correcciones son nuevos
registros" tiene sentido.

  Encaja: expediente clinico, asientos contables, actos notariales.
  No encaja: carrito de compras editable, borrador colaborativo.

La prueba para saber si un campo puede vivir en el Core:
  "Significaria lo mismo si VUHMIK fuera un CRM de restaurantes?"
  Si no, el campo no pertenece al Core.

### Contenido opaco: el campo `content`

El Core guarda un unico campo de contenido opaco: `content` (blob JSON).
El Core lo almacena, lo hashea y lo firma. NUNCA lo parsea ni lo interpreta.

El campo `notes` DESAPARECE del Core. Su contenido migra dentro del blob:

  Antes: notes = "Paciente refiere dolor abdominal..."
  Despues: content = {"type":"note","text":"Paciente refiere dolor..."}

### Discriminador de tipo: vive DENTRO del blob

El campo `type` que distingue una nota de una receta, una alergia o un
diagnostico vive DENTRO del blob `content`. El Core no lo conoce.

  Una nota:       content = {"type":"note",  "text":"..."}
  Una receta:     content = {"type":"prescription","medicamento":"..."}
  Una alergia:    content = {"type":"allergy","agente":"..."}
  Un diagnostico: content = {"type":"diagnosis","descripcion":"..."}
  Una vacuna:     content = {"type":"immunization","vacuna":"..."}
  Un laboratorio: content = {"type":"lab_result","estudio":"..."}

Es el Shader quien abre el blob, lee el `type` y sabe que hacer con el.
Los Asteroides muestran el contenido segun el tipo que el Shader expone.

### SubjectRef: clave de correlacion opaca

El campo `subject_id` se renombra a `subject_ref` para dejar claro que es
una clave de correlacion opaca, no una referencia a una entidad de dominio.

El Core la usa solo para indexar y agrupar registros del mismo sujeto.
NUNCA hay FK de `subject_ref` hacia `patients.id` en el Core.
El vinculo paciente <-> registro es LOGICO, resuelto por el Shader al
consultar por `tenant_id` + `subject_ref`.

### Hash sobre el blob opaco

El hash SHA-256 (ADR-0008) se calcula sobre el blob `content` como bytes
opacos mas la metadata del registro (id, tenant_id, subject_ref, state,
timestamps). El Core no necesita conocer el dominio para firmar: firma
el blob entero como unidad atomica.

### Garantias de append-only, lifecycle y hash: una sola vez

Las garantias no se re-implementan por modulo. El Core las da una sola vez
para todos los tipos de contenido. Una receta, una alergia y una nota son
el mismo registro generico — heredan append-only, lifecycle y hash gratis.

### Reinterpretacion de ADR-0011 a ADR-0015

Las secciones "Estado de implementacion" de los ADR-0011 a 0015 mencionan
tablas tipadas separadas (prescriptions, allergies, diagnoses, etc.).
Con este ADR se reinterpretan: NO son tablas en el Core. Son tipos de
contenido opaco que el Shader interpreta sobre el registro generico.

La migracion correspondiente es una sola tabla `registro` (o la tabla
`evidence` renombrada y limpiada) con el campo `content` blob, sin columnas
de dominio.

### Migracion forward-only

La migracion de `notes` a `content` es forward-only:
  1. Agregar columna `content` (nullable inicialmente).
  2. Poblar `content` envolviendo el valor de `notes`:
       content = '{"type":"note","text":"' || notes || '"}'
  3. Volver `content` NOT NULL.
  4. Eliminar columna `notes`.
  5. Renombrar `subject_id` a `subject_ref` (o agregar columna y migrar).

Cada paso es una migracion separada y forward-only.

## Dependencias

  - ADR-0006: el versionado silencioso (void+replace) se preserva; el Core
              sigue sin conocer el concepto de edicion.
  - ADR-0008: el hash se recalcula sobre el blob opaco.
  - ADR-0009: el traspaso exporta el contenido opaco; el Shader lo proyecta
              a IPS al generar el archivo de traspaso.
  - ADR-0010: el IPS vive en Shaders/export, no en el Core.
  - ADR-0011 a 0015: se reinterpretan sobre este modelo.
  - ADR-0017: el Shader consulta el registro de capacidades antes de
              interpretar el tipo de contenido.

## Estado de implementacion

  Implementado. El campo Content es un blob JSON opaco en entity.go.
  El Core nunca lo parsea. El discriminador de tipo vive dentro del blob;
  solo el Shader lo lee. subject_ref es el campo canónico; notes no existe.
  Ejecutado en el desarrollo post-MVP previo a esta sesión.
    - Hash: recalcular sobre content + metadata en lugar de fields
      nombrados de dominio.
    - Tests: actualizar fixtures y aserciones.

## Consecuencias

  El Core queda 100% agnostico: sin vocabulario de dominio en ninguna
  columna ni campo de la entidad.
  Las garantias de append-only, lifecycle y hash se dan una sola vez
  para todos los modulos clinicos futuros.
  Los modulos clinicos (receta, alergia, etc.) son tipos de contenido,
  no tablas separadas — sin retrabajo de garantias por modulo.
  El mismo Core puede ser reutilizado para un ERP o sistema notarial
  sin modificar una sola linea del nucleo.
  La migracion de notes a content es un cambio destructivo del esquema
  que requiere cuidado en produccion (pasos separados, forward-only).
