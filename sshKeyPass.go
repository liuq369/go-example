// https://github.com/islenbo/autossh
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
)

func main() {

	const (
		host = "192.168.100.3:24"          // 主机
		user = "user"                      // 用户
		keyf = "/home/liuq369/.ssh/id_rsa" // 私钥
		pass = "passwd"                    // 密码
		cmdd = "w"                         // 命令
	)

	// 读取密钥
	pemBytes, err := ioutil.ReadFile(keyf)
	ce(err)

	// 解析密钥
	signer, err := ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(pass))
	ce(err)
	method := ssh.PublicKeys(signer)

	// 鉴权方式
	auths := append([]ssh.AuthMethod{}, method)

	// 执行连接
	config := &ssh.ClientConfig{
		User: user,
		Auth: auths,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// 建立握手
	client, err := ssh.Dial("tcp", host, config)
	ce(err)
	defer client.Close()

	// 创建连接
	session, err := client.NewSession()
	ce(err)
	defer session.Close()

	// 执行命令
	buf, err := session.CombinedOutput(cmdd)
	ce(err)

	// 返回结果
	fmt.Print(string(buf))
}

// 错误处理
func ce(errs ...error) {

	l := len(errs)
	for i := 0; i < l; i++ {
		err := errs[i]
		if err != nil {
			log.Println("Error:", err)
		}
	}

	if errs[l-1] != nil {
		os.Exit(1)
	}
}
