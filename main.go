package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

type Config struct {
	Data []Data `json:"data"`
}

type Data struct {
	Host     string `json:"host"`
	IndexFmt string `json:"index_fmt"`
	Day      int    `json:"day"`
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

func delIndex(data Data) {
	currentTime := time.Now()
	oldTime := currentTime.AddDate(0, 0, data.Day)
	format := oldTime.Format(data.IndexFmt)
	index := getIndex(data)
	for k := range index {
		fmt.Println("key:", k, "format:", format)
		if find := strings.Contains(k, format); find {
			DelIndex(data, k)
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
