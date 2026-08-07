# Waktu Sholat - Product Requirements Document (PRD)

## 1. Product Vision

### 1.1 Vision Statement

A single-screen, Bahasa Indonesia prayer times app that shows the clock, the next prayer, and the qibla direction — and nothing else until you ask.

### 1.2 Target Users

- **Primary**: Indonesian Muslim who prays five times daily and wants the next prayer time without wading through tabs.
- **Secondary**: Indonesian diaspora Muslim living abroad who needs correct times in a non-Indonesian timezone and a qibla bearing from a non-Indonesian coordinate.
- **Tertiary**: Open-source contributor or self-hoster who wants a small, readable, no-API prayer times codebase.

### 1.3 Product Goals

1. Show the next prayer within 3 seconds of opening the app, on a single screen.
2. Compute prayer times and qibla locally — zero runtime external API calls.
3. Keep the interface in Bahasa Indonesia with a mission-control aesthetic that reads at a glance.

---

## 2. User Stories

### 2.1 First Run / Location

1. As a first-time user, I want to pick my city from a list, so that I get prayer times without enabling GPS.
2. As a first-time user, I want to use GPS for exact coordinates, so that my times are accurate even if my city isn't listed.
3. As a returning user, I want my location remembered, so that I don't re-pick my city every visit.

### 2.2 Prayer Schedule

4. As a daily user, I want to see today's five prayer times, so that I can plan my day.
5. As a daily user, I want to see the next prayer and a live countdown, so that I know how long I have.
6. As a user in Indonesia, I want to use the Kemenag calculation method, so that my times match local authority.
7. As a diaspora user, I want to switch calculation method, so that my times match the local convention.

### 2.3 Qibla

8. As a user, I want to see the qibla bearing in degrees, so that I can face the Ka'bah.
9. As a user, I want a compass label alongside the bearing, so that I can translate degrees into a direction I can act on.

### 2.4 Niat

10. As a user, I want the niat text for each prayer in Arabic, transliteration, and Indonesian, so that I can recite the intention correctly.

### 2.5 Presentation

11. As a user, I want a dark theme at night, so that the screen doesn't blind me at Subuh.
12. As a user with motion sensitivity, I want decorative animation disabled, so that the app is usable.
13. As a desktop user, I want a native window, so that I don't keep a browser tab open.

---

## 3. Functional Specifications

### 3.1 Location Picker (first run, no cookie)

#### 3.1.1 City dropdown
- **Type**: `<select>` with 23 options
- **Label**: "Kota"
- **Behavior**: On submit, POSTs `city` as `name|lat|lng`; server sets `ws_loc`, `ws_city`, `ws_method` cookies and reloads.
- **Validation**: Server parses `lat,lng` floats; on failure, no cookie is set and the picker re-renders.

#### 3.1.2 GPS button
- **Type**: Submit button with `onsubmit="return useGPS(this)"`
- **Behavior**: Calls `navigator.geolocation.getCurrentPosition`, fills hidden `lat`/`lng` inputs, triggers HTMX submit.
- **Error**: If geolocation unsupported or denied, alerts the user in Indonesian.

### 3.2 Hero Screen (location set)

#### 3.2.1 Clock
- **Type**: HTMX partial polled every 1 second
- **Endpoint**: `GET /partials/clock`
- **Behavior**: Renders current time in the user's timezone.

#### 3.2.2 Next Prayer
- **Type**: Static section rendered server-side on full page load
- **Behavior**: Shows prayer label (Indonesian), time (`HH:MM`), and "dalam HH:MM:SS" countdown.

#### 3.2.3 Qibla
- **Type**: Static section + HTMX partial (`GET /partials/qibla`)
- **Behavior**: Shows bearing in degrees and 16-wind compass label.

#### 3.2.4 Bottom nav
- **Type**: Three buttons ("Jadwal", "Niat", "Lokasi")
- **Behavior**: Each fetches a partial into `#drawer-body` and opens the drawer.

### 3.3 Schedule Drawer

- **Endpoint**: `GET /partials/schedule`
- **Behavior**: Lists all five prayers with times; highlights the next one.
- **Method picker**: `<select>` of eight methods; on change, POSTs `/location` to update `ws_method` and re-renders.

### 3.4 Niat Drawer

- **Endpoint**: `GET /partials/niat-list` → list of five prayers; `GET /partials/niat?prayer=<name>` → single niat.
- **Behavior**: Tap a prayer to see Arabic, transliteration, and Indonesian translation.

### 3.5 Location Drawer

- **Endpoint**: `GET /partials/location`
- **Behavior**: City picker + GPS + method picker, same as first-run but inside the drawer.

### 3.6 Theme Toggle

- **Type**: Fixed top-right button
- **Behavior**: Cycles dark ↔ light; persists to `localStorage.ws_theme`; syncs icon on load and on `prefers-color-scheme` change.

---

## 4. Non-Functional Specifications

### 4.1 Performance

