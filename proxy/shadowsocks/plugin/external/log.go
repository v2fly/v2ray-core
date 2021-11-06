package external

import "io"

var _ io.Writer = (*pluginOutWriter)(nil)

type pluginOutWriter struct {
	name string
}

func (w *pluginOutWriter) Write(p []byte) (n int, err error) {
	newError(w.name, "-stdout: ", string(p)).AtInfo().WriteToLog()
	return len(p), nil
}

var _ io.Writer = (*pluginErrWriter)(nil)

type pluginErrWriter struct {
	name string
}

func (w *pluginErrWriter) Write(p []byte) (n int, err error) {
	newError(w.name, "-stderr: ", string(p)).WriteToLog()
	return len(p), nil
}
