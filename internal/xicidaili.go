//----------------
//Func  :
//Author: xjh
//Date  : 2019/
//Note  :
//----------------
package internal

import (
	"fmt"
	"github.com/hkspirt/proxy/util"
	"regexp"
	"strings"
)

type Xicidaili struct {
	Proxyer
}

func (self *Xicidaili) Init() {
	self.regexp = regexp.MustCompile(`<td>(\d+\.\d+\.\d+\.\d+)</td>(?s:.*?)<td>(\d+)</td>(?s:.*?)<td>(HTTPS?)</td>`)
}

func (self *Xicidaili) load() {
	data := self.httpGet()
	proxies := self.regexp.FindAllSubmatch(data, -1)
	if len(proxies) < 1 {
		util.LogWarn("proxyer:%s regexp find failed", self.Url)
		return
	}

	var result []string
	for _, m := range proxies {
		result = append(result, fmt.Sprintf("%s://%s:%s", strings.ToLower(string(m[3])), string(m[1]), string(m[2])))
	}
	self.addrs = result
	util.LogInfo("proxyer:%s load over len:%d", self.Url, len(result))
}
