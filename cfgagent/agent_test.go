package cfgagent

import (
	"fmt"
	"testing"
)

func Test_ListAndListenConfigsInNamespace(t *testing.T) {

	cfg := &ConfigClient{
		Username:    "nacos",
		Password:    "fucku@2025",
		ServerAddr:  "106.75.239.178",
		NamespaceID: "vps-pool-1",
		GroupID:     "test-vps",
		Number:      10,
	}
	if err := cfg.NewClient(); err != nil {
		fmt.Println(err)
	}

	updLog, err := cfg.LoadAllConfig()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(updLog)
	select {}
}
