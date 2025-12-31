# Developer Agent Guide — Jenkins Go TUI

Este documento está orientado a agentes de IA y desarrolladores para entender rápidamente la arquitectura, estructura del repositorio, decisiones técnicas, prácticas de calidad y el estado del proyecto.

## Project Overview

**Jenkins Go TUI** es una interfaz de usuario en terminal (TUI) moderna para monitorizar y operar **Jenkins** de forma eficiente: estado general, ejecuciones (builds), nodos/agentes conectados, exploración por vistas y jobs, historial paginado y acceso a logs/artefactos.

Objetivos principales:

* Experiencia **rápida** (low-latency), **reactiva** (actualizaciones sin bloquear la UI) y **usable con teclado**.
* Operación **read-only por defecto**, con acciones explícitas (rebuild, abort, enable/disable job, etc.) protegidas por confirmaciones.
* Uso de la **API oficial de Jenkins** (REST / JSON endpoints) y mecanismos oficiales (crumb issuer para CSRF cuando aplique).

### Config

* **Config Path (Linux/macOS)**: `~/.config/jenkins-tui/config.toml`
* **Config Path (Windows)**: `%APPDATA%\jenkins-tui\config.toml` (o equivalente según librería)
* **Go Target Version**: Go 1.22+ (mínimo recomendado: 1.21)
* **Distribución**: binario único, cross-platform (Linux/macOS/Windows)

---

## Requisitos funcionales (TUI completa)

### Tabs obligatorias

1. **Dashboard (central)**

   * Estado actual del Jenkins (salud general, versión si disponible, uptime si se puede obtener).
   * Builds en ejecución (en curso) y últimos builds relevantes.
   * Cola (queue) y bloqueos.
   * Nodos/agents conectados y capacidad/ocupación (executors).
   * Indicadores: éxito/fallo reciente, builds “building”, nodos offline, trabajos deshabilitados, etc.

2. **Vistas (Views)**

   * Listado de vistas con **búsqueda rápida** (filtro incremental).
   * Al seleccionar una vista: listar jobs de esa vista (con estado/color, último build, “building”, etc.).
   * Acciones típicas: abrir job, ver último build, filtrar por nombre/estado, fijar vista favorita.

3. **Histórico de Builds**

   * Historial paginado con teclas:

     * **RePag** (PageUp): página anterior
     * **AvPag** (PageDown): página siguiente
   * Posibilidad de:

     * Ver detalles del build (result, timestamp, duration, causa si disponible, cambios si aplica).
     * Ver **logs** (console output) con scroll, búsqueda dentro del log, y opciones (tail/follow).
     * Enlaces a artefactos (si se habilita en acciones read-only: listar artefactos y descargar a ruta local opcional).

### Tabs recomendadas (para completar la TUI)

* **Job Details** (panel o tab secundaria): configuración básica, parámetros, últimos builds, salud.
* **Node Details**: detalles de un nodo (offline reason, labels, executors).
* **Settings / Connections**: configuración, perfiles (varios Jenkins), test de conexión, gestión credenciales.
* **Help / Keymap**: cheat-sheet de teclas, atajos por contexto.

---

## Technical Stack (equivalentes Go del stack Rust)

### Framework TUI (MVU)

* **TUI Framework**: `bubbletea` (modelo MVU: Model-View-Update).
* **Componentes**: `bubbles` (list, table, viewport, textinput, paginator, spinner).
* **Styling**: `lipgloss` (estilos), opcional `termenv` si se requiere control fino.
* **Input/teclas**: gestionado por Bubble Tea (captura de KeyMsg), con mapeo centralizado.

### Concurrencia / Async

* **Goroutines + context.Context** para tareas en background.
* **Canales** para emisión de mensajes al modelo (eventos, actualizaciones, errores).
* **Rate limiting / backoff**:

  * Recomendado: `golang.org/x/time/rate` para limitar llamadas.
  * Recomendado: `cenkalti/backoff/v4` (o implementación simple interna) para reintentos.

### Cliente HTTP

* `net/http` (cliente nativo) con:

  * timeouts estrictos (connect, TLS, overall request).
  * transporte con keep-alive, pooling.
  * soporte de proxy si el entorno lo requiere.
* Serialización: `encoding/json`
* Config: `toml` (p.ej. `BurntSushi/toml`) o `koanf` (con backends toml/yaml/env). Mantenerlo simple y determinista.

