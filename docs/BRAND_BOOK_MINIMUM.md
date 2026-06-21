# BRAND BOOK MINIMO — VuhmikGO post-MVP

## Fecha
2026-06-12

## Fuente
Derivado del Brand Book Maestro v2.0 con ajustes para modo claro
como primario y acentos desaturados para entorno medico/enterprise.

## Paleta principal

  Obsidiana      #090C10   sidebar, header, modo dark
  Fondo claro    #F8FAFB   fondo principal de la app
  Superficie     #FFFFFF   cards, formularios, modales
  Texto          #1A1F2E   texto principal (no negro puro)
  Texto suave    #64748B   labels, metadata, fechas, ayudas

## Paleta de acento

  Turquesa       #00C8D4   accion primaria, highlights, foco, info
  Jade           #00B885   success, confirmaciones, estado emitido
  Azul Clinico   #1E6FA8   links, acciones secundarias, datos

## Paleta de estado

  Warning        #FFB020   advertencias, estados pendientes
  Error          #FF4D6D   errores, anulaciones, acciones destructivas

## Tipografia

  Fuente principal: Sora (Google Fonts)
  Fuente campos de texto: Inter (Google Fonts)

  H1:      32px / 700 / lh 40px — titulos de pagina
  H2:      24px / 700 / lh 32px — secciones
  H3:      20px / 600 / lh 28px — subsecciones
  Body:    16px / 400-500 / lh 24px — texto principal
  Small:   14px / 400 / lh 20px — ayudas, labels
  Caption: 12px / 400 / lh 16px — notas legales, footers

## Reglas de uso

  Turquesa, Jade y Warning: usar con texto Obsidiana, no con blanco
  Azul Clinico: el mas seguro para texto clickeable sobre fondo claro
  Error: usar con texto blanco

## Tokens CSS

  :root {
    --color-obsidian:       #090C10;
    --color-background:     #F8FAFB;
    --color-surface:        #FFFFFF;
    --color-text:           #1A1F2E;
    --color-text-muted:     #64748B;
    --color-turquoise:      #00C8D4;
    --color-jade:           #00B885;
    --color-clinical-blue:  #1E6FA8;
    --color-warning:        #FFB020;
    --color-error:          #FF4D6D;
    --app-bg:               var(--color-background);
    --app-surface:          var(--color-surface);
    --app-header-bg:        var(--color-obsidian);
    --app-sidebar-bg:       var(--color-obsidian);
    --text-primary:         var(--color-text);
    --text-secondary:       var(--color-text-muted);
    --action-primary-bg:    var(--color-turquoise);
    --action-primary-text:  var(--color-obsidian);
    --state-success:        var(--color-jade);
    --state-warning:        var(--color-warning);
    --state-error:          var(--color-error);
    --state-info:           var(--color-turquoise);
    --font-brand:           "Sora", system-ui, sans-serif;
    --font-body:            "Inter", system-ui, sans-serif;
  }

## Estados de evidencia

  Draft:       Texto suave (#64748B)
  Emitida:     Jade (#00B885)
  Anulada:     Error (#FF4D6D)
  Reemplazada: Azul Clinico (#1E6FA8)
  Export:      Turquesa (#00C8D4)

## Regla de mantenimiento

  Ninguna vista Vue declara colores HEX directamente.
  Siempre usar variables CSS: var(--token-name)
