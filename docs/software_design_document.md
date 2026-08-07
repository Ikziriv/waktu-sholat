# Waktu Sholat - Software Design Document (SDD)

## 1. Introduction

### 1.1 Purpose

This SDD documents the architecture and component design of Waktu Sholat: a single-binary Go application that serves a prayer times web UI and, with a build tag, wraps the same server in a Wails desktop window.

### 1.2 Scope

The whole system — HTTP server, prayer/qibla computation, templ UI, HTMX partials, desktop shell. External libraries (`hablullah/go-prayer`, `a-h/templ`, Wails, Tailwind, HTMX, Three.js, GSAP) are described by role, not internally documented.

### 1.3 Definitions and Acronyms

| Term | Definition |
|------|------------|
| GOTH | Go + templ + Tailwind + HTMX stack |
| HTMX | Hypermedia-driven front-end via HTML attributes |
| templ | Type-safe HTML templating language for Go |
| Wails | Go framework that wraps a web UI in a native desktop window |
| Qibla | Direction to the Ka'bah from a given location |
| Niat | The intention recited before prayer |
| Twilight convention | Rules for computing Fajr/Isha angles (calculation method) |
| MWL | Muslim World League (default calculation method) |
| Kemenag | Kementerian Agama (Indonesian Ministry of Religion) |

---

## 2. System Architecture

### 2.1 High-Level Architecture

```
┌──────────────────────────────────────────────────────────┐
│                    Build target                           │
│                                                          │
│   go build .            →  waktu-sholat-web.exe          │
│   go build -tags desktop →  waktu-sholat-desktop.exe     │
│                                                          │
│   main.go     (build !desktop)  ──┐                      │
│   desktop.go  (build desktop)    ─┤  both call buildMux  │
└────────────────────────────────────┼─────────────────────┘
                                     │
                                     ▼
┌──────────────────────────────────────────────────────────┐
│              HTTP server (net/http ServeMux)             │
│                                                          │
│   GET /                    → handleIndex (full page)     │
│   GET /partials/clock      → handleClock                 │
│   GET /partials/schedule   → handleSchedule              │
│   GET /partials/qibla      → handleQibla                 │
│   GET /partials/niat       → handleNiat                  │
│   GET /partials/niat-list  → handleNiatList              │
│   GET /partials/location   → handleLocationDrawer        │
│   POST /location           → handleSetLocation           │
│   GET /static/*            → file server                 │
└──────────────────────────────────────────────────────────┘
          │                                  │
          ▼                                  ▼
┌─────────────────────┐         ┌────────────────────────┐
│  prayer.go          │         │  templ components      │
│  - TodaySchedule    │         │  Layout / IndexPage    │
│  - QiblaBearing     │         │  Clock / Schedule      │
│  - Cardinal         │         │  Qibla / Niat / Drawer │
│  - MethodByID       │         └────────────────────────┘
│  - runSelfCheck     │
└─────────────────────┘
          │
          ▼
┌─────────────────────┐
│  hablullah/go-prayer│
│  (schedule calc)    │
└─────────────────────┘
```

### 2.2 Design Patterns

- **Build-tag dual entrypoint** — `main.go` (`//go:build !desktop`) and `desktop.go` (`//go:build desktop`) each define `main()`, sharing `buildMux()` so the HTTP layer is identical across targets.
- **Server-rendered partials** — HTMX swaps server-rendered HTML fragments into the DOM. No client-side router, no JSON API.
- **Cookie-based state** — location, city, and method persisted in three cookies; no server-side session store.
- **Single self-check function** — `runSelfCheck` is the one runnable verification of the non-trivial math, invoked via `go run . check`.

### 2.3 Technology Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Language | Go 1.25 | Server, math, build |
| Templating | `a-h/templ` v0.3 | Type-safe HTML components, generated to `*_templ.go` |
| Styling | Tailwind CSS v4 | Utility-first CSS, compiled from `input.css` to `static/out.css` |
| Interactivity | HTMX 1.9 | AJAX swaps and polling via HTML attributes |
| 3D / motion | Three.js 0.159 + GSAP 3.12 | Decorative moon canvas (CDN, dark mode only) |
| Prayer calc | `hablullah/go-prayer` v1.1 | Twilight conventions, yearly schedule |
| Desktop | Wails v2.13 | Native window around the Go HTTP server |
| Build | PowerShell (`build.ps1`) | gen / css / web / desktop / run / check / all |

---

## 3. Component Design

### 3.1 `server.go` — HTTP layer

