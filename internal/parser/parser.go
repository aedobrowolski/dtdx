// Package parser provides a dtdx parser.  The output is an AST (abstract) of the
// dtdx document which can then be serialized as an XML DTD.
//
// The DTDX (DTD teXt) grammar mocks up an instance using plain text strings in
// order to define a document model that can then be used to generate an XML DTD.
//
// Definitions of elements appear as children of one of their parent elements.
// An element is never defined two times. Instead it is referenced using the
// name followed by a "..." suffix.  The definition does not need to come before
// the reference.
//
// Here is an example dtdx document that defines paragraph structures.
//
// 		# The first top level definition.
//		paragraph
//			# A definition with two references nested inside paragraph.
//			title?
//			line...+
//
// 		# A second top level definition.
// 		line
//			(PCDATA, bold)*
//
// The first element to be defined, paragraph, is the root of the DTD. The content
// models of title and bold default to (#PCDATA).
//
// This example document is equivalent to the DTD:
//
// 		<!ELEMENT paragraph 	(title?, line+)>
// 		<!ELEMENT title 		(#PCDATA)>
// 		<!ELEMENT line 			(#PCDATA,bold)* >
// 		<!ELEMENT bold 			(#PCDATA)>
//
// Attributes are defined after the element name as a list of name value pairs.
// The attribute name must be followed by '=' with no intervening space, followed
// by an optional type. If left off the type is derived from the name. This usually
// defaults to CDATA. If the name is id, idref or idrefs then the type is the upper
// case value of the name. The type can also be a list of NMTOKEN values separated
// by the vertical bar character '|' to create an enumerated attribute type.
//
// Here is an example of an element definition with three attributes.
//
// 		# Define paragraph element with three attributes
//		paragraph id= name= justify=(left|right|center)
//
// This example document is equivalent to the DTD:
//
// 		<!ELEMENT paragraph 	(#PCDATA)>
//		<!ATTLIST paragraph
//				id 		ID 					#IMPLIED
//				name 	CDATA 				#IMPLIED
// 				justify (left|right|center) #IMPLIED
//				>
package parser

import (
	"bytes"
	"fmt"
	"io"

	"github.com/adobrowolski/dtdx/internal/lexer"
)

/* --------------------------------------------------------------

dtdx            := (comment | elementDef)*
comment         := '#' text '\n'
element         := elementDef | elementRef
elementDef      := name attrs content
elementRef      := name Ellipsis
name            := identifier
attrs           := name '=' type?
type            := directive | enumeration
directive       := '#' identifier
enumeration     := '(' values ')'
values          := Value ( '|' values )
content         := contentStart contentBody
contentBody     := ( parenContent | nakedContent )
parenContent    := '(' contentBody ')'
nakedContent    := elementList
elementList     := elementChild (elementSep elementList)?
elementChild    := comment | element modifier?
modifier        := '*' | '+' | '?'
contentStart    := greaterIndent | '=>'
elementSep      := sameIndent | ','
greaterIndent   := '\n' indentTok             [len(tok.value)>parent.indent]
sameIndent      := '\n' indentTok             [len(tok.value)==parent.indent]

------------------------------------------------------------------ */

// elementMap maps element names to their definitions.
//
// While a new element is being defined the value is the placeholder.
type elementMap map[string]Element

var elements = elementMap{}

// Parser represents a parser.
type Parser struct {
	s   *lexer.Lex
	buf struct {
		tok lexer.Token // last read token
		lit string      // last read literal
		n   int         // buffer size (max=1)
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	input := buf.String()
	return &Parser{s: lexer.New(input, nil)}
}

// Parse parses a DTDX document
func (p *Parser) Parse() (*Element, error) {
	elem := Element{}

	// First token should be a "SELECT" keyword.
	if tok, lit := p.scan(); tok != identifierTok {
		return nil, fmt.Errorf("found %q, expected element identifier", lit)
	}

	return &elem, nil
}

func (p *Parser) scan() (lexer.TokenType, string) {
	token := p.s.NextToken()
	return token.Type, token.Value
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }
