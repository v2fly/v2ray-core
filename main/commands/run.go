package commands

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common/cmdarg"
	"github.com/v2fly/v2ray-core/v5/common/platform"
	"github.com/v2fly/v2ray-core/v5/main/commands/base"
)

var (
	configFiles          cmdarg.Arg
	configDirs           cmdarg.Arg
	configFormat         *string
	configDirRecursively *bool
)

func setConfigFlags(cmd *base.Command) {
	configFormat = cmd.Flag.String("format", core.FormatAuto, "")
	configDirRecursively = cmd.Flag.Bool("r", false, "")

	cmd.Flag.Var(&configFiles, "config", "")
	cmd.Flag.Var(&configFiles, "c", "")
	cmd.Flag.Var(&configDirs, "confdir", "")
	cmd.Flag.Var(&configDirs, "d", "")
}

func fileExists(file string) bool {
	info, err := os.Stat(file)
	return err == nil && !info.IsDir()
}

func dirExists(file string) bool {
	if file == "" {
		return false
	}
	info, err := os.Stat(file)
	return err == nil && info.IsDir()
}

func readConfDir(dirPath string, extension []string) cmdarg.Arg {
	confs, err := os.ReadDir(dirPath)
	if err != nil {
		base.Fatalf("failed to read dir %s: %s", dirPath, err)
	}
	files := make(cmdarg.Arg, 0)
	for _, f := range confs {
		ext := filepath.Ext(f.Name())
		for _, e := range extension {
			if strings.EqualFold(e, ext) {
				files.Set(filepath.Join(dirPath, f.Name()))
				break
			}
		}
	}
	return files
}

// getFolderFiles get files in the folder and it's children
func readConfDirRecursively(dirPath string, extension []string) cmdarg.Arg {
	files := make(cmdarg.Arg, 0)
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		ext := filepath.Ext(path)
		for _, e := range extension {
			if strings.EqualFold(e, ext) {
				files.Set(path)
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

func getConfigFilePath() cmdarg.Arg {
	extension, err := core.GetLoaderExtensions(*configFormat)
	if err != nil {
		base.Fatalf(err.Error())
	}
	dirReader := readConfDir
	if *configDirRecursively {
		dirReader = readConfDirRecursively
	}
	if len(configDirs) > 0 {
		for _, d := range configDirs {
			log.Println("Using confdir from arg:", d)
			configFiles = append(configFiles, dirReader(d, extension)...)
		}
	} else if envConfDir := platform.GetConfDirPath(); dirExists(envConfDir) {
		log.Println("Using confdir from env:", envConfDir)
		configFiles = append(configFiles, dirReader(envConfDir, extension)...)
	}
	if len(configFiles) > 0 {
		return configFiles
	}

	if len(configFiles) == 0 && len(configDirs) > 0 {
		base.Fatalf("no config file found with extension: %s", extension)
	}

	if workingDir, err := os.Getwd(); err == nil {
		configFile := filepath.Join(workingDir, "config.json")
		if fileExists(configFile) {
			log.Println("Using default config: ", configFile)
			return cmdarg.Arg{configFile}
		}
	}

	if configFile := platform.GetConfigurationPath(); fileExists(configFile) {
		log.Println("Using config from env: ", configFile)
		return cmdarg.Arg{configFile}
	}

	return nil
}

func startV2Ray() (core.Server, error) {
	config, err := core.LoadConfig(*configFormat, configFiles)
	if err != nil {
		if len(configFiles) == 0 {
			err = newError("failed to load config").Base(err)
		} else {
			err = newError(fmt.Sprintf("failed to load config: %s", configFiles)).Base(err)
		}
		return nil, err
	}

	server, err := core.New(config)
	if err != nil {
		return nil, newError("failed to create server").Base(err)
	}

	return server, nil
}
