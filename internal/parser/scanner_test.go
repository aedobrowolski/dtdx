package parser

import (
	"fmt"
	"testing"

	"github.com/adobrowolski/dtdx/internal/lexer"
)

func ExampleTokenTypeString() {
	for key := range lexer.TokenName {
		fmt.Printf("Key: %d Value: %s\n", key, key)
	}
	// Unordered output:
	// Key: -1 Value: ErrorTok
	// Key: 1 Value: indentTok
	// Key: 2 Value: equalsTok
	// Key: 3 Value: openTok
	// Key: 4 Value: closeTok
	// Key: 5 Value: separatorTok
	// Key: 6 Value: multiplicityTok
	// Key: 7 Value: identTok
	// Key: 8 Value: quoteTok
	// Key: 9 Value: referenceTok
	// Key: 10 Value: directiveTok
	// Key: 11 Value: commentTok
	// Key: 12 Value: eofTok
}

const test1 = `# The first top level definition.
paragraph
    # A definition with two references nested inside paragraph.
    title?
    line...+

# A second top level definition.
line
	(PCDATA, bold)*`

func ExampleTest1() {
	l := lexer.New(test1, NewlineState).Start()
	for tok := l.NextToken(); tok != nil; tok = l.NextToken() {
		fmt.Printf("%s\n", tok)
	}
	// Output:
	// {indentTok, ""}
	// {commentTok, "# The first top level definition."}
	// {indentTok, ""}
	// {identTok, "paragraph"}
	// {indentTok, "    "}
	// {commentTok, "# A definition with two references nested inside paragraph."}
	// {indentTok, "    "}
	// {identTok, "title"}
	// {multiplicityTok, "?"}
	// {indentTok, "    "}
	// {identTok, "line"}
	// {referenceTok, "..."}
	// {multiplicityTok, "+"}
	// {indentTok, ""}
	// {commentTok, "# A second top level definition."}
	// {indentTok, ""}
	// {identTok, "line"}
	// {indentTok, "	"}
	// {openTok, "("}
	// {identTok, "PCDATA"}
	// {separatorTok, ","}
	// {identTok, "bold"}
	// {closeTok, ")"}
	// {multiplicityTok, "*"}
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
	// {indentTok, ""}
	// {commentTok, "# Define paragraph element with three attributes"}
	// {indentTok, ""}
	// {identTok, "paragraph"}
	// {identTok, "id"}
	// {equalsTok, "="}
	// {directiveTok, "#ID"}
	// {identTok, "name"}
	// {equalsTok, "="}
	// {identTok, "justify"}
	// {equalsTok, "="}
	// {openTok, "("}
	// {identTok, "left"}
	// {separatorTok, "|"}
	// {identTok, "right"}
	// {separatorTok, "|"}
	// {identTok, "center"}
	// {closeTok, ")"}
	// {eofTok, ""}
}

func ExampleAttrScanner() {
	l := lexer.New("attr1=\"one\" attr2='2' attr3=", OuterState).Start()
	for tok := l.NextToken(); tok != nil; tok = l.NextToken() {
		fmt.Println(tok)
	}
	// Output:
	// {identTok, "attr1"}
	// {equalsTok, "="}
	// {quoteTok, "one"}
	// {identTok, "attr2"}
	// {equalsTok, "="}
	// {quoteTok, "2"}
	// {identTok, "attr3"}
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
		{identTok, "attr1"},
		{equalsTok, "="},
		{quoteTok, "one"},
		{identTok, "attr2"},
		{equalsTok, "="},
		{quoteTok, "2"},
		{identTok, "attr3"},
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
		{identTok, "attr1"},
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
