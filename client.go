package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"strings"
	"time"
)

type (
	ApiGatewayClient struct {
		Client    *fasthttp.Client
		Host      string
		AppKey    string
		AppSecret string
	}

	ApiGatewayConf Conf

	ApiReq struct {
		ctx         context.Context
		Path        string
		Body        interface{}
		Method      string
		ContentType string
		ApiGatewayClient
		req  *fasthttp.Request
		resp *fasthttp.Response
	}
)

func MustNewApiGatewayClient(conf ApiGatewayConf) *ApiGatewayClient {
	return NewApiGatewayClient(conf)
}

func NewApiGatewayClient(conf ApiGatewayConf) *ApiGatewayClient {
	client := &fasthttp.Client{
		// 读超时时间,不设置read超时,可能会造成连接复用失效
		ReadTimeout: time.Millisecond * time.Duration(conf.ReadTimeout),
		// 写超时时间
		WriteTimeout: time.Millisecond * time.Duration(conf.WriteTimeout),
		// 关闭空闲的活动连接
		MaxIdleConnDuration: time.Millisecond * time.Duration(conf.MaxConnDuration),
		// 当true时,从请求中去掉User-Agent标头
		NoDefaultUserAgentHeader: conf.NoDefaultUserAgentHeader,
		// 当true时，header中的key按照原样传输，默认会根据标准化转化
		DisableHeaderNamesNormalizing: conf.DisableHeaderNamesNormalizing,
		//当true时,路径按原样传输，默认会根据标准化转化
		DisablePathNormalizing: conf.DisablePathNormalizing,
		MaxResponseBodySize:    conf.MaxResponseBodySize,
		MaxConnsPerHost:        conf.MaxConnsPerHost,
		MaxConnDuration:        time.Millisecond * time.Duration(conf.MaxConnDuration),
		MaxConnWaitTimeout:     time.Millisecond * time.Duration(conf.MaxConnWaitTimeout),
		Dial: (&fasthttp.TCPDialer{
			// 最大并发数，0表示无限制
			Concurrency: 4096,
			// 将 DNS 缓存时间从默认分钟增加到一小时
			DNSCacheDuration: time.Hour,
		}).Dial,
	}

	s := &ApiGatewayClient{
		Client:    client,
		Host:      conf.Host,
		AppKey:    conf.AppKey,
		AppSecret: conf.AppSecret,
	}

	return s
}

func (s *ApiGatewayClient) NewApiWithCtx(ctx context.Context, contentType, path, method string, body interface{}) *ApiReq {
	return &ApiReq{
		ContentType:      contentType,
		Body:             body,
		Path:             path,
		ApiGatewayClient: *s,
		Method:           method,
		ctx:              ctx,
	}
}

func (s *ApiGatewayClient) NewApi(contentType, path, method string, body interface{}) *ApiReq {
	return &ApiReq{
		ContentType:      contentType,
		Body:             body,
		Path:             path,
		ApiGatewayClient: *s,
		Method:           method,
	}
}

func (a *ApiReq) NewRequest() {
	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	a.req = req
	a.resp = resp
}

func (a *ApiReq) setReqUri() string {
	return fmt.Sprintf("%s%s", a.Host, a.Path)
}

func (a *ApiReq) setBody(req *fasthttp.Request, contentType string, body interface{}) error {
	reqBody := make([]byte, 0)
	switch {
	case contentType == HTTPContentTypeJson:
		marshal, _ := json.Marshal(body)
		reqBody = marshal
	case contentType == HTTPContentTypeForm:
		strBody, ok := body.(string)
		if !ok {
			return errors.New(fmt.Sprintf("when contenType is %s, body content type must be str",
				contentType))
		}
		reqBody = []byte(strBody)
	case strings.HasPrefix(contentType, HTTPContentTypeMultipartForm):
		bufferBody, ok := body.(bytes.Buffer)
		if !ok {
			return errors.New(fmt.Sprintf("when contenType is %s, body content type must be buffer",
				contentType))
		}
		reqBody = bufferBody.Bytes()
	default:
		return errors.New(fmt.Sprintf("content type not support: %s", contentType))
	}

	req.SetBody(reqBody)
	return nil
}

func (a *ApiReq) doApi() ([]byte, error) {
	err := a.Client.Do(a.req, a.resp)
	if err != nil {
		return nil, err
	}

	var res ApiResult
	err = json.Unmarshal(a.resp.Body(), &res)

	if err != nil {
		return nil, err
	}

	if res.Code != "200" {
		return nil, errors.New(res.Message)
	}

	if res.Data != nil {
		jsonData, err := json.Marshal(res.Data)
		if err != nil {
			return nil, err
		}
		return jsonData, nil
	}
	return nil, nil
}

func (a *ApiReq) Do() ([]byte, error) {
	if a.ctx == nil {
		logx.Infof("start api without ctx, %s, %s, %+v", a.Method, a.Path, a.Body)
	} else {
		logc.Infof(a.ctx, "start api with ctx, %s, %s, %+v", a.Method, a.Path, a.Body)
	}

	start := time.Now()

	a.NewRequest()
	a.req.Header.SetMethod(a.Method)
	a.req.SetRequestURI(a.setReqUri())

	if a.Body != nil {
		err := a.setBody(a.req, a.ContentType, a.Body)
		if err != nil {
			return nil, err
		}
	}
	a.req.Header.SetContentType(a.ContentType)
	initialBasicHeader(a.req, a.AppKey, a.AppSecret, a.ContentType)

	data, err := a.doApi()

	end := time.Now()

	elapsed := end.Sub(start)
	milliseconds := elapsed.Milliseconds()

	if err != nil {
		if a.ctx != nil {
			logc.Errorf(a.ctx, "failed api with ctx, %s, %s, execution time: %d ms, err:%v",
				a.Method, a.Path, milliseconds, err)
		} else {
			logx.Errorf("failed api without ctx, %s, %s, execution time: %d ms, err:%v",
				a.Method, a.Path, milliseconds, err)
		}
		return nil, err
	}

	if a.ctx != nil {
		logc.Errorf(a.ctx, "success api with ctx, %s, %s, execution time: %d ms, data:%v",
			a.Method, a.Path, milliseconds, string(data))
	} else {
		logx.Errorf("success api without ctx, %s, %s, execution time: %d ms, data:%v",
			a.Method, a.Path, milliseconds, string(data))
	}
	return data, nil
}
