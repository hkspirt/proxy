//----------------
//Func  :
//Author: xjh
//Date  : 2019/
//Note  :
//----------------
package internal

import (
	"github.com/hkspirt/proxy/util"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

type IProxy interface {
	GetOne() string
}

type IProxyer interface {
	Init()
	Start(IProxy, IProxyer, <-chan bool)
	GetOne() string
	UseFailed(string)

	load()
	check(string) bool
}

const (
	AliveCheckTimeOut = 10 * time.Second         //单个代理检测超时
	ProxyGetTimeOut   = 30 * time.Second         //单个代理获取超时
	AliveCheckUrl     = "http://httpbin.org/get" //代理检测地址
	LoadSuccessTimer  = time.Minute * 30         //重新获取一次数据间隔-上一次获取成功
	LoadFailedTimer   = time.Second * 15         //重新获取一次数据间隔-上一次获取失败
)

type Proxyer struct {
	Url       string
	NeedProxy bool
	proxy     IProxy //通过代理去获取
	cookies   string

	regexp *regexp.Regexp
	addrs  []string
	canuse sync.Map
}

func (self *Proxyer) Start(pm IProxy, p IProxyer, cstop <-chan bool) {
	self.proxy = pm
	for {
		p.load()
		self.canuse = sync.Map{}
		for _, addr := range self.addrs {
			if p.check(addr) {
				util.LogInfo("proxyer:%s proxy url:%s check success", self.Url, addr)
				self.canuse.Store(addr, time.Now().Unix())
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
	var px func(*http.Request) (*url.URL, error) = nil
	paddr := self.proxy.GetOne()
	if paddr != "" && self.NeedProxy {
		px = func(*http.Request) (*url.URL, error) { return url.Parse(paddr) }
	}
	if px == nil && self.NeedProxy {
		return nil, "", ""
	}

	client := &http.Client{
		Transport: &http.Transport{Proxy: px},
		Timeout:   ProxyGetTimeOut,
	}
	if strings.HasPrefix(paddr, "http") {
		return client, "http://" + self.Url, paddr
	} else {
		return client, "https://" + self.Url, paddr
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

func (self *Proxyer) GetOne() string {
	addr := ""
	self.canuse.Range(func(key, value interface{}) bool {
		addr = key.(string)
		return false
	})
	return addr
}

func (self *Proxyer) UseFailed(addr string) {
	self.canuse.Delete(addr)
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
