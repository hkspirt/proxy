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

type Proxylistplus struct {
	Proxyer
}

func (self *Proxylistplus) Init() {
	self.regexp = regexp.MustCompile(`<td>(\d+\.\d+\.\d+\.\d+)</td>(?s:.*?)<td>(\d+)</td>(?s:.*?)<td>.*</td>(?s:.*?)<td>.*</td>(?s:.*?)<td>.*</td>(?s:.*?)<td>(.*)</td>`)
}

func (self *Proxylistplus) load() {
	data := self.httpGet()
	proxies := self.regexp.FindAllSubmatch(data, -1)
	if len(proxies) < 1 {
		util.LogWarn("proxyer:%s regexp find failed", self.Url)
		return
	}

	var result []string
	for _, m := range proxies {
		method := "https"
		if strings.Contains(strings.ToLower(string(m[3])), "no") {
			method = "http"
		}
		result = append(result, fmt.Sprintf("%s://%s:%s", method, string(m[1]), string(m[2])))
	}
	self.addrs = result
	util.LogInfo("proxyer:%s load over len:%d", self.Url, len(result))
}
