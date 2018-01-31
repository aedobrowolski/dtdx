package parser

// Attribute represents an attribute definition.
// The types legal in a DTD are ID, IDREF, IDREFS, NMTOKEN, NMTOKENS,
// ENTITY, ENTITIES, NOTATION, or an enumerated list of NMTOKEN.
type Attribute struct {
	Name    string // name of Attribute (future: qname)
	Type    string // type of Attribute
	Occur   Occur  // occurrence qualifier - default #IMPLIED
	Default string // default value of attribute or empty
}

// Occur represents the occurrence qualifier of an attribute.
type Occur string

const (
	implied  Occur = "#IMPLIED"
	required Occur = "#REQUIRED"
	fixed    Occur = "#FIXED"
)
