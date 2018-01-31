package parser

import (
	"fmt"
	"testing"

	"github.com/adobrowolski/dtdx/internal/lexer"
)

func ExampleTokenTypeString() {
	for key := range lexer.TokenName {
		fmt.Printf("Key: %2d Value: %s\n", key, key)
	}
	// Unordered output:
	// Key: -1 Value: ErrorTok
	// Key:  1 Value: indentTok
	// Key:  2 Value: dedentTok
	// Key:  3 Value: equalsTok
	// Key:  4 Value: openTok
	// Key:  5 Value: closeTok
	// Key:  6 Value: separatorTok
	// Key:  7 Value: multiplicityTok
	// Key:  8 Value: identifierTok
	// Key:  9 Value: quoteTok
	// Key: 10 Value: referenceTok
	// Key: 11 Value: directiveTok
	// Key: 12 Value: commentTok
	// Key: 13 Value: eofTok
}

const test1 = `# The first top level definition.
paragraph
    # A definition with two references nested inside paragraph.
    title?
    line...+

# A second top level definition.
line
	(PCDATA, bold)*
		# test double dedent`

func ExampleTest1() {
	l := lexer.New(test1, NewlineState).Start()
	for tok := l.NextToken(); tok != nil; tok = l.NextToken() {
		fmt.Printf("%s\n", tok)
	}
	// Output:
	// {commentTok, "# The first top level definition."}
	// {identifierTok, "paragraph"}
	// {indentTok, "    "}
	// {commentTok, "# A definition with two references nested inside paragraph."}
	// {identifierTok, "title"}
	// {multiplicityTok, "?"}
	// {identifierTok, "line"}
	// {referenceTok, "..."}
	// {multiplicityTok, "+"}
	// {dedentTok, ""}
	// {commentTok, "# A second top level definition."}
	// {identifierTok, "line"}
	// {indentTok, "	"}
	// {openTok, "("}
	// {identifierTok, "PCDATA"}
	// {separatorTok, ","}
	// {identifierTok, "bold"}
	// {closeTok, ")"}
	// {multiplicityTok, "*"}
	// {indentTok, "		"}
	// {commentTok, "# test double dedent"}
	// {dedentTok, ""}
	// {dedentTok, ""}
	// {eofTok, ""}
}

const test2 = `
# Define paragraph element with three attributes
paragraph id=#ID name= justify=(left|right|center)`

func ExampleTest2() {
	l := lexer.New(test2, NewlineState).Start()
	for tok := l.NextToken(); tok != nil; tok = l.NextToken() {
		fmt.Println(tok)
	}
	// Output:
	// {commentTok, "# Define paragraph element with three attributes"}
	// {identifierTok, "paragraph"}
	// {identifierTok, "id"}
	// {equalsTok, "="}
	// {directiveTok, "#ID"}
	// {identifierTok, "name"}
	// {equalsTok, "="}
	// {identifierTok, "justify"}
	// {equalsTok, "="}
	// {openTok, "("}
	// {identifierTok, "left"}
	// {separatorTok, "|"}
	// {identifierTok, "right"}
	// {separatorTok, "|"}
	// {identifierTok, "center"}
	// {closeTok, ")"}
	// {eofTok, ""}
}

func ExampleAttrScanner() {
	l := lexer.New("attr1=\"one\" attr2='2' attr3=", OuterState).Start()
	for tok := l.NextToken(); tok != nil; tok = l.NextToken() {
		fmt.Println(tok)
	}
	// Output:
	// {identifierTok, "attr1"}
	// {equalsTok, "="}
	// {quoteTok, "one"}
	// {identifierTok, "attr2"}
	// {equalsTok, "="}
	// {quoteTok, "2"}
	// {identifierTok, "attr3"}
	// {equalsTok, "="}
	// {eofTok, ""}
}

func TestReference(t *testing.T) {
	l := lexer.New("...", OuterState).Start()
	got := *l.NextToken()
	expect := lexer.Token{referenceTok, "..."}
	if got != expect {
		t.Errorf("Expected '%v', got '%v'\n", expect, got)
		t.Fail()
	}
}

func TestNextToken3(t *testing.T) {
	l := lexer.New("attr1=\"one\" attr2='2' attr3=", OuterState).Start()
	testCases := []lexer.Token{
		{identifierTok, "attr1"},
		{equalsTok, "="},
		{quoteTok, "one"},
		{identifierTok, "attr2"},
		{equalsTok, "="},
		{quoteTok, "2"},
		{identifierTok, "attr3"},
		{equalsTok, "="},
		{eofTok, ""},
	}
	for _, tC := range testCases {
		t.Run(tC.Value, func(t *testing.T) {
			if got, expect := *l.NextToken(), tC; got != expect {
				t.Errorf("Expected [%v], but found [%v]", expect, got)
			}
		})
	}
}

func TestRunawayQuote(t *testing.T) {
	l := lexer.New("attr1=\"one attr2='2' attr3=", OuterState).Start()
	testCases := []lexer.Token{
		{identifierTok, "attr1"},
		{equalsTok, "="},
		{lexer.ErrorTok, "Runaway quote: one attr2='2' attr3="},
	}
	for _, tC := range testCases {
		t.Run(tC.Value, func(t *testing.T) {
			if got, expect := l.NextToken(), tC; *got != expect {
				t.Errorf("Expected [%v], but found [%v]", expect, got)
			}
		})
	}
}
