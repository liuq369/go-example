package main

import (
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strings"
)

//SendMail 发送邮件的逻辑函数
func main() {

	cmd := os.Args
	if len(cmd) == 4 {

		// 重复 99 次，直到成功
		for i := 0; i < 6; i++ {
			err := faMailInit("admin@126.com", "passwd", "smtp.126.com", "25", cmd[1], cmd[2], cmd[3])
			if err == nil {
				fmt.Println("OK", i)
				os.Exit(0)
			}
		}
		fmt.Println("NO")
		os.Exit(1)
	} else {
		fmt.Println("'[1]:收件人;收件人'", "'[2]:邮件主题'", "'[3]:邮件正文'")
		os.Exit(1)
	}
}

func faMailInit(user, pass, host, port, to, head, body string) error {

	// 邮件主体
	emty := "Content-Type: text/plain; charset=UTF-8"
	msg := []byte("To: " + to + "\r\nFrom: " + user + "<" + user + ">\r\nSubject: " + head + "\r\n" + emty + "\r\n\r\n" + body)

	// TLS加密的用户密码验证
	auth := smtp.PlainAuth("", user, pass, host)

	// 发送邮件
	sendTo := strings.Split(to, ";")
	err := smtp.SendMail(net.JoinHostPort(host, port), auth, user, sendTo, msg)

	return err
}
