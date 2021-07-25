// Package validate provides routines for validating tweets
package validate

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/kylemcc/twitter-text-go/extract"
	"golang.org/x/text/unicode/norm"
)

type weightRange struct {
	start  int32
	end    int32
	weight int
}

const (
	currentMaxLength    = 280
	maxV1Length         = 140
	shortUrlLength      = 23
	shortHttpsUrlLength = 23
	invalidChars        = "\uFFFE\uFEFF\uFFFF\u202A\u202B\u202C\u202D\u202E"
	defaultWeight       = 200
	scale               = 100
)

var weightRanges = [...]weightRange{
	{start: 0, end: 4351, weight: 100},
	{start: 8192, end: 8205, weight: 100},
	{start: 8208, end: 8223, weight: 100},
	{start: 8242, end: 8247, weight: 100},
}

var formC = norm.NFC

type Tweet struct {
	WeightedLength int
	Permillage     int
	Valid          bool
	// TODO Not yet implemented
	// DisplayRangeEnd   int
	// DisplayRangeStart int
	// ValidRangeEnd     int
	// ValidRangeStart   int
}

// Validation error returned when text is too long to be a valid tweet.
// The value of the error is the actual length of the input string
type TooLongError int

func (e TooLongError) Error() string {
	return fmt.Sprintf("Length %d exceeds %d characters", int(e), maxV1Length)
}

// Validation error returned when text is empty
type EmptyError struct{}

func (e EmptyError) Error() string {
	return "Tweets may not be empty"
}

// Validation error returned when text contains an invalid character
//
// This error embeds the value of the invalid character, and its
// byte-offset within the input string
type InvalidCharacterError struct {
	Character rune
	Offset    int
}

func (e InvalidCharacterError) Error() string {
	return fmt.Sprintf("Invalid chararcter [%s] found at byte offset %d", string(e.Character), e.Offset)
}

// Returns the V1 length of the string as it would be displayed. This is equivalent to the length of the Unicode NFC
// (See: http://www.unicode.org/reports/tr15). This is needed in order to consistently calculate the length of a
// string no matter which actual form was transmitted. For example:
//
//     U+0065  Latin Small Letter E
// +   U+0301  Combining Acute Accent
// ----------
// =   2 bytes, 2 characters, displayed as é (1 visual glyph)
//     … The NFC of {U+0065, U+0301} is {U+00E9}, which is a single character and a +display_length+ of 1
//
// The string could also contain U+00E9 already, in which case the canonicalization will not change the value.
// Note that per twitter-text, this method is deprecated and will always return the v1 values
// New clients should use ParseTweet
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

func ParseTweet(text string) (Tweet, error) {
	length := adjustedWeighedLength(text)
	err := validateTweet(text, currentMaxLength)
	tweet := Tweet{
		WeightedLength: length,
		Permillage:     1000 * (length / currentMaxLength),
		Valid:          err == nil,
	}
	return tweet, err
}

// Returns the length of the weightedLength of a tweet, per the twitter algorithm
func adjustedWeighedLength(text string) int {
	length := weightedLength(text)
	urls := extract.ExtractUrls(text)
	adjustments := 0
	for _, url := range urls {
		length -= weightedLength(url.Text)
		if strings.HasPrefix(url.Text, "https://") {
			adjustments += shortHttpsUrlLength
		} else {
			adjustments += shortUrlLength
		}
	}
	return adjustments + (length / scale)
}

func weightedLength(text string) int {
	normalized := formC.String(text)
	weightedLength := 0
	for _, val := range normalized {
		length := defaultWeight
		for _, weightRange := range weightRanges {
			if val >= weightRange.start && val <= weightRange.end {
				length = weightRange.weight
				break
			}
		}
		weightedLength += length
	}
	return weightedLength
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
	return validateTweet(text, maxV1Length)
}

func validateTweet(text string, maxLength int) error {
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

// Returns true if the given text represents a valid @username
func UsernameIsValid(username string) bool {
	if username == "" {
		return false
	}

	extracted := extract.ExtractMentionedScreenNames(username)
	return len(extracted) == 1 && extracted[0].Text == username
}

// Returns true if the given text represents a valid
// @twitter/list
func ListIsValid(list string) bool {
	if list == "" {
		return false
	}

	extracted := extract.ExtractMentionsOrLists(list)
	if len(extracted) == 1 {
		e := extracted[0]
		if _, ok := e.ListSlug(); ok && e.Text == list {
			return true
		}
	}
	return false
}

// Returns true if the given text represents a valid #hashtag
func HashtagIsValid(hashtag string) bool {
	if hashtag == "" {
		return false
	}

	extracted := extract.ExtractHashtags(hashtag)
	return len(extracted) == 1 && extracted[0].Text == hashtag
}

// Returns true if the given text represents a valid URL
func UrlIsValid(url string, requireProtocol bool, allowUnicode bool) bool {
	if url == "" {
		return false
	}

	match := validateUrlUnencodedRe.FindStringSubmatchIndex(url)
	if match == nil || url[match[0]:match[1]] != url {
		return false
	}

	if requireProtocol {
		schemeStart := match[validateUrlUnencodedGroupScheme*2]
		schemeEnd := match[validateUrlUnencodedGroupScheme*2+1]
		if !protocolRe.MatchString(url[schemeStart:schemeEnd]) {
			return false
		}
	}

	pathStart := match[validateUrlUnencodedGroupPath*2]
	pathEnd := match[validateUrlUnencodedGroupPath*2+1]
	if !validateUrlPathRe.MatchString(url[pathStart:pathEnd]) {
		return false
	}

	queryStart := match[validateUrlUnencodedGroupQuery*2]
	queryEnd := match[validateUrlUnencodedGroupQuery*2+1]
	if queryStart > 0 && !validateUrlQueryRe.MatchString(url[queryStart:queryEnd]) {
		return false
	}

	fragmentStart := match[validateUrlUnencodedGroupFragment*2]
	fragmentEnd := match[validateUrlUnencodedGroupFragment*2+1]
	if fragmentStart > 0 && !validateUrlFragmentRe.MatchString(url[fragmentStart:fragmentEnd]) {
		return false
	}

	authorityStart := match[validateUrlUnencodedGroupAuthority*2]
	authorityEnd := match[validateUrlUnencodedGroupAuthority*2+1]
	authority := url[authorityStart:authorityEnd]

	if allowUnicode {
		return validateUrlUnicodeAuthorityRe.MatchString(authority)
	} else {
		return validateUrlAuthorityRe.MatchString(authority)
	}
}
