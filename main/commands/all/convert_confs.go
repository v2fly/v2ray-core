package all

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/v2fly/v2ray-core/v4/common/cmdarg"
	"github.com/v2fly/v2ray-core/v4/infra/conf/json"
	"github.com/v2fly/v2ray-core/v4/infra/conf/merge"
	"github.com/v2fly/v2ray-core/v4/main/commands/base"
)

func mergeConvertToMap(files []string, format string) map[string]interface{} {
	var (
		m   map[string]interface{}
		err error
	)
	switch inputFormat {
	case "json":
		m, err = merge.FilesToMap(files)
		if err != nil {
			base.Fatalf("failed to load json: %s", err)
		}
	case "toml":
		bs, err := tomlsToJSONs(files)
		if err != nil {
			base.Fatalf("failed to convert toml to json: %s", err)
		}
		m, err = merge.BytesToMap(bs)
		if err != nil {
			base.Fatalf("failed to merge converted json: %s", err)
		}
	case "yaml":
		bs, err := yamlsToJSONs(files)
		if err != nil {
			base.Fatalf("failed to convert yaml to json: %s", err)
		}
		m, err = merge.BytesToMap(bs)
		if err != nil {
			base.Fatalf("failed to merge converted json: %s", err)
		}
	default:
		base.Errorf("invalid input format: %s", format)
		base.Errorf("Run '%s help %s' for details.", base.CommandEnv.Exec, cmdConvert.LongName())
		base.Exit()
	}
	return m
}

// resolveFolderToFiles expands folder path (if any and it exists) to file paths.
// Any other paths, like file, even URL, it returns them as is.
func resolveFolderToFiles(paths []string, extensions []string, recursively bool) []string {
	dirReader := readConfDir
	if recursively {
		dirReader = readConfDirRecursively
	}
	files := make([]string, 0)
	for _, p := range paths {
		i, err := os.Stat(p)
		if err == nil && i.IsDir() {
			files = append(files, dirReader(p, extensions)...)
			continue
		}
		files = append(files, p)
	}
	return files
}

func readConfDir(dirPath string, extensions []string) []string {
	confs, err := ioutil.ReadDir(dirPath)
	if err != nil {
		base.Fatalf("failed to read dir %s: %s", dirPath, err)
	}
	files := make([]string, 0)
	for _, f := range confs {
		ext := filepath.Ext(f.Name())
		for _, e := range extensions {
			if strings.EqualFold(ext, e) {
				files = append(files, filepath.Join(dirPath, f.Name()))
				break
			}
		}
	}
	return files
}

// getFolderFiles get files in the folder and it's children
func readConfDirRecursively(dirPath string, extensions []string) []string {
	files := make([]string, 0)
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		ext := filepath.Ext(path)
		for _, e := range extensions {
			if strings.EqualFold(ext, e) {
				files = append(files, path)
				break
			}
		}
		return nil
	})
	if err != nil {
		base.Fatalf("failed to read dir %s: %s", dirPath, err)
	}
	return files
}

func yamlsToJSONs(files []string) ([][]byte, error) {
	jsons := make([][]byte, 0)
	for _, file := range files {
		bs, err := cmdarg.LoadArgToBytes(file)
		if err != nil {
			return nil, err
		}
		j, err := json.FromYAML(bs)
		if err != nil {
			return nil, err
		}
		jsons = append(jsons, j)
	}
	return jsons, nil
}

func tomlsToJSONs(files []string) ([][]byte, error) {
	jsons := make([][]byte, 0)
	for _, file := range files {
		bs, err := cmdarg.LoadArgToBytes(file)
		if err != nil {
			return nil, err
		}
		j, err := json.FromTOML(bs)
		if err != nil {
			return nil, err
		}
		jsons = append(jsons, j)
	}
	return jsons, nil
}
