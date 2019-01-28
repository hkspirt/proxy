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
	"github.com/robertkrimen/otto"
	"regexp"
	"strings"
)

type Ip66 struct {
	Proxyer
	param   *regexp.Regexp
	fun     *regexp.Regexp
	cookie  *regexp.Regexp
	jsRuner *otto.Otto
}

func (self *Ip66) Init() {
	self.regexp = regexp.MustCompile(`\d+\.\d+\.\d+\.\d+:\d+`)
	self.param = regexp.MustCompile(`setTimeout\("[a-z]+\(([0-9]+)\)"`)
	self.fun = regexp.MustCompile(`function ([a-z]+)(.|\n|\r)*}`)
	self.cookie = regexp.MustCompile(`_ydclearance=[0-9a-z_\-]*;`)
	self.jsRuner = otto.New()
}

func (self *Ip66) load() {
	data := self.httpGet()
	self.cookies = self.getCookie(string(data))
	proxies := self.regexp.FindAll(data, -1)
	if len(proxies) < 1 {
		util.LogWarn("proxyer:%s regexp find failed", self.Url)
		return
	}

	var result []string
	for _, line := range proxies {
		result = append(result, fmt.Sprintf("http://%s", string(line)))
	}
	self.addrs = result
	util.LogInfo("proxyer:%s load over len:%d", self.Url, len(result))
}

func (self *Ip66) getCookie(data string) string {
	fun := self.fun.FindSubmatch([]byte(data))
	if len(fun) < 2 {
		return self.cookies
	}
	funstr := strings.Replace(string(fun[0]), `eval("qo=eval;qo(po);");`, "return po;", 1)
	funstr = fmt.Sprintf("%s po = %s(%s);", funstr, fun[1], self.param.FindSubmatch([]byte(data))[1])
	self.jsRuner.Run(funstr)
	ret, err := self.jsRuner.Get("po")
	if err != nil {
		util.LogInfo("run js failed err:%v data:%s", err, data)
		return ""
	}
	cookie := self.cookie.Find([]byte(ret.String()))
	util.LogInfo("cookie:%s", cookie)
	return string(cookie)
}
