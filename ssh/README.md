```
		auth := &ssh.AuthInfo{
			Host: host,
			Port: port,
			User: user,
			Pass: pass,
		}

		dec, err := auth.Decryption(false)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		conn, err := auth.Connection(dec)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer conn.Close()

		err1 := auth.SCP(conn, s, "/tmp/init.sh")
		if err1 != nil {
			fmt.Println(err1)
			os.Exit(1)
		}

		cmd := []string{
			"sudo bash -o errexit -o nounset -o pipefail /tmp/init.sh" + " " + user + " " + newp + " " + name,
			"sudo shutdown -r 0",
		}
		out, err := auth.SSH(conn, cmd)
		if err != nil {
			fmt.Print(err, "\n", out[0])
			os.Exit(1)
		}
```
