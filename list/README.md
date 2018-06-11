package main

func main()  {

    list := &liuq.InputInfo{File: "./conf.list"}
    list.GetListStruct()

    for _, v := range list.Key {
        fmt.Println(v, list.Valu[v])
    }
}
