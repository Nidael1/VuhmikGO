# VUHMÍK — Bitácora de Cambios

Explicación en lenguaje llano de qué se cambió, por qué y qué problema resolvió.
Para el detalle técnico y el número de commit ver → [REGISTRO_ISSUES.md](REGISTRO_ISSUES.md)
Para el historial de capacidades por sprint ver → [CHANGELOG.md](../versiones/CHANGELOG.md)

---

## 2026-08-19

### Issue #164 — El widget de agenda ahora muestra el estado de cada consulta con colores

**¿Qué se agregó?**
Cada consulta en el widget "Hoy" del menú lateral ahora tiene un punto de color y una franja de color en el borde izquierdo que indica su estado:

- 🔴 **Rojo** — pendiente (el médico aún no la ha atendido).
- 🟢 **Verde** — atendida (la consulta fue emitida con registros).
- 🟡 **Amarillo** — cancelada (la consulta fue anulada). El nombre del paciente aparece tachado.

El sistema también fue mejorado por dentro: ahora existe un endpoint específico para traer las consultas del día en todos sus estados, en lugar de traer todas las del historial y filtrarlas en el navegador. Esto hace el widget más rápido y correcto.

---

### Issue #163 — Widget "Agenda de hoy" en el menú lateral

**¿Qué se agregó?**
En el menú lateral izquierdo, debajo de los enlaces de navegación (Pacientes, Consultas, Recetas), ahora aparece una sección llamada "Hoy" que muestra todas las consultas del día en curso.

Cada consulta muestra la hora en que fue registrada, el nombre corto del paciente (por ejemplo "Juan G.") y una palomita (✓) si ya fue emitida. Las consultas emitidas se ven más tenues para que visualmente el médico identifique de un vistazo cuántas le quedan pendientes. En el encabezado aparece el conteo "completadas/total" del día.

**¿Cómo se actualiza?**
Cada vez que el médico navega a una pantalla diferente — por ejemplo, después de registrar una nueva consulta — el widget se refresca automáticamente. No requiere recargar la página.

**¿Por qué importa?**
El médico puede saber en cualquier momento cuántas consultas tiene hoy y en qué estado están, sin salir de la pantalla donde está trabajando.

---

### Issue #156 — Corrección de bug: editar nombre y sexo del paciente al mismo tiempo podía corromper datos

**¿Qué estaba pasando?**

En la pantalla de detalle del paciente, el médico puede editar el nombre haciendo clic en el lápiz, y también puede editar el sexo de la misma forma. Cada vez que se guardaba uno de esos campos, el sistema mandaba al servidor *todos* los datos del paciente (nombre, sexo, fecha de nacimiento, CURP) aunque solo se hubiera cambiado uno.

Esto causaba un problema silencioso: si el médico editaba el nombre y el sexo casi al mismo tiempo, las dos peticiones llegaban al servidor con información desactualizada. La segunda petición sobreescribía el campo que ya había guardado la primera — y el sistema no avisaba ningún error. El expediente quedaba con un dato incorrecto en la base de datos sin que nadie se diera cuenta.

Además, al cambiar el sexo desde el selector desplegable, el sistema enviaba la petición dos veces seguidas por un error técnico en los eventos del selector (`@change` y `@blur` se disparaban juntos).

**¿Qué se cambió?**

- Ahora cuando se edita el nombre, solo se manda el nombre al servidor. Nada más.
- Cuando se edita el sexo, solo se manda el sexo al servidor. Nada más.
- Se eliminó el error que causaba la doble petición al cambiar el sexo.
- Si el servidor responde con un error al intentar guardar el sexo, el selector vuelve a abrirse para que el médico pueda intentarlo de nuevo sin tener que buscarlo.

**¿Por qué importa?**

El expediente clínico es un documento legal (NOM-004-SSA3-2012). Un dato incorrecto — incluso algo tan básico como el sexo o el nombre — puede tener implicaciones médicas y legales. La corrección garantiza que cada campo se guarda de forma independiente y sin riesgo de pisarse entre sí.

### Issue #157 — El selector de sexo enviaba la información dos veces al servidor

**¿Qué estaba pasando?**
Al cambiar el sexo del paciente en el selector, el sistema enviaba la petición al servidor dos veces seguidas porque había dos eventos activos al mismo tiempo (`@change` y `@blur`). Si una petición fallaba y la otra no, el estado de la pantalla quedaba diferente al de la base de datos.

**¿Qué se cambió?**
Se dejó solo un evento activo. Ahora el servidor recibe la petición una sola vez al seleccionar el valor.

---

### Issue #158 — El sistema podía fallar al editar el sexo si el médico navegaba muy rápido

**¿Qué estaba pasando?**
Si el médico abría el expediente y de inmediato cambiaba de pantalla mientras el selector de sexo estaba activo, el sistema intentaba usar datos del paciente que ya no estaban en memoria, causando un error técnico silencioso.

**¿Qué se cambió?**
Ahora el sistema verifica que los datos del paciente existan en memoria antes de enviar cualquier petición. Si no existen, cierra el selector sin hacer nada.

