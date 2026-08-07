# Waktu Sholat - Technical Design Document (TDD)

## 1. Technical Overview

### 1.1 System Description

Waktu Sholat is a Go 1.25 application that serves a server-rendered prayer times UI and, with a build tag, wraps the same HTTP server in a Wails desktop window. The front-end is templ components styled with Tailwind and made interactive with HTMX. Prayer times are computed locally via `hablullah/go-prayer`; qibla is a great-circle bearing formula. There is no JSON API and no client-side framework.

### 1.2 Technical Objectives

- One Go module, two build targets (web, desktop), one shared HTTP mux.
- Zero runtime external API calls for core features.
- Type-safe HTML via `templ`; no string-concatenated templates.
- One runnable self-check for the only non-trivial math (qibla).

---

## 2. Technology Stack

### 2.1 Core Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/a-h/templ` | v0.3.1020 | Type-safe HTML templating |
| `github.com/hablullah/go-prayer` | v1.1.1 | Prayer time calculation |
| `github.com/hablullah/go-juliandays` | v1.0.1 | Julian day conversion (transitive) |
| `github.com/hablullah/go-sampa` | v1.0.0 | Sun position (transitive) |
| `github.com/wailsapp/wails/v2` | v2.13.0 | Desktop shell (desktop build only) |
| `github.com/wailsapp/go-webview2` | v1.0.22 | WebView2 bindings (transitive) |
| `git.sr.ht/~jackmordaunt/go-toast/v2` | v2.0.3 | Toast notifications (transitive) |

### 2.2 Development Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `templ` CLI | latest | `templ generate` |
| `@tailwindcss/cli` | v4 | Compile `input.css` → `static/out.css` |
| `wails` CLI | v2 | `wails build` (optional) |
| Go | 1.25+ | Toolchain |

### 2.3 External Requirements

- **Runtime (web)**: a browser with HTMX 1.9, ES modules, and `Intl.DateTimeFormat` support. Three.js 0.159 and GSAP 3.12 are loaded from `unpkg.com` for the decorative moon only.
- **Runtime (desktop)**: Windows with WebView2 runtime installed.
- **Build**: Go 1.25+, Node.js (for the Tailwind CLI), PowerShell.

---

## 3. API Definitions

### 3.1 HTTP routes (`server.go`)

```go
// buildMux creates the HTTP route mux. Shared between web and desktop modes.
func buildMux() *http.ServeMux

// Routes (Go 1.22+ method-pattern mux):
//   GET  /                    handleIndex          — full page render
//   GET  /partials/clock      handleClock          — clock partial, polled every 1s
//   GET  /partials/schedule   handleSchedule       — today's five prayers
//   GET  /partials/qibla      handleQibla          — bearing + cardinal
//   GET  /partials/niat       handleNiat           — single niat (?prayer=subuh)
//   GET  /partials/niat-list  handleNiatList       — list of five prayers
//   GET  /partials/location   handleLocationDrawer — city/GPS/method picker
//   POST /location            handleSetLocation    — set cookies, return 200
//   GET  /static/*            file server          — static/out.css + assets
```

### 3.2 Cookie accessors (`server.go`)

```go
// locFromCookie reads lat/lng from the ws_loc cookie ("lat,lng").
// Returns ok=false if the cookie is missing or malformed.
func locFromCookie(r *http.Request) (lat, lng float64, ok bool)

// methodFromCookie returns the method id, defaulting to "mwl".
func methodFromCookie(r *http.Request) string

// cityFromCookie returns the display city name, or "" if unset.
func cityFromCookie(r *http.Request) string

// tzFor returns the request's *time.Location.
// Order: "tz" query param → Asia/Jakarta → UTC.
// ponytail: no tz-boundary lookup dep; falls back to Asia/Jakarta.
func tzFor(r *http.Request) *time.Location
```

### 3.3 Prayer computation (`prayer.go`)

