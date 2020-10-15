// +build !confonly

package admin

import "v2ray.com/core/common/log"

type writer struct{
}
func (*writer) Write(p []byte) (n int, err error) {
	log.Info("%v", string(p[:len(p)-1]))
	return len(p),nil
}
var ErrorLoggerWriter = &writer{}