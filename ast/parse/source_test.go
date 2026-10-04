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

// TestSourceRoundTrip checks that Source() output is valid, stable Luau for every file in the AST, conformance and benchmark test suites.
func TestSourceRoundTrip(t *testing.T) {
	sourceFiles(t, AstDir)
	sourceFiles(t, ConformanceDir)
	sourceFiles(t, BenchmarkDir)
}

// TestSourceFormatting checks the formatted output for a representative mix of declarations, functions, tables, if-local statements and types.
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

	expected := `export type Point = {
	x: number,
	y: number,
}
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
	print "no"
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

// TestSourceServiceSorting checks that a leading run of `game:GetService` declarations is sorted alphabetically by service name, including the guards that keep the rewrite behaviour-preserving.
func TestSourceServiceSorting(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "sorts locals by service name",
			src:  "local Workspace = game:GetService \"Workspace\"\nlocal Players = game:GetService \"Players\"\nlocal RunService = game:GetService \"RunService\"\n",
			want: "local Players = game:GetService \"Players\"\nlocal RunService = game:GetService \"RunService\"\nlocal Workspace = game:GetService \"Workspace\"",
		},
		{
			name: "comments move with their declaration but a file header stays put",
			src:  "--!strict\n-- header\nlocal Zed = game:GetService \"Zed\"\n-- players\nlocal Players = game:GetService \"Players\" -- the players\n",
			want: "--!strict\n-- header\n-- players\n\nlocal Players = game:GetService \"Players\" -- the players\nlocal Zed = game:GetService \"Zed\"",
		},
		{
			name: "parenthesised calls and sugar are both recognised",
			src:  "local C = game:GetService(\"Camera\")\nlocal A = game:GetService \"Accessory\"\nlocal B = game:GetService(\"Body\")\n",
			want: "local A = game:GetService \"Accessory\"\nlocal B = game:GetService \"Body\"\nlocal C = game:GetService \"Camera\"",
		},
		{
			name: "type annotations do not disqualify",
			src:  "local B: Instance = game:GetService \"B\"\nlocal A: Instance = game:GetService \"A\"\n",
			want: "local A: Instance = game:GetService \"A\"\nlocal B: Instance = game:GetService \"B\"",
		},
		{
			name: "the run stops at the first non-service statement",
			src:  "local B = game:GetService \"B\"\nlocal A = game:GetService \"A\"\nlocal C = game:GetService \"C\"\n",
			want: "local A = game:GetService \"A\"\nlocal B = game:GetService \"B\"\nlocal C = game:GetService \"C\"",
		},
		{
			name: "duplicate names still sort by service name",
			src:  "local A = game:GetService \"Zed\"\nlocal A = game:GetService \"Apple\"\n",
			want: "local A = game:GetService \"Apple\"\nlocal A = game:GetService \"Zed\"",
		},
		{
			name: "a non-game base is left unsorted",
			src:  "local B = Client:GetService \"B\"\nlocal A = Client:GetService \"A\"\n",
			want: "local B = Client:GetService \"B\"\nlocal A = Client:GetService \"A\"",
		},
		{
			name: "a non-string argument is left unsorted",
			src:  "local B = game:GetService(name)\nlocal A = game:GetService(other)\n",
			want: "local B = game:GetService(name)\nlocal A = game:GetService(other)",
		},
		{
			name: "multiple declared variables are left unsorted",
			src:  "local B, C = game:GetService \"B\", 1\nlocal A = game:GetService \"A\"\n",
			want: "local B, C = game:GetService \"B\", 1\nlocal A = game:GetService \"A\"",
		},
		{
			name: "only the leading same-kind run sorts",
			src:  "local B = game:GetService \"B\"\nlocal A = game:GetService \"A\"\nconst D = game:GetService \"D\"\nconst C = game:GetService \"C\"\n",
			want: "local A = game:GetService \"A\"\nlocal B = game:GetService \"B\"\n\nconst D = game:GetService \"D\"\nconst C = game:GetService \"C\"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, res := Parse(tc.src, Options{})
			if !ok {
				t.Fatal("error parsing source:", res.Errors)
			}

			got := res.Root.Source()
			if got != tc.want {
				t.Errorf("unexpected source:\n-- Expected\n%s\n-- Got\n%s\n", tc.want, got)
			}

			// the rewritten source must be stable
			ok, res2 := Parse(got, Options{})
			if !ok {
				t.Fatal("error parsing generated source:", res2.Errors)
			}
			if got2 := res2.Root.Source(); got2 != got {
				t.Errorf("generated source is not stable:\n-- First\n%s\n-- Second\n%s\n", got, got2)
			}
		})
	}
}

// TestSourceDirectiveSorting checks that comment directives at the top of the file are pulled together and sorted by category, that only the first type-check directive survives, and that non-header directives are left alone.
func TestSourceDirectiveSorting(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "sorts directives by category",
			src:  "--!strict\n--!nonstrict\n--!nocheck\n--!native\n--!nolint\n--!optimize 2\n\ntable.clear()\n",
			want: "--!strict\n--!nolint\n--!native\n--!optimize 2\n\ntable.clear()",
		},
		{
			name: "pulls interleaved directives above a normal comment",
			src:  "-- normal\n--!native\n--!strict\nlocal x = 1\n",
			want: "--!strict\n--!native\n-- normal\n\nlocal x = 1",
		},
		{
			name: "keeps the first type-check directive",
			src:  "--!nocheck\n--!strict\nlocal x = 1\n",
			want: "--!nocheck\n\nlocal x = 1",
		},
		{
			name: "leaves directives after the first statement alone",
			src:  "local x = 1\n--!native\n--!strict\nlocal y = 2\n",
			want: "local x = 1\n--!native\n--!strict\nlocal y = 2",
		},
		{
			name: "sorts directives when the file has no statements",
			src:  "--!native\n--!strict\n--!nolint\n",
			want: "--!strict\n--!nolint\n--!native",
		},
		{
			name: "keeps unknown directives after the known ones",
			src:  "--!wat\n--!native\n--!strict\nlocal x = 1\n",
			want: "--!strict\n--!native\n--!wat\n\nlocal x = 1",
		},
		{
			name: "keeps repeated lint and native directives",
			src:  "--!native\n--!native\n--!nolint UnknownGlobal\n--!strict\nlocal x = 1\n",
			want: "--!strict\n--!nolint UnknownGlobal\n--!native\n--!native\n\nlocal x = 1",
		},
		{
			name: "a space after the bang is not a directive",
			src:  "--! strict\n--!nocheck\nlocal x = 1\n",
			want: "--!nocheck\n--! strict\n\nlocal x = 1",
		},
		{
			name: "block and trailing comments are not directives",
			src:  "--[[!strict]]\nlocal x = 1 --!native\n",
			want: "--[[!strict]]\nlocal x = 1 --!native",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, res := Parse(tc.src, Options{})
			if !ok {
				t.Fatal("error parsing source:", res.Errors)
			}

			got := res.Root.Source()
			if got != tc.want {
				t.Errorf("unexpected source:\n-- Expected\n%s\n-- Got\n%s\n", tc.want, got)
			}

			ok, res2 := Parse(got, Options{})
			if !ok {
				t.Fatal("error parsing generated source:", res2.Errors)
			}
			if got2 := res2.Root.Source(); got2 != got {
				t.Errorf("generated source is not stable:\n-- First\n%s\n-- Second\n%s\n", got, got2)
			}
		})
	}
}

// TestSourceRequireSorting checks that a leading run of `require` declarations is sorted by filename, with string requires before instance-path ones, and that it composes after the service section.
func TestSourceRequireSorting(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "sorts string requires by filename",
			src:  "local Types = require \"../Types\"\nlocal logError = require \"../Logging/logError\"\nlocal isSimilar = require \"../Utility/isSimilar\"\n",
			want: "local logError = require \"../Logging/logError\"\nlocal Types = require \"../Types\"\nlocal isSimilar = require \"../Utility/isSimilar\"",
		},
		{
			name: "string requires come before instance requires",
			src:  "local inst = require(script.Parent.Services)\nlocal a = require \"../A\"\nlocal b = require(script.Parent.Alpha)\n",
			want: "local a = require \"../A\"\nlocal b = require(script.Parent.Alpha)\nlocal inst = require(script.Parent.Services)",
		},
		{
			name: "sorts instance requires by path",
			src:  "local z = require(script.Parent.Zed)\nlocal a = require(script.Parent.Alpha)\n",
			want: "local a = require(script.Parent.Alpha)\nlocal z = require(script.Parent.Zed)",
		},
		{
			name: "a single require is left alone",
			src:  "local A = require \"../A\"\nlocal B = 2\n",
			want: "local A = require \"../A\"\n\nlocal B = 2",
		},
		{
			name: "requires after a non-require statement are left alone",
			src:  "local B = require \"../B\"\nlocal A = require \"../A\"\nlocal x = 1\nlocal D = require \"../D\"\nlocal C = require \"../C\"\n",
			want: "local A = require \"../A\"\nlocal B = require \"../B\"\n\nlocal x = 1\nlocal D = require \"../D\"\nlocal C = require \"../C\"",
		},
		{
			name: "composes after the sorted service section",
			src:  "local Players = game:GetService \"Players\"\nlocal Workspace = game:GetService \"Workspace\"\nlocal B = require \"../B\"\nlocal A = require \"../A\"\nlocal rest = 1\n",
			want: "local Players = game:GetService \"Players\"\nlocal Workspace = game:GetService \"Workspace\"\n\nlocal A = require \"../A\"\nlocal B = require \"../B\"\n\nlocal rest = 1",
		},
		{
			name: "a shadowed require local is not sorted",
			src:  "local require = function(x)\n\treturn x\nend\nlocal B = require \"../B\"\nlocal A = require \"../A\"\n",
			want: "local require = function(x)\n\treturn x\nend\nlocal B = require \"../B\"\nlocal A = require \"../A\"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, res := Parse(tc.src, Options{})
			if !ok {
				t.Fatal("error parsing source:", res.Errors)
			}

			got := res.Root.Source()
			if got != tc.want {
				t.Errorf("unexpected source:\n-- Expected\n%s\n-- Got\n%s\n", tc.want, got)
			}

			ok, res2 := Parse(got, Options{})
			if !ok {
				t.Fatal("error parsing generated source:", res2.Errors)
			}
			if got2 := res2.Root.Source(); got2 != got {
				t.Errorf("generated source is not stable:\n-- First\n%s\n-- Second\n%s\n", got, got2)
			}
		})
	}
}

// TestSourceComments checks that comments are rendered near their original positions: leading comments stay above the statement they precede, trailing comments stay on the same line, and comments inside a block stay inside it.
func TestSourceComments(t *testing.T) {
	src := "-- leading comment\n" +
		"local x = 1 -- trailing comment\n" +
		"--[[ block comment ]]\n" +
		"local function f()\n" +
		"\t-- inner comment\n" +
		"\treturn x -- return comment\n" +
		"end\n" +
		"do\n" +
		"\t--[[ multi\n" +
		"\tline ]]\n" +
		"end\n"

	expected := "-- leading comment\n" +
		"local x = 1 -- trailing comment\n" +
		"\n" +
		"--[[ block comment ]]\n" +
		"local function f()\n" +
		"\t-- inner comment\n" +
		"\treturn x -- return comment\n" +
		"end\n" +
		"\n" +
		"do\n" +
		"\t--[[ multi\n" +
		"\tline ]]\n" +
		"end"

	ok, res := Parse(src, Options{})
	if !ok {
		t.Fatal("error parsing source:", res.Errors)
	}

	if got := res.Root.Source(); got != expected {
		t.Errorf("unexpected source:\n-- Expected\n%s\n-- Got\n%s\n", expected, got)
	}
}

// TestSourceFunctionSpacing checks that standalone function declarations are set off from the surrounding code in any scope with a single blank line, while a comment attached to the top of a function stays hugging it.
func TestSourceFunctionSpacing(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "blank lines around a function between statements",
			src:  "local x = 1\nlocal function f()\n\treturn x\nend\nlocal y = 2\n",
			want: "local x = 1\n\nlocal function f()\n\treturn x\nend\n\nlocal y = 2",
		},
		{
			name: "a sole function in a scope gets no blank lines",
			src:  "local function f()\n\treturn 1\nend\n",
			want: "local function f()\n\treturn 1\nend",
		},
		{
			name: "consecutive functions are separated by one blank line",
			src:  "local function a() end\nfunction b() end\n",
			want: "local function a()\nend\n\nfunction b()\nend",
		},
		{
			name: "a comment above a function stays hugging it",
			src:  "local x = 1\n-- doc\nlocal function f()\n\treturn x\nend\nlocal y = 2\n",
			want: "local x = 1\n\n-- doc\nlocal function f()\n\treturn x\nend\n\nlocal y = 2",
		},
		{
			name: "a function value assignment is not set off",
			src:  "local x = 1\nlocal f = function()\n\treturn x\nend\nlocal y = 2\n",
			want: "local x = 1\nlocal f = function()\n\treturn x\nend\nlocal y = 2",
		},
		{
			name: "function declarations inside a block are set off",
			src:  "do\n\tlocal x = 1\n\tlocal function f()\n\t\treturn x\n\tend\n\tlocal y = 2\nend\n",
			want: "do\n\tlocal x = 1\n\n\tlocal function f()\n\t\treturn x\n\tend\n\n\tlocal y = 2\nend",
		},
		{
			name: "an attributed function is set off",
			src:  "local x = 1\n@native\nfunction f() end\nlocal y = 2\n",
			want: "local x = 1\n\n@native\nfunction f()\nend\n\nlocal y = 2",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, res := Parse(tc.src, Options{})
			if !ok {
				t.Fatal("error parsing source:", res.Errors)
			}

			got := res.Root.Source()
			if got != tc.want {
				t.Errorf("unexpected source:\n-- Expected\n%s\n-- Got\n%s\n", tc.want, got)
			}

			ok, res2 := Parse(got, Options{})
			if !ok {
				t.Fatal("error parsing generated source:", res2.Errors)
			}
			if got2 := res2.Root.Source(); got2 != got {
				t.Errorf("generated source is not stable:\n-- First\n%s\n-- Second\n%s\n", got, got2)
			}
		})
	}
}

// TestSourceMethodCalls checks that a dot-notation call repeating its receiver identifier as the first argument is rewritten to the equivalent method call, and that forms which would change behaviour are left alone.
func TestSourceMethodCalls(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "repeated local receiver becomes a method call",
			src:  "x.func(x, whatever)",
			want: "x:func(whatever)",
		},
		{
			name: "a method with no remaining arguments",
			src:  "x.func(x)",
			want: "x:func()",
		},
		{
			name: "sugar is preserved for a single string argument",
			src:  "x.func(x, \"str\")",
			want: "x:func \"str\"",
		},
		{
			name: "sugar is preserved for a single table argument",
			src:  "x.func(x, { a = 1 })",
			want: "x:func {\n\ta = 1,\n}",
		},
		{
			name: "self is treated like any other identifier",
			src:  "self.method(self, 1, 2)",
			want: "self:method(1, 2)",
		},
		{
			name: "a global receiver is rewritten too",
			src:  "instance.func(instance, 1)",
			want: "instance:func(1)",
		},
		{
			name: "a grouped receiver is unwrapped",
			src:  "x.func((x), 1)",
			want: "x:func(1)",
		},
		{
			name: "an explicit type instantiation moves onto the method name",
			src:  "x.func<<number>>(x, 1)",
			want: "x:func<<number>>(1)",
		},
		{
			name: "a different first argument is left alone",
			src:  "x.func(y, whatever)",
			want: "x.func(y, whatever)",
		},
		{
			name: "an indexed first argument is left alone",
			src:  "x.func(x.y, whatever)",
			want: "x.func(x.y, whatever)",
		},
		{
			name: "an indexed receiver is left alone even when the argument repeats it",
			src:  "x.y.func(x.y, whatever)",
			want: "x.y.func(x.y, whatever)",
		},
		{
			name: "a call receiver is left alone so it is still evaluated twice",
			src:  "f().func(f(), whatever)",
			want: "f().func(f(), whatever)",
		},
		{
			name: "a type-asserted first argument is left alone",
			src:  "x.func(x :: T, 1)",
			want: "x.func(x :: T, 1)",
		},
		{
			name: "an existing method call is unchanged",
			src:  "x:func(1)",
			want: "x:func(1)",
		},
		{
			name: "an existing method call with type arguments is unchanged",
			src:  "x:func<<number>>(1)",
			want: "x:func<<number>>(1)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, res := Parse(tc.src, Options{})
			if !ok {
				t.Fatal("error parsing source:", res.Errors)
			}

			got := res.Root.Source()
			if got != tc.want {
				t.Errorf("unexpected source:\n-- Expected\n%s\n-- Got\n%s\n", tc.want, got)
			}

			ok, res2 := Parse(got, Options{})
			if !ok {
				t.Fatal("error parsing generated source:", res2.Errors)
			}
			if got2 := res2.Root.Source(); got2 != got {
				t.Errorf("generated source is not stable:\n-- First\n%s\n-- Second\n%s\n", got, got2)
			}
		})
	}
}

// TestSourceCommentsOption checks that comments are always rendered by Source(), while Result.CommentLocations still respects CaptureComments.
func TestSourceCommentsOption(t *testing.T) {
	src := "-- comment\nlocal x = 1\n"

	_, res := Parse(src, Options{})
	if len(res.CommentLocations) != 0 {
		t.Errorf("expected no comment locations without CaptureComments, got %d", len(res.CommentLocations))
	}
	if !strings.Contains(res.Root.Source(), "-- comment") {
		t.Errorf("expected Source() to include comments without CaptureComments")
	}

	_, res = Parse(src, Options{CaptureComments: true})
	if len(res.CommentLocations) != 1 {
		t.Fatalf("expected 1 comment location, got %d", len(res.CommentLocations))
	}
	if res.CommentLocations[0].Content != " comment" {
		t.Errorf("unexpected comment content: %q", res.CommentLocations[0].Content)
	}
}

// TestSourceDeclareNodes checks Source() for declaration nodes, which are part of the AST but aren't produced by this parser.
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
