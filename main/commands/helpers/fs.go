package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadDir finds files according to extensions in the dir
func ReadDir(dir string, extensions []string) ([]string, error) {
	confs, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0)
	for _, f := range confs {
		ext := filepath.Ext(f.Name())
		for _, e := range extensions {
			if strings.EqualFold(ext, e) {
				files = append(files, filepath.Join(dir, f.Name()))
				break
			}
		}
	}
	return files, nil
}

// ReadDirRecursively finds files according to extensions in the dir recursively
func ReadDirRecursively(dir string, extensions []string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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

// ResolveFolderToFiles expands folder path (if any and it exists) to file paths.
// Any other paths, like file, even URL, it returns them as is.
func ResolveFolderToFiles(paths []string, extensions []string, recursively bool) ([]string, error) {
	dirReader := ReadDir
	if recursively {
		dirReader = ReadDirRecursively
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