```go
// MethodByID resolves a method id to a TwilightConvention.
// Unknown ids default to MWL.
func MethodByID(id string) *prayer.TwilightConvention

// TodaySchedule returns the five prayer times for today at the
// given location, computed with the given twilight convention.
// Computes the full year then picks today's row.
func TodaySchedule(lat, lng float64, method *prayer.TwilightConvention, tz *time.Location) ([]PrayerTime, error)

// QiblaBearing returns the initial great-circle bearing in degrees
// (0–360, clockwise from true north) to the Ka'bah.
// Formula: atan2(sin(Δλ)·cos(φ2), cos(φ1)·sin(φ2) − sin(φ1)·cos(φ2)·cos(Δλ))
func QiblaBearing(lat, lng float64) float64

// Cardinal returns the 16-wind compass label for a bearing.
func Cardinal(b float64) string

// runSelfCheck verifies QiblaBearing against known references.
// Run with: go run . check
func runSelfCheck() error
```

### 3.4 Niat lookup (`niat.go`)

```go
// NiatByPrayer is the lookup table for the five daily prayers' niat.
var NiatByPrayer map[PrayerName]Niat

// PrayerLabel is the Indonesian display name for each prayer.
var PrayerLabel map[PrayerName]string
```

### 3.5 Desktop entrypoint (`desktop.go`)

```go
// App is the Wails application struct. Methods on it are callable from JS.
type App struct { ctx context.Context }

// OnStartup starts the Go HTTP server in a goroutine when the Wails app launches.
// Binds to 127.0.0.1:9090 so the WebView can hit it locally.
func (a *App) OnStartup(ctx context.Context)
```

---

## 4. Data Flow

### 4.1 Page load (location already set)

```
1. Browser GET /  (with ws_loc, ws_city, ws_method cookies)
2. handleIndex:
   a. locFromCookie → lat, lng
   b. cityFromCookie → "Jakarta"
   c. methodFromCookie → "mwl"
   d. tzFor → Asia/Jakarta (from query param or fallback)
   e. TodaySchedule(lat, lng, MWL, tz) → 5 PrayerTimes
   f. nextPrayer(schedule, now) → *PrayerTime, "00:39:53"
   g. QiblaBearing(lat, lng) → 295.2; Cardinal → "WNW"
   h. renderTempl(Layout("Waktu Sholat", IndexPage(data)))
3. Browser receives full HTML; HTMX attributes wire partials:
   - Clock: hx-get="/partials/clock" hx-trigger="load, every 1s"
   - Drawer buttons: hx-get="/partials/..." hx-on::after-request="openDrawer()"
```

### 4.2 Clock polling

```
Every 1s:
  GET /partials/clock?tz=Asia/Jakarta
    → tzFor(r) → *time.Location
    → ClockPartial(time.Now().In(tz))
    → HTML fragment swapped into the clock section
```

### 4.3 Location set (city)

```
1. User picks "Jakarta" in the picker, submits form
2. POST /location  (form: city="Jakarta|-6.208800|106.845600")
3. handleSetLocation:
   a. parse "name|lat|lng"
   b. set ws_loc = "-6.208800,106.845600"
   c. set ws_city = "Jakarta"
   d. set ws_method (from form or existing cookie)
   e. return 200, HX-Redirect: /
4. Browser reloads → handleIndex with cookies set
```

### 4.4 Location set (GPS)

```
1. User clicks "Gunakan GPS"
2. useGPS(form) → navigator.geolocation.getCurrentPosition
3. Fill form.lat, form.lng; htmx.trigger(form, 'submit')
4. POST /location  (form: lat, lng, method)
5. handleSetLocation sets cookies → 200, HX-Redirect: /
```

### 4.5 Qibla formula

```go
// Inputs: user lat/lng, Ka'bah lat/lng (21.4225247, 39.8261819)
φ1 = lat  × π/180
φ2 = kaabaLat × π/180
Δλ = (kaabaLng − lng) × π/180

y = sin(Δλ) × cos(φ2)
x = cos(φ1) × sin(φ2) − sin(φ1) × cos(φ2) × cos(Δλ)
θ = atan2(y, x) × 180/π

bearing = (θ + 360) mod 360
```

---

## 5. Implementation Guidelines

### 5.1 Thread Safety

- `Cities`, `Methods`, `NiatByPrayer`, `PrayerLabel` are read-only package vars — safe for concurrent access.
- `net/http` runs one goroutine per request; handlers do not mutate shared state.
- The desktop `OnStartup` runs `buildMux()` in a goroutine so the Wails UI thread is never blocked.

