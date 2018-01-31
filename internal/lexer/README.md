This package provides a Lexer that functions similarly to Rob Pike's discussion
about lexer design in this [talk](https://www.youtube.com/watch?v=HxaD_trXwRE).

You can define your token types by using the `lexer.TokenType` type (`int`) via

```go
const (
        StringToken lexer.TokenType = iota
        IntegerToken
        etc...
)
```

And then you define your own state functions (`lexer.StateFunc`) to handle
analyzing the string.

```go
func StringState(l *lexer.Lex) lexer.StateFunc {
        l.Next() // eat starting "
        l.Ignore() // drop current value
        while l.Peek() != '"' {
                l.Next()
        }
        l.Emit(StringToken)

        return SomeStateFunction
}
```

It should be easy to make this Lexer consumable by a parser generated by go yacc doing something alone the lines of the following:

```go
type MyLexer struct {
        lexer.Lex
}

func (m *MyLexer) Lex(lval *yySymType) int {
        tok := m.NextToken()
        if done {
                return EOFToken
        } else {
                lval.val = tok.Value
                return tok.Type
        }
}
```

# License

MIT