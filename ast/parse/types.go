package main

// bindings

type Binding struct {
	NodeLoc
	Name          AstName
	Annotation    *AstType
	ColonPosition *Position
}

type BindingList []Binding

// --------------------------------------------------------------------------------
// -- PARSER RESULT TYPES
// --------------------------------------------------------------------------------

type ParseError struct {
	Location Location
	Message  string
}

type Options struct {
	CaptureComments bool
	StoreCstData    bool
}
