package lexer_test

import (
	"testing"

	"github.com/adobrowolski/dtdx/internal/lexer"
)

const (
	NumberToken lexer.TokenType = iota
	OpToken
	IdentToken
)

func NumberState(l *lexer.Lex) lexer.StateFunc {
	l.AcceptRun("0123456789")
	l.Emit(NumberToken)
	if l.Peek() == '.' {
		l.Next()
		l.Emit(OpToken)
		return IdentState
	}

	return nil
}

func IdentState(l *lexer.Lex) lexer.StateFunc {
	r := l.Next()
	for (r >= 'a' && r <= 'z') || r == '_' {
		r = l.Next()
	}
	l.Backup()
	l.Emit(IdentToken)

	return WhitespaceState
}

func WhitespaceState(l *lexer.Lex) lexer.StateFunc {
	r := l.Next()
	if r == lexer.EOFRune {
		return nil
	}

	if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
		return l.Errorf("unexpected token %q", r)
	}

	l.AcceptRun(" \t\n\r")
	l.Ignore()

	return NumberState
}

func Test_LexerMovingThroughString(t *testing.T) {
	l := lexer.New("123", nil)
	run := []struct {
		s string
		r rune
	}{
		{"1", '1'},
		{"12", '2'},
		{"123", '3'},
		{"123", lexer.EOFRune},
	}

	for _, test := range run {
		r := l.Next()
		if r != test.r {
			t.Errorf("Expected %q but got %q", test.r, r)
			return
		}

		if l.Current() != test.s {
			t.Errorf("Expected %q but got %q", test.s, l.Current())
			return
		}
	}
}

func Test_LexingNumbers(t *testing.T) {
	l := lexer.New("123", NumberState)
	l.Start()
	tok := l.NextToken()
	if tok.Type != NumberToken {
		t.Errorf("Expected a %v but got %v", NumberToken, tok.Type)
		return
	}

	if tok.Value != "123" {
		t.Errorf("Expected %q but got %q", "123", tok.Value)
		return
	}

	tok = l.NextToken()
	if tok != nil {
		t.Errorf("Expected a nil token, but got %v", *tok)
		return
	}
}

func Test_LexerRewind(t *testing.T) {
	l := lexer.New("1", nil)
	r := l.Next()
	if r != '1' {
		t.Errorf("Expected %q but got %q", '1', r)
		return
	}

	if l.Current() != "1" {
		t.Errorf("Expected %q but got %q", "1", l.Current())
		return
	}

	l.Backup()
	if l.Current() != "" {
		t.Errorf("Expected empty string, but got %q", l.Current())
		return
	}
}

func Test_MultipleTokens(t *testing.T) {
	cases := []struct {
		tokType lexer.TokenType
		val     string
	}{
		{NumberToken, "123"},
		{OpToken, "."},
		{IdentToken, "hello"},
		{NumberToken, "675"},
		{OpToken, "."},
		{IdentToken, "world"},
	}

	l := lexer.New("123.hello  675.world", NumberState)
	l.Start()

	for _, c := range cases {
		tok := l.NextToken()
		if c.tokType != tok.Type {
			t.Errorf("Expected token type %v but got %v", c.tokType, tok.Type)
			return
		}

		if c.val != tok.Value {
			t.Errorf("Expected %q but got %q", c.val, tok.Value)
			return
		}
	}

	tok := l.NextToken()
	if tok != nil {
		t.Errorf("Did not expect a token, but got %v", *tok)
		return
	}
}

func Test_LexerError(t *testing.T) {
	l := lexer.New("1", WhitespaceState)
	l.Start()

	tok := lexer.Token{lexer.ErrorTok, "unexpected token '1'"}
	if got, expect := *l.NextToken(), tok; got != expect {
		t.Errorf("Expected %v but got %v", expect, got)
	}
}
