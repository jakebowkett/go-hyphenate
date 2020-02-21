/*
Package hyphenate provides a simple way to hyphenate text.
It uses github.com/speedata/hyphenation along with some
additional tweaks to better accomodate English text.
*/
package hyphenate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/speedata/hyphenation"
)

const shyHyphen = "­"

var (
	hyphens = []string{"-", "–", "—", shyHyphen}
	wordSep = "/" + strings.Join(hyphens, "")
)

type Hyphenator struct {
	hyphen string
	lang   *hyphenation.Lang
	custom map[string][]string
}

/*
New returns a Hyphenator that will use the hyphenation
patterns defined by the file at path. Patterns can
be found here:

	http://ctan.math.utah.edu/ctan/tex-archive/language/hyph-utf8/tex/generic/hyph-utf8/patterns/txt/

The hyphen parameter specifies the string that is
used to hyphenate words.

Callers may also override the ruleset at path by
supplying custom. When a word supplied to the
Hyphenate method is found as a key in custom the
corresponding value will be designate how the
word will be broken up.

While capitalisation will be preserved in calls
to the Hyphenate method, the custom map should
use lower case spellings otherwise they will not
be considered.

	custom := map[string][]string{
		"hello": []string{"h", "ello"},
	}

	h, err := hyphenate.New("en_us.txt", "-", custom)
	if err != nil {
		// handle err
	}

	println(h.Hyphenate("Hello")) // prints "H-ello"
*/
func New(path, hyphen string, custom map[string][]string) (h Hyphenator, err error) {
	path, err = filepath.Abs(path)
	if err != nil {
		return h, err
	}
	f, err := os.Open(path)
	if err != nil {
		return h, err
	}
	defer f.Close()
	lang, err := hyphenation.New(f)
	if err != nil {
		return h, err
	}
	h.hyphen = hyphen
	h.lang = lang
	h.custom = custom
	return h, nil
}

type subWord struct {
	word string
	sep  string
}

func subWords(s string) []subWord {
	var sw []subWord
	start := 0
	for pos, r := range s {
		if !strings.ContainsRune(wordSep, r) {
			continue
		}
		sep := string(r)
		sw = append(sw, subWord{
			word: s[start:pos],
			sep:  sep,
		})
		start = pos + len(sep)
	}
	sw = append(sw, subWord{word: s[start:]})
	return sw
}

/*
Hyphenate returns a hyphenated version of text, according
to the parameters provided to New.

Hyphenate deviates from the hyphenation patterns provided
to New in the following cases:

- If a word is 5 runes or less it will never be hyphenated.
- If a word segment is 1 rune it will not be hyphenated.
- Compound words are treated as separate words - e.g. "part-time" is two 4-letter words.
- Custom hyphenation patterns for words will override defaults.

Words are delineated according to the same criteria used
by strings.Fields
*/
func (h Hyphenator) Hyphenate(text string) string {

	ww := []string{}

	/*
		We trim off any whitespace before calling
		strings.Fields so that we can preserve and
		restore it later.
	*/
	text, textStart, textEnd := trimSpace(text)

	for _, s := range strings.Fields(text) {

		sw := ""

		for _, sub := range subWords(s) {

			/*
				Split any grammer off the word so that it's
				not factored into our minimum length rules.
				We save start/end grammar to re-attach later.

				Note s reassigned repeatedly in a loop below
				so we keep a copy of the original word.
			*/
			origWord, start, end := trim(sub.word, ",.;:?!()#")
			s := origWord

			// If there's a custom hyphen mapping for this word.
			if custom, ok := h.replace(origWord); ok {
				sw += start + custom + end + sub.sep
				continue
			}

			/*
				Segment the original word into parts according
				to the breakpoints we're supplied. It's easier
				to slice from the end of the word so we do that.
			*/
			breakpoints := h.lang.Hyphenate(s)
			parts := []string{}

			pos := 0
			for _, bp := range breakpoints {
				parts = append(parts, s[pos:bp])
				pos = bp
			}
			if s[pos:] != "" {
				parts = append(parts, s[pos:])
			}

			word := ""
			seen := 0
			for _, p := range parts {
				word += p
				seen += strLen(p)

				// Don't append a hyphen if there's already one.
				if endsWithHyphen(p) {
					continue
				}

				/*
					If word part begins with hyphen reset count
					on this iteration only. This prevents singular
					characters after the hyphen from being their
					own parts.
				*/
				partLen := seen
				if startsWithHyphen(p) {
					partLen = strLen(p) - 1
				}

				if addHyphen(partLen, origWord) {
					word += h.hyphen
				}
			}

			sw += start + word + end + sub.sep
		}

		ww = append(ww, sw)
	}

	return textStart + strings.Join(ww, " ") + textEnd
}

