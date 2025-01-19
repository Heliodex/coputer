package litecode

import (
	"fmt"
	"os/exec"
)

func Compile(filename string, oLevel ...uint8) (bytecode []byte, err error) {
	o := uint8(1)
	if len(oLevel) > 0 {
		o = oLevel[0]
	}

	cmd := exec.Command("luau-compile", "--binary", fmt.Sprintf("-O%d", o), filename)
	return cmd.Output()
}
