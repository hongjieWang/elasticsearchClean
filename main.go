package main

import (
	"encoding/json"
	"estools/tools"
	"fmt"
	"io/ioutil"
	"math"
	"strings"
	"time"

)

type Config struct {
	Data []tools.Data  `json:"data"`
}
var config Config

func Init() {
	JsonParse := NewJsonStruct()
	//下面使用的是相对路径，config.json文件和main.go文件处于同一目录下
	JsonParse.Load("config/config.json", &config)
}

type JsonStruct struct {
}

func NewJsonStruct() *JsonStruct {
	return &JsonStruct{}
}

func (jst *JsonStruct) Load(filename string, v interface{}) {
	//ReadFile函数会读取文件的全部内容，并将结果以[]byte类型返回
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	//读取的数据为json格式，需要进行解码
	err = json.Unmarshal(data, v)
	if err != nil {
		return
	}
}

func delIndex(data tools.Data) {
	currentTime := time.Now()

    stimptime := ""
	for i := 0; i < int(math.Abs(float64(data.Day))) ; i++ {
		oldTime := currentTime.AddDate(0, 0, data.Day + i)
		format := oldTime.Format(data.IndexFmt)
		stimptime = fmt.Sprintf("%s %s", stimptime, format)
	}

	index := tools.GetIndex(data)
	for k := range index {
		split := strings.Split(k, "-")
		fmt.Println("key:", k, "是否在里边:", stimptime)
		if len(split) > 1 {
			if find := strings.Contains(stimptime,split[1]); find {
				fmt.Println("-----0--")
				b := tools.DelIndex(data, k)
				fmt.Println(b)
			}
		}
	}
}

func main() {
	Init()
	for i, datum := range config.Data {
		fmt.Printf("config data Host is [%s], fmt is [%s]\n", datum.Host, datum.IndexFmt)
		println(i)
		delIndex(datum)
	}
}
