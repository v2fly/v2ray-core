package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// GetModuleName returns the value of module in `go.mod` file.
func GetModuleName(pathToProjectRoot string) (string, error) {
	var moduleName string
	loopPath := pathToProjectRoot
	for {
		if idx := strings.LastIndex(loopPath, string(filepath.Separator)); idx >= 0 {
			gomodPath := filepath.Join(loopPath, "go.mod")
			gomodBytes, err := ioutil.ReadFile(gomodPath)
			if err != nil {
				loopPath = loopPath[:idx]
				continue
			}

			gomodContent := string(gomodBytes)
			moduleIdx := strings.Index(gomodContent, "module ")
			newLineIdx := strings.Index(gomodContent, "\n")

			if moduleIdx >= 0 {
				if newLineIdx >= 0 {
					moduleName = strings.TrimSpace(gomodContent[moduleIdx+6 : newLineIdx])
					moduleName = strings.TrimSuffix(moduleName, "\r")
				} else {
					moduleName = strings.TrimSpace(gomodContent[moduleIdx+6:])
				}
				return moduleName, nil
			}
			return "", fmt.Errorf("can not get module path in `%s`", gomodPath)
		}
		break
	}
	return moduleName, fmt.Errorf("no `go.mod` file in every parent directory of `%s`", pathToProjectRoot)
}

func generateError(path string) {
	pkg := filepath.Base(path)
	if pkg == "v2ray-core" {
		pkg = "core"
	}

	moduleName, gmnErr := GetModuleName(path)
	if gmnErr != nil {
		fmt.Println("can not get module path", gmnErr)
		os.Exit(1)
	}

	file, err := os.OpenFile(path+"/errors.generated.go", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Failed to generate errors.generated.go: %v", err)
	}
	defer file.Close()

	fmt.Fprintln(file, "package", pkg)
	fmt.Fprintln(file, "")
	fmt.Fprintln(file, "import \""+moduleName+"/common/errors\"")
	fmt.Fprintln(file, "")
	fmt.Fprintln(file, "type errPathObjHolder struct{}")
	fmt.Fprintln(file, "")
	fmt.Fprintln(file, "func newError(values ...interface{}) *errors.Error {")
	fmt.Fprintln(file, "	return errors.New(values...).WithPathObj(errPathObjHolder{})")
	fmt.Fprintln(file, "}")
}

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("can not get current working directory")
		os.Exit(1)
	}

	generateError(pwd)

	genPkgs := []string{"app", "infra", "common", "features", "main", "proxy", "transport"}
	noGenPkgs := []string{
		"common/errors",
		"common/log",
		"common/platform",
		"common/serial",
		"common/signal",
		"common/signal/done",
		"common/signal/semaphore",
		"infra/errorgen",
		"infra/vprotogen"}

	for _, c := range genPkgs {
		walkErr := filepath.Walk(c, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Println(err)
				return err
			}

			if !info.IsDir() {
				return nil
			}

			for _, noGen := range noGenPkgs {
				if path == noGen {
					return nil
				}
			}

			println(path)
			generateError("./" + "/" + path)

			return nil
		})

		if walkErr != nil {
			fmt.Println(walkErr)
			os.Exit(1)
		}
	}
}