---

### Issue #159 — Al escribir mal la fecha, el error decía algo incorrecto

**¿Qué estaba pasando?**
Si se enviaba una fecha en formato incorrecto (por ejemplo `31/12/2000` en lugar de `2000-12-31`), el servidor respondía con el mensaje "la fecha no puede ser futura", lo cual era confuso porque el problema real era el formato, no la fecha en sí.

**¿Qué se cambió?**
Ahora el servidor distingue los dos casos y responde con el mensaje correcto: "debe tener formato YYYY-MM-DD" si el formato es incorrecto, o "no puede ser futura" si el formato es correcto pero la fecha es futura.

---

### Issue #160 — El servidor aceptaba un valor incorrecto de sexo y no avisaba

**¿Qué estaba pasando?**
Si alguien enviaba un valor inválido para el campo sexo (por ejemplo `"X"` en lugar de `M`, `F` o `I`), el servidor respondía con 200 (éxito) pero en realidad no guardaba nada. Quien hiciera la petición creería que el cambio se guardó cuando en realidad fue ignorado.

**¿Qué se cambió?**
Ahora el servidor responde con error 400 y el mensaje "sexo debe ser M, F o I" si el valor no es válido. Esto aplica tanto al crear como al actualizar un paciente.

---

### Issue #161 — El selector de fecha bloqueaba "hoy" en algunos casos

**¿Qué estaba pasando?**
El límite máximo del campo de fecha de nacimiento se calculaba usando la hora universal (UTC). En países con zona horaria positiva (UTC+), entre medianoche y la hora de su zona, el sistema mostraba el día de ayer como máximo y no dejaba seleccionar "hoy".

**¿Qué se cambió?**
Ahora el límite se calcula con la hora local del dispositivo del médico, no con UTC. Para México (UTC-6) no había problema, pero el cambio garantiza consistencia en cualquier contexto.

---

### Issue #162 — Las reglas de validación del paciente ahora viven donde deben

**¿Qué estaba pasando?**
La validación de que la fecha de nacimiento no puede ser futura estaba escrita dentro del código que maneja las peticiones HTTP, no en el lugar donde viven las reglas médicas del sistema (el Shader `med_basic`). Esto significa que si en el futuro se crea otra forma de registrar pacientes (por ejemplo, una importación masiva), esas validaciones se saltarían sin que nadie se diera cuenta.

El archivo CLAUDE.md del proyecto es explícito: las validaciones clínicas deben vivir en el Shader, no en los handlers HTTP.

**¿Qué se cambió?**
Se creó una función `ValidatePatientDemographics` dentro del Shader `med_basic`. Esta función es la responsable de validar que la fecha tenga el formato correcto, que no sea futura, y que el sexo sea un valor válido (M, F o I). Los handlers HTTP ahora solo llaman a esa función y, si regresa un error, responden con código 400 al cliente. La lógica médica y la lógica de red están separadas.

También se redactó el ADR-0030, que es el documento formal donde se registra esta decisión de arquitectura con su contexto y consecuencias, para que cualquier persona que trabaje en el proyecto en el futuro entienda por qué está así y no lo revierta sin pensar.

**¿Por qué importa?**
Cualquier ruta nueva de ingreso de datos de paciente que llame esta función heredará automáticamente todas las reglas clínicas. No hay riesgo de que alguien las duplique mal o las omita.

---

*Para agregar una entrada nueva: copiar el bloque de fecha, rellenar el número de issue y explicar en lenguaje llano.*

---

## 2026-08-20

### Issues #165–#170 — El médico ahora puede agendar citas antes de que llegue el paciente

**¿Qué se agregó?**
Se implementó el módulo completo de agendamiento de citas (`scheduler_ui`), que existía como módulo en la arquitectura desde el inicio pero nunca había sido construido.

**¿Qué problema resuelve?**
Antes, el widget de agenda siempre decía "Sin consultas registradas" porque no había manera de reservar un horario con anticipación. El médico solo podía crear la consulta en el momento de atender al paciente. Ahora puede planear su día con anticipación.

**¿Qué puede hacer el médico ahora?**

1. **Desde el menú lateral → "Citas"**: ver todas sus citas ordenadas, filtrar por Hoy/Próximas/Todas/Canceladas, y crear una nueva seleccionando al paciente por nombre.

2. **Desde el perfil de un paciente**: botón "Agendar cita" (azul) en la parte superior que abre el formulario con el paciente ya pre-seleccionado.

3. **En cada cita agendada**: tres acciones disponibles — "Iniciar consulta" (abre el flujo de consulta y marca la cita como atendida automáticamente), "No se presentó" y "Cancelar".

4. **En el widget de hoy del menú lateral**: las citas agendadas para el día aparecen ahora con un punto azul, distinguiéndose de las consultas ya iniciadas (verde/rojo/amarillo).

**¿Qué NO cambia?**
El flujo de crear una consulta directamente (sin cita previa) sigue funcionando igual. Las citas no son expediente clínico — no pasan por el Core de evidencia ni por los Shaders.

