package main

import (
	"bufio"
	"log"
	"os"
)

func main() {
	file, err := os.OpenFile(output.lof, os.O_WRONLY|os.O_TRUNC|os.O_EXCL|os.O_CREATE, 0600)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	outputWriter := bufio.NewWriter(file)

	for i := 0; i < 2; i++ {
		outputWriter.WriteString("hello world!\n")
	}
	outputWriter.Flush()
}
