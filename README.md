# ali-apigateway-sdk

## 注意事项

本sdk目前只支持 `multipart/form-data`,  `application/json`, `application/x-www-form-urlencoded`
只支持 `post`, `get`
## 配置

```text
Host                          string // host
AppKey                        string // ak
AppSecret                     string // sk
ReadTimeout                   int    `json:",optional,default=5000"`      // 读超时时间，单位毫秒，默认5s
WriteTimeout                  int    `json:",optional,default=5000"`      // 写超时，单位毫秒,默认5s
MaxConnections                int    `json:",optional,default=10000"`     // 关闭空闲活动连接时间，单位 毫秒，默认10s
NoDefaultUserAgentHeader      bool   `json:",optional,default=true"`      // 是否去掉请求中去掉User-Agent标头, true为去掉
DisableHeaderNamesNormalizing bool   `json:",optional,default=true"`      // header中的key按照原样传输，否则会根据标准化转化,默认按原样传输
DisablePathNormalizing        bool   `json:",optional,default=true"`      // 是否按照原样传输路径，否则会根据标准取转化，true时，按原样传输
MaxResponseBodySize           int    `json:",optional,default=104857600"` // 响应体最大字节，单位字节。默认100M
MaxConnsPerHost               int    `json:",optional,default=300"`       // 每个主机最大并发限制数，为空的时候不限制
MaxConnDuration               int    `json:",optional,default=600000"`    // 每个连接存活时间，单位毫秒， 默认 10分钟
MaxConnWaitTimeout            int    `json:",optional,default=60000"`     // 每个请求超时时间，单位毫秒，默认10s
```

## 使用方式

### 初始化客户端

```go
config := sdk.ApiGatewayConf{
    Host:      "",
    AppKey:    "",
    AppSecret: "",
}
s := sdk.MustNewApiGatewayClient(config)
```

### application/json
```go

body := make(map[string]interface{})
body["categoryId"] = ""

data, err := s.NewApiWithCtx(ctx, sdk.HTTPMethodPost, "/xxxxx", sdk.HTTPContentTypeJson, body).Do()
```

### application/x-www-form-urlencoded
```go
body := "key1=value1&key2=value2"

data, err := s.NewApiWithCtx(ctx,sdk.HTTPMethodPost, "/xxxx", sdk.HTTPContentTypeForm, body).Do()
```

### multipart/form-data

```go
var b bytes.Buffer
writer := multipart.NewWriter(&b)
// 添加字段
writer.WriteField("field1", "value1")
writer.WriteField("field2", "value2")
// 如果有文件
// 创建一个 form 文件部分
formFile, err := body.CreateFormFile("file", fmt.Sprintf("0%s", extStr))
if err != nil {
    return nil, err
}

// 将文件流直接写入到 form 文件部分
if _, err := io.Copy(formFile, file); err != nil {
    return nil, err
}

// 结束 multipart 写入
if err := writer.Close(); err != nil {
    return nil, err
}

data, err := s.NewApiWithCtx(ctx, sdk.HTTPMethodPost, "/xxxx", writer.FormDataContentType(), b).Do()

```


### get

```go
path := "/xxxxx?key1=value1&key2=value2"
data, err := s.NewApiWithCtx(ctx, sdk.HTTPMethodPost, path, sdk.HTTPContentTypeForm, nil).Do()
```