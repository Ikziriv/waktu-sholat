package main

// Surah holds a Quranic surah for display (onboarding welcome, etc.).
type Surah struct {
	Number          int
	Name            string
	Arabic          string
	Transliteration string
	Translation     string
}

// SurahAlAsr — Surah 103, shown on the first-run onboarding screen.
// Arabic + transliteration + Indonesian translation (Kemenag).
var SurahAlAsr = Surah{
	Number:          103,
	Name:            "Al-'Asr",
	Arabic:          "بِسْمِ اللَّهِ الرَّحْمَٰنِ الرَّحِيمِ ۝ وَالْعَصْرِ ۝ إِنَّ الْإِنْسَانَ لَفِي خُسْرٍ ۝ إِلَّا الَّذِينَ آمَنُوا وَعَمِلُوا الصَّالِحَاتِ وَتَوَاصَوْا بِالْحَقِّ وَتَوَاصَوْا بِالصَّبْرِ",
	Transliteration: "Bismillāhir-raḥmānir-raḥīm. Wal-'aṣr. Innal-insāna lafī khusr. Illalladzīna āmanū wa 'amilush-shāliḥāti wa tawāṣau bil-ḥaqqi wa tawāṣau bish-shabr.",
	Translation:     "Dengan nama Allah Yang Maha Pengasih, Maha Penyayang. Demi masa. Sesungguhnya manusia benar-benar berada dalam kerugian, kecuali orang-orang yang beriman dan mengerjakan amal saleh serta saling menasihati untuk kebenaran dan saling menasihati untuk kesabaran.",
}
