package sdk

type Conf struct {
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
}
