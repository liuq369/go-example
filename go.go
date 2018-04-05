package main

import (
	"log"
)

// 参考 https://gobyexample.com/signals
func main() {

	// 定义并初始化一个channel
	do := make(chan bool)

	// 并发顺序执行里面的函数，执行完成发送信号，main 函数在此阻塞
	go func() {
		log.Println("")
		do <- true
	}()

	// 收到完成信号，main 函数在此结束
	<-do
}
