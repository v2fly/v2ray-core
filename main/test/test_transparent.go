package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
)

func main() {
	// 代理服务器地址和端口
	proxyAddr := "106.75.239.178:3456"
	// 目标服务器地址

	// 通过 TCP 连接到代理服务器
	conn, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		fmt.Printf("Failed to connect to proxy server: %v\n", err)
		return
	}
	defer conn.Close()

	// 构建 HTTP 请求
	req, err := http.NewRequest("GET", "https://ifconfig.me", nil)
	if err != nil {
		fmt.Printf("Failed to create HTTP request: %v\n", err)
		return
	}
	req.Header.Set("Host", "ifconfig.me")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	// 发送 HTTP 请求到目标服务器
	err = req.Write(conn)
	if err != nil {
		fmt.Printf("Failed to send HTTP request: %v\n", err)
		return
	}

	// 读取服务器响应
	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		fmt.Printf("Failed to read HTTP response: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应体内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body: %v\n", err)
		return
	}

	// 输出响应内容
	fmt.Println(string(body))
}