func (h Hyphenator) replace(word string) (replaced string, ok bool) {
	ss, ok := h.custom[strings.ToLower(word)]
	if !ok {
		return replaced, ok
	}
	pos := 0
	var parts []string
	for _, s := range ss {
		parts = append(parts, word[pos:pos+len(s)])
		pos += len(s)
	}
	return strings.Join(parts, h.hyphen), ok
}

func startsWithHyphen(s string) bool {
	for _, c := range hyphens {
		if strings.HasPrefix(s, c) {
			return true
		}
	}
	return false
}
func endsWithHyphen(s string) bool {
	for _, c := range hyphens {
		if strings.HasSuffix(s, c) {
			return true
		}
	}
	return false
}

func addHyphen(pLen int, full string) bool {
	fLen := strLen(full)
	if fLen <= 5 {
		return false
	}
	if pLen < 2 {
		return false
	}
	if fLen-pLen < 2 {
		return false
	}
	return true
}

func strLen(s string) int {
	return len([]rune(s))
}

func trimSpace(s string) (new, start, end string) {

	orig := s

	s = strings.TrimLeftFunc(s, unicode.IsSpace)
	diff := len(orig) - len(s)
	start = orig[0:diff]

	sLen := len(s)
	s = strings.TrimRightFunc(s, unicode.IsSpace)
	diff = sLen - len(s)
	end = orig[len(orig)-diff : len(orig)]

	return s, start, end
}

func trim(s string, cutset string) (new, start, end string) {

	orig := s

	s = strings.TrimLeft(s, cutset)
	diff := len(orig) - len(s)
	start = orig[0:diff]

	sLen := len(s)
	s = strings.TrimRight(s, cutset)
	diff = sLen - len(s)
	end = orig[len(orig)-diff : len(orig)]

	return s, start, end
}

func in(ss []string, s string) bool {
	for i := range ss {
		if ss[i] == s {
			return true
		}
	}
	return false
}

type fields struct {
	prefixed bool
	seps     []string
	words    []string
}

/*
newFields splits contiguous spaces (as defined by
unicode.IsSpace) into their own slice separate from
a words slice.
*/
func newFields(s string) fields {
	f := fields{}
	if s == "" {
		return f
	}
	f.prefixed = unicode.IsSpace([]rune(s)[0])
	inField := !f.prefixed
	start := 0
	for pos, r := range s {
		if unicode.IsSpace(r) {
			if inField {
				f.words = append(f.words, s[start:pos])
				inField = false
				start = pos
			}
		} else {
			if !inField {
				f.seps = append(f.seps, s[start:pos])
				inField = true
				start = pos
			}
		}
	}
	if inField {
		f.words = append(f.words, s[start:])
	} else {
		f.seps = append(f.seps, s[start:])
	}
	return f
}

func (f fields) replaceWords(newWords []string) (string, error) {
	if len(newWords) != len(f.words) {
		return "", fmt.Errorf(
			"hyphenate: mismatch in number of words supplied to fields.words:"+
				"\n\tgot %d,"+
				"\n\twanted %d",
			len(newWords), len(f.words))
	}
	n := len(f.seps) + len(newWords)
	combined := make([]string, n, n)
	for i := range combined {
		if f.prefixed {
			if i%2 == 0 {
				combined[i] = f.seps[i/2]
			} else {
				combined[i] = newWords[i/2]
			}
		} else {
			if i%2 == 0 {
				combined[i] = newWords[i/2]
			} else {
				combined[i] = f.seps[i/2]
			}
		}
	}
	return strings.Join(combined, ""), nil
}
