package slug

import (
	"strings"
	"unicode"
)

var stopWords = map[string]bool{
	// English
	"a": true, "an": true, "the": true, "to": true, "for": true,
	"of": true, "in": true, "on": true, "at": true, "with": true,
	"using": true, "via": true,
	// Portuguese
	"o": true, "os": true, "as": true, "de": true, "do": true,
	"da": true, "em": true, "para": true, "com": true, "por": true,
	"uma": true, "um": true,
}

const prefix = "branchly/"
const maxChars = 50

// GenerateSlug derives a short, semantic branch slug from a user prompt.
// The result is always prefixed with "branchly/" and at most 50 characters.
func GenerateSlug(prompt string) string {
	// Lowercase and replace non-alphanumeric characters with spaces.
	var buf strings.Builder
	for _, r := range strings.ToLower(prompt) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			buf.WriteRune(r)
		} else {
			buf.WriteByte(' ')
		}
	}

	// Split into words and remove stop words.
	words := strings.Fields(buf.String())
	filtered := make([]string, 0, len(words))
	for _, w := range words {
		if !stopWords[w] {
			filtered = append(filtered, w)
		}
	}

	if len(filtered) == 0 {
		return prefix + "task"
	}

	// Try 4, 3, 2, 1 words until the full slug fits within maxChars.
	for n := 4; n >= 1; n-- {
		take := filtered
		if len(take) > n {
			take = filtered[:n]
		}
		full := prefix + strings.Join(take, "-")
		if len([]rune(full)) <= maxChars {
			return full
		}
	}

	// Last resort: truncate single word to fit.
	word := []rune(filtered[0])
	max := maxChars - len([]rune(prefix))
	if len(word) > max {
		word = word[:max]
	}
	return prefix + string(word)
}
