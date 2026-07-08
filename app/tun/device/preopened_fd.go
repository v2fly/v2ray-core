package device

import "os"

func ClosePreopenedFD(fd int) error {
	if fd < 0 {
		return nil
	}
	file := os.NewFile(uintptr(fd), "preopened-tun")
	if file == nil {
		return nil
	}
	return file.Close()
}