```
Functions:
  buildMux() *http.ServeMux
  locFromCookie(r) (lat, lng float64, ok bool)
  methodFromCookie(r) string
  cityFromCookie(r) string
  tzFor(r) *time.Location
  renderTempl(w, comp templ.Component)
  handleIndex, handleClock, handleSchedule, handleQibla,
  handleNiat, handleNiatList, handleLocationDrawer,
  handleSetLocation  (handlers)

Types:
  City { Name, Lat, Lng }
  PageData { Cities, HasLocation, Lat, Lng, CityName,
             MethodID, Methods, Schedule, QiblaBearing,
             QiblaCardinal, NextPrayer, NextIn }

Constants:
  cookieLoc    = "ws_loc"
  cookieMethod = "ws_method"
  cookieCity   = "ws_city"
```

**Key Design Decisions:**
- One `PageData` struct feeds every page render and partial — no per-handler DTOs.
- `tzFor` reads the `tz` query param injected by `htmx:configRequest` on every request, so partials get the right timezone without re-deriving it.
- `ponytail: no tz-boundary lookup dep; falls back to Asia/Jakarta` — the fallback chain is explicit and documented at the call site.

### 3.2 `prayer.go` — Computation

```
Constants:
  kaabaLat = 21.4225247
  kaabaLng = 39.8261819

Types:
  PrayerName string  (Subuh | Zuhur | Ashar | Maghrib | Isya)
  PrayerTime { Name PrayerName, Time time.Time }
  Method { ID, Name string }

Functions:
  MethodByID(id string) *prayer.TwilightConvention
  TodaySchedule(lat, lng float64, method, tz) ([]PrayerTime, error)
  QiblaBearing(lat, lng float64) float64
  Cardinal(b float64) string
  runSelfCheck() error
```

**Key Design Decisions:**
- `TodaySchedule` computes the whole year then picks today's row — the library doesn't expose a single-day API, and the year-long computation is fast enough (< 100 ms).
- `QiblaBearing` uses the standard initial-bearing formula (`atan2(sin(Δλ)cos(φ2), cos(φ1)sin(φ2) − sin(φ1)cos(φ2)cos(Δλ))`), normalized to 0–360°.
- `runSelfCheck` is the only test. It covers three reference cities plus the Ka'bah-itself edge case, with a ±2° tolerance.

### 3.3 `niat.go` — Static content

```
Types:
  Niat { Name, Arabic, Transliteration, Translation string }

Tables:
  NiatByPrayer map[PrayerName]Niat   (5 entries)
  PrayerLabel  map[PrayerName]string (5 entries)
```

**Key Design Decisions:**
- Hardcoded table, not a JSON file — five rows, no need for a loader.
- Indonesian translations, since the app is "waktu sholat".

### 3.4 `*.templ` — UI components

| File | Components |
|------|-----------|
| `layout.templ` | `Layout`, `MoonCanvas`, `ThemeToggle`, `Drawer`, `Scripts` |
| `index.templ` | `IndexPage`, `LocationPicker`, `ClockSection`, `NextPrayerSection`, `LocationDrawer`, `NiatListDrawer`, `MethodPicker` |
| `clock.templ` | `ClockPartial` |
| `schedule.templ` | `ScheduleSection` |
| `qibla.templ` | `QiblaSection` |
| `niat.templ` | `NiatDetail` |

**Key Design Decisions:**
- Single hero screen (100dvh) with drawers for secondary content — the three essentials (clock, next prayer, qibla) are always visible.
- Moon canvas is decorative only: dark-mode-only, skipped under `prefers-reduced-motion`, paused when the tab is hidden.
- HTMX attributes live in the templ files; the only JS is the GPS, theme, and drawer helpers in `Scripts`.

### 3.5 `desktop.go` — Wails shell

```
Types:
  App { ctx context.Context }

Methods:
  App.OnStartup(ctx)  // starts buildMux() in a goroutine on 127.0.0.1:9090

Embed:
  //go:embed frontend/dist
  frontendAssets embed.FS
```

**Key Design Decisions:**
- The desktop app runs the same HTTP server on localhost and points the WebView at it — no IPC bridge, no duplicated logic.
- `frontend/dist` is embedded so the binary is self-contained.

---

## 4. Data Models

### 4.1 Cookie state

```json
{
  "ws_loc": "lat,lng",
  "ws_city": "Jakarta",
  "ws_method": "mwl"
}
```

No server-side persistence. Cookies are set on `POST /location` and read on every request.

### 4.2 PrayerTime

```json
{
  "Name": "ashar",
  "Time": "2026-08-05T15:12:00+07:00"
}
```

### 4.3 PageData (render context)

