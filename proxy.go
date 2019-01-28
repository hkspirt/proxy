//----------------
//Func  :
//Author: xjh
//Date  : 20189/
//Note  :
//----------------
package proxy

import (
	"github.com/hkspirt/proxy/internal"
	"github.com/hkspirt/proxy/util"
)

var (
	Proxy = &proxy{proxies: map[string]internal.IProxyer{}}
)

func init() {
	Proxy.start()
}

type proxy struct {
	proxies map[string]internal.IProxyer
}

func (self *proxy) GetOne() string {
	for _, p := range self.proxies {
		addr := p.GetOne()
		if addr != "" {
			return addr
		}
	}
	return ""
}

func (self *proxy) UseFailed(addr string) {
	for _, p := range self.proxies {
		p.UseFailed(addr)
	}
	util.LogWarn("proxy use failed:%s", addr)
}

func (self *proxy) start() {
	self.proxies["Kuaidaili_intr"] = &internal.Kuaidaili{Proxyer: internal.Proxyer{Url: "www.kuaidaili.com/free/intr/", NeedProxy: false}}                      //ok
	self.proxies["Kuaidaili_inha"] = &internal.Kuaidaili{Proxyer: internal.Proxyer{Url: "www.kuaidaili.com/free/inha/", NeedProxy: false}}                      //ok
	self.proxies["Data5u"] = &internal.Data5u{Proxyer: internal.Proxyer{Url: "www.data5u.com/free/index.shtml", NeedProxy: false}}                              //ok
	self.proxies["Ip66"] = &internal.Ip66{Proxyer: internal.Proxyer{Url: "www.66ip.cn/mo.php?tqsl=50", NeedProxy: true}}                                        //ok
	self.proxies["Proxylistplus"] = &internal.Proxylistplus{Proxyer: internal.Proxyer{Url: "list.proxylistplus.com/Fresh-HTTP-Proxy-List-1", NeedProxy: false}} //ok
	self.proxies["Xicidaili_nn"] = &internal.Xicidaili{Proxyer: internal.Proxyer{Url: "www.xicidaili.com/nn", NeedProxy: false}}                                //ok
	self.proxies["Xicidaili_nt"] = &internal.Xicidaili{Proxyer: internal.Proxyer{Url: "www.xicidaili.com/nt", NeedProxy: false}}                                //ok

	for _, px := range self.proxies {
		tmp := px
		tmp.Init()
		util.GoChan(func(cstop <-chan bool) {
			tmp.Start(self, tmp, cstop)
		})
	}
}
