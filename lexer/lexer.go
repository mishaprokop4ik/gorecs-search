package lexer

import (
	"strings"
	"unicode"
)

type Lexer struct {
	Terms []rune
}

func NewLexer(content ...string) *Lexer {
	t := make([]rune, 0)
	for _, word := range content {
		t = append(t, []rune(word)...)
	}
	return &Lexer{Terms: t}
}

func (l *Lexer) chop(n int) []rune {
	term := l.Terms[:n]
	l.Terms = l.Terms[n:]

	return term
}

func (l *Lexer) chopWhile(predicate func(component rune) bool) []rune {
	n := 0
	for n < len(l.Terms) && predicate(l.Terms[n]) {
		n++
	}

	return l.chop(n)
}

func (l *Lexer) trimLeft() {
	for len(l.Terms) != 0 && (unicode.IsSpace(l.Terms[0]) || unicode.IsPunct(l.Terms[0])) {
		l.Terms = l.Terms[1:]
	}
}

// All returns all tokens
func (l *Lexer) All() []string {
	tokens := []string{}
	for t := l.Next(); t != ""; t = l.Next() {
		tokens = append(tokens, t)
	}

	return tokens
}

func (l *Lexer) Next() string {
	l.trimLeft()

	if len(l.Terms) == 0 {
		return ""
	}

	if unicode.IsNumber(l.Terms[0]) {
		return strings.ToLower(string(l.chopWhile(unicode.IsNumber)))
	}

	if unicode.IsLetter(l.Terms[0]) {
		token := l.chopWhile(func(c rune) bool {
			return unicode.IsNumber(c) || unicode.IsLetter(c)
		})

		return strings.ToLower(string(token))
	}

	return string(l.chop(1))
}