### Logging / Observabilidad (local)

* `log/slog` (Go estándar) con niveles y output configurable (texto/JSON).
* Un “debug panel” opcional para inspeccionar últimos errores y latencias.

---

## Jenkins Official API — Principios y endpoints clave

**Normas:**

* Priorizar endpoints JSON oficiales (`/api/json`) y campos `tree=...` para reducir payload.
* Para acciones que modifican estado: respetar **CSRF** mediante **crumb issuer** cuando Jenkins lo requiera.
* Autenticación: típicamente **Basic Auth** con usuario + **API Token**. Evitar almacenar contraseñas.

### Endpoints típicos (orientativos)

* **Root info**: `/api/json`
* **Vistas**: `/api/json?tree=views[name,url]`
* **Vista detalle**: `/view/{viewName}/api/json?tree=jobs[name,url,color]`
* **Job detalle**: `/job/{jobName}/api/json?...`
* **Build detalle**: `/job/{jobName}/{buildNumber}/api/json?...`
* **Console log**: `/job/{jobName}/{buildNumber}/consoleText`
* **Queue**: `/queue/api/json`
* **Nodes/Computers**: `/computer/api/json?tree=computer[displayName,offline,temporarilyOffline,numExecutors,executors[currentExecutable[url]],assignedLabels[name],offlineCauseReason]`
* **Crumb** (si aplica): `/crumbIssuer/api/json`

Nota: la modelización debe ser tolerante a instalaciones con plugins, campos faltantes y diferencias de versión.

---

## UX / Interacción (teclado)

### Navegación general

* `Tab` / `Shift+Tab`: siguiente/anterior tab
* `1..9`: salto directo a tab (si se habilita)
* `Esc`: volver atrás / cerrar modal
* `?`: abrir ayuda contextual (keymap)

### Dashboard

* `r`: refresh manual
* `Enter`: detalle del item seleccionado (build/job/node)
* `f`: filtros rápidos (por ejemplo: solo fallos, solo building)

### Vistas

* `/`: activar búsqueda rápida
* `Enter`: seleccionar vista / abrir jobs
* `Backspace`: limpiar filtro
* `Ctrl+p`: command palette (opcional recomendado)

### Histórico Builds

* `RePag` (PageUp): página anterior
* `AvPag` (PageDown): página siguiente
* `Enter`: ver detalles / logs
* `l`: abrir logs del build seleccionado
* `s`: toggle “follow/tail” si el build está en ejecución
* `g/G`: inicio/fin (si se usa viewport)

### Modales / Acciones

* Confirmaciones obligatorias para acciones destructivas: `y/n`
* Mensajes de error no bloqueantes, con panel de “últimos eventos”.

---

## Arquitectura recomendada

### Patrón MVU (Bubble Tea)

* `Model` contiene el estado: tab actual, selección, filtros, datos cacheados, estado de red (loading/errors), paginación.
* `Update(msg)` procesa:

  * input (teclas)
  * resultados de background (mensajes `DataUpdate`, `ErrMsg`, `ToastMsg`)
  * ticks (auto-refresh opcional)
* `View()` renderiza el layout: tabs + panel principal + status bar + overlays.

### Background workers

* Un “scheduler” por tipo de dato:

  * Dashboard refresh (queue, running builds, nodes) con frecuencia configurable.
  * Vistas y jobs: cache + refresh bajo demanda o al entrar en tab.
  * Historial: paginación basada en parámetros internos (offset/limit) y, si Jenkins no soporta paginación directa, paginación local sobre builds recuperados incrementalmente.
* Todos los workers deben:

  * respetar `context.Context` (cancelación al cambiar de tab, salir, o cambiar conexión).
  * aplicar rate limit y backoff ante errores transitorios.

### Caching y coherencia

* Cache en memoria por:

  * vistas
  * jobs por vista
  * builds por job (últimos N)
  * nodes snapshot
* Estrategia:

  * “stale-while-revalidate”: mostrar cache inmediato y refrescar en background.
  * marcadores de “última actualización” en UI.

---

## Directory Structure (Go)

Sugerencia de estructura orientada a escalabilidad y tests:

* `cmd/jenkins-tui/`

  * `main.go` — bootstrap (flags, config, logging, init tea program)
