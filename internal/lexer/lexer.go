// Package lexer provides a generic Lexer framework that can be extended for any grammar.
// See Rob Pike's lexer design [talk](https://www.youtube.com/watch?v=HxaD_trXwRE).
//
// You can define your token types by using the `lexer.TokenType` type (`int`) via
//
//     const (
//             StringToken lexer.TokenType = iota
//             IntegerToken
//             // etc...
//     )
//
// And then you define your own state functions (`lexer.StateFunc`) to handle
// analyzing the string.
//
//     func stringState(l *lexer.Lex) lexer.StateFunc {
//             l.Next() // eat starting "
//             l.Ignore() // drop current value
//             for l.Peek() != '"' {
//                     l.Next()
//             }
//             l.Emit(StringToken)
//
//             return nextStateFunction
//     }
//
// Then start your lexer and hook up your parser (calling lex.NextToken).
//
// 		lex := lexer.New("string ToScan = `here`;", stringState)
//		lex.Start()
// 		...
// 		tok := lex.NextToken()
//
// Credits: this is a modified version of github.com/bbuck/go-lexer (MIT license).
package lexer

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

/* ----------------------------------------------------------------------------------- */
/* Parser API */

// TokenType defines the categories of token.
//
// Token types are categories of token that should be treated similarly by the
// parser.  For example, the 'operator' type may include the values '+', '-',
// '*' and '/' in a calculator grammar.  Some tokens types are open ended. For
// example the number token type may include values that match any signed or
// unsigned, integer or float, base 8, 10, or 16, with or without exponent.
// The lexer should scan the value, letting the parser do the validation.
type TokenType int

// ErrorTok is used to send back errors to the parser.
const ErrorTok = TokenType(-1)

// Token represents a lexeme detected by the lexer.  It has a type and a value.
// The value is always a slice of the input string.
type Token struct {
	Type  TokenType
	Value string
}

// TokenTypeStringer is used to print token types if not nil.
var TokenTypeStringer func(TokenType) string

func (t Token) String() string {
	var typ string
	if TokenTypeStringer == nil {
		typ = fmt.Sprintf("%v", t.Type)
	} else {
		typ = TokenTypeStringer(t.Type)
	}
	return fmt.Sprintf("{%s, \"%s\"}", typ, t.Value)
}

const (
	// EOFRune is pseudo-rune signaling the end of the input string
	EOFRune rune = 0
)

// Lex encapsulates the lexer state.
type Lex struct {
	source          string
	startState      StateFunc
	start, position int
	atEOF           bool
	tokens          chan Token
}

// New returns a lexer ready to parse the given string.
func New(src string, startState StateFunc) *Lex {
	return &Lex{
		source:     src,
		startState: startState,
		start:      0,
		position:   0,
	}
}

// Start begins executing the Lexer in a goroutine.
func (l *Lex) Start() *Lex {
	l.tokens = make(chan Token, 2)
	go func() {
		defer close(l.tokens)
		state := l.startState
		for state != nil {
			state = state(l)
		}
	}()
	return l
}

// NextToken gets and returns the next token from the lexer or nil if finished.
func (l *Lex) NextToken() *Token {
	if tok, ok := <-l.tokens; ok {
		return &tok
	}
	return nil
}

/* ----------------------------------------------------------------------------------- */
/* Scanner API */

// StateFunc is the type to be implemented in the scanner when scanning longer
// constructions.
//
// What is a lexer state?  State represents where we are and what we expect.
// Lexers have little state compared to parsers, which do the heavy lifting.
//
// State is needed when the meaning of a string might be ambiguous (for example
// the string ']]>' in a CDATA section vs outside one). Typically states
// are needed for numeric constants, quotes, comments or other mini-grammars.
// If these can be nested, then a state stack may be needed in the scanner.
type StateFunc func(*Lex) StateFunc

// Current returns the value being being analyzed at this moment.
func (l *Lex) Current() string {
	return l.source[l.start:l.position]
}

// Emit will receive a token TYPE and push a new token with the current analyzed
// value into the tokens channel.
func (l *Lex) Emit(t TokenType) {
	tok := Token{
		Type:  t,
		Value: l.Current(),
	}
	l.Ignore()
	l.tokens <- tok
}

// Errorf is a state function that formats an error message and returns it as
// an ErrorTok token.  The scan terminates.
func (l *Lex) Errorf(format string, args ...interface{}) StateFunc {
	tok := Token{
		ErrorTok,
		fmt.Sprintf(format, args...),
	}
	l.tokens <- tok
	return nil
}

// Ignore skips over the current string to ignore the section of the source
// being analyzed.
func (l *Lex) Ignore() {
	l.start = l.position
}

// Peek performs a Next operation immediately followed by a Backup returning the
// peeked rune.
func (l *Lex) Peek() rune {
	r := l.Next()
	l.Backup()

	return r
}

// Next pulls the next rune from the Lexer and returns it, moving the position
// forward in the source.
func (l *Lex) Next() rune {
	var (
		r rune
		s int
	)
	str := l.source[l.position:]
	if len(str) == 0 {
		if l.atEOF {
			panic("Next attempted to move past end of source.")
		}
		l.atEOF = true
		r, s = EOFRune, 0
	} else {
		r, s = utf8.DecodeRuneInString(str)
	}
	l.position += s
	return r
}

// Backup will undo Next by changing the lexer's current position. This can
// occur more than once per call to Next but you can never backup past the
// last point a token was emitted.
func (l *Lex) Backup() {
	str := l.source[l.start:l.position]
	if l.atEOF {
		l.atEOF = false
	} else if len(str) > 0 {
		_, s := utf8.DecodeLastRuneInString(str)
		l.position -= s
	}
}

// Accept consumes a rune that appears in CHARS and returns true.
func (l *Lex) Accept(chars string) bool {
	if strings.ContainsRune(chars, l.Next()) {
		return true
	}
	l.Backup() // last next wasn't a match
	return false
}

// AcceptRun consumes any sequence of runes that appear in CHARS.
func (l *Lex) AcceptRun(chars string) {
	for strings.ContainsRune(chars, l.Next()) {
	}
	l.Backup()
}

// AcceptTo accepts runes not found in CHARS or the line ends.
func (l *Lex) AcceptTo(chars string) {
	chars = chars + "\n\x00"
	for !strings.ContainsRune(chars, l.Next()) {
	}
	l.Backup()
}

// LookingAt returns true if the current position starts with PREFIX.
func (l *Lex) LookingAt(prefix string) bool {
	if strings.HasPrefix(l.source[l.position:], prefix) {
		return true
	}
	return false
}
