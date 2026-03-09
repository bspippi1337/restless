package engine

import (
	"io"
	"os/exec"
)

func RenderDOT(dot string, outfile string) error {

	cmd := exec.Command("dot", "-Tsvg", "-o", outfile)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, dot)
	}()

	return cmd.Run()
}
