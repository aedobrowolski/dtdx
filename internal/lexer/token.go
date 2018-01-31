package lexer

import (
	"fmt"
	"strconv"
)

// TokenType defines the categories of token.
//
// Token types are categories of token that should be treated similarly by the
// parser.  For example, the 'operator' type may include the values '+', '-',
// '*' and '/' in a calculator grammar.  Some tokens types are open ended. For
// example the number token type may include values that match any signed or
// unsigned, integer or float, base 8, 10, or 16, with or without exponent.
// The lexer should scan the value, letting the parser do the validation.
type TokenType int

func (tt TokenType) String() string {
	if tokString, ok := TokenName[tt]; ok {
		return tokString
	}
	return "tok_" + strconv.Itoa(int(tt))
}

// TokenName maps token values to strings.  Add const values defined in other packages.
var TokenName = map[TokenType]string{}

// Token represents a lexeme detected by the lexer.  It has a type and a value.
// The value is always a slice of the input string.
type Token struct {
	Type  TokenType
	Value string
}

func (t Token) String() string {
	return fmt.Sprintf("{%s, \"%s\"}", t.Type, t.Value)
}
