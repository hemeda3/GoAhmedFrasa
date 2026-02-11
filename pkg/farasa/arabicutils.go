package farasa

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Arabic letter constants
const (
	ALEF             = '\u0627'
	ALEF_MADDA       = '\u0622'
	ALEF_HAMZA_ABOVE = '\u0623'
	ALEF_HAMZA_BELOW = '\u0625'

	HAMZA         = '\u0621'
	HAMZA_ON_NABRA = '\u0624'
	HAMZA_ON_WAW   = '\u0626'

	YEH         = '\u064A'
	DOTLESS_YEH = '\u0649'

	TEH_MARBUTA = '\u0629'
	HEH         = '\u0647'
)

// String constants
const (
	AllArabicLetters = "\u0621\u0622\u0623\u0624\u0625\u0626\u0627\u0628\u0629\u062A\u062B\u062C\u062D\u062E\u062F" +
		"\u0630\u0631\u0632\u0633\u0634\u0635\u0636\u0637\u0638\u0639\u063A\u0641\u0642\u0643\u0644\u0645\u0646\u0647\u0648\u0649\u064A"

	AllHindiDigits = "\u0660\u0661\u0662\u0663\u0664\u0665\u0666\u0667\u0668\u0669"

	AllArabicLettersAndHindiDigits = "\u0621\u0622\u0623\u0624\u0625\u0626\u0627\u0628\u0629\u062A\u062B\u062C\u062D\u062E\u062F" +
		"\u0630\u0631\u0632\u0633\u0634\u0635\u0636\u0637\u0638\u0639\u063A\u0641\u0642\u0643\u0644\u0645\u0646\u0647\u0648\u0649\u064A" +
		"\u0660\u0661\u0662\u0663\u0664\u0665\u0666\u0667\u0668\u0669"

	AllDigits     = "0123456789"
	ALLDelimiters = "\u0020\u0000-\u002F\u003A-\u0040\u007B-\u00BB\u005B-\u005D\u005F-\u0060\u005E\u0600-\u060C\u06D4-\u06ED\ufeff"
)

// Prefixes and suffixes used in Arabic morphology
var Prefixes = []string{
	"\u0627\u0644", "\u0648", "\u0641", "\u0628", "\u0643", "\u0644", "\u0644\u0644", "\u0633",
}

var Suffixes = []string{
	"\u0647", "\u0647\u0627", "\u0643", "\u064a", "\u0647\u0645\u0627", "\u0643\u0645\u0627", "\u0646\u0627", "\u0643\u0645", "\u0647\u0645", "\u0647\u0646", "\u0643\u0646",
	"\u0627", "\u0627\u0646", "\u064a\u0646", "\u0648\u0646", "\u0648\u0627", "\u0627\u062a", "\u062a", "\u0646", "\u0629",
}

// Compiled regex patterns
var (
	emailRegex        = regexp.MustCompile(`[a-zA-Z0-9\-\._]+@[a-zA-Z0-9\-\._]+`)
	pAllDiacritics    = regexp.MustCompile("[\u0640\u064b\u064c\u064d\u064e\u064f\u0650\u0651\u0652\u0670]")
	pAllNonCharacters = regexp.MustCompile("[\u0020\u2000-\u200F\u2028-\u202F\u205F-\u206F\uFEFF]+")
	pAllDelimiters    = regexp.MustCompile("[" + ALLDelimiters + "]+")
	reZeroWidthSpaces = regexp.MustCompile("[\u200B\ufeff]+")
	reTabNewline      = regexp.MustCompile("[\t\n\r]")
)

