package hyphenate

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/speedata/hyphenation"
)

const wordSep = "-–—/"

type Hyphenator struct {
	hyphen string
	lang   *hyphenation.Lang
	custom map[string][]string
}

func New(path, hyphen string, custom map[string][]string) (h Hyphenator, err error) {
	path, err = filepath.Abs(path)
	if err != nil {
		return h, err
	}
	f, err := os.Open(path)
	if err != nil {
		return h, err
	}
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

func (h Hyphenator) Hyphenate(text string) string {

	ww := []string{}

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

	return strings.Join(ww, " ")
}

func (h Hyphenator) replace(word string) (replaced string, ok bool) {

	// ss, ok := h.custom[strings.ToLower(word)]
	// if !ok {
	// 	return replaced, ok
	// }
	// return strings.Join(ss, h.hyphen), ok

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

	// mapping, ok := h.custom[strings.ToLower(word)]
	// if !ok {
	// 	return replaced, false
	// }
	// mapping = append(mapping, len(word))

	// var segs [][]rune
	// rr := []rune(word)
	// pos := 0
	// for _, m := range mapping {
	// 	segs = append(segs, rr[pos:m])
	// 	pos = m
	// }

	// for _, seg := range segs {
	// 	replaced += string(seg) + h.hyphen
	// }
	// replaced = strings.TrimSuffix(replaced, h.hyphen)

	// return replaced, true
}

func startsWithHyphen(s string) bool {
	for _, c := range []string{"-", "–", "—"} {
		if strings.HasPrefix(s, c) {
			return true
		}
	}
	return false
}
func endsWithHyphen(s string) bool {
	for _, c := range []string{"-", "–", "—"} {
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
