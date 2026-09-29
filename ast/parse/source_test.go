package main

import (
	"os"
	"strings"
	"testing"

	"github.com/Heliodex/coputer/ast/lex"
)

func sourceFiles(t *testing.T, dir string) {
	files, err := os.ReadDir("../" + dir)
	if err != nil {
		t.Fatal("error reading directory:", err)
	}

	for _, f := range files {
		fn := f.Name()
		if !strings.HasSuffix(fn, Ext) {
			continue
		}
		name := trimext(fn)

		content, err := os.ReadFile("../" + dir + "/" + name + Ext)
		if err != nil {
			t.Fatal("error reading file:", err)
		}

		ok, res := Parse(string(content), Options{})
		if !ok {
			t.Errorf("%s/%s: source failed to parse: %v", dir, name, res.Errors)
			continue
		}

		source := res.Root.Source()

		ok, res2 := Parse(source, Options{})
		if !ok {
			t.Errorf("%s/%s: generated source failed to parse: %v\n--- source\n%s", dir, name, res2.Errors, source)
			continue
		}

		// rendering the re-parsed AST should produce identical source
		source2 := res2.Root.Source()
		if source != source2 {
			t.Errorf("%s/%s: generated source is not stable\n--- first\n%s\n--- second\n%s", dir, name, source, source2)
		}
	}
}

// TestSourceRoundTrip checks that Source() output is valid, stable Luau for
// every file in the AST, conformance and benchmark test suites.
func TestSourceRoundTrip(t *testing.T) {
	sourceFiles(t, AstDir)
	sourceFiles(t, ConformanceDir)
	sourceFiles(t, BenchmarkDir)
}

// TestSourceFormatting checks the formatted output for a representative mix of
// declarations, functions, tables, if-local statements and types.
func TestSourceFormatting(t *testing.T) {
	src := `export type Point = { x: number, y: number }
type Callback<T> = (T) -> ()
local function map<T, U>(list: {T}, f: (T) -> U): {U}
	local out = table.create(#list)
	for i, v in list do
		out[i] = f(v)
	end
	return out
end
local t = {
	a = 1,
	[2] = "two",
}
if local n = #t then
	print(n, t.a)
elseif false then
	print("no")
else
	print("n is " .. tostring(t.a))
end
`

	expected := `export type Point = { x: number, y: number }
type Callback<T> = (T) -> ()
local function map<T, U>(list: { T }, f: (T) -> U): { U }
	local out = table.create(#list)
	for i, v in list do
		out[i] = f(v)
	end
	return out
end
local t = {
	a = 1,
	[2] = "two",
}
if local n = #t then
	print(n, t.a)
elseif false then
	print("no")
else
	print("n is " .. tostring(t.a))
end`

	ok, res := Parse(src, Options{})
	if !ok {
		t.Fatal("error parsing source:", res.Errors)
	}

	if got := res.Root.Source(); got != expected {
		t.Errorf("unexpected source:\n-- Expected\n%s\n-- Got\n%s\n", expected, got)
	}
}

// TestSourceDeclareNodes checks Source() for declaration nodes, which are part
// of the AST but aren't produced by this parser.
func TestSourceDeclareNodes(t *testing.T) {
	ref := func(name string) AstTypeReference {
		return AstTypeReference{NodeLoc: &NodeLoc{}, Name: name}
	}
	pack := func(p AstTypePack) *AstTypePack { return &p }

	t.Run("declare function", func(t *testing.T) {
		node := AstStatDeclareFunction{
			NodeLoc:    &NodeLoc{},
			Name:       "foo",
			Generics:   []AstGenericType{{NodeLoc: &NodeLoc{}, Name: "T"}},
			Params:     AstTypeList{Types: []AstType{ref("number")}, TailType: pack(AstTypePackVariadic{NodeLoc: &NodeLoc{}, VariadicType: ref("string")})},
			ParamNames: []AstArgumentName{{Name: "a"}},
			Vararg:     true,
			RetTypes:   AstTypePackExplicit{NodeLoc: &NodeLoc{}, Types: AstTypeList{Types: []AstType{ref("string"), ref("number")}}},
		}

		expected := "declare function foo<T>(a: number, ...: string): (string, number)"
		if got := node.Source(); got != expected {
			t.Errorf("unexpected source:\n-- Expected\n%s\n-- Got\n%s\n", expected, got)
		}
	})

	t.Run("declare extern type", func(t *testing.T) {
		super := "Bar"
		node := AstStatDeclareExternType{
			NodeLoc:   &NodeLoc{},
			Name:      "Foo",
			SuperName: &super,
			Props: []AstDeclaredExternTypeProperty{
				{Name: lex.AstName{Value: "prop"}, Ty: ref("number")},
				{Name: lex.AstName{Value: "method"}, IsMethod: true, Ty: AstTypeFunction{
					NodeLoc:     &NodeLoc{},
					ArgTypes:    AstTypeList{Types: []AstType{ref("number")}},
					ArgNames:    []*AstArgumentName{{Name: "foo"}},
					ReturnTypes: AstTypePackExplicit{NodeLoc: &NodeLoc{}, Types: AstTypeList{Types: []AstType{ref("string")}}},
				}},
			},
			Indexer: &AstTableIndexer{IndexType: ref("string"), ResultType: ref("number"), Access: "ReadWrite"},
		}

		expected := "declare extern type Foo extends Bar with\n" +
			"\tprop: number\n" +
			"\tfunction method(self, foo: number): string\n" +
			"\t[string]: number\n" +
			"end"
		if got := node.Source(); got != expected {
			t.Errorf("unexpected source:\n-- Expected\n%s\n-- Got\n%s\n", expected, got)
		}
	})

	t.Run("declare global", func(t *testing.T) {
		node := AstStatDeclareGlobal{NodeLoc: &NodeLoc{}, Name: "x", Type: ref("number")}

		expected := "declare x: number"
		if got := node.Source(); got != expected {
			t.Errorf("unexpected source:\n-- Expected\n%s\n-- Got\n%s\n", expected, got)
		}
	})

	t.Run("attribute with args", func(t *testing.T) {
		name := "deprecated"
		node := AstAttr{NodeLoc: &NodeLoc{}, Name: &name, Args: []AstExpr{AstExprConstantString{NodeLoc: &NodeLoc{}, Value: "use"}}}

		expected := `@[deprecated("use")]`
		if got := node.Source(); got != expected {
			t.Errorf("unexpected source:\n-- Expected\n%s\n-- Got\n%s\n", expected, got)
		}
	})
}
