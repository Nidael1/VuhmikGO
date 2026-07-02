# VUHMÍK — Documento de estado para continuación en Cowork
**Fecha:** 28 de junio de 2026  
**Sprint activo:** 9.4  
**Repo:** github.com/Nidael1/VuhmikGO  
**Ruta local:** /Volumes/D/VuhmikGO  

---

## Descripción del proyecto

VUHMÍK es un SaaS de expediente clínico electrónico para médicos independientes en México.  
Cumple NOM-024-SSA3-2012 y NOM-004-SSA3-2012 desde el primer registro.

**Stack:**
- Backend: Go 1.26.3 — arquitectura hexagonal, Core agnóstico, CQRS con proyecciones
- Frontend: Vue 3 + TypeScript + Vite
- BD: PostgreSQL con migraciones secuenciales (go-migrate)
- Cache: Redis (refresh tokens)
- Patrón: Core append-only → Shaders (contexto clínico) → Asteroides (módulos del médico)

**Modelo de negocio:** SaaS tipo Odoo — núcleo base + módulos activables por tenant desde panel admin.

---

## Comandos de arranque

```bash
# Terminal 1 — Backend
cd /Volumes/D/VuhmikGO
lsof -ti :8080 | xargs kill -9 2>/dev/null; sleep 1
DATABASE_URL="postgres://localhost:5432/vuhmik_dev?sslmode=disable" \
JWT_SECRET="vuhmik-dev-secret-2026" \
REDIS_URL="redis://localhost:6379" \
go run ./cmd/vuhmik-api/ &

# Terminal 2 — Frontend
export NVM_DIR="/Volumes/D/nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
cd /Volumes/D/VuhmikGO/frontend && npm run dev

# Migraciones (si hace falta)
migrate -path database/migrations \
  -database "postgres://localhost:5432/vuhmik_dev?sslmode=disable" up
```

**Nota importante:** NVM está en /Volumes/D/nvm, NO en ~/.nvm. Siempre usar `export NVM_DIR="/Volumes/D/nvm"` antes de npm.

---

## Cuentas de desarrollo

| Email | Password | Rol |
|-------|----------|-----|
| dev@vuhmik.com | dev123456 | Admin → redirige a /admin |
| user3@prueba.com | password123 | Médico → redirige a /patients |

**Módulos activos para user3:** allergy, prescription, note, consultation  
(activados en BD; en producción el admin los activa desde /admin)

---

## Migraciones aplicadas

| # | Nombre | Contenido |
|---|--------|-----------|
| 000001–000009 | Core base | Evidence, users, patients, modules, capabilities |
| 000010 | professional_profiles | Perfil médico por rubro |
| 000011 | projections | note/allergy/prescription_projections |
| 000012 | admin_flags | is_admin, is_suspended en users |
| 000013 | metrics_snapshot | metrics_snapshot + activity_log |
| 000014 | profile_fields | universidad, direccion, telefono en professional_profiles |
| 000015 | note_vitals | signos vitales en note_projections + clinical_note_id en prescription_projections |
| 000016 | consultations | consultation_projections + consultation_id en note_projections |

---

## Arquitectura de archivos clave

### Backend
```
cmd/vuhmik-api/main.go                          — punto de entrada, inyección de dependencias
internal/auth/auth.go                            — JWT (Claims incluye IsAdmin), bcrypt
internal/core/evidence/                          — Core agnóstico append-only
internal/shaders/                               — AllergyShader, PrescriptionShader, ConsultationShader
internal/application/                           — Services: AllergyService, PrescriptionService, ConsultationService
internal/application/ports/                     — Interfaces: repositorios, proyecciones
internal/infrastructure/postgres/               — Adaptadores PostgreSQL
internal/infrastructure/redis/                  — Refresh tokens
internal/delivery/http/api/
  ├── router.go                                 — mux, JWTMiddleware, JWTOrQueryMiddleware, AdminMiddleware
  ├── deps.go                                   — struct Deps con todos los repositorios
  ├── auth_handlers.go                          — login, register, refresh, logout
  ├── patient_handlers.go                       — CRUD pacientes
  ├── evidence_handlers.go                      — draft, emit, void, replace, export
  ├── allergy_handlers.go                       — alergias
  ├── prescription_handlers.go                  — recetas
  ├── prescription_print.go                     — GET /prescriptions/:id/print → HTML imprimible
  ├── consultation_handlers.go                  — consultas
  ├── profile_handlers.go                       — perfil profesional
  └── admin_handlers.go                         — panel admin: tenants, capabilities, suspend
database/migrations/                            — migraciones secuenciales SQL
docs/adr/ADR-0001 a ADR-0024.md                — decisiones de arquitectura
```

