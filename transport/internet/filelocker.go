package internet

import (
	"os"
)

type FileLocker struct {
	path string
	file *os.File
}