```json
{
  "Cities": [{"Name":"Jakarta","Lat":-6.2088,"Lng":106.8456}],
  "HasLocation": true,
  "Lat": -6.2088,
  "Lng": 106.8456,
  "CityName": "Jakarta",
  "MethodID": "mwl",
  "Methods": [{"ID":"mwl","Name":"Muslim World League"}],
  "Schedule": [{"Name":"subuh","Time":"..."}],
  "QiblaBearing": 295.2,
  "QiblaCardinal": "WNW",
  "NextPrayer": {"Name":"ashar","Time":"..."},
  "NextIn": "00:39:53"
}
```

---

## 5. State Management

### 5.1 First-run → Location-set flow

```
[No ws_loc cookie]
        │
        ▼
  LocationPicker shown
        │
   ┌────┴────┐
   │ City    │ GPS
   ▼         ▼
  POST /location (city|lat|lng  OR  lat,lng from geolocation)
        │
        ▼
  Set ws_loc, ws_city, ws_method cookies
        │
        ▼
  Reload → handleIndex → hero screen
```

### 5.2 Drawer open/close flow

```
Click nav button (Jadwal / Niat / Lokasi)
        │
        ▼
  hx-get="/partials/..."  →  #drawer-body
        │
        ▼
  hx-on::after-request="openDrawer()"
        │
        ▼
  Drawer slides up, backdrop fades in
        │
   click backdrop / ✕ / "tutup"
        │
        ▼
  closeDrawer()
```

### 5.3 Clock polling

```
ClockSection loads with hx-trigger="load, every 1s"
        │
        ▼
  GET /partials/clock?tz=Asia/Jakarta  (every 1s)
        │
        ▼
  ClockPartial(now) rendered into the section
```

---

## 6. Error Handling

### 6.1 Exception hierarchy

There is no exception hierarchy (Go). Errors are returned and handled at the call site.

### 6.2 Error propagation

| Location | Error | Handling |
|----------|-------|----------|
| `TodaySchedule` | library error or no row for today | `handleIndex`/`handleSchedule` render an empty `ScheduleSection`; `NextPrayer` stays nil |
| `locFromCookie` | malformed cookie | returns `ok=false`; picker re-renders |
| `tzFor` | unknown tz string | falls back to `Asia/Jakarta`, then UTC |
| `handleSetLocation` | bad lat/lng | no cookie set; user stays on picker |
| `useGPS` | geolocation denied | `alert("Gagal mengambil lokasi: ...")` |
| Moon canvas | Three.js/GSAP CDN fail | `console.warn`, app continues without the moon |

---

## 7. Performance Considerations

### 7.1 Multi-threading

- The web server is `net/http`, one goroutine per request.
- The desktop app starts `buildMux()` in a goroutine on `OnStartup` so the Wails UI thread is never blocked.

### 7.2 Memory Management

- No caches — every request recomputes. `TodaySchedule` computes a full year on each call; this is < 100 ms and avoids cache invalidation logic.
- `Cities`, `Methods`, `NiatByPrayer`, `PrayerLabel` are package-level vars, allocated once.

### 7.3 Scalability Targets

Single-user, localhost. Not designed for multi-tenant concurrency. The desktop build is strictly single-user; the web build is intended for personal or family self-hosting.

---

## 8. Security Considerations

- [ ] No user accounts, no database, no PII stored server-side.
- [ ] Cookies carry only `lat,lng`, city name, and method id — no sensitive data.
- [ ] Desktop server binds to `127.0.0.1` only — not exposed on the LAN.
- [ ] No `eval`, no `innerHTML` of user input — all HTML is templ-generated.
- [ ] Three.js / GSAP loaded from `unpkg.com` CDN over HTTPS — review CSP if hardening for production hosting.
- [ ] `go-prayer` and `templ` are the only faith-sensitive dependencies — pin versions in `go.mod`.

---

## 9. Deployment Architecture

### 9.1 Build Process

```
templ generate
    → *_templ.go from *.templ
npx @tailwindcss/cli -i input.css -o static/out.css --minify
    → static/out.css
go build [-tags desktop] -o <binary> .
    → web or desktop binary
```

`build.ps1` wraps these into named actions: `gen`, `css`, `web`, `desktop`, `wails-build`, `run`, `check`, `all`.

### 9.2 Distribution

- **Web**: single static binary + `static/` directory. Run on any host that can serve a Go binary. Bind to localhost for personal use, or behind a reverse proxy for public deployment.
- **Desktop**: `waktu-sholat-desktop.exe` — self-contained, embeds `frontend/dist`. Distribute as a standalone `.exe` or via a Wails installer.

---

## 10. Document Control

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-08-05 | waktu-sholat maintainer | Initial release |
