# Go语言Elasticsearch数据清理工具

微服务架构中收集通常大家都采用ELK进行日志收集，同时我们还采用了SkyWalking进行链路跟踪，而SkyWalking数据存储也用到了ES，SkyWalking每天产生大量的索引数据，如下：

![WX20211008-104751@2x](https://img-1258527903.cos.ap-beijing.myqcloud.com/img/WX20211008-104751@2x.png)

这里一天大概产生了700左右个索引数据。对历史的链路数据我们不做过多的保留。

这里我整理了个小工具，可以定期清理es数据。

## 一、清理思路

可以看到索引数据都是以日期结尾，我们可以根据日期去匹配索引数据，并对索引进行删除。这里需要考虑一点，有的Es服务开启了索引保护机制，不能通过`*index`去删除，只能通过索引的全名称去删除。所以我们整体流程如下：

1、获取es服务中全部索引数据。

2、根据当前时间-保留天数，获取要删除的日期。

3、通过字符串匹配，判断索引中是否包含要删除的日期，如果包含则进行删除。

4、工具友好性，我们可以通过配置文件配置ES服务地址、日期格式化类型、保留天数等信息。

## 二、代码实现

### 2.1、获取ES服务中全部索引数据

要获取Es服务中全部索引数据，我们首先连接Es服务器，这里我们使用`github.com/olivere/elastic/v7`库操作Es。

- 连接ES:

```go
func GetEsClient(data Data) *elastic.Client {
	Init()
	file := "./eslog.log"
	logFile, _ := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766) 
	client, err := elastic.NewClient(
		elastic.SetURL(data.Host),
		elastic.SetSniff(false),
		elastic.SetInfoLog(log.New(logFile, "ES-INFO: ", 0)),
		elastic.SetTraceLog(log.New(logFile, "ES-TRACE: ", 0)),
		elastic.SetErrorLog(log.New(logFile, "ES-ERROR: ", 0)),
	)
	if err != nil {
		return nil
	}
	return client
}
```

我们通过GetEsClient方法，连接ES,并返回client，供后续方法使用。这里的Data是包含了ES服务地址等信息，我们后面会给出Data的数据结构。

- 获取全部索引数据

```go
func getIndex(data Data) map[string]interface{} {
	client := GetEsClient(data)
	mapping := client.GetMapping()
	service := mapping.Index("*")
	result, err := service.Do(context.Background())
	if err != nil {
		fmt.Printf("create index failed, err: %v\n", err)
		return nil
	}
	return result
}
```

通过`client.GetMapping().Index("*")`API获取es服务中全部的索引数据，并返回，数据格式如下：

![WX20211008-110537@2x](https://img-1258527903.cos.ap-beijing.myqcloud.com/img/WX20211008-110537@2x.png)

这次我们获取全部索引完成。

### 2.2、根据当前时间-保留天数，获取要删除的日期

我们根据当前时间-保留天数，获取当前需要删除的日期数据。我们通过GoLang内置的函数库`time`完成该功能的实现。

```go
currentTime := time.Now()//获取当前时间
oldTime := currentTime.AddDate(0, 0, data.Day)//通过配置文件获取保留天数
format := oldTime.Format(data.IndexFmt)//通过配置文件获取序列化日期格式
```

### 2.3、通过字符串匹配，判断索引中是否包含要删除的日期，如果包含则进行删除

这里通过字符串匹配进行判断是否需要删除索引数据。

```go
func delIndex(data Data) {
	currentTime := time.Now()
	oldTime := currentTime.AddDate(0, 0, data.Day)
	format := oldTime.Format(data.IndexFmt)
	index := getIndex(data)//获取全部索引
	for k := range index {//遍历索引数据
		fmt.Println("key:", k, "format:", format)
		if find := strings.Contains(k, format); find { //判断索引中是否包含要删除的日期格式，
			DelIndex(data, k)//如果包含则调用DelIndex方法删除
		}
	}
}
```

```go
// DelIndex 删除 index
func DelIndex(data Data, index ...string) bool {
	client := GetEsClient(data)
	response, err := client.DeleteIndex(index...).Do(context.Background())
	if err != nil {
		fmt.Printf("delete index failed, err: %v\n", err)
		return false
	}
	return response.Acknowledged
}
```

通过`DeleteIndex`API删除指定的数据。

### 2.4、通过配置文件灵活配置数据

这里我们定义了Config和Data对象，对象结构如下：

```go
type Config struct {
	Data []Data `json:"data"`
}

type Data struct {
	Host     string `json:"host"`
	IndexFmt string `json:"index_fmt"`
	Day      int    `json:"day"`
}
```

配置文件内容如下：

```json
{
  "data": [
    {
      "host": "http://ip1:9200",//服务IP
      "index_fmt": "20060102",//日期格式化
      "day": -1 //保留天数 保留1天
    },
    {
      "host": "http://ip2:9200/",
      "index_fmt": "20060102",
      "day": -1
    },
    {
      "host": "http://ip3:32093",
      "index_fmt": "2006.01.02",
      "day": -7  //保留天数 保留7天
    }
  ]
}
```

我们通过Init方法加载配置文件到Config;

```go
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
```

编写Main方法运行程序：

```go
func main() {
	Init()
	for i, datum := range config.Data {
		fmt.Printf("config data Host is [%s], fmt is [%s]\n", datum.Host, datum.IndexFmt)
		println(i)
		delIndex(datum)
	}
}
```

这里我们依然遍历配置文件中的多个服务配置。可以同时管理多个Es服务。

## 三、完整代码

```go
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

```

```go
package main

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
	"log"
	"os"
	"time"
)

// GetEsClient 初始化客户端
func GetEsClient(data Data) *elastic.Client {
	Init()
	file := "./eslog.log"
	logFile, _ := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766) // 应该判断error，此处简略
	client, err := elastic.NewClient(
		elastic.SetURL(data.Host),
		elastic.SetSniff(false),
		elastic.SetInfoLog(log.New(logFile, "ES-INFO: ", 0)),
		elastic.SetTraceLog(log.New(logFile, "ES-TRACE: ", 0)),
		elastic.SetErrorLog(log.New(logFile, "ES-ERROR: ", 0)),
	)
	if err != nil {
		return nil
	}
	return client
}

// IsDocExists 判断索引是否存储
func IsDocExists(data Data, id string, index string) bool {
	client := GetEsClient(data)
	defer client.Stop()
	exist, _ := client.Exists().Index(index).Id(id).Do(context.Background())
	if !exist {
		log.Println("ID may be incorrect! ", id)
		return false
	}
	return true
}

// PingNode 是否联通
func PingNode(data Data) {
	start := time.Now()
	client := GetEsClient(data)
	info, code, err := client.Ping(data.Host).Do(context.Background())
	if err != nil {
		fmt.Printf("ping es failed, err: %v", err)
	}
	duration := time.Since(start)
	fmt.Printf("cost time: %v\n", duration)
	fmt.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)
}

// GetDoc 获取文档
func GetDoc(data Data, id string, index string) (*elastic.GetResult, error) {
	client := GetEsClient(data)
	defer client.Stop()
	if !IsDocExists(data, id, index) {
		return nil, fmt.Errorf("id不存在")
	}
	esResponse, err := client.Get().Index(index).Id(id).Do(context.Background())
	if err != nil {
		return nil, err
	}
	return esResponse, nil
}

// CreateIndex 创建 index
func CreateIndex(data Data, index, mapping string) bool {
	client := GetEsClient(data)
	result, err := client.CreateIndex(index).BodyString(mapping).Do(context.Background())
	if err != nil {
		fmt.Printf("create index failed, err: %v\n", err)
		return false
	}
	return result.Acknowledged
}

// DelIndex 删除 index
func DelIndex(data Data, index ...string) bool {
	client := GetEsClient(data)
	response, err := client.DeleteIndex(index...).Do(context.Background())
	if err != nil {
		fmt.Printf("delete index failed, err: %v\n", err)
		return false
	}
	return response.Acknowledged
}

func getIndex(data Data) map[string]interface{} {
	client := GetEsClient(data)
	mapping := client.GetMapping()
	service := mapping.Index("*")
	result, err := service.Do(context.Background())
	if err != nil {
		fmt.Printf("create index failed, err: %v\n", err)
		return nil
	}
	return result
}

```
