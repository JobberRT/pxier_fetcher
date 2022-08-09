package core

import (
	"crypto/tls"
	"github.com/JobberRT/pxier_fetcher/public"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"strings"
	"time"
)

type strExecutor struct {
	httpUrl   string
	socks5Url string
	client    *fasthttp.Client
	timeout   time.Duration
}

func newSTRExecutor() *strExecutor {
	logrus.Info("creating str executor")
	hu := viper.GetString("executor.str.http_url")
	if len(hu) == 0 {
		hu = "https://raw.githubusercontent.com/shiftytr/proxy-list/master/http.txt"
	}
	su := viper.GetString("executor.str.socks5_url")
	if len(su) == 0 {
		su = "https://raw.githubusercontent.com/shiftytr/proxy-list/master/socks5.txt"
	}
	timeout := viper.GetInt64("executor.str.timeout")
	if timeout == 0 {
		timeout = 5
	}
	f := &strExecutor{
		httpUrl:   hu,
		socks5Url: su,
		timeout:   time.Duration(timeout) * time.Second,
		client:    &fasthttp.Client{TLSConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	proxy := viper.GetString("executor.str.proxy")
	if len(proxy) != 0 {
		if strings.Contains(proxy, "http") {
			f.client.Dial = fasthttpproxy.FasthttpHTTPDialer(proxy)
		} else {
			f.client.Dial = fasthttpproxy.FasthttpSocksDialer(proxy)
		}
	}
	return f
}

func (f *strExecutor) Fetch() []*proxy {
	logrus.WithField("provider", f.Type()).Info("fetch")
	return append(f.fetchHttpProxy(), f.fetchSocks5Proxy()...)
}

func (f *strExecutor) Type() string {
	return public.ExecutorTypeSTR
}

func (f *strExecutor) fetchHttpProxy() []*proxy {
	logrus.WithField("provider", f.Type()).Info("fetching http proxy")
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(f.httpUrl)
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.SetContentEncoding("gzip")
	if err := f.client.DoTimeout(req, res, f.timeout); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"url":      f.httpUrl,
			"provider": f.Type(),
			"type":     public.DialTypeHttp,
		}).Error("failed to fetch http proxy")
		return nil
	}

	body, err := readBody(res)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"raw":      string(body),
			"url":      f.httpUrl,
			"provider": f.Type(),
			"type":     public.DialTypeHttp,
		}).Error("failed to unGzip body")
		return nil
	}
	rawSlice := strings.Split(string(body), "\n")
	proxies := make([]*proxy, 0)
	for _, each := range rawSlice {
		if len(each) == 0 {
			continue
		}
		proxies = append(proxies, &proxy{
			Address:   each,
			ErrTimes:  0,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			Provider:  public.ExecutorTypeSTR,
			DialType:  public.DialTypeHttp,
		})
	}
	return proxies
}

func (f *strExecutor) fetchSocks5Proxy() []*proxy {
	logrus.WithField("provider", f.Type()).Info("fetching socks5 proxy")
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(f.socks5Url)
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.SetContentEncoding("gzip")
	if err := f.client.DoTimeout(req, res, f.timeout); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"url":      f.httpUrl,
			"provider": f.Type(),
			"type":     public.DialTypeHttp,
		}).Error("failed to fetch http proxy")
		return nil
	}

	body, err := readBody(res)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"raw":      string(body),
			"url":      f.httpUrl,
			"provider": f.Type(),
			"type":     public.DialTypeHttp,
		}).Error("failed to unGzip body")
		return nil
	}
	rawSlice := strings.Split(string(body), "\n")
	proxies := make([]*proxy, 0)
	for _, each := range rawSlice {
		if len(each) == 0 {
			continue
		}
		proxies = append(proxies, &proxy{
			Address:   each,
			ErrTimes:  0,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			Provider:  public.ExecutorTypeSTR,
			DialType:  public.DialTypeSocks5,
		})
	}
	return proxies
}
