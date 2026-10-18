package vm

import (
	"testing"

	. "github.com/Heliodex/coputer/litecode/types"
	"github.com/Heliodex/coputer/litecode/vm/compile"
	"github.com/Heliodex/coputer/litecode/vm/std"
)

var scriptSink []Val

// benchmarkScript compiles one benchmark script once and runs it in a fresh coroutine per iteration. Compilation and coroutine setup are excluded from the measurement, matching what TestBenchmark times.
func benchmarkScript(b *testing.B, name string, o uint8) {
	b.Helper()

	p, err := compile.Compile(compile.MakeCompiler(o), benchDir+"/"+name)
	if err != nil {
		b.Fatal(err)
	}

	luau_print := std.MakeFn("print", func(args std.Args) (r []Val, err error) {
		return
	})
	var env Env
	env.AddFn(luau_print)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		b.StopTimer()
		co, _ := Load(p, env, TestArgs{})
		b.StartTimer()

		r, err := co.Resume()
		if err != nil {
			b.Fatal(err)
		}
		scriptSink = r
	}
}

func BenchmarkScriptStrings(b *testing.B)  { benchmarkScript(b, "strings", 1) }
func BenchmarkScriptHash(b *testing.B)     { benchmarkScript(b, "hash", 1) }
func BenchmarkScriptPairs(b *testing.B)    { benchmarkScript(b, "pairs", 1) }
func BenchmarkScriptClosures(b *testing.B) { benchmarkScript(b, "closures", 1) }
func BenchmarkScriptMethods(b *testing.B)  { benchmarkScript(b, "methods", 1) }

func BenchmarkScript3(b *testing.B)          { benchmarkScript(b, "3", 1) }
func BenchmarkScript4(b *testing.B)          { benchmarkScript(b, "4", 1) }
func BenchmarkScript5(b *testing.B)          { benchmarkScript(b, "5", 1) }
func BenchmarkScript10(b *testing.B)         { benchmarkScript(b, "10", 1) }
func BenchmarkScript12(b *testing.B)         { benchmarkScript(b, "12", 1) }
func BenchmarkScript14(b *testing.B)         { benchmarkScript(b, "14", 1) }
func BenchmarkScriptBase64(b *testing.B)     { benchmarkScript(b, "base64", 1) }
func BenchmarkScriptBigstring(b *testing.B)  { benchmarkScript(b, "bigstring", 1) }
func BenchmarkScriptLargealloc(b *testing.B) { benchmarkScript(b, "largealloc", 1) }
func BenchmarkScriptLuauception(b *testing.B) {
	benchmarkScript(b, "luauception", 1)
}