* `internal/app/`

  * `model.go` — Model global (tabs, estado, routing)
  * `update.go` — Update(msg) y dispatch de comandos
  * `view.go` — composición de vistas
  * `keymap/` — mapeo de teclas, ayuda contextual
* `internal/ui/`

  * `tabs/`

    * `dashboard/` (componentes, view, update)
    * `views/`
    * `builds/`
  * `components/` (table wrappers, statusbar, toasts, modales)
  * `theme/` (paleta y estilos lipgloss)
* `internal/jenkins/`

  * `client.go` — cliente HTTP, auth, crumb, middlewares
  * `endpoints.go` — construcción de URLs, helpers `tree=`
  * `models/` — structs JSON (root, view, job, build, node, queue)
* `internal/config/`

  * `config.go` — carga/validación config, perfiles
* `internal/testutil/`

  * `httptest_server.go` — helpers de mock Jenkins
  * `golden/` — snapshots/golden tests para vistas (si se adopta)
* `Makefile`
* `.github/workflows/ci.yml` (recomendado)
* `AGENTS.md` (este documento)

---

## Config Schema (sugerido)

Ejemplo conceptual (toml):

* `active_profile = "prod"`
* `[profiles.prod]`

  * `base_url = "https://jenkins.company.com"`
  * `username = "user"`
  * `api_token = "..."` (idealmente referenciado desde keychain/env, no plano)
  * `insecure_skip_tls_verify = false` (evitar salvo entornos controlados)
  * `timeout_seconds = 15`
  * `auto_refresh_seconds = 10`
  * `max_builds_per_job = 200`
  * `max_log_bytes = 200000` (para evitar explotar memoria)
  * `rate_limit_rps = 5`

**Recomendación de seguridad**:

* Permitir `api_token_env = "JENKINS_TOKEN"` y no persistir tokens en disco por defecto.
* En entornos corporativos, considerar integración con keychain (opcional, no bloqueante para MVP).

---

## Feature Set recomendado (ampliación)

### Dashboard

* KPIs:

  * builds en curso (count)
  * queue length
  * nodos offline
  * ratio de fallos últimos N builds (si se agrega)
* Panel “Running Builds”:

  * job, build number, duración parcial, ejecutor/nodo (si se conoce), link.
* Panel “Nodes”:

  * nombre, offline, executors total/ocupados, labels, reason.
* Panel “Alerts”:

  * últimas fallas
  * nodos recién caídos
  * cola bloqueada

### Views tab

* Lista de vistas con búsqueda incremental.
* Al seleccionar vista:

  * tabla de jobs con columnas: nombre, estado/color, last build result, building, last duration.
* Acciones read-only:

  * abrir job details
  * abrir “último build”
* Acciones opcionales (con confirmación):

  * habilitar/deshabilitar job (si API y permisos lo permiten)
  * trigger build (parametrizado si se soporta)

### Builds History tab

* Selector de “scope”:

  * por job actual
  * por vista (agregado: builds de jobs de la vista; requiere estrategia de recolección)
* Paginación:

  * PageUp/PageDown mueve páginas; mostrar “Página X / Y (estimado)” si Y no es conocido.
* Logs:

  * cargar incremental (si se implementa) o truncado por bytes.
  * búsqueda dentro del log.
  * follow/tail para builds running: refresco cada N segundos.

---

## Key Technical Decisions (Go)

1. **MVU consistente**: Bubble Tea como núcleo. Cualquier operación remota se modela como “Command” que devuelve mensajes; la UI nunca bloquea.
2. **Cliente Jenkins robusto**:

   * timeouts, retries con backoff y clasificación de errores (4xx vs 5xx).
   * soporte crumb automático cuando sea requerido.
3. **Minimización de payload**: uso de `tree=` para limitar campos.
4. **Compatibilidad Jenkins**:

   * tolerancia a nulls/campos ausentes.
   * parsing defensivo.
5. **Seguridad por defecto**:

   * read-only (MVP).
   * acciones mutantes siempre con confirmación y chequeo de permisos/errores claros.
6. **Calidad obligatoria: tests por cada función relevante** (ver sección de Testing).

---

## Testing Strategy (OBLIGATORIO)

Regla del proyecto:

> **Toda nueva función relevante debe venir acompañada de tests.**
> Relevante = lógica de negocio, parsing/modelos, cliente HTTP, paginación, filtros, construcción de endpoints, gestión de errores, y cualquier comportamiento observable por el usuario.