- Clock partial: < 50 ms response.
- Schedule computation: < 100 ms per request.
- First paint after cookie set: < 1 second.
- Moon canvas: pauses when tab hidden; skips entirely under `prefers-reduced-motion`.

### 4.2 Reliability

- No runtime external API calls for core features.
- Timezone fallback chain: `tz` query param → `Asia/Jakarta` → UTC.
- Missing location cookie → location picker, not an error.

### 4.3 Usability

- Single hero screen (100dvh), no scroll for essentials.
- Bahasa Indonesia throughout.
- Drawer pattern for secondary content.
- Monospace numerals for tabular alignment.

### 4.4 Compatibility

- Modern browsers (HTMX 1.9, ES modules, `Intl.DateTimeFormat`).
- Windows desktop via Wails WebView2.
- Go 1.25+ for builds.

---

## 5. User Interface Mockups

### 5.1 First Run — Location Picker

```
┌─────────────────────────────────────┐
│              ☾  (toggle)             │
│                                     │
│         ┌─────────────────┐         │
│         │   PILIH LOKASI   │         │
│         │                   │         │
│         │  Kota             │         │
│         │  [Jakarta      ▼] │         │
│         │  [ Gunakan Kota ] │         │
│         │      — atau —     │         │
│         │  [ Gunakan GPS  ] │         │
│         └─────────────────┘         │
└─────────────────────────────────────┘
```

### 5.2 Hero Screen (location set)

```
┌─────────────────────────────────────┐
│ WAKTU SHOLAT        Jakarta ●  ☾    │
│                                     │
│            14:32:07                 │
│                                     │
│        SHOLAT BERIKUTNYA            │
│        Ashar  15:12                 │
│        dalam 00:39:53               │
│                                     │
│        295°  WNW                    │
│                                     │
│  ─────────────────────────────      │
│   JADWAL    NIAT    LOKASI          │
└─────────────────────────────────────┘
```

### 5.3 Schedule Drawer (slides up from bottom)

```
┌─────────────────────────────────────┐
│  JADWAL                       ✕     │
│  Subuh    04:43                     │
│  Zuhur    12:08                     │
│  Ashar    15:12  ← berikutnya       │
│  Maghrib 18:21                     │
│  Isya     19:34                     │
│                                     │
│  Metode: [MWL                  ▼]   │
└─────────────────────────────────────┘
```

---

## 6. Acceptance Criteria

### 6.1 Location — City Picker

```gherkin
Given the user has no ws_loc cookie
When they open the app
Then the location picker is shown with 23 city options
And the "Gunakan Kota" button is enabled
```

### 6.2 Location — GPS

```gherkin
Given the user clicks "Gunakan GPS"
And the browser supports geolocation
And the user grants permission
Then the lat/lng hidden inputs are populated
And the form is submitted via HTMX
And the ws_loc cookie is set
```

### 6.3 Next Prayer Countdown

```gherkin
Given the user's location is set
When the page loads
Then the next prayer name, time, and "dalam HH:MM:SS" countdown are displayed
And the countdown updates as the clock partial polls every 1 second
```

### 6.4 Qibla Bearing

```gherkin
Given the user's location is Jakarta (-6.2088, 106.8456)
When the qibla section renders
Then the bearing is approximately 295°
And the compass label is "WNW"
```

### 6.5 Method Switch

```gherkin
Given the user opens the schedule drawer
When they change the method to "kemenag"
Then the ws_method cookie is updated to "kemenag"
And the schedule re-renders with Kemenag times
```

### 6.6 Theme Toggle

```gherkin
Given the user clicks the theme toggle
When the current theme is dark
Then the theme switches to light
And ws_theme is persisted as "light" in localStorage
And the sun icon is shown
```

### 6.7 Self-Check

```gherkin
Given the developer runs `go run . check`
When the qibla formula is evaluated for Jakarta, London, New York, and the Ka'bah
Then all four cases pass within ±2° tolerance
And "self-check ok" is printed
```

---

## 7. Out of Scope

- Push notifications / adhan audio playback.
- Monthly or yearly schedule export.
- User accounts or cloud sync.
- Hijri calendar display.
- Mosque finder.
- Mobile native apps (iOS/Android) — the web build is mobile-responsive.
- Multi-language UI beyond Bahasa Indonesia.

---

## 8. Release Criteria

### 8.1 Release Blockers

- [ ] `go run . check` passes.
- [ ] Schedule renders for at least one Indonesian city and one diaspora city.
- [ ] Qibla bearing matches the self-check references.
- [ ] No runtime external API calls for core features.
- [ ] Location, method, and city cookies persist across reloads.

### 8.2 Release Checklist

- [ ] `.\build.ps1 web` produces a working binary.
- [ ] `.\build.ps1 desktop` produces a working Wails binary.
- [ ] Light and dark themes both render correctly.
- [ ] `prefers-reduced-motion` disables the moon canvas.
- [ ] README and docs/ are up to date.

---

## 9. Document Control

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-08-05 | waktu-sholat maintainer | Initial release |
