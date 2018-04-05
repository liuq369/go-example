package main

import (
	"io"
	"log"
	"net"
	"strconv"
)

func main() {

	// 用户名 密码
	var userPass map[string]string
	userPass = make(map[string]string)
	userPass["username"] = "password"

	for {

		// 接受等待并将下一个连接返回给侦听器
		client, err := listener.Accept()
		if err != nil {
			log.Println("Error:", err)
		}

		go handleAuthRequest(client, userPass)
	}
}

// 验证用户正确，就转发数据
// 参考 https://github.com/luSunn/rserver
func handleAuthRequest(c net.Conn, uP map[string]string) {
	// 适当时候关闭tcp连接
	defer c.Close()

	// 获取请求正文
	var b [1024]byte
	c.Read(b[:])

	// 只处理SOCKS5协议
	if b[0] == 0x05 {

		// 回应客户端验证
		c.Write([]byte{0x05, 0x02})

		// 获取请求正文
		c.Read(b[:])

		// 获取用户名密码
		var u, p string
		uLen := int(b[1])
		u = string(b[2 : uLen+2])
		pLen := int(b[uLen+2])
		p = string(b[uLen+3 : uLen+pLen+3])

		// 验证用户名密码
		pass, ok := uP[u]
		if ok == true {
			if p != pass {

				// 响应身份验证失败
				c.Write([]byte{0x01, 0x01})
			} else {

				// 响应身份验证成功
				c.Write([]byte{0x01, 0x00})

				// 获取请求正文
				n, err := c.Read(b[:])
				if err != nil {
					log.Println("Error:", err)
				}

				var host, port string
				switch b[3] {
				case 0x01: // ipv4
					host = net.IPv4(b[4], b[5], b[6], b[7]).String()
				case 0x03: // 域名
					host = string(b[5 : n-2]) // b[4]表示域名的长度
				case 0x04: // ipv6
					host = net.IP{b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15], b[16], b[17], b[18], b[19]}.String()
				}

				// 检查端口合法性
				port = strconv.Itoa(int(b[n-2])<<8 | int(b[n-1]))
				if port != "0" {

					// 响应远程服务器
					server, err := net.Dial("tcp", net.JoinHostPort(host, port))
					if nil != err {
						log.Println("Error:", err)
					}
					defer server.Close()

					// 响应客户端连接成功
					c.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

					// 开始转发数据
					go io.Copy(server, c)
					io.Copy(c, server)
				}
			}
		}
	}
}
