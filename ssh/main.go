package liuq

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"

	"golang.org/x/crypto/ssh"
)

// AuthInfo 基本验证信息
// key 及 pass 可选其中之一，也可全选
type AuthInfo struct {
	Host string
	Port string
	User string
	Key  string
	Pass string
}

// Decryption 必须首先解密
func (a AuthInfo) Decryption(c bool) ([]ssh.AuthMethod, error) {

	var auths []ssh.AuthMethod
	if c {

		// 读取密钥
		pemBytes, err := ioutil.ReadFile(a.Key)
		if err != nil {
			return nil, err
		}

		// 解析密钥
		signer, err := ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(a.Pass))
		if err != nil {
			return nil, err
		}
		method := ssh.PublicKeys(signer)

		// 鉴权方式
		sshs := []ssh.AuthMethod{}
		auths = append(sshs, method)
	} else {

		// 鉴权方式
		auths = []ssh.AuthMethod{
			ssh.Password(a.Pass),
		}
	}

	return auths, nil
}

// Connection 创建一个连接
func (a AuthInfo) Connection(dec []ssh.AuthMethod) (*ssh.Client, error) {

	// 准备连接
	config := &ssh.ClientConfig{
		User: a.User,
		Auth: dec,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// 建立握手
	h := net.JoinHostPort(a.Host, a.Port)
	client, err := ssh.Dial("tcp", h, config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// SSH ssh 执行命令
func (a AuthInfo) SSH(conn *ssh.Client, cmd []string) ([]string, error) {

	// defer conn.Close()

	var out []string

	for _, k := range cmd {

		// 创建会话
		session, err := conn.NewSession()
		if err != nil {
			return nil, err
		}

		// 执行命令
		buf, err := session.CombinedOutput(k)
		if err != nil {
			var out = []string{string(buf)}
			return out, err
		}

		out = append(out, string(buf))
	}

	// 返回结果
	return out, nil
}

// File 读取上传的文件
func (a AuthInfo) File(source string) ([]byte, error) {

	// 打开文件
	file, err := os.Open(source)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 获得字节
	contentsBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return contentsBytes, nil
}

// SCP scp 上传文件
func (a AuthInfo) SCP(conn *ssh.Client, s []byte, d string) error {

	// defer conn.Close()

	// 创建会话
	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	bytesReader := bytes.NewReader(s)
	// 获得路径及文件名
	remoteDir, remoteFile := path.Split(d)

	go func() {

		// 创建管道
		w, _ := session.StdinPipe()
		defer w.Close()

		// 写入管道 目录umask码 长度 目录
		// fmt.Fprintln(w, "D0755", 0, "") // mkdir

		// 写入管道 文件umask码 长度 文件
		fmt.Fprintln(w, "C0644", len(s), remoteFile)
		io.Copy(w, bytesReader)
		fmt.Fprint(w, "\x00") // 移除以 \x00 结尾
	}()
	err1 := session.Run("/usr/bin/scp -tr " + remoteDir)
	if err1 != nil {
		return err1
	}

	return nil
}
