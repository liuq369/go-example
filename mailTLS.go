package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strings"
)

func main() {

	commands := os.Args
	if len(commands) != 7 {
		fmt.Println("[1]:用户名@...", "[2]:密码", "[3]:SMTP服务器:465", "[4]:收件人_收件人", "[5]:邮件主题", "[6]:邮件正文")
		os.Exit(1)
	}

	info := &infoMta{
		smtpUser:   commands[1],
		smtpPass:   commands[2],
		smtpServer: commands[3],
		to:         commands[4],
		mailHead:   commands[5],
		mailBody:   commands[6],
	}

	for i := 0; i < 6; i++ {
		if smtpTLSMta(info) == nil {
			fmt.Println("OK", i)
			os.Exit(0)
		}
	}
	fmt.Println("NO")
	os.Exit(1)
}

type infoMta struct {
	smtpUser   string
	smtpPass   string
	smtpServer string
	to         string
	mailHead   string
	mailBody   string
}

// SSL/TLS Email Example
func smtpTLSMta(info *infoMta) error {

	str := strings.Replace(info.to, "_", ";", -1)
	sendTo := strings.Split(str, ";")

	message := []byte("Content-Type: text/plain; charset=UTF-8 " +
		"\r\nFrom: " + info.smtpUser +
		"\r\nTo: " + str +
		"\r\nSubject: " + info.mailHead +
		"\r\n\r\n" + info.mailBody)

	onlyHost, _, err := net.SplitHostPort(info.smtpServer)
	if err != nil {
		return err
	}

	auth := smtp.PlainAuth("",
		info.smtpUser,
		info.smtpPass,
		onlyHost,
	)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         onlyHost,
	}

	// 这里是关键，你需要调用tls.Dial，而不是smtp.Dial在465上运行的smtp服务器，从一开始就需要一个ssl连接（没有starttls）
	conn, err := tls.Dial("tcp", info.smtpServer, tlsconfig)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, onlyHost)
	if err != nil {
		return err
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		return err
	}

	// To && From
	if err = c.Mail(info.smtpUser); err != nil {
		return err
	}

	// 每次只能调用 Rcpt 一个参数，但是可以调用多次 Rcpt 函数
	r := len(sendTo)
	for i := 0; i < r; i++ {
		if err = c.Rcpt(sendTo[i]); err != nil {
			return err
		}
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(message)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}
