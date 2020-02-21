package hyphenate

import (
	"testing"
)

func TestNewFields(t *testing.T) {

	wants := []struct {
		s      string
		seps   []string
		words  []string
		prefix bool
	}{
		{
			s:      "hello there friendo",
			seps:   []string{" ", " "},
			words:  []string{"hello", "there", "friendo"},
			prefix: false,
		},
		{
			s:      "   hello  there friendo ",
			seps:   []string{"   ", "  ", " ", " "},
			words:  []string{"hello", "there", "friendo"},
			prefix: true,
		},
		{
			s:      "   hello \t there \nfriendo ",
			seps:   []string{"   ", " \t ", " \n", " "},
			words:  []string{"hello", "there", "friendo"},
			prefix: true,
		},
		{
			s:      " ",
			seps:   []string{" "},
			words:  nil,
			prefix: true,
		},
		{
			s:      "a",
			seps:   nil,
			words:  []string{"a"},
			prefix: false,
		},
		{
			s:      "",
			seps:   nil,
			words:  nil,
			prefix: false,
		},
	}

	for _, w := range wants {
		got := newFields(w.s)
		showErr := false
		showErr = got.prefixed != w.prefix
		showErr = !sameSlice(got.seps, w.seps)
		showErr = !sameSlice(got.words, w.words)
		if showErr {
			t.Errorf(`
newFields(%q)
    return
        seps:   %#v,
        words:  %#v,
        prefix: %v
    wanted
        seps:   %#v,
        words:  %#v,
        prefix: %v`,
				w.s,
				got.seps, got.words, got.prefixed,
				w.seps, w.words, w.prefix,
			)
		}
	}
}

func sameSlice(ss1, ss2 []string) bool {
	if len(ss1) != len(ss2) {
		return false
	}
	for i := range ss1 {
		if ss1[i] != ss2[i] {
			return false
		}
	}
	return true
}

func TestReplaceWords(t *testing.T) {
	wants := []struct {
		s     string
		sNew  string
		words []string
		err   bool
	}{
		{
			s:     "hello there my friendy friend",
			sNew:  "howdy there my friendly friends",
			words: []string{"howdy", "there", "my", "friendly", "friends"},
			err:   false,
		},
		{
			s:     "hello  there my friendy  friend",
			sNew:  "howdy  there my friendly  friends",
			words: []string{"howdy", "there", "my", "friendly", "friends"},
			err:   false,
		},
		{
			s:     "  hello  there my friend  ",
			sNew:  "  howdy  there my friends  ",
			words: []string{"howdy", "there", "my", "friends"},
			err:   false,
		},
	}
	for _, w := range wants {
		f := newFields(w.s)
		errStr := "nil"
		if w.err {
			errStr = "error"
		}
		if got, err := f.replaceWords(w.words); got != w.sNew || w.err && err == nil {
			t.Errorf(`
fields.replaceWords(%#v)
    return
        %q,
        %v,
    wanted
        %q,
        %s`,
				w.words, got, err, w.sNew, errStr)
		}
	}
}
