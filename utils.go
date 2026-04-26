package main

var languageNames = map[string]string{
	"ja-JP":  "Japanese",
	"en-US":  "English",
	"en-IN":  "English (India)",
	"id-ID":  "Bahasa Indonesia",
	"ms-MY":  "Bahasa Melayu",
	"ca-ES":  "Català",
	"de-DE":  "Deutsch",
	"es-419": "Español (América Latina)",
	"es-ES":  "Español (España)",
	"fr-FR":  "Français",
	"it-IT":  "Italiano",
	"pl-PL":  "Polski",
	"pt-BR":  "Português (Brasil)",
	"pt-PT":  "Português (Portugal)",
	"vi-VN":  "Tiếng Việt",
	"tr-TR":  "Türkçe",
	"ru-RU":  "Русский",
	"ar-SA":  "العربية",
	"hi-IN":  "हिंदी",
	"ta-IN":  "தமிழ்",
	"te-IN":  "తెలుగు",
	"zh-CN":  "中文 (普通话)",
	"zh-HK":  "中文 (粵語)",
	"zh-TW":  "中文 (國語)",
	"ko-KR":  "한국어",
	"th-TH":  "ไทย",
}

// isoLanguageCodes maps Crunchyroll locale codes to ISO 639-2/B codes for FFmpeg metadata
var isoLanguageCodes = map[string]string{
	"ja-JP":  "jpn",
	"en-US":  "eng",
	"en-IN":  "eng",
	"id-ID":  "ind",
	"ms-MY":  "may",
	"ca-ES":  "cat",
	"de-DE":  "ger",
	"es-419": "spa",
	"es-ES":  "spa",
	"fr-FR":  "fre",
	"it-IT":  "ita",
	"pl-PL":  "pol",
	"pt-BR":  "por",
	"pt-PT":  "por",
	"vi-VN":  "vie",
	"tr-TR":  "tur",
	"ru-RU":  "rus",
	"ar-SA":  "ara",
	"hi-IN":  "hin",
	"ta-IN":  "tam",
	"te-IN":  "tel",
	"zh-CN":  "chi",
	"zh-HK":  "chi",
	"zh-TW":  "chi",
	"ko-KR":  "kor",
	"th-TH":  "tha",
}

// getISOCode returns the ISO 639-2/B code for a Crunchyroll locale, falling back to the locale itself
func getISOCode(locale string) string {
	if code, ok := isoLanguageCodes[locale]; ok {
		return code
	}
	return locale
}
