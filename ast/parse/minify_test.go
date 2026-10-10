package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Heliodex/coputer/ast/lex"
)

// MinifiedOutSuffix is appended to a formatting test's name to get the file containing its expected minified output.
const MinifiedOutSuffix = "_minified"

// TestMinification parses each {name}.luau file in test/formatting and checks that its Minify() output matches the expected {name}_minified.luau file, mirroring TestFormatting.
// Set MINIFY_UPDATE=1 to regenerate the expected files from the current implementation.
func TestMinification(t *testing.T) {
	files, err := os.ReadDir("../" + FormattingDir)
	if err != nil {
		t.Fatal("error reading formatting tests directory:", err)
	}

	for _, f := range files {
		fn := f.Name()
		if !strings.HasSuffix(fn, Ext) || strings.HasSuffix(fn, FormattingOutSuffix+Ext) || strings.HasSuffix(fn, MinifiedOutSuffix+Ext) {
			continue
		}
		name := trimext(fn)

		t.Log(" -- Testing", name, "--")
		filename := fmt.Sprintf("../%s/%s", FormattingDir, name)

		content, err := os.ReadFile(filename + Ext)
		if err != nil {
			t.Fatal("error reading test file:", err)
		}

		ok, res := Parse(string(content), Options{})
		if !ok {
			t.Errorf("%s: error parsing file: %v", name, res.Errors)
			continue
		}

		// minified files are expected to end with a newline
		got := formatAST(res).Root.Minify() + "\n"

		expectedBytes, err := os.ReadFile(filename + MinifiedOutSuffix + Ext)
		if err != nil {
			t.Fatal("error reading expected output:", err)
		}
		expected := strings.ReplaceAll(string(expectedBytes), "\r\n", "\n")

		if !compareSource(t, name, got, expected) {
			continue
		}

		// the expected output should itself be valid Luau
		if ok, res := Parse(got, Options{}); !ok {
			t.Errorf("%s: minified output failed to parse: %v", name, res.Errors)
		}
	}
}

// TestMinificationIdempotent checks that parsing already-minified files and rendering them again reproduces them exactly.
func TestMinificationIdempotent(t *testing.T) {
	files, err := os.ReadDir("../" + FormattingDir)
	if err != nil {
		t.Fatal("error reading formatting tests directory:", err)
	}

	for _, f := range files {
		fn := f.Name()
		if !strings.HasSuffix(fn, MinifiedOutSuffix+Ext) {
			continue
		}
		name := strings.TrimSuffix(trimext(fn), MinifiedOutSuffix)

		t.Log(" -- Testing", name, "--")
		filename := fmt.Sprintf("../%s/%s%s", FormattingDir, name, MinifiedOutSuffix)

		content, err := os.ReadFile(filename + Ext)
		if err != nil {
			t.Fatal("error reading minified file:", err)
		}
		expected := strings.ReplaceAll(string(content), "\r\n", "\n")

		ok, res := Parse(expected, Options{})
		if !ok {
			t.Errorf("%s: error parsing minified file: %v", name, res.Errors)
			continue
		}

		compareSource(t, name, formatAST(res).Root.Minify()+"\n", expected)
	}
}

// TestMinifyStatementSeparators checks the three ways statements are joined: directly when Luau accepts it, by a newline when the tokens would merge, and by a semicolon when a newline still wouldn't separate them.
func TestMinifyStatementSeparators(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "calls concatenate directly",
			src:  "f()\ng()\n",
			want: "f()g()",
		},
		{
			name: "a table value concatenates directly with the next statement",
			src:  "x = {}\ny = 2\n",
			want: "x={}y=2",
		},
		{
			name: "a string value concatenates directly with the next statement",
			src:  "local s = \"x\"\nlocal b = 1\n",
			want: "local s=\"x\"local b=1",
		},
		{
			name: "word tokens are split by a newline",
			src:  "local a = 1\nlocal b = 2\n",
			want: "local a=1\nlocal b=2",
		},
		{
			name: "a block end and a name are split by a newline",
			src:  "if x then end\nz = 2\n",
			want: "if x then end\nz=2",
		},
		{
			name: "a callable followed by a parenthesised statement takes a semicolon",
			src:  "f()\n;(a or b)()\n",
			want: "f();(a or b)()",
		},
		{
			name: "a value followed by a parenthesised statement takes a semicolon",
			src:  "local x = y\n;(a or b)()\n",
			want: "local x=y;(a or b)()",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, res := Parse(tc.src, Options{})
			if !ok {
				t.Fatal("error parsing source:", res.Errors)
			}

			got := res.Root.Minify()
			if got != tc.want {
				t.Errorf("unexpected minified output:\n-- Expected\n%s\n-- Got\n%s\n", tc.want, got)
			}

			ok, res2 := Parse(got, Options{})
			if !ok {
				t.Fatal("error parsing minified output:", res2.Errors)
			}
			if got2 := res2.Root.Minify(); got2 != got {
				t.Errorf("minified output is not stable:\n-- First\n%s\n-- Second\n%s\n", got, got2)
			}
			// without comments, Source() should be identical if minification preserved the parse
			if res.Root.Source() != res2.Root.Source() {
				t.Errorf("minification changed the parse:\n-- Original\n%s\n-- Reparsed\n%s\n", res.Root.Source(), res2.Root.Source())
			}
		})
	}
}

