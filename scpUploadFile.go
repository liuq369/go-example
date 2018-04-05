// https://gist.github.com/jedy/3357393 and https://github.com/bramvdbogaerde/go-scp
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"

	"golang.org/x/crypto/ssh"
)

func main() {

	const (
		host = "192.168.100.3:24"                    // 主机
		user = "user"                                // 用户
		keyf = "/home/liuq369/.ssh/id_rsa"           // 私钥
		pass = "passswd"                             // 密码
		loFi = "/home/liuq369/.Me/go/src/test/test0" // 本地 "文件"
		reFi = "/home/user/test"                     // 远程 "文件"
	)

	// 读取密钥
	pemBytes, _ := ioutil.ReadFile(keyf)

	// 解析密钥
	signer, _ := ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(pass))
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
	clientt, _ := ssh.Dial("tcp", host, config)
	defer clientt.Close()

	// 创建连接
	session, _ := clientt.NewSession()
	defer session.Close()

	// 打开文件
	file, _ := os.Open(loFi)
	defer file.Close()

	// 获得字节
	contentsBytes, _ := ioutil.ReadAll(file)
	bytesReader := bytes.NewReader(contentsBytes)

	// 获得路径及文件名
	remoteDir, remoteFile := path.Split(reFi)

	go func() {

		// 创建管道
		w, _ := session.StdinPipe()
		defer w.Close()

		// 写入管道 目录umask码 长度 目录
		// fmt.Fprintln(w, "D0755", 0, "") // mkdir

		// 写入管道 文件umask码 长度 文件
		fmt.Fprintln(w, "C0644", len(contentsBytes), remoteFile)
		io.Copy(w, bytesReader)
		fmt.Fprint(w, "\x00") // 移除以 \x00 结尾
	}()
	ce(session.Run("/usr/bin/scp -tr " + remoteDir))
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
