package main

import (
	"fmt"
	"math"
	"time"

	"github.com/hablullah/go-prayer"
)

// Ka'bah coordinates (Masjid al-Haram, Mecca).
const (
	kaabaLat = 21.4225247
	kaabaLng = 39.8261819
)

// PrayerName labels the five obligatory prayers plus sunrise for display.
type PrayerName string

const (
	Subuh   PrayerName = "subuh"
	Zuhur   PrayerName = "zuhur"
	Ashar   PrayerName = "ashar"
	Maghrib PrayerName = "maghrib"
	Isya    PrayerName = "isya"
)

// PrayerTime pairs a prayer name with its computed time for today.
type PrayerTime struct {
	Name PrayerName
	Time time.Time
}

// MethodByID resolves a calculation method id from a cookie value to a
// TwilightConvention. Defaults to MWL.
func MethodByID(id string) *prayer.TwilightConvention {
	switch id {
	case "kemenag":
		return prayer.Kemenag()
	case "isna":
		return prayer.ISNA()
	case "egypt":
		return prayer.Egypt()
	case "karachi":
		return prayer.Karachi()
	case "umm-al-qura":
		return prayer.UmmAlQura()
	case "tehran":
		return prayer.Tehran()
	case "jafari":
		return prayer.Jafari()
	case "mwl":
		fallthrough
	default:
		return prayer.MWL()
	}
}

// Methods is the user-facing list of calculation methods.
type Method struct {
	ID   string
	Name string
}

var Methods = []Method{
	{"mwl", "Muslim World League"},
	{"kemenag", "Kemenag (Indonesia)"},
	{"isna", "ISNA (North America)"},
	{"egypt", "Egyptian General Authority"},
	{"karachi", "Univ. of Karachi"},
	{"umm-al-qura", "Umm al-Qura (Saudi)"},
	{"tehran", "Tehran"},
	{"jafari", "Jafari (Shia)"},
}

// TodaySchedule returns the five prayer times for today at the given
// location, computed with the given twilight convention.
func TodaySchedule(lat, lng float64, method *prayer.TwilightConvention, tz *time.Location) ([]PrayerTime, error) {
	cfg := prayer.Config{
		Latitude:           lat,
		Longitude:          lng,
		Timezone:           tz,
		TwilightConvention: method,
		AsrConvention:      prayer.Shafii,
	}
	schedules, err := prayer.Calculate(cfg, time.Now().Year())
	if err != nil {
		return nil, err
	}
	today := time.Now().In(tz).Format("2006-01-02")
	for _, s := range schedules {
		if s.Date == today {
			return []PrayerTime{
				{Subuh, s.Fajr.In(tz)},
				{Zuhur, s.Zuhr.In(tz)},
				{Ashar, s.Asr.In(tz)},
				{Maghrib, s.Maghrib.In(tz)},
				{Isya, s.Isha.In(tz)},
			}, nil
		}
	}
	return nil, fmt.Errorf("no schedule for today (%s)", today)
}

// QiblaBearing returns the initial great-circle bearing in degrees
// (0–360, clockwise from true north) from the user's location to the Ka'bah.
func QiblaBearing(lat, lng float64) float64 {
	φ1 := lat * math.Pi / 180
	φ2 := kaabaLat * math.Pi / 180
	Δλ := (kaabaLng - lng) * math.Pi / 180
	y := math.Sin(Δλ) * math.Cos(φ2)
	x := math.Cos(φ1)*math.Sin(φ2) - math.Sin(φ1)*math.Cos(φ2)*math.Cos(Δλ)
	θ := math.Atan2(y, x) * 180 / math.Pi
	b := math.Mod(θ+360, 360)
	return b
}

// Cardinal returns the 16-wind compass label for a bearing in degrees.
func Cardinal(b float64) string {
	sectors := []string{
		"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE",
		"S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW",
	}
	idx := int(math.Mod(b+11.25, 360) / 22.5)
	return sectors[idx]
}

// demo is a runnable self-check for the qibla formula — the only non-trivial
// math in this file. Run with: go run -tags=check . check
// Verified against known references:
//   - Jakarta (~6.2S, 106.8E) → ~295° (WNW)
//   - London  (~51.5N, 0.0)   → ~119° (ESE)
//   - Mecca   (Ka'bah itself) → 0° (bearing undefined, but should be ~0)
func runSelfCheck() error {
	cases := []struct {
		name       string
		lat, lng   float64
		wantApprox float64
	}{
		{"Jakarta", -6.2088, 106.8456, 295},
		{"London", 51.5074, -0.1278, 119},
		{"New York", 40.7128, -74.0060, 58},
	}
	for _, c := range cases {
		got := QiblaBearing(c.lat, c.lng)
		diff := math.Abs(got - c.wantApprox)
		if diff > 2 {
			return fmt.Errorf("qibla %s: got %.1f, want ~%.1f (diff %.1f)", c.name, got, c.wantApprox, diff)
		}
		fmt.Printf("  %s: %.1f° %s  ok\n", c.name, got, Cardinal(got))
	}
	// Ka'bah itself: bearing should be ~0 (sin and cos of Δλ both ~0).
	k := QiblaBearing(kaabaLat, kaabaLng)
	if k > 1 && k < 359 {
		return fmt.Errorf("qibla at Ka'bah: got %.1f, want ~0", k)
	}
	fmt.Printf("  Ka'bah: %.1f°  ok\n", k)
	if err := ramadanSelfCheck(time.Now()); err != nil {
		return fmt.Errorf("ramadan: %w", err)
	}
	return nil
}
