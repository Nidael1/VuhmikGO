# ADR-0020 — Super-admin: plano de control off-web (diferido)

## Estado
Propuesto

## Fecha
2026-06-24

## Contexto

El registro de capacidades (ADR-0017) separa dos planos de escritura:

  Plano de datos (admin, ADR-0018): activa modulos por cuenta. Vive en
  la web, lo opera el administrador comercial.

  Plano de control (super-admin): define que modulos EXISTEN y estan
  publicados, y a que rubro pertenecen. Es lo que convierte el mismo
  Core agnostico (ADR-0016) en un ECE medico hoy, o en un ERP manana.

En v1 el plano de control no tiene UI: la publicacion de modulos se
hace por migracion forward-only desde una maquina de confianza.

Este ADR documenta la decision de diferir el super-admin con UI, las
razones de seguridad que lo justifican, y como debe construirse cuando
sea necesario.

## Decision

### En v1: configuracion como codigo, sin superficie web

El super-admin no es una aplicacion web en v1. Publicar un modulo es
una migracion forward-only que el dueno corre desde su maquina de
confianza por SSH.

La app en produccion NO tiene ningun endpoint HTTP que escriba en
MODULES. Solo la lee. Aunque alguien comprometiera la app publica
por completo, no podria publicar modulos ni alterar el catalogo.

Ventajas:
  - Superficie de ataque del plano de control = cero en la web.
  - Toda publicacion queda como migracion versionada en git.
  - Sin costo de desarrollo de una UI que aun no se necesita.

### Cuando escale: app interna aislada, nunca en la web publica

Restricciones invariantes cuando exista UI:
  - Otro binario, otro proceso, otro despliegue.
  - Nunca fusionado con la app publica ni con el panel admin.
  - Accesible solo desde localhost o detras de VPN.
  - Sin autenticacion compartida con la app publica.

Leccion de SingHealth (2018): un control privilegiado mal aislado
sobre la misma superficie de ataque fue el vector de entrada.

### Lo que el super-admin controla

  - Publicar / despublicar modulos en MODULES.
  - Cambiar el rubro de un modulo (medico / erp / crm).
  - Cambiar el estado: en_desarrollo / publicado / deprecado.
  - Ver el catalogo completo incluidos los no publicados.

Lo que NO controla:
  - Activacion por cuenta (eso es el admin comercial, ADR-0018).
  - Datos clinicos o PHI de cualquier tipo.
  - Usuarios medicos o sus sesiones.

### Seed inicial en v1

Modulos pre-publicados en la migracion inicial de MODULES:

  note           Notas clinicas (ya implementado)
  prescription   Receta electronica (ADR-0011)
  allergy        Alergias e intolerancias (ADR-0012)
  diagnosis      Diagnosticos / lista de problemas (ADR-0013)
  immunization   Inmunizaciones / vacunacion (ADR-0014)
  lab_result     Resultados de laboratorio (ADR-0015)

## Dependencias

  - ADR-0016: el super-admin define los tipos del blob opaco.
  - ADR-0017: unico escritor de MODULES (plano de control).
  - ADR-0018: el panel de toggles opera sobre lo ya publicado.

## Estado de implementacion

  Diferido. No se construye en v1.
  En v1: seed de MODULES por migracion (rubro medico, 6 modulos).
  Futuro: app interna aislada (localhost/VPN).
  Requiere su propio ADR cuando se construya.

## Consecuencias

  Superficie de ataque del plano de control = cero en la web en v1.
  Toda publicacion de modulos es auditada en git como migracion.
  El diferimiento ahorra desarrollo que hoy no se necesita.
  El seed deja listos los 6 modulos del rubro medico desde el dia uno.
