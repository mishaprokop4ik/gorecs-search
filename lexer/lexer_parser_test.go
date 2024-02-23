package lexer_test

import (
	"github.com/mishaprokop4ik/gorecs-search/lexer"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLexerParse(t *testing.T) {
	content := "hello, it is Misha. Why are you here...?"
	l := lexer.NewLexer(content)

	expected := []string{
		"hello",
		"it",
		"is",
		"misha",
		"why",
		"are",
		"you",
		"here",
	}

	terms := []string{}
	for term := l.Next(); term != ""; term = l.Next() {
		terms = append(terms, term)
	}

	t.Log("comparing terms")
	assert.Equal(t, expected, terms)
}