### Frontend
```
frontend/src/
  app/stores/auth.ts                            — Pinia store (token, isAdmin, isAuthenticated)
  domain/types/                                 — auth.ts, patient.ts, evidence.ts, allergy.ts,
                                                  prescription.ts, consultation.ts
  infrastructure/api/httpClient.ts              — cliente HTTP con Authorization header
  infrastructure/repositories/                  — allergyRepository, evidenceRepository,
                                                  prescriptionRepository, consultationRepository,
                                                  patientRepository, authRepository
  presentation/layouts/AppLayout.vue            — sidebar: Pacientes · Consultas · Recetas · Mi perfil
  presentation/views/
    LoginView.vue                               — login único, redirige según is_admin
    AdminView.vue                               — panel admin: lista tenants, toggles módulos, búsqueda
    PatientListView.vue                         — lista de pacientes
    PatientDetailView.vue                       — expediente: alergias, recetas, consultas
    ConsultationListView.vue                    — lista global de consultas
    ConsultationNewView.vue                     — formulario unificado: signos vitales + nota + receta
    PrescriptionListView.vue                    — lista global de recetas
    EvidenceDraftView.vue                       — nueva nota clínica con signos vitales
    ProfileView.vue                             — perfil profesional del médico
  router/index.ts                               — rutas con guards requiresAuth y requiresAdmin
```

---

## Estado del Sprint 9.4

### Issues cerrados ✓
- **#149** — Migración 000012: is_admin + is_suspended
- **#150** — JWT con is_admin + AdminMiddleware + suspensión en login
- **#151** — Panel admin: toggles módulos, búsqueda, colapsable, redirección por rol
- **#152** — Migración 000013: metrics_snapshot + activity_log
- **#153** — Migración 000014: universidad, dirección, teléfono en perfil
- **#154** — Signos vitales en notas + fix CURP nullable + fix login/registro
- **#155** — Módulo consultas: backend completo + frontend (ADR-0024)

### HEAD actual
```
97e0e06 — [sprint 9.4][issue #155] frontend consultas lista nueva vista sidebar router adr-0024
```

### Cambios sin commitear (pendientes)
Los siguientes archivos tienen cambios desde el último commit:
- `frontend/src/presentation/views/ConsultationNewView.vue` — modal confirmación sin receta + window.open PDF
- `frontend/src/presentation/views/PatientDetailView.vue` — sección consultas con signos vitales
- `frontend/src/domain/types/prescription.ts` — campo consultation_id
- `internal/application/ports/prescription_projection_repository.go` — campo ConsultationID
- `internal/infrastructure/postgres/prescription_projection_repository.go` — consultation_id en queries
- `internal/delivery/http/api/prescription_print.go` — handler HTML imprimible (NUEVO)
- `internal/delivery/http/api/router.go` — JWTOrQueryMiddleware + prescriptionDispatcher + ruta /print

**Commit pendiente sugerido:**
```bash
git add -A
git commit -m "[sprint 9.4] pdf receta html imprimible modal sin receta patron consultas vitals adr-0024"
git push origin main
```

---

## Tarea en progreso al momento de transferir

**PDF de receta imprimible** — casi terminado, hay un bug de autenticación:

El handler `GET /api/v1/prescriptions/:id/print` devuelve HTML imprimible.  
El frontend llama `window.open('/api/v1/prescriptions/${rxId}/print?token=${authStore.token}', '_blank')`.

El middleware `JWTOrQueryMiddleware` en router.go ya está implementado para leer el token del query param:
```go
func JWTOrQueryMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("Authorization") == "" {
            if t := r.URL.Query().Get("token"); t != "" {
                r.Header.Set("Authorization", "Bearer "+t)
            }
        }
        JWTMiddleware(next)(w, r)
    }
}
```

La ruta en router.go usa:
```go
mux.HandleFunc("/api/v1/prescriptions/", JWTOrQueryMiddleware(prescriptionDispatcher))
```

**El bug:** el servidor sigue respondiendo UNAUTHORIZED cuando se abre la URL con ?token=...  
Posible causa: el JWTMiddleware interno del prescriptionDispatcher podría estar revalidando sin ver el header modificado, o el mux está enrutando mal.

**Para debuggear:**
```bash
# Obtener un token y probar directo
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user3@prueba.com","password":"password123"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")

# Listar recetas para obtener un ID
curl -s http://localhost:8080/api/v1/prescriptions \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool

# Probar print con query param
curl -v "http://localhost:8080/api/v1/prescriptions/RX_ID_AQUI/print?token=$TOKEN"
```

---

## Pendientes del sprint (por orden de prioridad)

1. **Fix bug PDF receta** — resolver UNAUTHORIZED en /print?token=
2. **PDF receta** — verificar que el HTML muestre datos del médico, paciente y signos vitales correctos
3. **PatientDetailView** — verificar que la sección Consultas muestra correctamente signos vitales + nota vinculada
4. **Admin: crear usuarios** — formulario en AdminView para dar de alta médicos con todos los datos obligatorios: email, password, nombre, cédula, especialidad, universidad, dirección, teléfono
5. **Panel de métricas** — frontend de /admin/metrics (ADR-0019): MRR, cuentas activas, conteos; el backend de GET /admin/metrics no está implementado aún
6. **MetricsWorker WAR-A** — precálculo cada 4h de la tabla metrics_snapshot

---

## Issues futuros (sprint 9.5+)

- Diagnósticos CIE-10 (ADR-0013)
- Inmunizaciones (ADR-0014)
- Resultados de laboratorio (ADR-0015)
- Perfil del médico editable desde el frontend (ProfileView.vue existe pero es solo lectura)
- Botón suspender/reactivar médico desde AdminView

---

## Decisiones de arquitectura relevantes (ADRs)

| ADR | Decisión |
|-----|----------|
| ADR-0006 | Void+replace: correcciones nunca borran el original |
| ADR-0016 | Core agnóstico: no conoce la clínica, solo registra evidencia |
| ADR-0018 | Panel de toggles: admin activa módulos por tenant |
| ADR-0019 | Panel de métricas: precálculo en worker, no en vivo |
| ADR-0021 | Perfil profesional por rubro (medicine) |
| ADR-0022 | CQRS con proyecciones de lectura |
| ADR-0024 | Módulo consulta: agrupa signos vitales + nota + receta bajo consultation_id |

---

## Notas importantes

- **CURP nullable:** el índice idx_users_curp_unique es UNIQUE. Si CURP está vacío se pasa NULL (no '') para evitar colisiones. Esto está resuelto en user_repository.go.
- **Token JWT expira en 15 minutos.** El frontend usa refresh tokens (7 días) para renovar.
- **Módulos:** un médico sin módulos activos recibe 422 al intentar usar funciones clínicas. El admin debe activarlos desde /admin.
- **Signos vitales:** se guardan en consultation_projections y note_projections. El PDF de receta los jala de la consulta vinculada via ConsultationID.
- **Frontend/frontend:** existe una carpeta duplicada `/Volumes/D/VuhmikGO/frontend/frontend/` con un HomeView.vue residual. Es basura histórica, ignorar.
