// Package validate provides routines for validating tweets
package validate

import (
	"code.google.com/p/go.text/unicode/norm"
	"fmt"
	"github.com/kylemcc/twitter-text-go/extract"
	"strings"
	"unicode/utf8"
)

const (
	maxLength           = 140
	shortUrlLength      = 22
	shortHttpsUrlLength = 23
	invalidChars        = "\uFFFE\uFEFF\uFFFF\u202A\u202B\u202C\u202D\u202E"
)

var formC = norm.NFC

type TooLongError int

func (e TooLongError) Error() string {
	return fmt.Sprintf("Length %d exceeds %d characters", int(e), maxLength)
}

type EmptyError struct{}

func (e EmptyError) Error() string {
	return "Tweets may not be empty"
}

type InvalidCharacterError struct {
	Character rune
	Offset    int
}

func (e InvalidCharacterError) Error() string {
	return fmt.Sprintf("Invalid chararcter [%s] found at byte offset %d", e.Character, e.Offset)
}

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

// Checks whether a string is a valid tweet and returns true or false
func TweetIsValid(text string) bool {
	err := ValidateTweet(text)
	return err == nil
}

// Checks whether a string is a valid tweet. Returns nil if the string
// is valid. Otherwise, it returns an error in the following cases:
//
// - The text is too long
// - The text is empty
// - The text contains invalid characters
func ValidateTweet(text string) error {
	if text == "" {
		return EmptyError{}
	} else if length := TweetLength(text); length > maxLength {
		return TooLongError(length)
	} else if i := strings.IndexAny(text, invalidChars); i > -1 {
		r, _ := utf8.DecodeRuneInString(text[i:])
		return InvalidCharacterError{Offset: i, Character: r}
	}
	return nil
}

func UsernameIsValid(username string) bool {
	if username == "" {
		return false
	}

	extracted := extract.ExtractMentionedScreenNames(username)
	return len(extracted) == 1 && extracted[0].Text == username
}

func ListIsValid(list string) bool {
	if list == "" {
		return false
	}

	extracted := extract.ExtractMentionsOrLists(list)
	return len(extracted) == 1 && extracted[0].Text == list && extracted[0].ListSlug != ""
}

func HashtagIsValid(hashtag string) bool {
    if hashtag == "" {
        return false
    }

    extracted := extract.ExtractHashtags(hashtag)
    return len(extracted) == 1 && extracted[0].Text == hashtag
}

func UrlIsValid(url string) bool {
	return false
}
