package all

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func readDir(dirPath string, extensions []string) ([]string, error) {
	confs, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
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
	return files, nil
}

func readDirRecursively(dirPath string, extensions []string) ([]string, error) {
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
		return nil, err
	}
	return files, nil
}

// resolveFolderToFiles expands folder path (if any and it exists) to file paths.
// Any other paths, like file, even URL, it returns them as is.
func resolveFolderToFiles(paths []string, extensions []string, recursively bool) ([]string, error) {
	dirReader := readDir
	if recursively {
		dirReader = readDirRecursively
	}
	files := make([]string, 0)
	for _, p := range paths {
		i, err := os.Stat(p)
		if err == nil && i.IsDir() {
			fs, err := dirReader(p, extensions)
			if err != nil {
				return nil, fmt.Errorf("failed to read dir %s: %s", p, err)
			}
			files = append(files, fs...)
			continue
		}
		files = append(files, p)
	}
	return files, nil
}
