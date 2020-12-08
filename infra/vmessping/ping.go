package vmessping

import (
	"fmt"
	"os"
	"time"

	"v2ray.com/core/infra/conf"
	"v2ray.com/core/infra/miniv2ray"
)

// PrintVersion prints version info of vmessping
func PrintVersion() {
	fmt.Fprintf(os.Stderr,
		"Vmessping, A prober for v2ray (v2ray-core: %s)\n", miniv2ray.CoreVersion())
}

// Ping pings dest via outbound
func Ping(outbound *conf.OutboundDetourConfig, count uint, dest string, timeoutsec, inteval, quit uint, stopCh <-chan os.Signal, showNode, verbose bool) (*PingStat, error) {
	server, err := miniv2ray.StartV2Ray(outbound, verbose)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	if err := server.Start(); err != nil {
		fmt.Println("Failed to start", err)
		return nil, err
	}
	defer server.Close()

	if showNode {
		go func() {
			info, err := miniv2ray.GetNodeInfo(server, time.Second*10)
			if err != nil {
				return
			}

			fmt.Printf("Node Outbound: %s/%s\n", info["loc"], info["ip"])
		}()
	}

	ps := &PingStat{}
	ps.StartTime = time.Now()
	round := count
L:
	for round > 0 {
		seq := count - round + 1
		ps.ReqCounter++

		chDelay := make(chan int64)
		go func() {
			delay, err := miniv2ray.MeasureDelay(server, time.Second*time.Duration(timeoutsec), dest)
			if err != nil {
				ps.ErrCounter++
				fmt.Printf("Ping %s: seq=%d err %v\n", dest, seq, err)
			}
			chDelay <- delay
		}()

		select {
		case delay := <-chDelay:
			if delay > 0 {
				ps.Delays = append(ps.Delays, delay)
				fmt.Printf("Ping %s: seq=%d time=%d ms\n", dest, seq, delay)
			}
		case <-stopCh:
			break L
		}

		if quit > 0 && ps.ErrCounter >= quit {
			break
		}

		if round--; round > 0 {
			select {
			case <-time.After(time.Second * time.Duration(inteval)):
				continue
			case <-stopCh:
				break L
			}
		}
	}

	ps.CalStats()
	return ps, nil
}
