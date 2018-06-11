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
	"golang.org/x/crypto/ssh/terminal"
)

// Auth 基本验证信息
type Auth struct {
	Host string // 远程主机IP
	Port string // 远程主机端口
	User string // 远程主机用户名
	Key  string // 本地密钥路径
	Pass string // 验证密码
}

// Decryption 必须首先解密
func (in Auth) Decryption() ([]ssh.AuthMethod, error) {

	// 鉴权方式
	var auths []ssh.AuthMethod
	if len(in.Key) == 0 {

		// 鉴权方式
		auths = []ssh.AuthMethod{
			ssh.Password(in.Pass),
		}
	} else {

		// 读取密钥
		pemBytes, err := ioutil.ReadFile(in.Key)
		if err != nil {
			return nil, err
		}

		// 解析密钥
		signer, err := ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(in.Pass))
		if err != nil {
			return nil, err
		}
		method := ssh.PublicKeys(signer)

		// 鉴权方式
		sshs := []ssh.AuthMethod{}
		auths = append(sshs, method)
	}

	// 鉴权完毕
	return auths, nil
}

// Connection 创建一个连接
func (in Auth) Connection(dec []ssh.AuthMethod) (*ssh.Client, error) {

	// 连接信息
	config := &ssh.ClientConfig{
		User: in.User,
		Auth: dec,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// 建立握手
	connect, err := ssh.Dial("tcp", net.JoinHostPort(in.Host, in.Port), config)
	if err != nil {
		return nil, err
	}

	// 握手完毕
	return connect, nil
}

// SecureShellCommand 执行非交互式命令
func (in Auth) SecureShellCommand(conn *ssh.Client, cmd []string) ([][]byte, error) {

	var out [][]byte
	for _, v := range cmd {

		session, err := conn.NewSession()
		if err != nil {
			return nil, err
		}
		defer session.Close()

		// 执行命令
		buf, err := session.CombinedOutput(v)
		if err != nil {
			return nil, err
		}

		out = append(out, bytes.TrimRight(buf, "\n"))
	}

	// 返回命令输出
	return out, nil
}

// SecureShellBash ssh 执行交互式shell
func (in Auth) SecureShellBash(conn *ssh.Client) error {

	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// 创建文件描述符
	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer terminal.Restore(fd, oldState)

	// 获取窗口宽高
	width, height, err := terminal.GetSize(fd)
	if err != nil {
		return err
	}

	// 配置窗口宽高
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	// 交互式shell准备读写操作
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	// 创建终端
	if err := session.RequestPty("xterm-256color", height, width, modes); err != nil {
		return err
	}

	// 获取shell
	if err := session.Shell(); err != nil {
		return err
	}

	// 等待远程命令退出
	return session.Wait()
}

// SecureCopyWrite scp 写覆盖文件
func (in Auth) SecureCopyWrite(conn *ssh.Client, s []byte, d string) error {

	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// 读取源内容字节
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

	// 写入文件到
	return session.Run("/usr/bin/scp -tr " + remoteDir)
}

// SecureCopyFile scp 上传文件
func (in Auth) SecureCopyFile(conn *ssh.Client, s string, d string) error {

	file, err := ioutil.ReadFile(s)
	if err != nil {
		return err
	}

	return in.SecureCopyWrite(conn, file, d)
}
