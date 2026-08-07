package main

import (
	"fmt"
	"time"

	"github.com/hablullah/go-hijri"
	"github.com/hablullah/go-prayer"
)

// imsakOffset is the safety margin before Fajr that marks the start of the
// fast. 10 minutes is the convention used by Kemenag and MWL schedules.
const imsakOffset = 10 * time.Minute

// HijriMonthName is the Indonesian display name for each Hijri month.
// Ramadan is month 9; only the months we display are filled, the rest are
// kept for completeness since the lookup is by index.
var HijriMonthName = map[int64]string{
	1: "Muharram", 2: "Safar", 3: "Rabiul Awal", 4: "Rabiul Akhir",
	5: "Jumadil Awal", 6: "Jumadil Akhir", 7: "Rajab", 8: "Sya'ban",
	9: "Ramadhan", 10: "Syawal", 11: "Dzulqaidah", 12: "Dzulhijjah",
}

// RamadanDay is one day of Ramadhan with its fasting + prayer schedule.
// Imsak = Fajr - 10m (start of fast); Maghrib = end of fast.
type RamadanDay struct {
	HijriDay  int
	Gregorian time.Time
	Imsak     time.Time
	Subuh     time.Time
	Zuhur     time.Time
	Ashar     time.Time
	Maghrib   time.Time
	Isya      time.Time
}

// RamadanInfo bundles a Ramadhan window with its schedule for the template.
type RamadanInfo struct {
	HijriYear int
	Start     time.Time // Gregorian 1 Ramadhan
	End       time.Time // Gregorian last day of Ramadhan
	Days      []RamadanDay
	// TodayIdx is the index into Days for today, or -1 if today is not in Ramadhan.
	TodayIdx int
	// Years is the list of Gregorian years offered in the lookahead selector.
	Years []int
	// Year is the Gregorian year currently selected.
	Year int
}

// RamadanRange returns the Gregorian start and end dates (inclusive) of
// Ramadhan for the Hijri year whose Ramadhan falls in or after the given
// Gregorian year, using the Umm al-Qura calendar. Also returns that Hijri year.
//
// ponytail: Umm al-Qura is tabulated to 2077; beyond that CreateUmmAlQuraDate
// errors. Ceiling is fine for a prayer app — upgrade to astronomical hilal
// calc if multi-century lookahead is ever needed.
func RamadanRange(gregorianYear int) (start, end time.Time, hijriYear int) {
	sample := time.Date(gregorianYear, 1, 1, 12, 0, 0, 0, time.UTC)
	if hd, err := hijri.CreateUmmAlQuraDate(sample); err == nil {
		hijriYear = int(hd.Year)
	} else {
		hijriYear = gregorianYear - 579 // rough fallback: 1 AH ≈ 622 CE
	}
	start = hijri.UmmAlQuraDate{Year: int64(hijriYear), Month: 9, Day: 1}.ToGregorian()
	if start.Year() < gregorianYear {
		hijriYear++
		start = hijri.UmmAlQuraDate{Year: int64(hijriYear), Month: 9, Day: 1}.ToGregorian()
	}
	shawwal1 := hijri.UmmAlQuraDate{Year: int64(hijriYear), Month: 10, Day: 1}.ToGregorian()
	end = shawwal1.AddDate(0, 0, -1)
	return start, end, hijriYear
}

// RamadanSchedule computes the daily Imsak + five prayer times for every day
// of the Ramadhan window [start, end] at the given location, in tz.
// prayer.Calculate is invoked per Gregorian year the window spans (Ramadhan
// can cross a year boundary in late-December edge cases).
func RamadanSchedule(lat, lng float64, method *prayer.TwilightConvention, tz *time.Location, start, end time.Time) ([]RamadanDay, error) {
	lookup := map[int]map[string]prayer.Schedule{}
	for y := start.In(tz).Year(); y <= end.In(tz).Year(); y++ {
		cfg := prayer.Config{
			Latitude:           lat,
			Longitude:          lng,
			Timezone:           tz,
			TwilightConvention: method,
			AsrConvention:      prayer.Shafii,
		}
		schedules, err := prayer.Calculate(cfg, y)
		if err != nil {
			return nil, err
		}
		m := make(map[string]prayer.Schedule, len(schedules))
		for _, s := range schedules {
			m[s.Date] = s
		}
		lookup[y] = m
	}

	var days []RamadanDay
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		ld := d.In(tz)
		s, ok := lookup[ld.Year()][ld.Format("2006-01-02")]
		if !ok {
			continue
		}
		hd, _ := hijri.CreateUmmAlQuraDate(ld)
		days = append(days, RamadanDay{
			HijriDay:  int(hd.Day),
			Gregorian: ld,
			Imsak:     s.Fajr.In(tz).Add(-imsakOffset),
			Subuh:     s.Fajr.In(tz),
			Zuhur:     s.Zuhr.In(tz),
			Ashar:     s.Asr.In(tz),
			Maghrib:   s.Maghrib.In(tz),
			Isya:      s.Isha.In(tz),
		})
	}
	return days, nil
}

// ramadanSelfCheck verifies the Hijri conversion round-trips structurally:
// the computed start converts back to 1 Ramadhan, the day after end converts
// to 1 Syawal, and the month length is 29 or 30 days. Run via `go run . check`.
func ramadanSelfCheck(now time.Time) error {
	start, end, hy := RamadanRange(now.Year())
	if hy == 0 {
		return fmt.Errorf("ramadan: no hijri year resolved for %d", now.Year())
	}
	hd, err := hijri.CreateUmmAlQuraDate(start)
	if err != nil {
		return fmt.Errorf("ramadan start convert: %w", err)
	}
	if hd.Month != 9 || hd.Day != 1 {
		return fmt.Errorf("ramadan start: got %d-%d-%d, want 9-1", hd.Year, hd.Month, hd.Day)
	}
	shawwal, err := hijri.CreateUmmAlQuraDate(end.AddDate(0, 0, 1))
	if err != nil {
		return fmt.Errorf("ramadan end+1 convert: %w", err)
	}
	if shawwal.Month != 10 || shawwal.Day != 1 {
		return fmt.Errorf("ramadan end: day after end got %d-%d-%d, want 10-1", shawwal.Year, shawwal.Month, shawwal.Day)
	}
	length := int(end.Sub(start).Hours()/24) + 1
	if length < 29 || length > 30 {
		return fmt.Errorf("ramadan length: got %d days, want 29 or 30", length)
	}
	fmt.Printf("  Ramadhan %dH: %s → %s (%d days)  ok\n", hy, start.Format("2006-01-02"), end.Format("2006-01-02"), length)
	return nil
}
