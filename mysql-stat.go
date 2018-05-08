package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type josnOut struct {
	Throughput  map[string]int
	Performance map[string]int
	Threads     map[string]int
	Buffer      map[string]int
}

type authInfo struct {
	host string
	port string
	user string
	pass string
}

func average(s1, s2, v map[string]int) ([]byte, error) {

	throughput := make(map[string]int)
	performance := make(map[string]int)
	threads := make(map[string]int)
	buffer := make(map[string]int)

	throughput["qps/s"] = s2["Com_select"] - s1["Com_select"]
	throughput["tps/1"] = s2["Com_insert"] + s2["Com_delete"] + s2["Com_update"] - (s1["Com_insert"] + s1["Com_delete"] + s1["Com_update"])

	performance["slow-time"] = v["long_query_time"]
	performance["slow/s"] = s2["Slow_queries"] - s1["Slow_queries"]

	threads["used"] = s2["Threads_connected"] - 1
	threads["run"] = s2["Threads_running"] - 1
	threads["max"] = v["max_connections"]
	threads["exit/s"] = s2["Aborted_connects"] - s1["Aborted_connects"]

	buffer["used/%"] = s2["Innodb_buffer_pool_pages_data"] * 100 / s2["Innodb_buffer_pool_pages_total"]
	buffer["used/s"] = s2["Innodb_buffer_pool_pages_data"] - s1["Innodb_buffer_pool_pages_data"]
	buffer["hit/s"] = s2["Innodb_buffer_pool_read_requests"] - s1["Innodb_buffer_pool_read_requests"]
	buffer["missed/s"] = s2["Innodb_buffer_pool_reads"] - s1["Innodb_buffer_pool_reads"]

	out, err := json.MarshalIndent(josnOut{
		Throughput:  throughput,
		Performance: performance,
		Threads:     threads,
		Buffer:      buffer,
	}, "", "    ")
	return out, err
}

// 获取
func (a authInfo) rows() (map[string]int, map[string]int, map[string]int, error) {

	// 连接
	db, err := sql.Open("mysql", a.user+":"+a.pass+"@tcp("+a.host+":"+a.port+")/?charset=utf8&timeout=2s")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	// 查询
	rows0, err0 := db.Query("SHOW GLOBAL STATUS")
	time.Sleep(1 * time.Second)
	rows1, err1 := db.Query("SHOW GLOBAL STATUS")
	rows2, err2 := db.Query("SHOW VARIABLES")
	if err0 != nil || err1 != nil || err2 != nil {
		fmt.Println(err)
	}

	// 获取列名
	columns0, err0 := rows0.Columns()
	columns1, err1 := rows0.Columns()
	columns2, err2 := rows0.Columns()
	if err0 != nil || err1 != nil || err2 != nil {
		fmt.Println(err)
	}

	// 分割数据
	values0 := make([]sql.RawBytes, len(columns0))
	values1 := make([]sql.RawBytes, len(columns1))
	values2 := make([]sql.RawBytes, len(columns2))
	scanArgs0 := make([]interface{}, len(values0))
	scanArgs1 := make([]interface{}, len(values1))
	scanArgs2 := make([]interface{}, len(values2))
	for i := range values0 {
		scanArgs0[i] = &values0[i]
	}
	for j := range values1 {
		scanArgs1[j] = &values1[j]
	}
	for k := range values2 {
		scanArgs2[k] = &values2[k]
	}

	var onlyInfo = [12]string{
		"Com_select", // select 语句统计数量
		"Com_insert", // insert 语句统计数量
		"Com_delete", // delete 语句统计数量
		"Com_update", // update 语句统计数量

		"Slow_queries", // 慢查询语句统计数量

		"Threads_connected", // 已连接线程当前数量
		"Threads_running",   // 运行中线程当前数量
		"Aborted_connects",  // 连接失败统计数量

		"Innodb_buffer_pool_pages_total",   // 缓冲池总大小
		"Innodb_buffer_pool_pages_data",    // 缓冲池使用情况
		"Innodb_buffer_pool_read_requests", // 缓冲命中成功统计次数
		"Innodb_buffer_pool_reads",         // 缓冲命中失败统计次数
	}
	var onlyInfo2 = [2]string{
		"long_query_time", // 慢查询触发时间
		"max_connections", // 最大连接数
	}

	// 提取行
	valu0 := make(map[string]int)
	valu1 := make(map[string]int)
	valu2 := make(map[string]int)
	for rows0.Next() {
		rows0.Scan(scanArgs0...)
		for _, v := range onlyInfo {
			if string(values0[0]) == v {
				i, _ := strconv.ParseFloat(string(values0[1]), 32)
				valu0[string(values0[0])] = int(i)
			}
		}
	}
	for rows1.Next() {
		rows1.Scan(scanArgs1...)
		for _, v := range onlyInfo {
			if string(values1[0]) == v {
				i, _ := strconv.ParseFloat(string(values1[1]), 32)
				valu1[string(values1[0])] = int(i)
			}
		}
	}

	for rows2.Next() {
		rows2.Scan(scanArgs2...)
		for _, v := range onlyInfo2 {
			if string(values2[0]) == v {
				i, _ := strconv.ParseFloat(string(values2[1]), 32)
				valu2[string(values2[0])] = int(i)
			}
		}
	}

	return valu0, valu1, valu2, nil
}

func main() {

	auth := &authInfo{
		host: "192.168.100.10",
		port: "3306",
		user: "root",
		pass: "Mama3860!",
	}

	s1, s2, v, err := auth.rows()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	out, err := average(s1, s2, v)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(string(out))
}
