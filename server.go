package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
)

//go:embed static
var staticAssets embed.FS

// City is a hardcoded location preset for the picker.
type City struct {
	Name string
	Lat  float64
	Lng  float64
}

// Cities covers major Indonesian population centers plus a few common
// diaspora cities so the picker is useful outside ID too.
var Cities = []City{
	{"Jakarta", -6.2088, 106.8456},
	{"Surabaya", -7.2575, 112.7521},
	{"Medan", 3.5952, 98.6722},
	{"Bandung", -6.9175, 107.6191},
	{"Makassar", -5.1477, 119.4327},
	{"Semarang", -6.9667, 110.4167},
	{"Palembang", -2.9761, 104.7754},
	{"Yogyakarta", -7.7956, 110.3695},
	{"Denpasar", -8.6705, 115.2126},
	{"Banjarmasin", -3.3193, 114.5908},
	{"Pontianak", -0.0263, 109.3425},
	{"Balikpapan", -1.2379, 116.8529},
	{"Aceh", 5.5483, 95.3238},
	{"Padang", -0.9471, 100.4172},
	{"Pekanbaru", 0.5071, 101.4478},
	{"Manado", 1.4748, 124.8421},
	{"Kupang", -10.1772, 123.6070},
	{"Jayapura", -2.5916, 140.6690},
	{"Mecca", 21.4225, 39.8262},
	{"Medina", 24.5247, 39.5692},
	{"Cairo", 30.0444, 31.2357},
	{"Istanbul", 41.0082, 28.9784},
	{"London", 51.5074, -0.1278},
	{"New York", 40.7128, -74.0060},
}

const (
	cookieLoc    = "ws_loc"    // "lat,lng"
	cookieMethod = "ws_method" // method id
	cookieCity   = "ws_city"   // city name (display only)
)

// PageData holds all data for the index page render.
type PageData struct {
	Cities        []City
	HasLocation   bool
	Lat, Lng      float64
	CityName      string
	MethodID      string
	Methods       []Method
	Schedule      []PrayerTime
	QiblaBearing  float64
	QiblaCardinal string
	NextPrayer    *PrayerTime
	NextIn        string
}

// buildMux creates the HTTP route mux. Shared between web and desktop modes.
func buildMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleIndex)
	mux.HandleFunc("GET /partials/clock", handleClock)
	mux.HandleFunc("GET /partials/schedule", handleSchedule)
	mux.HandleFunc("GET /partials/qibla", handleQibla)
	mux.HandleFunc("GET /partials/niat", handleNiat)
	mux.HandleFunc("GET /partials/niat-list", handleNiatList)
	mux.HandleFunc("GET /partials/jamak", handleJamak)
	mux.HandleFunc("GET /partials/jamak-list", handleJamakList)
	mux.HandleFunc("GET /partials/ramadan", handleRamadan)
	mux.HandleFunc("GET /partials/location", handleLocationDrawer)
	mux.HandleFunc("POST /location", handleSetLocation)
	staticFS, _ := fs.Sub(staticAssets, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	return mux
}

// locFromCookie reads lat/lng from the ws_loc cookie.
func locFromCookie(r *http.Request) (lat, lng float64, ok bool) {
	c, err := r.Cookie(cookieLoc)
	if err != nil {
		return 0, 0, false
	}
	parts := strings.Split(c.Value, ",")
	if len(parts) != 2 {
		return 0, 0, false
	}
	lat, err1 := strconv.ParseFloat(parts[0], 64)
	lng, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return lat, lng, true
}

func methodFromCookie(r *http.Request) string {
	if c, err := r.Cookie(cookieMethod); err == nil && c.Value != "" {
		return c.Value
	}
	return "mwl"
}

func cityFromCookie(r *http.Request) string {
	if c, err := r.Cookie(cookieCity); err == nil {
		return c.Value
	}
	return ""
}

// tzFor returns a *time.Location for the request. The browser sends its
// local tz via the "tz" query param (set by htmx:configRequest in JS).
// ponytail: no tz-boundary lookup dep; falls back to Asia/Jakarta.
func tzFor(r *http.Request) *time.Location {
	if tz := r.URL.Query().Get("tz"); tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil {
			return loc
		}
	}
	if loc, err := time.LoadLocation("Asia/Jakarta"); err == nil {
		return loc
	}
	return time.UTC
}

