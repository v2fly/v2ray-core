package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"v2ray.com/core/common"
)

// protocMap is the map of paths to `protoc` binary excutable files of specific platform
var protocMap = map[string]string{
	"windows": filepath.Join(".dev", "protoc", "windows", "protoc.exe"),
	"darwin":  filepath.Join(".dev", "protoc", "macos", "protoc"),
	"linux":   filepath.Join(".dev", "protoc", "linux", "protoc"),
}

func main() {
	pwd, wdErr := os.Getwd()
	if wdErr != nil {
		fmt.Println("Can not get current working directory.")
		os.Exit(1)
	}

	GOBIN := common.GetGOBIN()
	protoc := protocMap[runtime.GOOS]

	protoFilesMap := make(map[string][]string)
	walkErr := filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}

		if info.IsDir() {
			return nil
		}

		dir := filepath.Dir(path)
		filename := filepath.Base(path)
		if strings.HasSuffix(filename, ".proto") {
			protoFilesMap[dir] = append(protoFilesMap[dir], path)
		}

		return nil
	})
	if walkErr != nil {
		fmt.Println(walkErr)
		os.Exit(1)
	}

	moduleName, gmnErr := common.GetModuleName(pwd)
	if gmnErr != nil {
		fmt.Println(gmnErr)
		os.Exit(1)
	}

	protocGenGoPath := filepath.Join(GOBIN, "protoc-gen-go")
	protocGenGoGrpcPath := filepath.Join(GOBIN, "protoc-gen-go-grpc")

	for _, files := range protoFilesMap {
		for _, relProtoFile := range files {
			args := []string{"--go_out=module=" + moduleName + ":" + pwd, "--go-grpc_out=module=" + moduleName + ":" + pwd, "--plugin", "protoc-gen-go=" + protocGenGoPath, "--plugin", "protoc-gen-go-grpc=" + protocGenGoGrpcPath}
			args = append(args, relProtoFile)
			cmd := exec.Command(protoc, args...)
			cmd.Env = append(cmd.Env, os.Environ()...)
			cmd.Env = append(cmd.Env, "GOBIN="+GOBIN)
			output, cmdErr := cmd.CombinedOutput()
			if len(output) > 0 {
				fmt.Println(string(output))
			}
			if cmdErr != nil {
				fmt.Println(cmdErr)
				os.Exit(1)
			}
		}
	}
	fmt.Println("All pb.go files are generated successfully!")
}
