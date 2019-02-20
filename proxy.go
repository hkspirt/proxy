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
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

var (
	Proxy = &proxy{snatchs: map[string]internal.IProxyer{}, canuse: map[string]int{}, usefailed: map[string]int{}}
)

const (
	FailedMaxTime = 50 //代理连续失败的最大次数(超过该次数则从池子中移除)
)

func init() {
	Proxy.start()
}

type proxy struct {
	sync.RWMutex
	snatchs   map[string]internal.IProxyer
	canuse    map[string]int
	usefailed map[string]int
	running   int32
}

func (self *proxy) Run() {
	Proxy.start()
}

func (self *proxy) Add(addr string) {
	self.Lock()
	defer self.Unlock()

	_, a1 := self.canuse[addr]
	_, a2 := self.usefailed[addr]
	if !a1 && !a2 {
		self.canuse[addr] = 1
	}
}

func (self *proxy) Has(addr string) bool {
	self.RLock()
	defer self.RUnlock()

	_, a1 := self.canuse[addr]
	_, a2 := self.usefailed[addr]
	return a1 || a2
}

func (self *proxy) Get() string {
	return self.GetOne()
}

func (self *proxy) GetOne() string {
	self.RLock()
	defer self.RUnlock()

	for addr, _ := range self.canuse {
		return addr
	}
	for addr, _ := range self.usefailed {
		util.LogWarn("no can use proxy")
		return addr
	}
	return ""
}

func (self *proxy) UseSuccess(addr string) {
	self.Lock()
	defer self.Unlock()

	last := self.canuse[addr]
	last += 128
	self.canuse[addr] = last
	delete(self.usefailed, addr)
}

func (self *proxy) UseFailed(addr string) {
	self.Lock()
	defer self.Unlock()

	last, ok := self.canuse[addr]
	if !ok {
		self.usefailed[addr] -= 1
		return
	}

	if last > 0 {
		last /= 2
	} else {
		last -= 1
	}

	if last < FailedMaxTime {
		delete(self.canuse, addr)
		self.usefailed[addr] = last
	} else {
		self.canuse[addr] = last
		delete(self.usefailed, addr)
	}
}

func (self *proxy) start() {
	if !atomic.CompareAndSwapInt32(&self.running, 0, 1) {
		return
	}

	self.snatchs["Kuaidaili_intr"] = &internal.Kuaidaili{Proxyer: internal.Proxyer{Url: "www.kuaidaili.com/free/intr/"}}                      //ok
	self.snatchs["Kuaidaili_inha"] = &internal.Kuaidaili{Proxyer: internal.Proxyer{Url: "www.kuaidaili.com/free/inha/"}}                      //ok
	self.snatchs["Data5u"] = &internal.Data5u{Proxyer: internal.Proxyer{Url: "www.data5u.com"}}                                               //ok
	self.snatchs["Data5u_in"] = &internal.Data5u{Proxyer: internal.Proxyer{Url: "www.data5u.com/free/index.shtml"}}                           //ok
	self.snatchs["Data5u_gngn"] = &internal.Data5u{Proxyer: internal.Proxyer{Url: "www.data5u.com/free/gngn/index.shtml"}}                    //ok
	self.snatchs["Data5u_gnpt"] = &internal.Data5u{Proxyer: internal.Proxyer{Url: "www.data5u.com/free/gnpt/index.shtml"}}                    //ok
	self.snatchs["Data5u_gwgn"] = &internal.Data5u{Proxyer: internal.Proxyer{Url: "www.data5u.com/free/gwgn/index.shtml"}}                    //ok
	self.snatchs["Data5u_gwpt"] = &internal.Data5u{Proxyer: internal.Proxyer{Url: "www.data5u.com/free/gwpt/index.shtml"}}                    //ok
	self.snatchs["Ip66_mo"] = &internal.Ip66{Proxyer: internal.Proxyer{Url: "www.66ip.cn/mo.php?tqsl=100"}}                                   //ok
	self.snatchs["Ip66_nmtq"] = &internal.Ip66{Proxyer: internal.Proxyer{Url: "www.66ip.cn/nmtq.php?getnum=100"}}                             //ok
	self.snatchs["Proxylistplus"] = &internal.Proxylistplus{Proxyer: internal.Proxyer{Url: "list.proxylistplus.com/Fresh-HTTP-Proxy-List-1"}} //ok
	self.snatchs["Xicidaili_nn"] = &internal.Xicidaili{Proxyer: internal.Proxyer{Url: "www.xicidaili.com/nn"}}                                //ok
	self.snatchs["Xicidaili_nt"] = &internal.Xicidaili{Proxyer: internal.Proxyer{Url: "www.xicidaili.com/nt"}}                                //ok

	for _, snatch := range self.snatchs {
		snatch.Init()
		tmp := snatch
		util.GoChan(func(cstop <-chan bool) {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(5000)))
			tmp.Start(self, tmp, cstop)
		})
	}
}
