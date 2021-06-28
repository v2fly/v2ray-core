package main

import (
	"github.com/v2fly/v2ray-core/v4/main/v2binding"
	"time"
)

func main() {
	v2binding.StartAPIInstance(".")
	for {
		time.Sleep(time.Minute)
	}
}
