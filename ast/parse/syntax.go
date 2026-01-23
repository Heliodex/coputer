package main

import "github.com/Heliodex/coputer/ast/lex"

// ------------------------------------------------------------------------------
// AST NAME & LOCAL
// ------------------------------------------------------------------------------

// Represents a simple name/identifier in the AST.
// Used for variable names, function names, type names, etc.
type AstName struct {
	Value string
}

// Represents a local variable declaration in the AST.
// Tracks the variable's name, location, shadowing information, scope depth, and optional type annotation.
type AstLocal struct {
	// The name of the local variable
	Name string
	// Source location of the variable declaration
	Location lex.Location
	shadow   *AstLocal
	// Reference to a shadowed local variable with the same name (if any)
	FunctionDepth int
	// The nesting depth of the function containing this local
	LoopDepth int
	// The nesting depth of loops containing this local
	Annotation *AstType
	// Optional type annotation for the local variable
}

// ------------------------------------------------------------------------------
// TYPE ANNOTATION UNION TYPES
// ------------------------------------------------------------------------------

// Union type representing all possible type annotation nodes.
// Type annotations specify the expected types of values in Luau's type system.
type AstType interface {
	isAstType()
	Kind() string
}

// ------------------------------------------------------------------------------
// TYPE ANNOTATION IMPLEMENTATIONS
// ------------------------------------------------------------------------------

// A type reference: `TypeName` or `Module.TypeName<Params>`.
type AstTypeReference struct {
	Location lex.Location
	// True if type parameters are provided (even if empty)
	HasParameterList bool
	// Optional module/namespace prefix
	Prefix *string
	// Location of the prefix
	PrefixLocation *lex.Location
	// The type name
	Name string
	// Location of the type name
	NameLocation lex.Location
	// Type arguments passed to generic types
	Parameters []AstTypeOrPack
}

func (t *AstTypeReference) isAstType()   {}
func (t *AstTypeReference) Kind() string { return "TypeReference" }

// A table type: `{ prop: Type, [Key]: Value }`.
type AstTypeTable struct {
	Location lex.Location
	// The named properties of the table type
	Properties []AstTableProp
	// Optional indexer signature `[KeyType]: ValueType`
	Indexer *AstTableIndexer
}

func (t *AstTypeTable) isAstType()   {}
func (t *AstTypeTable) Kind() string { return "TypeTable" }

// A function type: `(ParamTypes) -> (ReturnTypes)`.
type AstTypeFunction struct {
	Location lex.Location
	// Function type attributes
	Attributes []AstAttr
	// Generic type parameters
	Generics []AstGenericType
	// Parameter types
}
