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
	"strings"

	"golang.org/x/crypto/ssh"
)

type authInfo struct {
	host string // 主机
	port int    // 端口
	user string // 用户名
	key  string // 密钥文件路径
	pass string // 解密密钥的密码
}

func (a authInfo) dec(c bool) []ssh.AuthMethod {

	var auths []ssh.AuthMethod
	if c {

		// 读取密钥
		pemBytes, err := ioutil.ReadFile(a.key)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// 解析密钥
		signer, err := ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(a.pass))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		method := ssh.PublicKeys(signer)

		// 鉴权方式
		sshs := []ssh.AuthMethod{}
		auths = append(sshs, method)
	} else {

		// 鉴权方式
		auths = []ssh.AuthMethod{ssh.Password(a.pass)}
	}

	return auths
}

func (a authInfo) conn(dec []ssh.AuthMethod) (*ssh.Session, error) {

	// 执行连接
	config := &ssh.ClientConfig{
		User: a.user,
		Auth: dec,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// 建立握手
	client, err := ssh.Dial("tcp", a.host, config)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// 创建连接
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	return session, nil
}

func (a authInfo) ssh(s *ssh.Session, cmd string) (string, error) {

	// 执行命令
	buf, err := s.CombinedOutput(cmd)
	if err != nil {
		return "", err
	}

	// 返回结果
	return string(buf), nil
}

func (a authInfo) scp(s *ssh.Session, f string) error {

	// 打开文件
	fileFilter := strings.Replace(f, ":", " ", -1)
	files := strings.Fields(fileFilter)

	file, err := os.Open(files[0])
	if err != nil {
		return err
	}
	defer file.Close()

	// 获得字节
	contentsBytes, _ := ioutil.ReadAll(file)
	bytesReader := bytes.NewReader(contentsBytes)

	// 获得路径及文件名
	remoteDir, remoteFile := path.Split(files[1])

	go func() {

		// 创建管道
		w, _ := s.StdinPipe()
		defer w.Close()

		// 写入管道 目录umask码 长度 目录
		// fmt.Fprintln(w, "D0755", 0, "") // mkdir

		// 写入管道 文件umask码 长度 文件
		fmt.Fprintln(w, "C0644", len(contentsBytes), remoteFile)
		io.Copy(w, bytesReader)
		fmt.Fprint(w, "\x00") // 移除以 \x00 结尾
	}()
	err1 := s.Run("/usr/bin/scp -tr " + remoteDir)
	if err1 != nil {
		return err1
	}

	return nil
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

func main() {

	// 验证信息
	auth := &authInfo{
		host: "192.168.1.1",
		port: 22,
		user: "root",
		key:  "/home/user/.ssh/releases",
		pass: "passwd",
	}

	// ssh 密钥执行命令
	session1, err1 := auth.conn(auth.dec(true))
	sshKeyOut, err2 := auth.ssh(session1, "who")
	ce(err1, err2)
	fmt.Println(sshKeyOut)

	// ssh 密码执行命令
	session2, err3 := auth.conn(auth.dec(false))
	sshPassOut, err4 := auth.ssh(session2, "who")
	ce(err3, err4)
	fmt.Println(sshPassOut)

	// scp 密钥执行命令
	session3, err5 := auth.conn(auth.dec(true))
	ce(err5)
	err6 := auth.scp(session3, "/etc/hosts:/tmp/hosts")
	if err6 != nil {
		fmt.Println("ok")
	}

	// scp 密码执行命令
	session4, err7 := auth.conn(auth.dec(false))
	ce(err7)
	err8 := auth.scp(session4, "/etc/hosts:/tmp/hosts")
	if err8 != nil {
		fmt.Println("ok")
	}
}
