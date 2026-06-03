# BOOTSTRAP — Inicialización excepcional del repositorio

## Naturaleza

Este documento deja constancia **permanente** de que el primer cambio del
repositorio `github.com/Nidael1/VuhmikGO` fue una **inicialización
excepcional** que precede a la disciplina normal de ejecución
`un issue = una rama = un PR = un commit`.

Este bootstrap **no es un issue**. No tiene número de issue. No cierra
issue alguno.

Después del bootstrap, **toda** modificación al repositorio sigue
estrictamente la disciplina de issues definida en `12_execution.md` y los
documentos canónicos del proyecto.

---

## Alcance funcional (precisión obligatoria)

Este bootstrap no implementa funcionalidad de producto, Core, Shaders ni
Asteroides. No crea entidades de dominio, endpoints, migraciones ni
lógica de negocio. Solo inicializa el repositorio para permitir
ejecución trazable de issues posteriores.

---

## Justificación

El plan de ejecución (`12_execution.md`) define los issues `#1`–`#66`.
Issue #1 (`Definir entidades Core base`) requiere que el repositorio
exista, que `go.mod` declare el módulo `github.com/Nidael1/VuhmikGO` y
que las decisiones arquitectónicas (ADRs) estén versionadas, para que su
criterio de validación "el modelo compila / corre" pueda evaluarse.

Crear un "Issue de bootstrap" en GitHub alteraría la numeración (GitHub
asigna números secuencialmente desde `#1`) y desincronizaría el plan
documental. Por eso el bootstrap se ejecuta como cambio **visible y
documentado** pero **sin número de issue**, mediante commit inicial
directo a `main` y este documento como constancia.

---

## Alcance del bootstrap (lo que el commit inicial incluye)

Únicamente lo necesario para que Issue #1 pueda ejecutarse:

- `go.mod` declarando `module github.com/Nidael1/VuhmikGO`.
- `.gitignore` mínimo Go.
- `README.md` con punteros a documentación canónica.
- `BOOTSTRAP.md` (este documento).
- `docs/adr/ADR-0001-stack-go.md`.
- `docs/adr/ADR-0002-shaders-por-pais-y-modo-generico.md`.
- `docs/adr/ADR-0003-estructura-go-idiomatica.md`.

---

## Prohibiciones explícitas del bootstrap

El bootstrap **NO** incluye:

- Código de dominio, aplicación, infraestructura ni delivery.
- Migraciones SQL.
- Entidades Core (eso es Issue #1).
- Estados Core (eso es Issue #2).
- Modelo de tenant o configuración.
- Catálogos de Shaders o Asteroides.
- API, controladores HTTP, workers.
- Dependencias externas en `go.mod` (queda en stdlib pura).

---

## Disciplina a partir del bootstrap

Después del commit inicial de bootstrap a `main`, toda modificación
posterior cumple:

1. Existe un Issue en GitHub con número.
2. Existe una rama dedicada con prefijo `issue/<n>-<slug>`.
3. Existe un PR contra `main` cuyo cuerpo inicia con `Closes #<n>`.
4. El alcance del cambio coincide exactamente con el alcance del issue
   en `12_execution.md`.
5. El commit usa la convención `[sprint X.Y][issue #N] descripción`.
6. El PR se mergea solo si la validación documentada del issue se
   cumple.

Cualquier cambio futuro que no cumpla estas reglas viola el freeze y
requiere ADR.

---

## Registro

- Fecha del bootstrap: 2026-06-02.
- Rama del bootstrap: `main` (commit inicial directo).
- Commit del bootstrap: `[sprint 0][bootstrap] repository initialization`.
- Mecanismo: **commit inicial directo a `main`, sin Pull Request**.

### Por qué commit directo y no Pull Request

En el momento del bootstrap el repositorio estaba recién creado y `main`
no tenía ningún commit. Un Pull Request requiere una rama base existente
contra la cual comparar; una rama sin commits no existe como referencia en
el remoto. Por lo tanto, el primer commit de un repositorio no puede pasar
por un Pull Request: alguien debe establecer `main` con el commit inicial.

Este commit inicial directo a `main` ES la inicialización excepcional que
este documento registra. La disciplina de Pull Requests
(`un issue = una rama = un PR = un commit = un cierre`) aplica de forma
completa **a partir del Issue #1**, que sí se ejecuta con rama dedicada,
Pull Request y `Closes #1`.

Este documento permanece en el repositorio como evidencia auditable de
que la excepción ocurrió una sola vez y bajo condiciones controladas.
