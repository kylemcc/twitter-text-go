// Package validate provides routines for validating tweets
package validate

import (
	"code.google.com/p/go.text/unicode/norm"
	"github.com/kylemcc/twitter-text-go/extract"
	"strings"
	"unicode/utf8"
)

const (
	maxLength           = 140
	shortUrlLength      = 22
	shortHttpsUrlLength = 23
)

var formC = norm.NFC

// Returns the length of the string as it would be displayed. This is equivilent to the length of the Unicode NFC
// (See: http://www.unicode.org/reports/tr15). This is needed in order to consistently calculate the length of a
// string no matter which actual form was transmitted. For example:
//
//     U+0065  Latin Small Letter E
// +   U+0301  Combining Acute Accent
// ----------
// =   2 bytes, 2 characters, displayed as é (1 visual glyph)
//     … The NFC of {U+0065, U+0301} is {U+00E9}, which is a single chracter and a +display_length+ of 1
//
// The string could also contain U+00E9 already, in which case the canonicalization will not change the value.
func TweetLength(text string) int {
	length := utf8.RuneCountInString(formC.String(text))

	urls := extract.ExtractUrls(text)
	for _, url := range urls {
		length -= url.Range.Length()
		if strings.HasPrefix(url.Text, "https://") {
			length += shortHttpsUrlLength
		} else {
			length += shortUrlLength
		}
	}
	return length
}

func TweetIsValid(text string) bool {
	return false
}

func UsernameIsValid(username string) bool {
	return false
}

func ListIsValid(list string) bool {
	return false
}

func HashtagIsValid(hashtag string) bool {
	return false
}

func UrlIsValid(url string) bool {
	return false
}