### Tipos de tests

1. **Unit tests (rápidos, deterministas)**

   * `internal/jenkins/endpoints.go`: construcción de URLs y encoding de rutas.
   * `internal/jenkins/models`: parseo JSON con casos reales y edge cases.
   * Filtros (búsqueda rápida, predicados building/failure).
   * Paginación: cálculo de offsets, límites, y estado UI.

2. **HTTP client tests con `httptest`**

   * Mock server que simule:

     * 200 con payloads típicos
     * 401/403 (auth)
     * 500/502 (reintentos)
     * crumb requerido (403 + header/indicador según Jenkins), luego crumb OK
   * Verificar:

     * headers auth
     * timeouts no se “cuelgan”
     * backoff y límites de reintentos

3. **Golden tests (opcional pero recomendado)**

   * Snapshot del render (View()) para estados clave:

     * dashboard vacío/cargando/error
     * vista con filtro
     * logs truncados y con búsqueda
   * Los golden deben ser estables (evitar timestamps dinámicos sin normalizar).

4. **Integration tests (recomendado en CI)**

   * Jenkins en contenedor (docker) para validar endpoints básicos.
   * Si no es viable en todos los entornos, al menos un pipeline nightly.

### Cobertura mínima sugerida

* Cliente Jenkins: >80% en paquetes críticos (`internal/jenkins`, filtros, paginación).
* UI: cubrir estados y rutas principales; golden selectivo.

---

## Development Commands (propuestos)

* `make run` — ejecutar en local
* `make test` — `go test ./...`
* `make test-race` — `go test -race ./...`
* `make lint` — `golangci-lint run`
* `make fmt` — `gofmt -w .` + `goimports` (si se adopta)
* `make vet` — `go vet ./...`
* `make build` — build multi-OS (si se configura)
* `make ci` — fmt + lint + test

---

## Current State & Known Issues (plantilla viva)

1. **Riesgo de “model bloat”**: si toda la navegación/acciones se concentran en `internal/app/model.go`, dividir por tabs y componentes desde el principio.
2. **Diferencias entre Jenkins**: campos y permisos varían mucho (plugins, seguridad). Mantener parsing defensivo y mensajes de error accionables.
3. **Logs grandes**: la consola puede ser enorme. Debe existir truncado por bytes y/o carga incremental para evitar consumo excesivo.
4. **Paginación real**: Jenkins no ofrece paginación uniforme para todo. Donde no exista, implementar paginación local sobre datasets incrementales (con límites configurables).
5. **Auto-refresh**: debe ser configurable y suspenderse cuando el usuario está leyendo logs (para no “pelear” con el viewport).

---

## Roadmap sugerido (orden recomendado)

1. MVP read-only:

   * Conexión + auth + root + views + jobs list
   * Dashboard con queue/nodes/running builds
   * Historial por job + logs básicos con truncado
2. UX avanzada:

   * Command palette, toasts, help panel, búsqueda mejorada
   * Cache SWR y auto-refresh estable
3. Acciones seguras:

   * trigger build (parametrizado), abort build, enable/disable (con confirmación)
4. Integración CI + Jenkins docker tests

---

## Definition of Done (DoD) por feature

Una funcionalidad se considera “terminada” cuando:

* Está implementada con UI consistente y navegación clara.
* Tiene **tests** (unit + http mock si aplica).
* Maneja errores (auth, red, permisos) con mensajes claros.
* No bloquea la UI (operaciones remotas en background).
* Respeta límites de memoria (logs truncados / incremental).
* Está documentada (actualización de este AGENTS.md si cambia arquitectura o keybindings).

---

## Nota final para agentes de IA

Cuando añadas o modifiques comportamiento:

* Prioriza **robustez** (Jenkins heterogéneo), **responsividad** (no bloquear UI) y **testabilidad** (interfaces pequeñas, cliente mockeable).
* Evita “smart features” sin telemetría local y sin tests.
* Si dudas entre dos diseños: elige el que sea más fácil de testear y de mantener por tabs.

Si quieres, puedo devolverte esta guía ya “lista para pegar” como `AGENTS.md` con una sección de **Keybindings detallada** (tabla por tab) y una propuesta de **interfaces Go** (`JenkinsClient` + comandos Bubble Tea) para asegurar testabilidad desde el primer commit.

