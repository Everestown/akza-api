package slug

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	reDashes  = regexp.MustCompile(`-{2,}`)
	reNonSlug = regexp.MustCompile(`[^a-z0-9-]`)
)

var cyrillicMap = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d",
	'е': "e", 'ё': "yo", 'ж': "zh", 'з': "z", 'и': "i",
	'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n",
	'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t",
	'у': "u", 'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch",
	'ш': "sh", 'щ': "shch", 'ъ': "", 'ы': "y", 'ь': "",
	'э': "e", 'ю': "yu", 'я': "ya",
}

func transliterate(s string) string {
	var b strings.Builder
	for _, r := range s {
		if lat, ok := cyrillicMap[unicode.ToLower(r)]; ok {
			b.WriteString(lat)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func removeAccents(s string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(func(r rune) bool {
		return unicode.Is(unicode.Mn, r)
	}), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

// Generate converts any text into a URL-safe slug.
func Generate(text string) string {
	s := strings.ToLower(text)
	s = transliterate(s)
	s = removeAccents(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = reNonSlug.ReplaceAllString(s, "-")
	s = reDashes.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// GenerateUnique appends numeric suffix until exists() returns false.
func GenerateUnique(text string, exists func(string) bool) string {
	base := Generate(text)
	candidate := base
	for i := 2; exists(candidate); i++ {
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
	return candidate
}
