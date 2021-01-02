// +build !android

package conf

const bootstrapDNS = ""

func BootstrapDNS() bool {
	return false
}
