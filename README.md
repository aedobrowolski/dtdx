## Parsing DTDX Documents

The DTDX (DTD teXt) grammar mocks up an instance using plain text strings in
order to define a document model that can then be used to generate an XML DTD.

### Elements
Definitions of elements appear as children of one of their parent elements.
An element is never defined two times. Instead it is referenced using the
name followed by a "..." suffix.  The definition does not need to come before
the reference.

### Elements Example
Here is an example dtdx document that defines paragraph structures.

```
# The first top level definition.
paragraph
    # A definition with two references nested inside paragraph.
    title?
    line...+

# A second top level definition.
line
    (#PCDATA, bold)*
```

The first element defined, paragraph, is the root of the DTD. The content
models of title and bold default to (#PCDATA). 
This example document is equivalent to the DTD:

```xml
<!ELEMENT paragraph     (title?, line+)>
<!ELEMENT title         (#PCDATA)>
<!ELEMENT line          (#PCDATA,bold)* >
<!ELEMENT bold          (#PCDATA)>
```

### Attributes

Attributes are defined after the element name as a list of name value pairs.
The attribute name must be followed by '=' followed by an optional type
directive (a #VALUE). If left off the type is derived from the name. This
usually defaults to CDATA. However, if the name is id, idref or idrefs then the
type is the upper case value of the name. If the name is 'number' then the type
is NMTOKEN. The type can also be a list of NMTOKEN values separated by the
vertical bar character '|' to create an enumerated attribute type.

### Attributes Example

Here is an example of an element with three attributes.  The first is typed
ID implicitly while the last two are given explicit types. 

```        
# Define paragraph element with three attributes
paragraph id= name=#CDATA justify=(left|right|center)
```

This example document is equivalent to the DTD:
```xml
<!ELEMENT paragraph     (#PCDATA)>
<!ATTLIST paragraph
        id      ID                  #IMPLIED
        name    CDATA               #IMPLIED
        justify (left|right|center) #IMPLIED
        >
```

## DTDX Grammar

```
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
greaterIndent   := '\n' WS             [WS.lit.len>parent.indent]
sameIndent      := '\n' WS             [WS.lit.len==parent.indent]
```