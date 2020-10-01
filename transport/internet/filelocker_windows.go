package internet

func (fl *FileLocker) Acquire() error {
	return nil
}

func (fl *FileLocker) Release() {
	return
}
