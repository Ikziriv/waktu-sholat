package main

// Niat holds the intention (niat) text for one obligatory prayer.
type Niat struct {
	Name            PrayerName
	Arabic          string
	Transliteration string
	Translation     string
}

// NiatByPrayer is the lookup table for the five daily prayers' niat.
// Arabic text uses standard sholat niat wording; translations in Indonesian
// since the app is "waktu sholat".
var NiatByPrayer = map[PrayerName]Niat{
	Subuh: {
		Name:            Subuh,
		Arabic:          "أُصَلِّي فَرْضَ الصُّبْحِ رَكْعَتَيْنِ مُسْتَقْبِلَ الْقِبْلَةِ لِلَّهِ تَعَالَى",
		Transliteration: "Ushallii fardhash-shubhi rak'ataini mustaqbilal-qiblati lillaahi ta'aalaa",
		Translation:     "Aku niat sholat fardu Subuh dua rakaat menghadap kiblat karena Allah Ta'ala",
	},
	Zuhur: {
		Name:            Zuhur,
		Arabic:          "أُصَلِّي فَرْضَ الظُّهْرِ أَرْبَعَ رَكَعَاتٍ مُسْتَقْبِلَ الْقِبْلَةِ لِلَّهِ تَعَالَى",
		Transliteration: "Ushallii fardhazh-zhuhri arba'a raka'aatin mustaqbilal-qiblati lillaahi ta'aalaa",
		Translation:     "Aku niat sholat fardu Zuhur empat rakaat menghadap kiblat karena Allah Ta'ala",
	},
	Ashar: {
		Name:            Ashar,
		Arabic:          "أُصَلِّي فَرْضَ الْعَصْرِ أَرْبَعَ رَكَعَاتٍ مُسْتَقْبِلَ الْقِبْلَةِ لِلَّهِ تَعَالَى",
		Transliteration: "Ushallii fardhal-'ashri arba'a raka'aatin mustaqbilal-qiblati lillaahi ta'aalaa",
		Translation:     "Aku niat sholat fardu Ashar empat rakaat menghadap kiblat karena Allah Ta'ala",
	},
	Maghrib: {
		Name:            Maghrib,
		Arabic:          "أُصَلِّي فَرْضَ الْمَغْرِبِ ثَلَاثَ رَكَعَاتٍ مُسْتَقْبِلَ الْقِبْلَةِ لِلَّهِ تَعَالَى",
		Transliteration: "Ushallii fardhal-maghribi tsalaatsa raka'aatin mustaqbilal-qiblati lillaahi ta'aalaa",
		Translation:     "Aku niat sholat fardu Maghrib tiga rakaat menghadap kiblat karena Allah Ta'ala",
	},
	Isya: {
		Name:            Isya,
		Arabic:          "أُصَلِّي فَرْضَ الْعِشَاءِ أَرْبَعَ رَكَعَاتٍ مُسْتَقْبِلَ الْقِبْلَةِ لِلَّهِ تَعَالَى",
		Transliteration: "Ushallii fardhal-'isyaa-i arba'a raka'aatin mustaqbilal-qiblati lillaahi ta'aalaa",
		Translation:     "Aku niat sholat fardu Isya empat rakaat menghadap kiblat karena Allah Ta'ala",
	},
}

// PrayerLabel is the display name (Indonesian) for each prayer.
var PrayerLabel = map[PrayerName]string{
	Subuh:   "Subuh",
	Zuhur:   "Zuhur",
	Ashar:   "Ashar",
	Maghrib: "Maghrib",
	Isya:    "Isya",
}

// JamakMode distinguishes the two jamak performance modes.
type JamakMode string

const (
	JamakTaqdim JamakMode = "taqdim"  // earlier prayer performed at its time, later joined in
	JamakTakhir JamakMode = "taakhir" // later prayer performed at its time, earlier made up
)

// Jamak holds the niat text for one jamak combination. Subuh has no jamak,
// so only Zuhur+Ashar and Maghrib+Isya exist, each in two modes.
type Jamak struct {
	First, Second   PrayerName
	Mode            JamakMode
	Arabic          string
	Transliteration string
	Translation     string
}

// Label is the drawer header for a jamak entry, e.g. "Jamak Taqdim · Zuhur → Ashar".
func (j Jamak) Label() string {
	mode := "Taqdim"
	if j.Mode == JamakTakhir {
		mode = "Ta'khir"
	}
	return "Jamak " + mode + " · " + PrayerLabel[j.First] + " → " + PrayerLabel[j.Second]
}

// JamakList is the ordered list of the four valid jamak combinations.
// Ordered (not a map) because pairs have a natural display order.
var JamakList = []Jamak{
	{
		First: Zuhur, Second: Ashar, Mode: JamakTaqdim,
		Arabic:          "أُصَلِّي فَرْضَ الظُّهْرِ جَمْعَ تَقْدِيمٍ مَعَ الْعَصْرِ لِلَّهِ تَعَالَى",
		Transliteration: "Ushallii fardhazh-zhuhri jam'a taqdiimin ma'al-'ashri lillaahi ta'aalaa",
		Translation:     "Aku niat sholat fardu Zuhur dijamak taqdim dengan Ashar karena Allah Ta'ala",
	},
	{
		First: Zuhur, Second: Ashar, Mode: JamakTakhir,
		Arabic:          "أُصَلِّي فَرْضَ الْعَصْرِ جَمْعَ تَأْخِيرٍ مَعَ الظُّهْرِ لِلَّهِ تَعَالَى",
		Transliteration: "Ushallii fardhal-'ashri jam'a ta'khiirin ma'azh-zhuhri lillaahi ta'aalaa",
		Translation:     "Aku niat sholat fardu Ashar dijamak ta'khir dengan Zuhur karena Allah Ta'ala",
	},
	{
		First: Maghrib, Second: Isya, Mode: JamakTaqdim,
		Arabic:          "أُصَلِّي فَرْضَ الْمَغْرِبِ جَمْعَ تَقْدِيمٍ مَعَ الْعِشَاءِ لِلَّهِ تَعَالَى",
		Transliteration: "Ushallii fardhal-maghribi jam'a taqdiimin ma'al-'isyaa-i lillaahi ta'aalaa",
		Translation:     "Aku niat sholat fardu Maghrib dijamak taqdim dengan Isya karena Allah Ta'ala",
	},
	{
		First: Maghrib, Second: Isya, Mode: JamakTakhir,
		Arabic:          "أُصَلِّي فَرْضَ الْعِشَاءِ جَمْعَ تَأْخِيرٍ مَعَ الْمَغْرِبِ لِلَّهِ تَعَالَى",
		Transliteration: "Ushallii fardhal-'isyaa-i jam'a ta'khiirin ma'al-maghribi lillaahi ta'aalaa",
		Translation:     "Aku niat sholat fardu Isya dijamak ta'khir dengan Maghrib karena Allah Ta'ala",
	},
}

// JamakByQuery looks up a jamak entry by its query params (first, second, mode).
func JamakByQuery(first, second PrayerName, mode JamakMode) (Jamak, bool) {
	for _, j := range JamakList {
		if j.First == first && j.Second == second && j.Mode == mode {
			return j, true
		}
	}
	return Jamak{}, false
}
