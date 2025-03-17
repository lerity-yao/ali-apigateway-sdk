package sdk

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ApiResult struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Success bool        `json:"success"`
}

// HTTP header keys.
const (
	HTTPHeaderAccept      = "Accept"
	HTTPHeaderContentMD5  = "Content-MD5"
	HTTPHeaderContentType = "Content-Type"
	HTTPHeaderDate        = "Date"
	HTTPHeaderUserAgent   = "User-Agent"
)

// HTTP header keys used for Aliyun API gateway signature.
const (
	HTTPHeaderCAPrefix           = "X-Ca-"
	HTTPHeaderCAKey              = "X-Ca-Key"
	HTTPHeaderCANonce            = "X-Ca-Nonce"
	HTTPHeaderCASignature        = "X-Ca-Signature"
	HTTPHeaderCASignatureHeaders = "X-Ca-Signature-Headers"
	HTTPHeaderCASignatureMethod  = "X-Ca-Signature-Method"
	HTTPHeaderCATimestamp        = "X-Ca-Timestamp"
)

// HTTP header content-type values.
const (
	HTTPContentTypeForm                      = "application/x-www-form-urlencoded"
	HTTPContentTypeMultipartForm             = "multipart/form-data"
	HTTPContentTypeMultipartFormWithBoundary = "multipart/form-data; boundary="
	HTTPContentTypeStream                    = "application/octet-stream"
	HTTPContentTypeJson                      = "application/json"
	HTTPContentTypeXml                       = "application/xml"
	HTTPContentTypeText                      = "text/plain"
)

// HTTP method values.
const (
	HTTPMethodGet     = "GET"
	HTTPMethodPost    = "POST"
	HTTPMethodPut     = "PUT"
	HTTPMethodDelete  = "DELETE"
	HTTPMethodPatch   = "PATCH"
	HTTPMethodHead    = "HEAD"
	HTTPMethodOptions = "OPTIONS"
)

// default values.
const (
	defaultUserAgent  = "Go-Aliyun-Sign-Client"
	defaultAccept     = "*/*"
	defaultSignMethod = "HmacSHA256"
	defaultLF         = "\n"
	defaultSep        = ","
)

var signHeaders = map[string]struct{}{
	http.CanonicalHeaderKey(HTTPHeaderCAKey):             {},
	http.CanonicalHeaderKey(HTTPHeaderCANonce):           {},
	http.CanonicalHeaderKey(HTTPHeaderCASignatureMethod): {},
	http.CanonicalHeaderKey(HTTPHeaderCATimestamp):       {},
}

// currentTimeMillis returns the millisecond representation of the current time.
func currentTimeMillis() string {
	t := time.Now().UnixNano() / 1000000

	return strconv.FormatInt(t, 10)
}

// currentGMTDate returns the GMT date representation of the current time.
func currentGMTDate() string {
	return time.Now().UTC().Format(http.TimeFormat)
}

//

// uuid4 returns random generated UUID string.
func uuid4() string {
	u, err := uuid.NewRandom()
	if err != nil {
		return ""
	}

	return u.String()
}

// hmacSHA256 returns the string encrypted with HmacSHA256 method.
func hmacSHA256(b, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(b)

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// xmd5 returns the string hashed with MD5 method.
func xmd5(b []byte) string {
	m := md5.New()
	m.Write(b)

	return base64.StdEncoding.EncodeToString(m.Sum(nil))
}

func buildParamStringToSign(req *fasthttp.Request) string {

	queryArgs := req.URI().QueryArgs()

	queryKeys := make([][]byte, 0)

	queryArgs.VisitAll(func(key, value []byte) {
		queryKeys = append(queryKeys, key)
	})

	paramKeys := getSortKeys(queryKeys)

	paramList := make([]string, 0)
	for _, key := range paramKeys {
		value := queryArgs.Peek(key)
		if string(value) == "" {
			paramList = append(paramList, key)
		} else {
			paramList = append(paramList, key+"="+string(value))
		}
	}

	params := strings.Join(paramList, "&")
	if params != "" {
		params = "?" + params
	}

	return string(req.URI().Path()) + params
}

func getSortKeys(m [][]byte, needFormat ...bool) []string {
	nf := false
	if len(needFormat) > 0 {
		nf = needFormat[0]
	}

	keys := make([]string, 0, len(m))
	for _, item := range m {
		itemX := string(item)
		if nf {
			itemX = http.CanonicalHeaderKey(itemX)
		}
		keys = append(keys, itemX)
	}

	sort.Strings(keys)

	return keys
}

func buildHeaderStringToSign(req *fasthttp.Request) string {
	var builder strings.Builder

	headerAllKey := req.Header.PeekKeys()

	signHeaderKeys := make([]string, 0)

	headerKeys := getSortKeys(headerAllKey, true)

	for _, key := range headerKeys {
		if _, ok := signHeaders[key]; ok {
			signHeaderKeys = append(signHeaderKeys, key)
			builder.WriteString(key + ":" + string(req.Header.Peek(key)) + defaultLF)
		}
	}

	req.Header.Set(HTTPHeaderCASignatureHeaders, strings.Join(signHeaderKeys, defaultSep))

	return builder.String()
}

func buildStringToSign(req *fasthttp.Request) string {
	var s strings.Builder
	s.WriteString(strings.ToUpper(string(req.Header.Method())) + defaultLF)

	s.WriteString(string(req.Header.Peek(HTTPHeaderAccept)) + defaultLF)
	s.WriteString(string(req.Header.Peek(HTTPHeaderContentMD5)) + defaultLF)
	s.WriteString(string(req.Header.Peek(HTTPHeaderContentType)) + defaultLF)
	s.WriteString(string(req.Header.Peek(HTTPHeaderDate)) + defaultLF)

	s.WriteString(buildHeaderStringToSign(req))

	paramStr := buildParamStringToSign(req)

	s.WriteString(paramStr)
	//
	return s.String()
}

func initialBasicHeader(req *fasthttp.Request, ak, sk, contentType string) {
	req.Header.Set(HTTPHeaderContentType, contentType)
	req.Header.Set("Connection", "Keep-Alive")

	req.Header.Set(HTTPHeaderCAKey, ak)
	req.Header.Set(HTTPHeaderCANonce, uuid4())
	req.Header.Set(HTTPHeaderCASignatureMethod, defaultSignMethod)
	req.Header.Set(HTTPHeaderCATimestamp, currentTimeMillis())

	httpHeaderAccept := string(req.Header.Peek(HTTPHeaderAccept))
	if httpHeaderAccept == "" {
		req.Header.Set(HTTPHeaderAccept, defaultAccept)
	}

	httpHeaderDate := string(req.Header.Peek(HTTPHeaderDate))
	if httpHeaderDate == "" {
		req.Header.Set(HTTPHeaderDate, currentGMTDate())
	}

	httpHeaderUserAgent := string(req.Header.Peek(HTTPHeaderUserAgent))
	if httpHeaderUserAgent == "" {
		req.Header.Set(HTTPHeaderUserAgent, defaultUserAgent)
	}

	ct := string(req.Header.Peek(HTTPHeaderContentType))
	if req.Body() != nil && ct != HTTPContentTypeForm &&
		!strings.HasPrefix(ct, HTTPContentTypeMultipartFormWithBoundary) {
		req.Header.Set(HTTPHeaderContentMD5, xmd5(req.Body()))
	}

	stringToSign := buildStringToSign(req)

	req.Header.Set(HTTPHeaderCASignature, hmacSHA256([]byte(stringToSign), []byte(sk)))
}
