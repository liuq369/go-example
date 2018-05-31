```
package main

import (
	"fmt"

	list "github.com/liuq369/go-example/list"
)

func main() {

	l := &list.InputInfo{File: "./conf.list"}
	l.GetListStruct()

	for _, v := range l.Key {
		fmt.Println(v, l.Valu[v])
	}
}

```
