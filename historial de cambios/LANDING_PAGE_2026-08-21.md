# Landing Page VUHMÍK — 2026-08-21

**Autora:** Nida (NDT — Next Dev Tech)  
**Archivo:** `landing/index.html`  
**Rama:** `main` (commits directos)

---

## Resumen

Creación de la landing page pública de VUHMÍK para médicos independientes y vendedores.
Página autocontenida en HTML/CSS/JS, sin dependencias externas salvo Google Fonts.
Publicada también como artifact en claude.ai para revisión en vivo.

---

## Decisiones de diseño

**Paleta:** Dark-first — Obsidiana `#090C10` como ground principal (coherente con la app).
Turquesa `#00C8D4` reservado solo para CTAs principales. Jade `#00B885` para estados sellados.

**Tipografía:** Sora (display/headings — brand existente) + Inter (body/datos) +
JetBrains Mono para las huellas de verificación y etiquetas técnicas. El mono actúa
como motivo visual de la inmutabilidad sin requerir que el lector entienda tecnología.

**Tema:** Toggle light/dark en el nav con persistencia en localStorage.

**Demo card animada:** El hero muestra una nota de consulta real (paciente ficticio)
con la huella de verificación apareciendo con animación CSS. Grunda el claim abstracto
"inalterable" en algo concreto y visible.

---

## Secciones

| Sección | Propósito |
|---|---|
| Hero | Claim principal + demo card animada |
| Stats bar | 3 datos duros: 71% hospitales privados, 2027, 642 startups |
| El problema | 3 hechos concretos sobre el decreto y el mercado |
| La solución | 3 diferenciadores en lenguaje clínico |
| Cómo funciona | 3 pasos: registrar → sellar → compartir |
| Editable vs. sellado | Comparativa directa con la competencia |
| Módulos | 9 módulos disponibles + 1 en roadmap |
| Precio | $600/mes apertura, $1,000 regular tachado |
| CTA / Lista de espera | Formulario de correo con confirmación visual |
| Footer | Brand + nota regulatoria |

---

## Iteraciones de la sesión

### v1 — Estructura base
- Hero bipartito, secciones de problema/solución/diferenciador/módulos/precio/CTA.
- Precio inicial: referencia de mercado $183 MXN (competencia).

### v2 — Lenguaje sin jerga técnica
Feedback: "no cuenta con lenguaje que cualquier doctor o vendedor pueda entender sin ser de TI".
- Eliminado: `sha256:...`, `IPS/FHIR R4`, `ADR-0006`, `Core append-only`, `NOM-024-SSA3-2012` como código.
- Reemplazado por: "huella de verificación", "compatible con el sistema nacional", "diseñado para el ritmo del consultorio".
- Precio actualizado: $600/mes apertura (antes $183 referencia de mercado), $1,000 precio regular.
- Badge "Precio de apertura — plazas limitadas" + ahorro calculado $4,800/año.

### v3 — Lenguaje regulatorio ambiguo
Feedback: no tenemos certificación legal — el lenguaje no debe afirmar lo que no está formalmente reconocido.

Regla aplicada: nunca "cumple", nunca "válida bajo" — siempre orientativo.

| Reemplazado | Por |
|---|---|
| "Cumple NOM-024" | "Orientado a NOM-024" |
| "Cumple NOM-004" | "Orientado a NOM-004" |
| "Cumple la norma oficial" | "Estructurado bajo los criterios de" |
| "Receta con validez legal" | "En línea con los campos que la regulación establece" |
| "Cada módulo cumple un requerimiento" | "Diseñado bajo los criterios que la regulación vigente establece" |
| "Certificación CENETEC en proceso" | "En proceso de certificación ante autoridades competentes" |

---

## Decisión de precio documentada

- **Precio regular:** $1,000 MXN/mes (posicionamiento por encima del mercado — diferenciador legal justifica el premium).
- **Precio de apertura:** $600 MXN/mes — para los primeros médicos registrados, se mantiene fijo mientras la suscripción esté activa.
- **Argumento de ventas:** "Entra hoy a $600 y ese precio no sube mientras sigas activo."

---

## Commits

| Hash | Descripción |
|---|---|
| `01be93f` | feat: landing page inicial — hero, secciones, demo card animada |
| `f54b6c9` | landing: lenguaje sin jerga técnica + precio $600/$1,000 |
| `b2628c9` | landing: lenguaje regulatorio ambiguo sin afirmar certificación |

---

## Archivo

```
landing/
└── index.html    ← página completa autocontenida (~1,200 líneas)
```

---

*NDT — Next Dev Tech. 2026-08-21.*
