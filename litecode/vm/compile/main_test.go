package compile

import (
	"fmt"
	"os"
	"testing"

	"slices"
)

func Expect(t *testing.T, got, want any) {
	t.Helper()
	if got != want {
		t.Errorf("Expected %v, got %v", want, got)
	}
}

func TestDeserialise(t *testing.T) {
	const file = "hello.bytecode"

	bytecode, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("Failed to read bytecode file: %v", err)
	}

	d, err := Deserialise(bytecode)
	if err != nil {
		t.Fatalf("Failed to deserialize bytecode: %v", err)
	}

	Expect(t, len(d.ProtoList), 1)
	Expect(t, d.ProtoList[0], d.MainProto)
	Expect(t, d.MainProto.Dbgname, "(main)")
	Expect(t, d.MainProto.Code[0].Opcode, uint8(65))
	Expect(t, slices.Equal(d.MainProto.InstLineInfo, []uint32{1, 1, 1, 1, 1, 2}), true)

	fmt.Println(d.MainProto.InstLineInfo)
}