// renderTempl writes a templ.Component to the ResponseWriter.
func renderTempl(w http.ResponseWriter, comp templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := comp.Render(context.Background(), w); err != nil {
		log.Printf("templ render: %v", err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data := PageData{Cities: Cities, Methods: Methods}
	lat, lng, ok := locFromCookie(r)
	if ok {
		data.HasLocation = true
		data.Lat, data.Lng = lat, lng
		data.CityName = cityFromCookie(r)
		data.MethodID = methodFromCookie(r)
		tz := tzFor(r)
		if sched, err := TodaySchedule(lat, lng, MethodByID(data.MethodID), tz); err == nil {
			data.Schedule = sched
			data.NextPrayer, data.NextIn = nextPrayer(sched, time.Now().In(tz))
		}
		b := QiblaBearing(lat, lng)
		data.QiblaBearing = b
		data.QiblaCardinal = Cardinal(b)
	}
	renderTempl(w, Layout("Waktu Sholat", IndexPage(data)))
}

func handleClock(w http.ResponseWriter, r *http.Request) {
	tz := tzFor(r)
	now := time.Now().In(tz)
	renderTempl(w, ClockPartial(now))
}

func handleSchedule(w http.ResponseWriter, r *http.Request) {
	lat, lng, ok := locFromCookie(r)
	if !ok {
		renderTempl(w, ScheduleSection(PageData{}))
		return
	}
	tz := tzFor(r)
	methodID := methodFromCookie(r)
	sched, err := TodaySchedule(lat, lng, MethodByID(methodID), tz)
	if err != nil {
		renderTempl(w, ScheduleSection(PageData{}))
		return
	}
	next, in := nextPrayer(sched, time.Now().In(tz))
	renderTempl(w, ScheduleSection(PageData{
		Schedule:   sched,
		NextPrayer: next,
		NextIn:     in,
		MethodID:   methodID,
		Methods:    Methods,
	}))
}

func handleQibla(w http.ResponseWriter, r *http.Request) {
	lat, lng, ok := locFromCookie(r)
	if !ok {
		renderTempl(w, QiblaSection(PageData{}))
		return
	}
	b := QiblaBearing(lat, lng)
	renderTempl(w, QiblaSection(PageData{
		QiblaBearing:  b,
		QiblaCardinal: Cardinal(b),
	}))
}

func handleNiat(w http.ResponseWriter, r *http.Request) {
	name := PrayerName(r.URL.Query().Get("prayer"))
	n, ok := NiatByPrayer[name]
	if !ok {
		http.Error(w, "unknown prayer", http.StatusBadRequest)
		return
	}
	renderTempl(w, NiatPartial(PrayerLabel[n.Name], n.Arabic, n.Transliteration, n.Translation))
}

func handleNiatList(w http.ResponseWriter, r *http.Request) {
	renderTempl(w, NiatListDrawer())
}

func handleJamak(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	j, ok := JamakByQuery(PrayerName(q.Get("first")), PrayerName(q.Get("second")), JamakMode(q.Get("mode")))
	if !ok {
		http.Error(w, "unknown jamak combination", http.StatusBadRequest)
		return
	}
	renderTempl(w, NiatPartial(j.Label(), j.Arabic, j.Transliteration, j.Translation))
}

func handleJamakList(w http.ResponseWriter, r *http.Request) {
	renderTempl(w, JamakListDrawer())
}

// handleRamadan renders the Ramadhan drawer. ?year= selects a Gregorian year
// to look ahead from (defaults to next year, since the request is "see Ramadhan
// on the next year"). No location set → empty placeholder, same as schedule.
func handleRamadan(w http.ResponseWriter, r *http.Request) {
	lat, lng, ok := locFromCookie(r)
	if !ok {
		renderTempl(w, RamadanSection(RamadanInfo{Year: time.Now().Year() + 1, Years: ramadanYearOptions(time.Now().Year())}))
		return
	}
	tz := tzFor(r)
	methodID := methodFromCookie(r)
	now := time.Now().In(tz)

	year := now.Year() + 1
	if q := r.URL.Query().Get("year"); q != "" {
		if v, err := strconv.Atoi(q); err == nil {
			year = v
		}
	}

	start, end, hy := RamadanRange(year)
	days, err := RamadanSchedule(lat, lng, MethodByID(methodID), tz, start, end)
	if err != nil {
		renderTempl(w, RamadanSection(RamadanInfo{Year: year, Years: ramadanYearOptions(now.Year())}))
		return
	}
	todayIdx := -1
	for i, d := range days {
		if d.Gregorian.Format("2006-01-02") == now.Format("2006-01-02") {
			todayIdx = i
			break
		}
	}
	renderTempl(w, RamadanSection(RamadanInfo{
		HijriYear: hy,
		Start:     start.In(tz),
		End:       end.In(tz),
		Days:      days,
		TodayIdx:  todayIdx,
		Years:     ramadanYearOptions(now.Year()),
		Year:      year,
	}))
}

// ramadanYearOptions builds the lookahead selector: current year + next 4.
func ramadanYearOptions(currentYear int) []int {
	years := make([]int, 5)
	for i := range years {
		years[i] = currentYear + i
	}
	return years
}

func handleLocationDrawer(w http.ResponseWriter, r *http.Request) {
	data := PageData{Cities: Cities, Methods: Methods, MethodID: methodFromCookie(r)}
	renderTempl(w, LocationDrawer(data))
}

func handleSetLocation(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	lat, err1 := strconv.ParseFloat(r.FormValue("lat"), 64)
	lng, err2 := strconv.ParseFloat(r.FormValue("lng"), 64)
	city := r.FormValue("city")
	// City picker submits "Name|Lat|Lng" in the city field; GPS submits
	// separate lat/lng. Handle both.
	if err1 != nil || err2 != nil {
		if strings.Contains(city, "|") {
			parts := strings.SplitN(city, "|", 3)
			if len(parts) == 3 {
				city = parts[0]
				lat, err1 = strconv.ParseFloat(parts[1], 64)
				lng, err2 = strconv.ParseFloat(parts[2], 64)
			}
		}
	}
	// Method-only update: keep existing location cookie, just change method.
	if err1 != nil || err2 != nil {
		if c, e := r.Cookie(cookieLoc); e == nil {
			parts := strings.Split(c.Value, ",")
			if len(parts) == 2 {
				lat, err1 = strconv.ParseFloat(parts[0], 64)
				lng, err2 = strconv.ParseFloat(parts[1], 64)
			}
		}
	}
	if err1 != nil || err2 != nil {
		http.Error(w, "bad coords", http.StatusBadRequest)
		return
	}
	method := r.FormValue("method")
	if method == "" {
		method = "mwl"
	}
	setCookie(w, cookieLoc, fmt.Sprintf("%f,%f", lat, lng))
	setCookie(w, cookieMethod, method)
	if city != "" && !strings.Contains(city, "|") {
		setCookie(w, cookieCity, city)
	} else {
		http.SetCookie(w, &http.Cookie{Name: cookieCity, Value: "", MaxAge: -1, Path: "/"})
	}
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusNoContent)
}

func setCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 365,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// nextPrayer returns the next upcoming prayer and a humanized "in Xh Ym"
// duration string. If all of today's prayers have passed, signals "besok".
func nextPrayer(sched []PrayerTime, now time.Time) (*PrayerTime, string) {
	for i := range sched {
		if sched[i].Time.After(now) {
			return &sched[i], humanDur(sched[i].Time.Sub(now))
		}
	}
	if len(sched) > 0 {
		t := &sched[0]
		return t, "besok"
	}
	return nil, ""
}

func humanDur(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	switch {
	case h == 0 && m == 0:
		return "sekarang"
	case h == 0:
		return fmt.Sprintf("%d mnt", m)
	case m == 0:
		return fmt.Sprintf("%d jam", h)
	default:
		return fmt.Sprintf("%d jam %d mnt", h, m)
	}
}
