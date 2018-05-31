package liuq

import (
	"bufio"
	"os"
	"strings"
)

// InputInfo 基本信息
type InputInfo struct {
	File string
	List [][]string
	Key  []string
	Valu map[string][]string
}

// GetList 获得数据
func (i *InputInfo) get() error {

	// 打开文件
	file, err := os.Open(i.File)
	if err != nil {
		return err
	}
	defer file.Close()

	// 琢行扫描
	var outList [][]string
	confScanner := bufio.NewScanner(file)
	for confScanner.Scan() {
		line := strings.SplitN(confScanner.Text(), "#", 2)[0]
		fields := strings.Fields(line)
		if len(fields) != 0 {
			// fmt.Println(fields[0], "11111", fields[1], "22222", fields[2], "333333", fields[3])
			outList = append(outList, fields)
		}
	}

	i.List = outList
	return nil
}

// GetListStruct 结构
func (i *InputInfo) GetListStruct() error {

	err := i.get()
	if err != nil {
		return err
	}

	var key []string
	valu := make(map[string][]string)
	for _, v := range i.List {
		key = append(key, v[0])
		var tmp []string

		le := len(v)
		for i := 1; i < le; i++ {
			tmp = append(tmp, v[i])
		}
		valu[v[0]] = tmp
	}

	i.Key = key
	i.Valu = valu
	return nil
}
