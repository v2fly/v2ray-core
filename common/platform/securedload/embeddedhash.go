package securedload

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"strings"

	"github.com/v2fly/VSign/insmgr"
	"github.com/v2fly/VSign/signerVerify"

	"github.com/v2fly/v2ray-core/v4/common/platform"
	"github.com/v2fly/v2ray-core/v4/common/platform/filesystem"
)

type EmbeddedHashProtectedLoader struct {
	checkedFile map[string]string
}

func (e EmbeddedHashProtectedLoader) VerifyAndLoad(filename string) ([]byte, error) {
	platformFileName := filepath.FromSlash(filename)
	fileContent, err := filesystem.ReadFile(platform.GetAssetLocation(platformFileName))
	if err != nil {
		return nil, newError("Cannot find file", filename).Base(err)
	}
	fileHash := sha256.Sum256(fileContent)
	fileHashAsString := hex.EncodeToString(fileHash[:])
	if fileNameVerified, ok := e.checkedFile[fileHashAsString]; ok {
		for _, filenameVerifiedIndividual := range strings.Split(fileNameVerified, ";") {
			if strings.HasSuffix(filenameVerifiedIndividual, filename) {
				return fileContent, nil
			}
		}
	}
	return nil, newError("Unrecognized file at ", filename, " can not be loaded for execution")
}

func NewEmbeddedHashProtectedLoader() *EmbeddedHashProtectedLoader {
	instructions := insmgr.ReadAllIns(bytes.NewReader([]byte(allowedHashes)))
	checkedFile, _, ok := signerVerify.CheckAsClient(instructions, "v2fly", true)
	if !ok {
		panic("Embedded Hash data is invalid")
	}
	return &EmbeddedHashProtectedLoader{checkedFile: checkedFile}
}

func init() {
	RegisterProtectedLoader("embedded", NewEmbeddedHashProtectedLoader())
}
