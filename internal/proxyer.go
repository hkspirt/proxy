//----------------
//Func  :
//Author: xjh
//Date  : 2019/
//Note  :
//----------------
package internal

import (
	"context"
	"github.com/hkspirt/proxy/util"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type IProxy interface {
	GetOne() string
	Add(string)
	Has(string) bool
}

type IProxyer interface {
	Init()
	Start(IProxy, IProxyer, <-chan bool)
	load()
	check(string) bool
}

const (
	AliveCheckTimeOut = 10 * time.Second         //单个代理检测超时
	ProxyGetTimeOut   = 30 * time.Second         //单个代理获取超时
	AliveCheckUrl     = "http://httpbin.org/get" //代理检测地址
	LoadSuccessTimer  = time.Minute * 10         //重新获取一次数据间隔-上一次获取成功
	LoadFailedTimer   = time.Second * 10         //重新获取一次数据间隔-上一次获取失败
)

type Proxyer struct {
	Url     string
	proxy   IProxy //通过代理去获取
	cookies string

	regexp *regexp.Regexp
	addrs  []string
}

func (self *Proxyer) Start(pm IProxy, p IProxyer, cstop <-chan bool) {
	self.proxy = pm
	for {
		p.load()
		for _, addr := range self.addrs {
			if pm.Has(addr) {
				continue
			}
			if self.check(addr) {
				pm.Add(addr)
			}
		}
		if len(self.addrs) > 0 {
			time.Sleep(LoadSuccessTimer)
		} else {
			time.Sleep(LoadFailedTimer)
		}
	}
}

func (self *Proxyer) newHttpClient() (*http.Client, string, string) {
	var px func(*http.Request) (*url.URL, error)
	paddr := ""
	if rand.Intn(10)%2 == 0 { //随机选择是否用代理
		paddr = self.proxy.GetOne()
	}

	if paddr != "" {
		px = func(*http.Request) (*url.URL, error) { return url.Parse(paddr) }
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: px,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(network, addr, time.Second*5)
				if err != nil {
					return nil, err
				}
				conn.SetDeadline(time.Now().Add(time.Second * 15))
				return conn, nil
			},
			TLSHandshakeTimeout:   time.Second * 5,
			ResponseHeaderTimeout: time.Second * 10,
			IdleConnTimeout:       time.Second * 5,
		},
		Timeout: ProxyGetTimeOut,
	}
	if strings.HasPrefix(paddr, "https") {
		return client, "https://" + self.Url, paddr
	} else {
		return client, "http://" + self.Url, paddr
	}
}

func (self *Proxyer) httpGet() []byte {
	client, requrl, paddr := self.newHttpClient()
	if client == nil {
		util.LogWarn("proxyer:%s need prox but no", self.Url)
		return []byte{}
	}

	req, err := http.NewRequest(http.MethodGet, requrl, nil)
	if err != nil {
		util.LogWarn("proxyer:%s paddr:%s request err:%v", self.Url, paddr, err)
		return []byte{}
	}
	req.Header.Set("Cookie", self.cookies)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	//req.Header.Set("Accept-Encoding", "gzip, deflate, sdch")

	resp, err := client.Do(req)
	if err != nil {
		util.LogWarn("proxyer:%s proxy:%s get err:%v", self.Url, paddr, err)
		return []byte{}
	}

	if resp.StatusCode != http.StatusOK {
		util.LogWarn("proxyer:%s proxy:%s status err:%v", self.Url, paddr, resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		util.LogWarn("proxyer:%s proxy:%s read err:%v", self.Url, paddr, err)
		return []byte{}
	}
	return data
}

func (self *Proxyer) check(addr string) bool {
	proxyUrl, err := url.Parse(addr)
	if err != nil {
		util.LogWarn("proxyer:%s addr:%s parse err:%v", self.Url, addr, err)
		return false
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
		Timeout: AliveCheckTimeOut,
	}
	resp, err := client.Get(AliveCheckUrl)
	if err != nil {
		return false
	}
	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}