// replaceChars replaces each character in 'from' with the corresponding character in 'to'.
// This is equivalent to Apache Commons StringUtils.replaceChars.
func replaceChars(input, from, to string) string {
	fromRunes := []rune(from)
	toRunes := []rune(to)

	charMap := make(map[rune]rune, len(fromRunes))
	for i, r := range fromRunes {
		if i < len(toRunes) {
			charMap[r] = toRunes[i]
		}
	}

	result := make([]rune, 0, utf8.RuneCountInString(input))
	for _, r := range input {
		if replacement, ok := charMap[r]; ok {
			result = append(result, replacement)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// Buck2Morph converts Buckwalter to morphological representation
func Buck2Morph(input string) string {
	buck := "$Y'|&}*<>&}"
	morph := "PyAAAAOAAAA"
	return replaceChars(input, buck, morph)
}

// UTF82Buck converts UTF-8 Arabic to Buckwalter transliteration
func UTF82Buck(input string) string {
	ar := "\u0627\u0625\u0622\u0623\u0621\u0628\u062a\u062b\u062c\u062d\u062e\u062f\u0630\u0631\u0632\u0633\u0634\u0635\u0636\u0637\u0638\u0639\u063a\u0641\u0642\u0643\u0644\u0645\u0646\u0647\u0648\u064a\u0649\u0629\u0624\u0626\u064e\u064b\u064f\u064c\u0650\u064d\u0652\u0651"
	buck := "A<|>'btvjHxd*rzs$SDTZEgfqklmnhwyYp&}aFuNiKo~"
	return replaceChars(input, ar, buck)
}

// Buck2UTF8 converts Buckwalter transliteration to UTF-8 Arabic
func Buck2UTF8(input string) string {
	ar := "\u0627\u0625\u0622\u0623\u0621\u0628\u062a\u062b\u062c\u062d\u062e\u062f\u0630\u0631\u0632\u0633\u0634\u0635\u0636\u0637\u0638\u0639\u063a\u0641\u0642\u0643\u0644\u0645\u0646\u0647\u0648\u064a\u0649\u0629\u0624\u0626\u064e\u064b\u064f\u064c\u0650\u064d\u0652\u0651"
	buck := "A<|>'btvjHxd*rzs$SDTZEgfqklmnhwyYp&}aFuNiKo~"
	return replaceChars(input, buck, ar)
}

// RemoveDiacritics removes all Arabic diacritical marks
func RemoveDiacritics(s string) string {
	return pAllDiacritics.ReplaceAllString(s, "")
}

// RemoveNonCharacters replaces non-printable unicode characters with spaces
func RemoveNonCharacters(s string) string {
	return pAllNonCharacters.ReplaceAllString(s, " ")
}

// Normalize normalizes Arabic text (diacritics removal + lam-lam expansion)
func Normalize(s string) string {
	// IF Starts with lam-lam
	if strings.HasPrefix(s, "\u0644\u0644") {
		s = "\u0644\u0627\u0644" + s[len("\u0644\u0644"):]
	}
	// If starts with waw-lam-lam
	if strings.HasPrefix(s, "\u0648\u0644\u0644") {
		s = "\u0648\u0644\u0627\u0644" + s[len("\u0648\u0644\u0644"):]
	}
	s = pAllDiacritics.ReplaceAllString(s, "")
	return s
}

// NormalizeFull performs full normalization including hamza, ta marbuta, etc.
func NormalizeFull(s string) string {
	// IF Starts with lam-lam
	if strings.HasPrefix(s, "\u0644\u0644") {
		s = "\u0644\u0627\u0644" + s[len("\u0644\u0644"):]
	}
	// If starts with waw-lam-lam
	if strings.HasPrefix(s, "\u0648\u0644\u0644") {
		s = "\u0648\u0644\u0627\u0644" + s[len("\u0648\u0644\u0644"):]
	}

	s = strings.ReplaceAll(s, string(ALEF_MADDA), string(ALEF))
	s = strings.ReplaceAll(s, string(ALEF_HAMZA_ABOVE), string(ALEF))
	s = strings.ReplaceAll(s, string(ALEF_HAMZA_BELOW), string(ALEF))
	s = strings.ReplaceAll(s, string(DOTLESS_YEH), string(YEH))
	s = strings.ReplaceAll(s, string(HAMZA_ON_NABRA), string(HAMZA))
	s = strings.ReplaceAll(s, string(HAMZA_ON_WAW), string(HAMZA))
	s = strings.ReplaceAll(s, string(TEH_MARBUTA), string(HEH))

	s = pAllDiacritics.ReplaceAllString(s, "")
	return s
}

// containsRune checks if a string contains a specific substring (single char optimization)
func containsStr(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}

// runeAt returns the substring of length 1 rune at position i
func runeSlice(s string, start, end int) string {
	runes := []rune(s)
	if start < 0 || end > len(runes) {
		return ""
	}
	return string(runes[start:end])
}

// runeLen returns the rune count of a string
func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}

// charBasedTokenizer splits a word based on delimiter/punctuation characters
func charBasedTokenizer(s string) string {
	runes := []rune(s)
	var sb strings.Builder
	extendedLetters := AllArabicLettersAndHindiDigits + "\u0640\u064b\u064c\u064d\u064e\u064f\u0650\u0651\u0652\u0670" +
		"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" +
		"\u00C0\u00C1\u00C2\u00C3\u00C4\u00C5\u00C6\u00C7\u00C8\u00C9\u00CB\u00CC\u00CD\u00CE\u00CF\u00D0\u00D1\u00D2\u00D3\u00D4\u00D5\u00D6\u00D8\u00D9\u00DA\u00DB\u00DC\u00DD\u00DE\u00DF" +
		"\u00E0\u00E1\u00E2\u00E3\u00E4\u00E5\u00E6\u00E7\u00E8\u00E9\u00EA\u00EB\u00EC\u00ED\u00EE\u00EF\u00F0\u00F1\u00F2\u00F3\u00F4\u00F5\u00F8\u00F9\u00FA\u00FB\u00FC\u00FD\u00FE\u00FF"

	for i := 0; i < len(runes); i++ {
		ch := string(runes[i])

		if pAllDelimiters.MatchString(ch) {
			sb.WriteString(" " + ch + " ")
		} else if ch == "." || ch == "," {
			if i == 0 {
				sb.WriteString(ch + " ")
			} else if i == len(runes)-1 {
				sb.WriteString(" " + ch)
			} else if strings.Contains(AllDigits, string(runes[i-1])) && strings.Contains(AllDigits, string(runes[i+1])) {
				sb.WriteString(ch)
			} else {
				sb.WriteString(" " + ch + " ")
			}
		} else if !strings.Contains(extendedLetters, ch) {
			sb.WriteString(" " + ch + " ")
		} else {
			if i == 0 {
				sb.WriteString(ch)
			} else {
				prevCh := string(runes[i-1])
				if (strings.Contains(AllDigits, ch) && strings.Contains(AllArabicLetters, prevCh)) ||
					(strings.Contains(AllDigits, prevCh) && strings.Contains(AllArabicLetters, ch)) {
					sb.WriteString(" " + ch)
				} else {
					sb.WriteString(ch)
				}
			}
		}
	}
	return sb.String()
}

// Tokenize splits Arabic text into tokens with normalization
func Tokenize(s string) []string {
	s = RemoveNonCharacters(s)
	s = RemoveDiacritics(s)
	s = reTabNewline.ReplaceAllString(s, " ")

	var output []string
	words := strings.Split(s, " ")
	for _, w := range words {
		if len(w) == 0 {
			continue
		}
		if strings.HasPrefix(w, "#") ||
			strings.HasPrefix(w, "@") ||
			strings.HasPrefix(w, ":") ||
			strings.HasPrefix(w, ";") ||
			strings.HasPrefix(w, "http://") ||
			emailRegex.MatchString(w) {
			output = append(output, w)
		} else {
			tokenized := charBasedTokenizer(w)
			for _, ss := range strings.Split(tokenized, " ") {
				ss = strings.TrimSpace(ss)
				if len(ss) > 0 {
					if strings.HasPrefix(ss, "\u0644\u0644") {
						output = append(output, "\u0644\u0627\u0644"+ss[len("\u0644\u0644"):])
					} else {
						output = append(output, ss)
					}
				}
			}
		}
	}
	return output
}