### 5.2 Error Handling Strategy

- Return errors from computation functions; do not panic.
- In handlers, fall back to an empty `PageData`/partial rather than returning 500 — the UI degrades to "no schedule" instead of an error page.
- Log render errors via `log.Printf` in `renderTempl`; do not surface to the user.

### 5.3 Logging

- `log.Printf` for server startup (`"waktu-sholat listening on %s"`) and render errors.
- No structured logging, no levels — the app is single-user; `log` is sufficient.

### 5.4 Configuration

- Listen address: `--addr` flag (default `127.0.0.1:9090`).
- Desktop address: hardcoded `127.0.0.1:9090` in `desktop.go`.
- No env vars, no config file — the build script and flags are the configuration.

---

## 6. Testing Strategy

### 6.1 Unit Tests

| Component | Coverage Target | Key Scenarios |
|-----------|----------------|---------------|
| `QiblaBearing` | 100% (via self-check) | Jakarta ~295°, London ~119°, New York ~58°, Ka'bah ~0° |
| `Cardinal` | implicit | Each of the 16 sectors |
| `MethodByID` | implicit | Each id resolves; unknown id defaults to MWL |
| `locFromCookie` | implicit | Well-formed, malformed, missing |

### 6.2 Integration Tests

- `go run . check` — runs `runSelfCheck`, the runnable verification of the qibla formula. Exits 0 on success, 1 on failure.
- Manual: `.\build.ps1 run`, open `http://127.0.0.1:9090`, pick Jakarta, verify schedule and qibla render.

### 6.3 Performance Tests

No automated benchmarks. Manual targets:
- Schedule computation: < 100 ms per request.
- Clock partial: < 50 ms response.
- First paint after cookie set: < 1 second.

---

## 7. Build and Deployment

### 7.1 Build Configuration

`build.ps1` actions:

| Action | Steps |
|--------|-------|
| `gen` | `templ generate` |
| `css` | `npx @tailwindcss/cli -i input.css -o static/out.css --minify` |
| `css-watch` | Tailwind watch mode |
| `web` | gen + css + `go build -o waktu-sholat.exe .` |
| `desktop` | gen + css + `go build -tags desktop -o waktu-sholat-desktop.exe .` |
| `wails-build` | gen + css + `wails build` |
| `run` | gen + css + `go run . --addr=127.0.0.1:9090` |
| `check` | `go run . check` |
| `all` | gen + css + web |

### 7.2 Platform-Specific Notes

- **Windows**: primary target. Desktop build uses Wails WebView2 (preinstalled on Windows 11). PowerShell build script.
- **Linux/macOS**: web binary cross-compiles via `GOOS=linux go build .`; the build script is PowerShell but the steps translate to bash. Desktop build requires the Wails CLI and platform WebView runtime.
- **Browser**: HTMX 1.9, ES modules (for Three.js/GSAP), `Intl.DateTimeFormat`. IE and old Safari are not supported.

---

## 8. Version Control

### 8.1 Branch Strategy

Single-`main` workflow for a solo project. Feature branches optional; tag releases.

### 8.2 Commit Convention

Conventional Commits style, e.g.:

```
feat(qibla): add 16-wind cardinal label
fix(schedule): handle missing today row gracefully
docs(prd): add product requirements document
chore(build): pin tailwind cli version
```

---

## 9. Security Checklist

- [ ] No user accounts, no database, no PII stored server-side.
- [ ] Cookies carry only `lat,lng`, city name, method id — no sensitive data.
- [ ] Desktop server binds to `127.0.0.1` only.
- [ ] No `eval`, no `innerHTML` of user input — all HTML is templ-generated.
- [ ] Three.js / GSAP loaded from `unpkg.com` over HTTPS — set a CSP if hardening for public hosting.
- [ ] Pin `hablullah/go-prayer` and `a-h/templ` versions in `go.mod`.
- [ ] No secrets in the repo — the app has none.

---

## 10. Document Control

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-08-05 | waktu-sholat maintainer | Initial release |
