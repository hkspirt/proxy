//----------------
//Func  :
//Author: xjh
//Date  : 2019/
//Note  :
//----------------
package internal

import (
	"github.com/hkspirt/proxy/util"
	"fmt"
	"io/ioutil"
	"net/http"
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
	req, err := http.NewRequest(http.MethodGet, self.Url, nil)
	if err != nil {
		util.LogWarn("proxyer:%s new request err:%v", self.Url, err)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		util.LogWarn("proxyer:%s http get err:%v", self.Url, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		util.LogWarn("proxyer:%s http status err:%v", self.Url, resp.StatusCode)
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		util.LogWarn("proxyer:%s http read err:%v", self.Url, err)
		return
	}

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
