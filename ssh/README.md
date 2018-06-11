```
package main

import (
	rssh "github.com/liuq369/go-example/ssh"
)

type auth struct {
	host, port, user, pass string
}

func main() {

	a := &auth{
		host: "192.168.100.10",
		port: "22",
		user: "root",
		pass: "Mama3860",
	}
	a.shell()

}

// 执行脚本
func (in auth) shell() error {

	auth := &rssh.AuthBasic{
		Host: in.host,
		Port: in.port,
		User: in.user,
		Pass: in.pass,
	}

	dec, err := auth.Decryption()
	if err != nil {
		return err
	}

	conn, err := auth.Connection(dec)
	if err != nil {
		return err
	}
	defer conn.Close()

	sess, err := auth.Session(conn)
	if err != nil {
		return err
	}
	defer sess.Close()

	return auth.SecureShellBash(sess)
}
```
