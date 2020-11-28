package all

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/v2fly/v2ray-core/v4/commands/base"
	"github.com/v2fly/v2ray-core/v4/infra/conf/merge"
)

var cmdMerge = &base.Command{
	UsageLine: "{{.Exec}} merge [-r] [c1.json] [url] [dir1] ...",
	Short:     "Merge json files into one",
	Long: `
Merge JSON files into one.

Arguments:

	-r
		Load confdir recursively.

Examples:

	{{.Exec}} {{.LongName}} c1.json c2.json 
	{{.Exec}} {{.LongName}} c1.json https://url.to/c2.json 
	{{.Exec}} {{.LongName}} "path/to/json_dir"
`,
}

func init() {
	cmdMerge.Run = executeMerge
}

var mergeReadDirRecursively = cmdMerge.Flag.Bool("r", false, "")

func executeMerge(cmd *base.Command, args []string) {
	unnamed := cmd.Flag.Args()
	files := resolveFolderToFiles(unnamed, *mergeReadDirRecursively)
	if len(files) == 0 {
		base.Fatalf("empty config list")
	}

	data, err := merge.FilesToJSON(files)
	if err != nil {
		base.Fatalf(err.Error())
	}
	if _, err := os.Stdout.Write(data); err != nil {
		base.Fatalf(err.Error())
	}
}

// resolveFolderToFiles expands folder path (if any and it exists) to file paths.
// Any other paths, like file, even URL, it returns them as is.
func resolveFolderToFiles(paths []string, recursively bool) []string {
	dirReader := readConfDir
	if recursively {
		dirReader = readConfDirRecursively
	}
	files := make([]string, 0)
	for _, p := range paths {
		i, err := os.Stat(p)
		if err == nil && i.IsDir() {
			files = append(files, dirReader(p)...)
			continue
		}
		files = append(files, p)
	}
	return files
}

func readConfDir(dirPath string) []string {
	confs, err := ioutil.ReadDir(dirPath)
	if err != nil {
		base.Fatalf("failed to read dir %s: %s", dirPath, err)
	}
	files := make([]string, 0)
	for _, f := range confs {
		ext := filepath.Ext(f.Name())
		if ext == ".json" || ext == ".jsonc" {
			files = append(files, filepath.Join(dirPath, f.Name()))
		}
	}
	return files
}

// getFolderFiles get files in the folder and it's children
func readConfDirRecursively(dirPath string) []string {
	files := make([]string, 0)
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		ext := filepath.Ext(path)
		if ext == ".json" || ext == ".jsonc" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		base.Fatalf("failed to read dir %s: %s", dirPath, err)
	}
	return files
}
