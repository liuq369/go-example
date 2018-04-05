package example

import (
	"Zcontrol/lib"
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "Zcontrol"
	app.Usage = "Final control platform"
	app.Version = "20171028"
	app.Author = "Elijah"
	app.Email = "https://wiki.liuq.org"
	app.Commands = []cli.Command{
		{
			Name:  "http1",
			Usage: "Run HTTP Service to (http/1.1)(http://*:8080)",
			Action: func(c *cli.Context) error {
				zcontrol.Httpcheck("http1")
				return nil
			},
		},
		{
			Name:  "https2",
			Usage: "Run HTTPS Service to (http/2)(https://*:8080)",
			Action: func(c *cli.Context) error {
				zcontrol.Httpcheck("https2")
				return nil
			},
		},
		{
			Name:  "sftp",
			Usage: "Run SFTP Service to (SSL/TLS)(*:2121)",
			Action: func(c *cli.Context) error {
				zcontrol.Sftp()
				return nil
			},
		},
		{
			Name:  "proxy_http",
			Usage: "Run HTTP Proxy to (*:8080)",
			Action: func(c *cli.Context) error {
				zcontrol.Proxyhttp()
				return nil
			},
		},
		{
			Name:  "proxy_socks5",
			Usage: "Run SOCKS5 Proxy to (*:8080)",
			Action: func(c *cli.Context) error {
				zcontrol.Socks5()
				return nil
			},
		},
		{
			Name:  "port_proxy",
			Usage: "Run Port forwarding to (*:8081 --> *:8080)",
			Action: func(c *cli.Context) error {
				zcontrol.PortProxy(8081, 8080) // 监听本机端口，转发到目的端口
				return nil
			},
		},
		{
			Name:  "remote_ssh",
			Usage: "Run SSH remote command (config /etc/Zcontrol.conf)",
			Action: func(c *cli.Context) error {
				zcontrol.Remotessh("/etc/Zcontrol.conf") // 最大主机数 每台主机最大命令数 配置文件
				return nil
			},
		},
		{
			Name:    "test",
			Aliases: []string{"t"},
			Usage:   "Test features",
			Action: func(c *cli.Context) error {
				zcontrol.Test()
				return nil
			},
		},
		{
			Name:  "other",
			Usage: "options for task templates",
			Subcommands: []cli.Command{
				{
					Name:  "one",
					Usage: "add a new template",
					Action: func(c *cli.Context) error {
						fmt.Println("new task template: ", c.Args().First())
						return nil
					},
				},
				{
					Name:  "two",
					Usage: "remove an existing template",
					Action: func(c *cli.Context) error {
						fmt.Println("removed task template: ", c.Args().First())
						return nil
					},
				},
			},
		},
	}
	app.Run(os.Args)
}
