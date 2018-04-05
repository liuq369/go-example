package main

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql" // 导入数据库驱动模块
)

func exitInfo() {
	fmt.Println("[keepPing] [portPing] [httpGet] [mysqlPing] [mysqlSync] [memcacheGet] [redisGet]")
	os.Exit(0)
}

func main() {
	// err3 := mysqlPing("192.168.100.128:3306", "status", "Q3iYL3Huj67AdyNaF|O+")
	command := os.Args
	if len(command) >= 3 && len(command) <= 5 {
		switch {
		case command[1] == "httpGet":
			err := httpGet(command[2])
			if err != nil {
				os.Exit(1)
			} else {
				os.Exit(0)
			}

		case command[1] == "portPing":
			err := portPing(command[2])
			if err != nil {
				os.Exit(1)
			} else {
				os.Exit(0)
			}

		case command[1] == "memcacheGet":
			err := memcacheGet(command[2])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else {
				os.Exit(0)
			}

		case command[1] == "redisGet":
			err := redisGet(command[2])
			if err != nil {
				os.Exit(1)
			} else {
				os.Exit(0)
			}

		case command[1] == "keepPing":

			err := keepPing(command[2])
			if err != nil {
				os.Exit(1)
			} else {
				os.Exit(0)
			}

		case command[1] == "mysqlPing":
			if len(command) == 5 {
				err := mysqlPing(command[2], command[3], command[4])
				if err != nil {
					os.Exit(1)
				} else {
					os.Exit(0)
				}
			} else {
				exitInfo()
			}

		case command[1] == "mysqlSync":
			if len(command) == 5 {
				err := mysqlSync(command[2], command[3], command[4])
				if err != nil {
					os.Exit(1)
				} else {
					os.Exit(0)
				}
			} else {
				exitInfo()
			}

		default:
			exitInfo()
		}
	} else {
		exitInfo()
	}
}

// linux ping 主机检测
func keepPing(h string) error {

	out, err := exec.Command("ping", h, "-4", "-c 2", "-w 2").Output()
	if err != nil {
		return err
	}
	if strings.Contains(string(out), "rtt min/avg/max/mdev") {
		// fmt.Println("YES")
		return nil
	} else {
		// fmt.Println("NO")
		var err error = errors.New("destination host unreachable")
		return err
	}
}

// 检查端口是否开放
func portPing(hp string) error {

	h, _, _ := net.SplitHostPort(hp)
	err := keepPing(h)
	if err != nil {
		return err
	}

	_, err1 := net.Dial("tcp", hp)
	return err1
}

// 用 http 协议 get 连接 url 地址，获取状态码
// 状态码不是 200/302 那就是失败
func httpGet(hp string) error {

	err := portPing(hp)
	if err != nil {
		return err
	}

	resp, err := http.Get("http://" + hp)
	if err != nil {
		var err error = errors.New("get url is error")
		return err
	}
	stat := resp.StatusCode
	if stat == 200 || stat == 302 {
		return err
	} else {
		var err error = errors.New("get status is error")
		return err
	}
}

// 返回 是否存活
// sql 用户至少需要这些权限 "SUPER, REPLICATION CLIENT"
func mysqlPing(hp, u, p string) error {

	h, _, _ := net.SplitHostPort(hp)
	err := keepPing(h)
	if err != nil {
		return err
	}

	// 连接数据库
	db, err := sql.Open("mysql", u+":"+p+"@tcp("+hp+")/?charset=utf8&timeout=2s")
	defer db.Close()
	if err != nil {
		return err
	}

	// 查询语句
	_, err1 := db.Query("SHOW STATUS")
	return err1
}

// 返回 同步是否成功
// sql 用户至少需要这些权限 "SUPER, REPLICATION CLIENT"
func mysqlSync(hp, u, p string) error {

	h, _, _ := net.SplitHostPort(hp)
	err := keepPing(h)
	if err != nil {
		return err
	}

	// 连接数据库
	db, err := sql.Open("mysql", u+":"+p+"@tcp("+hp+")/?charset=utf8&timeout=2s")
	defer db.Close()
	if err != nil {
		return err
	}

	// 查询语句
	rows, err := db.Query("SHOW SLAVE STATUS")
	if err != nil {
		return err
	}

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// 分割数据
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var onek, twok string

	// 提取行
	for rows.Next() {

		// 从数据获取 RawBytes
		rows.Scan(scanArgs...)

		// 现在用数据做一些事情：字段名 值
		onek = string(values[10])
		twok = string(values[11])
	}

	var errNil error
	if onek == "Yes" && twok == "Yes" {
		return errNil
	} else {
		var err error = errors.New("double master synchronization failed")
		return err
	}
}

// memcache set/get 检测
func memcacheGet(hp string) error {
	name := randStr(12)
	mc := memcache.New(hp)
	mc.Set(&memcache.Item{
		Key:        name,
		Value:      []byte(name),
		Expiration: 1,
	})

	_, err := mc.Get(name)
	return err
}

// redis set/get 检测
func redisGet(hp string) error {
	client := redis.NewClient(&redis.Options{
		Addr:     hp,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		return err
	}
	name := randStr(12)
	err1 := client.Set(name, name, time.Second).Err()
	if err1 != nil {
		return err1
	}

	_, err2 := client.Get(name).Result()
	if err2 != nil {
		return err2
	} else {
		return nil
	}
}

// 生成随机字符串
// 字符串长度 大小写英文
func randStr(strlen int) string {
	rand.Seed(time.Now().Unix())
	data := make([]byte, strlen)
	var num int
	for i := 0; i < strlen; i++ {
		num = rand.Intn(57) + 65
		for {
			if num > 90 && num < 97 {
				num = rand.Intn(57) + 65
			} else {
				break
			}
		}
		data[i] = byte(num)
	}
	return string(data)
}
