//go:build !coverage
// +build !coverage

package scenarios

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func BuildV2Ray() error {
	genTestBinaryPath()
	if _, err := os.Stat(testBinaryPath); err == nil {
		return nil
	}

	fmt.Printf("Building V2Ray into path (%s)\n", testBinaryPath)
	cmd := exec.Command("go", "build", "-v", "-o="+testBinaryPath, "./main")
	cmd.Env = os.Environ()
	cmd.Dir, _ = filepath.Abs("../..")
	println(cmd.Dir)
	out, err := cmd.CombinedOutput()
	println(string(out))
	return err
}

func RunV2RayProtobuf(config []byte) *exec.Cmd {
	genTestBinaryPath()
	proc := exec.Command(testBinaryPath, "-config=stdin:", "-format=pb")
	proc.Stdin = bytes.NewBuffer(config)
	proc.Stderr = os.Stderr
	proc.Stdout = os.Stdout

	return proc
}
