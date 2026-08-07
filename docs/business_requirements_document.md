# Waktu Sholat - Business Requirements Document (BRD)

## 1. Executive Summary

### 1.1 Project Overview

Waktu Sholat is a free, open-source prayer times application for Indonesian Muslims and the diaspora. It computes prayer schedules and qibla direction locally, with no tracking, no accounts, and no external API dependencies at runtime. The app ships as a web server and a native desktop application from a single Go codebase.

### 1.2 Business Problem

Indonesian Muslims rely on prayer times daily, yet most existing apps share a set of friction points:

- **Cluttered interfaces** that bury the next prayer time behind tabs, ads, and notifications.
- **Privacy concerns** — many apps require accounts, request excessive permissions, or phone home to ad networks.
- **External API dependency** — prayer times are fetched from remote services that can be slow, rate-limited, or offline.
- **Indonesian-language gap** — global apps default to English or Arabic; Indonesian users get partial translations.
- **Diaspora disconnect** — Indonesians abroad juggle one app for prayer times and another for qibla, with method presets that don't include Kemenag.

### 1.3 Business Opportunity

A focused, single-screen app that computes everything locally removes the friction points above. The opportunity is not feature parity with bloated prayer suites — it is the opposite: show the three things a Muslim needs at prayer time, hide everything else, and respect privacy by construction. Open source builds trust in a faith-sensitive audience where correctness matters.

---

## 2. Business Objectives

### 2.1 Primary Objectives

| Objective | Success Metric | Target |
|-----------|---------------|--------|
| Provide accurate prayer times for any location | Schedule accuracy vs. Kemenag published times | ±2 minutes for Indonesian cities |
| Provide accurate qibla direction | Bearing accuracy vs. known references | ±2° for Jakarta, London, New York |
| Zero runtime tracking | External network calls at runtime | 0 |
| Single-screen UX | Time-to-next-prayer on first load | < 3 seconds after location set |

### 2.2 Secondary Objectives

- Cover the Indonesian diaspora with city presets beyond Indonesia.
- Ship a desktop build for users who want a native window.
- Keep the codebase small enough for one developer to maintain.

---

## 3. Stakeholder Requirements

### 3.1 Primary Stakeholders

#### 3.1.1 Indonesian Muslim end user
- **Role**: Daily user of the app, in Indonesia.
- **Needs**: Correct prayer times for their city, qibla direction, niat text, Bahasa Indonesia interface.
- **Pain Points**: Existing apps are cluttered, ad-supported, and default to non-Indonesian calculation methods.
- **Success Criteria**: Opens the app and sees the next prayer within 3 seconds, in Indonesian, with Kemenag as a one-tap method.

#### 3.1.2 Indonesian diaspora Muslim end user
- **Role**: Indonesian Muslim living abroad (Mecca, London, New York, etc.).
- **Needs**: Prayer times for non-Indonesian coordinates, qibla, a calculation method that matches local convention.
- **Pain Points**: Apps that only cover Indonesia; no diaspora city presets.
- **Success Criteria**: Picks a diaspora city or uses GPS, gets correct times in the local timezone.

#### 3.1.3 Open-source contributor / maintainer
- **Role**: Developer who extends or self-hosts the app.
- **Needs**: Small, readable codebase; clear build instructions; no proprietary dependencies.
- **Pain Points**: Prayer apps that are closed-source or depend on paid APIs.
- **Success Criteria**: Clones the repo, runs `.\build.ps1 run`, and gets a working server in under 5 minutes.

### 3.2 Stakeholder Matrix

| Stakeholder | Influence | Interest | Engagement Strategy |
|-------------|-----------|----------|-------------------|
| Indonesian end user | Med | High | Bahasa Indonesia UI, Kemenag default-adjacent method |
| Diaspora end user | Low | High | Diaspora city presets, GPS fallback, local timezone |
| Maintainer | High | High | Single-file modules, PowerShell build script, self-check |
| Faith accuracy reviewers | Low | High | Open-source math, verified qibla references |

---

## 4. Functional Requirements

### 4.1 Prayer Schedule

| ID | Requirement | Priority |
|----|------------|----------|
| FR-01 | Compute today's five prayer times (Subuh, Zuhur, Ashar, Maghrib, Isya) for the user's coordinates | Must Have |
| FR-02 | Support eight calculation methods (MWL, Kemenag, ISNA, Egypt, Karachi, Umm al-Qura, Tehran, Jafari) | Must Have |
| FR-03 | Display times in the user's local timezone | Must Have |
| FR-04 | Show the next prayer with a live countdown | Must Have |
| FR-05 | Persist location, city, and method choice across sessions via cookies | Must Have |