// TestMinifyDeclareNodes checks Minify() for declaration nodes, which are part of the AST but aren't produced by this parser, mirroring TestSourceDeclareNodes.
func TestMinifyDeclareNodes(t *testing.T) {
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

		expected := "declare function foo<T>(a:number,...:string):(string,number)"
		if got := node.Minify(); got != expected {
			t.Errorf("unexpected minified code:\n-- Expected\n%s\n-- Got\n%s\n", expected, got)
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

		expected := "declare extern type Foo extends Bar with prop:number function method(self,foo:number):string [string]:number end"
		if got := node.Minify(); got != expected {
			t.Errorf("unexpected minified code:\n-- Expected\n%s\n-- Got\n%s\n", expected, got)
		}
	})

	t.Run("declare global", func(t *testing.T) {
		node := AstStatDeclareGlobal{NodeLoc: &NodeLoc{}, Name: "x", Type: ref("number")}

		expected := "declare x:number"
		if got := node.Minify(); got != expected {
			t.Errorf("unexpected minified code:\n-- Expected\n%s\n-- Got\n%s\n", expected, got)
		}
	})

	t.Run("attribute with args", func(t *testing.T) {
		name := "deprecated"
		node := AstAttr{NodeLoc: &NodeLoc{}, Name: &name, Args: []AstExpr{AstExprConstantString{NodeLoc: &NodeLoc{}, Value: "use"}}}

		expected := `@[deprecated("use")]`
		if got := node.Minify(); got != expected {
			t.Errorf("unexpected minified code:\n-- Expected\n%s\n-- Got\n%s\n", expected, got)
		}
	})
}

// TestMinifyDirectives checks that directive comments survive minification (including inside a block) while ordinary comments are removed.
func TestMinifyDirectives(t *testing.T) {
	src := "--!strict\n-- ordinary\nlocal x = 1\nlocal function f()\n\t--!native\n\treturn 2\nend\n"

	ok, res := Parse(src, Options{})
	if !ok {
		t.Fatal("error parsing source:", res.Errors)
	}

	got := formatAST(res).Root.Minify()

	for _, want := range []string{"--!strict\n", "--!native\n", "local x=1", "local function f()"} {
		if !strings.Contains(got, want) {
			t.Errorf("minified output missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "ordinary") {
		t.Errorf("minified output kept an ordinary comment:\n%s", got)
	}

	ok, res2 := Parse(got, Options{})
	if !ok {
		t.Fatal("error parsing minified output:", res2.Errors)
	}
	if got2 := formatAST(res2).Root.Minify(); got2 != got {
		t.Errorf("minified output is not stable:\n-- First\n%s\n-- Second\n%s\n", got, got2)
	}
}

// minifyFiles checks that Minify() output for every file in dir is valid Luau and stable.
func minifyFiles(t *testing.T, dir string) {
	files, err := os.ReadDir("../" + dir)
	if err != nil {
		t.Fatal("error reading directory:", err)
	}

	for _, f := range files {
		fn := f.Name()
		if !strings.HasSuffix(fn, Ext) || strings.HasSuffix(fn, FormattingOutSuffix+Ext) || strings.HasSuffix(fn, MinifiedOutSuffix+Ext) {
			continue
		}
		name := trimext(fn)

		content, err := os.ReadFile("../" + dir + "/" + name + Ext)
		if err != nil {
			t.Fatal("error reading file:", err)
		}

		ok, res := Parse(string(content), Options{})
		if !ok {
			t.Errorf("%s/%s: minify failed to parse: %v", dir, name, res.Errors)
			continue
		}

		minified := formatAST(res).Root.Minify()

		ok, res2 := Parse(minified, Options{})
		if !ok {
			t.Errorf("%s/%s: generated minified code failed to parse: %v\n--- minified\n%s", dir, name, res2.Errors, minified)
			continue
		}

		minified2 := formatAST(res2).Root.Minify()
		if minified != minified2 {
			t.Errorf("%s/%s: generated minified code is not stable\n--- first\n%s\n--- second\n%s", dir, name, minified, minified2)
		}
	}
}

// TestMinifyRoundTrip checks that Minify() output is valid, stable Luau for every file in the AST, conformance, benchmark and formatting test suites.
func TestMinifyRoundTrip(t *testing.T) {
	minifyFiles(t, AstDir)
	minifyFiles(t, ConformanceDir)
	minifyFiles(t, BenchmarkDir)
	minifyFiles(t, FormattingDir)
}
