package main

import (
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"
)

func main() {
	// 主机
	host := "192.168.100.3"

	// 端口
	port := "22"

	// 用户
	user := "user"

	// 密码
	pass := "user"

	// 命令
	cmdd := "sudo ip a"

	// 鉴权方式
	method := ssh.Password(pass)
	auths := []ssh.AuthMethod{method}

	// 执行连接
	config := ssh.ClientConfig{
		User: user,
		Auth: auths,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// 建立握手
	addr := host + ":" + port
	Client, _ := ssh.Dial("tcp", addr, &config)
	defer Client.Close()

	// 创建连接
	session, _ := Client.NewSession()
	defer session.Close()

	// 	执行命令
	buf, _ := session.CombinedOutput(cmdd)

	// 返回结果
	fmt.Print(string(buf))
}
