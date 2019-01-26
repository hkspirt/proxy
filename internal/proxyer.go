//----------------
//Func  :
//Author: xjh
//Date  : 2019/
//Note  :
//----------------
package internal

import (
	"github.com/hkspirt/proxy/util"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"
)

type IProxyer interface {
	Init()
	Start(IProxyer, <-chan bool)
	GetOne() string
	UseFailed(string)

	load()
	check(string) bool
}

const (
	AliveCheckTimeOut = 10                       //单个代理检测超时
	AliveCheckUrl     = "http://httpbin.org/get" //代理检测地址
	LoadSuccessTimer  = time.Minute * 30         //重新获取一次数据间隔-上一次获取成功
	LoadFailedTimer   = time.Second * 15         //重新获取一次数据间隔-上一次获取失败
)

type Proxyer struct {
	Url string

	regexp *regexp.Regexp
	addrs  []string
	canuse sync.Map
}

func (self *Proxyer) Start(p IProxyer, cstop <-chan bool) {
	for {
		p.load()
		self.canuse = sync.Map{}
		for _, addr := range self.addrs {
			if p.check(addr) {
				util.LogInfo("host:%s proxy:%s check success", self.Url, addr)
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
		Timeout: AliveCheckTimeOut * time.Second,
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
