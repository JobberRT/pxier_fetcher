package core

import (
	"crypto/tls"
	"fmt"
	"github.com/JobberRT/pxier_fetcher/public"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"regexp"
	"strings"
	"time"
)

var (
	keyPattern = regexp.MustCompile("[a-z\\d]{32}")
	ipPattern  = regexp.MustCompile("\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}:\\d{1,5}")
)

type ihuanExecutor struct {
	httpUrl       string
	statisticsUrl string
	keyUrl        string
	key           string
	zone          string
	statistics    string
	eachFetchNum  int
	timeout       time.Duration
	client        *fasthttp.Client
}

func newIHuanExecutor() *ihuanExecutor {
	logrus.Info("creating ihuan executor")
	hu := viper.GetString("executor.ihuan.http_url")
	if len(hu) == 0 {
		hu = "https://ip.ihuan.me/tqdl.html"
	}
	su := viper.GetString("executor.ihuan.statistics_url")
	if len(su) == 0 {
		su = "https://ip.ihuan.me/ti.html"
	}
	ku := viper.GetString("executor.ihuan.key_url")
	if len(ku) == 0 {
		ku = "https://ip.ihuan.me/mouse.do"
	}
	timeout := viper.GetInt64("executor.ihuan.timeout")
	if timeout == 0 {
		timeout = 15
	}
	efn := viper.GetInt("executor.ihuan.each_fetch_num")
	if efn == 0 {
		efn = 100
	}
	zone := viper.GetString("executor.ihuan.zone")
	f := &ihuanExecutor{
		httpUrl:       hu,
		statisticsUrl: su,
		keyUrl:        ku,
		eachFetchNum:  efn,
		zone:          zone,
		timeout:       time.Duration(timeout) * time.Second,
		client:        &fasthttp.Client{TLSConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	proxy := viper.GetString("executor.ihuan.proxy")
	if len(proxy) != 0 {
		if strings.Contains(proxy, "http") {
			f.client.Dial = fasthttpproxy.FasthttpHTTPDialer(proxy)
		} else {
			f.client.Dial = fasthttpproxy.FasthttpSocksDialer(proxy)
		}
	}
	return f
}

func (f *ihuanExecutor) Fetch() []*proxy {
	logrus.WithField("provider", f.Type()).Info("fetching proxy")
	if len(f.statistics) == 0 {
		f.generateStatistics()
	}
	if len(f.key) == 0 {
		f.generateKey()
	}
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	postData := fmt.Sprintf("num=%d&port=&kill_port=&address=%s&kill_address=&anonymity=&type=&post=&sort=1&key=%s", f.eachFetchNum, f.zone, f.key)
	req.SetRequestURI(f.httpUrl)
	req.SetBodyString(postData)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.Set("Accept-Encoding", "br")
	req.Header.SetContentType("application/x-www-form-urlencoded")
	req.Header.SetUserAgent("Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.85 Safari/537.36")
	req.Header.SetReferer("https://ip.ihuan.me/ti.html")
	if err := f.client.DoTimeout(req, res, f.timeout); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"url":      f.httpUrl,
			"provider": f.Type(),
		}).Error("failed to fetch proxy")
		return nil
	}

	body, err := readBody(res)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"raw":      string(body),
			"url":      f.httpUrl,
			"provider": f.Type(),
		}).Error("failed to unGzip body")
		return nil
	}
	Ips := ipPattern.FindAll(body, -1)
	if Ips == nil {
		logrus.WithFields(logrus.Fields{
			"raw":      string(body),
			"url":      f.httpUrl,
			"provider": f.Type(),
		}).Error("empty ips")
		return nil
	}

	proxies := make([]*proxy, 0)
	for _, ip := range Ips {
		if ip == nil {
			continue
		}
		proxies = append(proxies, &proxy{
			Address:   string(ip),
			Provider:  public.ExecutorTypeIHuan,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			ErrTimes:  0,
			DialType:  public.DialTypeHttp,
		})
	}
	return proxies
}

func (f *ihuanExecutor) Type() string {
	return public.ExecutorTypeIHuan
}

func (f *ihuanExecutor) generateStatistics() {
	logrus.WithField("provider", f.Type()).Info("generate statistics")
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(f.statisticsUrl)
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.SetUserAgent("Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.85 Safari/537.36")
	req.Header.Set("Accept-Encoding", "br")
	if err := f.client.DoTimeout(req, res, f.timeout); err != nil {
		logrus.WithError(err).WithField("url", f.statisticsUrl).Error("failed to get statistics")
		return
	}
	if res.Header.Peek("Set-Cookie") == nil {
		logrus.WithField("raw", res.Header.String()).Error("empty statistics")
		return
	}
	f.statistics = string(res.Header.Peek("Set-Cookie"))
}

func (f *ihuanExecutor) generateKey() {
	logrus.WithField("provider", f.Type()).Info("generate key")
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(f.keyUrl)
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.SetUserAgent("Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.85 Safari/537.36")
	req.Header.Set("Accept-Encoding", "br")
	req.Header.SetReferer(f.statisticsUrl)
	req.Header.Set("Cookie", f.statistics)
	if err := f.client.DoTimeout(req, res, f.timeout); err != nil {
		logrus.WithError(err).WithField("url", f.statisticsUrl).Error("failed to get statistics")
		return
	}

	bodyBytes, err := readBody(res)
	if err != nil {
		logrus.WithError(err).Error("failed to get response body bytes")
		return
	}
	key := keyPattern.Find(bodyBytes)
	if key == nil {
		logrus.WithField("raw", string(bodyBytes)).Error("empty key")
		return
	}
	f.key = string(key)
}
