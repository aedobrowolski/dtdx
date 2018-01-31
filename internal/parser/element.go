package parser

import (
	"bytes"
)

// Elements and Attributes are fundamental components of schemas for XML.
// In addition to Element and Attribute, a DTD based schema can also have
// general entities, parameter entities, notations, and entity references.

// Element represents an Element definition.
type Element struct {
	Name    string       // name of the element (future: qname)
	Attrs   []Attribute  // the elements Attribute list
	Content ContentModel // content model
}

// ContentModel is either a content model or content model fragment
type ContentModel struct {
	children     []*ContentModel // non-nil for groups
	element      *Element        // non-nil for elementModelType
	modelType    modelType
	multiplicity multiplicity
}

// modelType identifies the type of content model fragment
type modelType int

const (
	unknownModelType  modelType = iota
	pcdataModelType             // #PCDATA
	elementModelType            // ident
	groupModelType              // ()
	sequenceModelType           // (,)
	choiceModelType             // (|)
	allModelType                // (&) - not supported by XML DTD's
)

// multiplicity of the model fragment
type multiplicity string

const (
	singleMultiplicity     = "" // default
	optionalMultiplicity   = "?"
	zeroOrMoreMultiplicity = "*"
	oneOrMoreMultiplicity  = "+"
)

// Convert the content model to a string
func (c *ContentModel) String() string {
	if c == nil {
		return "EMPTY"
	}
	return baseString(c) + string(c.multiplicity)
}

func baseString(c *ContentModel) string {
	switch c.modelType {
	case pcdataModelType:
		return "(#PCDATA)"
	case elementModelType:
		return c.element.Name
	case groupModelType:
		if len(c.children) == 1 {
			return "(" + c.children[0].String() + ")"
		}
		fallthrough
	case choiceModelType:
	case allModelType:
	case sequenceModelType:
		var result bytes.Buffer
		sep := getSep(c.modelType)
		result.WriteRune('(')
		result.WriteString(c.children[0].String())
		for i := 1; i < len(c.children); i++ {
			result.WriteString(sep)
			result.WriteString(c.children[i].String())
		}
		result.WriteRune(')')
		return result.String()
	}
	return "EMPTY"
}

func getSep(mt modelType) string {
	switch mt {
	case choiceModelType:
		return " | "
	case allModelType:
		return " & "
	}
	return ", "
}
