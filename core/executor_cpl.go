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

type cplExecutor struct {
	url     string
	timeout time.Duration
	client  *fasthttp.Client
}

func newCPLExecutor() *cplExecutor {
	logrus.Info("creating cpl executor")
	url := viper.GetString("executor.cpl.url")
	if len(url) == 0 {
		url = "https://raw.githubusercontent.com/clarketm/proxy-list/master/proxy-list-raw.txt"
	}
	timeout := viper.GetInt64("executor.cpl.timeout")
	if timeout == 0 {
		timeout = 5
	}
	f := &cplExecutor{
		url:     url,
		timeout: time.Duration(timeout) * time.Second,
		client:  &fasthttp.Client{TLSConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	proxy := viper.GetString("executor.cpl.proxy")
	if len(proxy) != 0 {
		if strings.Contains(proxy, "http") {
			f.client.Dial = fasthttpproxy.FasthttpHTTPDialer(proxy)
		} else {
			f.client.Dial = fasthttpproxy.FasthttpSocksDialer(proxy)
		}
	}
	return f
}

func (f *cplExecutor) Fetch() []*proxy {
	logrus.WithField("provider", f.Type()).Info("fetching proxy")
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(f.url)
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.SetContentEncoding("gzip")
	if err := f.client.DoTimeout(req, res, f.timeout); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"url":      f.url,
			"provider": f.Type(),
		}).Error("failed to fetch proxy")
		return nil
	}

	body, err := readBody(res)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"raw":      string(body),
			"url":      f.url,
			"provider": f.Type(),
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
			Provider:  public.ExecutorTypeCPL,
			DialType:  public.DialTypeHttp,
		})
	}
	return proxies
}

func (f *cplExecutor) Type() string {
	return public.ExecutorTypeCPL
}
