package mergers

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/v2fly/v2ray-core/v4/common/cmdarg"
	"github.com/v2fly/v2ray-core/v4/infra/conf/merge"
)

type jsonConverter func(v []byte) ([]byte, error)

func makeLoader(name string, extensions []string, converter jsonConverter) *MergeableFormat {
	return &MergeableFormat{
		Name:       name,
		Extensions: extensions,
		Loader:     makeConvertToJSONLoader(converter),
	}
}

func makeConvertToJSONLoader(converter func(v []byte) ([]byte, error)) MergeLoader {
	return func(input interface{}, target map[string]interface{}) error {
		if target == nil {
			panic("merge target is nil")
		}
		switch v := input.(type) {
		case string:
			err := loadFile(v, target, converter)
			if err != nil {
				return err
			}
		case []string:
			err := loadFiles(v, target, converter)
			if err != nil {
				return err
			}
		case cmdarg.Arg:
			err := loadFiles(v, target, converter)
			if err != nil {
				return err
			}
		case []byte:
			err := loadBytes(v, target, converter)
			if err != nil {
				return err
			}
		case io.Reader:
			err := loadReader(v, target, converter)
			if err != nil {
				return err
			}
		default:
			return newError("unknow merge input type")
		}
		return nil
	}
}

func loadFiles(files []string, target map[string]interface{}, converter func(v []byte) ([]byte, error)) error {
	for _, file := range files {
		err := loadFile(file, target, converter)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadFile(file string, target map[string]interface{}, converter func(v []byte) ([]byte, error)) error {
	bs, err := cmdarg.LoadArgToBytes(file)
	if err != nil {
		return fmt.Errorf("fail to load %s: %s", file, err)
	}
	if converter != nil {
		bs, err = converter(bs)
		if err != nil {
			return fmt.Errorf("error convert to json '%s': %s", file, err)
		}
	}
	_, err = merge.ToMap(bs, target)
	return err
}

func loadReader(reader io.Reader, target map[string]interface{}, converter func(v []byte) ([]byte, error)) error {
	bs, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	return loadBytes(bs, target, converter)
}

func loadBytes(bs []byte, target map[string]interface{}, converter func(v []byte) ([]byte, error)) error {
	var err error
	if converter != nil {
		bs, err = converter(bs)
		if err != nil {
			return fmt.Errorf("fail to convert to json: %s", err)
		}
	}
	_, err = merge.ToMap(bs, target)
	return err
}
