package parser

import (
	"unicode"

	"github.com/adobrowolski/dtdx/internal/lexer"
)

/* ----------------------------------------------------------------------------- */
// Define the scanner for the lexical tokens of this grammar.
// Start by defining the token types, then the lexical states.

// Lexical types used in the grammar
const (
	illegalTok lexer.TokenType = iota

	indentTok       // whitespace at start of line
	equalsTok       // =
	openTok         // (
	closeTok        // )
	separatorTok    // , or | or &
	multiplicityTok // * or + or ?
	identTok        // value
	quoteTok        // 'value' or "value"
	referenceTok    // ...
	directiveTok    // #VALUE
	commentTok      // # value
	eofTok          // signals end of input
)

func init() {
	lexer.TokenName[indentTok] = "indentTok"
	lexer.TokenName[equalsTok] = "equalsTok"
	lexer.TokenName[openTok] = "openTok"
	lexer.TokenName[closeTok] = "closeTok"
	lexer.TokenName[separatorTok] = "separatorTok"
	lexer.TokenName[multiplicityTok] = "multiplicityTok"
	lexer.TokenName[identTok] = "identTok"
	lexer.TokenName[quoteTok] = "quoteTok"
	lexer.TokenName[referenceTok] = "referenceTok"
	lexer.TokenName[directiveTok] = "directiveTok"
	lexer.TokenName[commentTok] = "commentTok"
	lexer.TokenName[eofTok] = "eofTok"
}

/* -----------------------------------------------------------------------------

The initial state is NewLineState. Like python indents matter. Whitespace is
scanned up until the first non-blank token. An 'indent' token will be emitted.
Then the state will change to OuterState.

In the OuterState whitespace is ignored. The state goes to
- NewLineState 		after a newline
- CommentState		after a comment start not followed by a letter
- SingleQuoteState 	after a single quote
- DoubleQuoteState 	after a double quote
- IdentifierState 	after a hash or alphanumeric
All the single character tokens will be emitted while in this state.

SingleQuoteState and DoubleQuoteState lex the body of a quote and emit it,
either as the sQuote or dQuote token, going back to OuterState when done.

IdentifierState scans tokens that start with an alphanumeric value or a #
immediately followed by an alphanumeric value. It can emit 'ident' and
'reference' tokens.

CommentState eats an initial # and parses the remaining characters up to
the newline, emiting a 'comment'.  It then goes to the OuterState.

*/

// OuterState handles all single letter tokens and delegates to other states.
func OuterState(l *lexer.Lex) lexer.StateFunc {
	for {
		switch r := l.Next(); r {
		case ' ', '\t':
			l.Ignore()
		case '\n':
			return NewlineState
		case '=':
			l.Emit(equalsTok)
		case '(':
			l.Emit(openTok)
		case ')':
			l.Emit(closeTok)
		case ',', '|', '&':
			l.Emit(separatorTok)
		case '*', '+', '?':
			l.Emit(multiplicityTok)
		case '"':
			return DoubleQuoteState
		case '\'':
			return SingleQuoteState
		case '.':
			return ReferenceState
		case '#':
			r = l.Peek()
			if 'A' <= r && r <= 'Z' {
				return DirectiveState
			}
			return CommentState
		case lexer.EOFRune:
			l.Emit(eofTok)
			return nil
		default:
			if unicode.IsLetter(r) || r == '_' || r == ':' {
				return IdentifierState
			}
			return l.Errorf("Unexpected unicode character (%#U) in outer context.", r)
		}
	}
}

// NewlineState handles a \n and emits an 'indent'.
func NewlineState(l *lexer.Lex) lexer.StateFunc {
	l.Ignore() // drop the newline (if any)
	l.AcceptRun("\t ")
	if l.LookingAt("\n") { // empty line?
		l.Next() // move past the newline and try again
		return NewlineState
	}
	l.Emit(indentTok)

	return OuterState
}

// DoubleQuoteState handles values of the form "..."
func DoubleQuoteState(l *lexer.Lex) lexer.StateFunc {
	return quoteHelper(l, "\"")
}

// SingleQuoteState handles values of the form '...'
func SingleQuoteState(l *lexer.Lex) lexer.StateFunc {
	return quoteHelper(l, "'")
}

func quoteHelper(l *lexer.Lex, quote string) lexer.StateFunc {
	l.Ignore() // drop the initial quote
	l.AcceptTo(quote)
	if l.LookingAt(quote) {
		l.Emit(quoteTok)
		l.Next() // skip the final quote
		return OuterState
	}

	return l.Errorf("Runaway quote: %s", l.Current())
}

const uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// DirectiveState handles #UPPERCASE directives
func DirectiveState(l *lexer.Lex) lexer.StateFunc {
	l.AcceptRun(uppercase)
	l.Emit(directiveTok)
	return OuterState
}

// CommentState handles #... comments, but not directives
func CommentState(l *lexer.Lex) lexer.StateFunc {
	l.AcceptTo("") // newline or eof
	l.Emit(commentTok)
	return OuterState
}

// IdentifierState handles identifiers (NMTOKEN)
func IdentifierState(l *lexer.Lex) lexer.StateFunc {
	for {
		r := l.Next()
		if !isAlphaNumeric(r) {
			l.Backup()
			break
		}
	}
	l.Emit(identTok)
	return OuterState
}

func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// ReferenceState handles a reference ellipsis (...)
func ReferenceState(l *lexer.Lex) lexer.StateFunc {
	// l.Backup()
	l.AcceptRun(".")
	if l.Current() == "..." {
		l.Emit(referenceTok)
		return OuterState
	}

	return l.Errorf("Malformed reference ellipsis: %s", l.Current())
}