### 4.2 Qibla Direction

| ID | Requirement | Priority |
|----|------------|----------|
| FR-06 | Compute the great-circle bearing from the user's location to the Ka'bah | Must Have |
| FR-07 | Display the bearing in degrees and a 16-wind compass label | Must Have |

### 4.3 Niat Sholat

| ID | Requirement | Priority |
|----|------------|----------|
| FR-08 | Provide niat text (Arabic, transliteration, Indonesian translation) for each of the five obligatory prayers | Must Have |

### 4.4 Location

| ID | Requirement | Priority |
|----|------------|----------|
| FR-09 | Offer 23 city presets covering Indonesian population centers and diaspora cities | Must Have |
| FR-10 | Support GPS via `navigator.geolocation` for exact coordinates | Should Have |

### 4.5 Platform

| ID | Requirement | Priority |
|----|------------|----------|
| FR-11 | Ship a web server binary that runs on Windows (and cross-compilable to Linux/macOS) | Must Have |
| FR-12 | Ship a Wails desktop binary for Windows | Should Have |
| FR-13 | Light/dark theme with persistence and `prefers-color-scheme` respect | Should Have |

---

## 5. Non-Functional Requirements

### 5.1 Performance

| ID | Requirement | Target |
|----|------------|--------|
| NFR-01 | First paint after location is set | < 1 second |
| NFR-02 | Clock partial response time | < 50 ms |
| NFR-03 | Schedule computation for one day | < 100 ms |
| NFR-04 | Binary size (web) | < 10 MB |

### 5.2 Usability

- Single hero screen (100dvh) — all essentials visible without scrolling.
- Drawer pattern for secondary content (schedule, niat, location).
- Bahasa Indonesia throughout the interface.
- Mission-control monospace aesthetic for readability at a glance.

### 5.3 Reliability

- No external API calls at runtime — prayer times and qibla are computed locally.
- Graceful fallback: if timezone lookup fails, fall back to `Asia/Jakarta`, then UTC.
- If no location cookie is set, show the location picker instead of an error.

### 5.4 Compatibility

- Modern browsers with HTMX, ES modules, and `Intl.DateTimeFormat` support.
- Windows desktop via Wails (WebView2).
- Go 1.25+ for builds.

---

## 6. Success Metrics and KPIs

### 6.1 Operational Metrics

| Metric | Baseline | Target | Measurement |
|--------|----------|--------|-------------|
| Schedule accuracy (Indonesian cities) | n/a | ±2 min vs. Kemenag | Manual spot check across 5 cities |
| Qibla accuracy | n/a | ±2° vs. known references | `go run . check` self-check |
| Runtime external calls | n/a | 0 | Network audit during a session |
| First-load time-to-next-prayer | n/a | < 3 s | Local benchmark after cookie set |

### 6.2 Technical Metrics

- Self-check passes on every build.
- Single Go module, no transitive faith-related dependencies beyond `hablullah/go-prayer`.

---

## 7. Constraints and Assumptions

### 7.1 Constraints

- No runtime internet dependency for core features (prayer times, qibla, niat).
- No user accounts, no database, no server-side state beyond cookies.
- Interface language is Bahasa Indonesia.
- Desktop target is Windows (Wails WebView2).

### 7.2 Assumptions

- The user's browser supports `Intl.DateTimeFormat().resolvedOptions().timeZone` for timezone detection.
- The `hablullah/go-prayer` library's twilight conventions are acceptable for the supported methods.
- Users in the diaspora are comfortable selecting a calculation method if the local convention differs from Kemenag.

---

## 8. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Prayer time drift vs. local authority | Med | High | Default to MWL, offer Kemenag; document the ±2 min tolerance |
| Qibla formula edge cases (poles, antimeridian) | Low | Med | Self-check covers three references + Ka'bah; document the ceiling |
| Timezone detection fails in older browsers | Low | Med | Fallback chain: query param → Asia/Jakarta → UTC |
| Three.js / GSAP CDN unavailable | Low | Low | Moon is decorative; app works without it |
| Wails WebView2 missing on Windows | Low | Med | Web binary is the primary target; desktop is secondary |

---

## 9. Timeline and Milestones

| Phase | Deliverables | Duration |
|-------|-------------|----------|
| Core computation | prayer.go (schedule, qibla, cardinal, self-check) | Complete |
| Web UI | templ components, Tailwind, HTMX partials, cookies | Complete |
| Desktop shell | Wails integration, embed frontend/dist | Complete |
| Documentation | README, BRD, PRD, SDD, TDD | Current |

---

## 10. Document Control

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-08-05 | waktu-sholat maintainer | Initial release |
