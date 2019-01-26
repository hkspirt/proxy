# proxy
采用定时轮询的方式从各个网站公开数据爬取可用http代理地址，并进行检查可用性后，放入可用代理池(内存)。

发布时(2019-01-26)支持：
[Kuaidaili](https://www.kuaidaili.com/free/intr/)
[Data5u](http://www.data5u.com/free/index.shtml)
[66IP](http://www.66ip.cn/mo.php?tqsl=50) 
[Proxylistplus](https://list.proxylistplus.com/Fresh-HTTP-Proxy-List-1)
[Xicidaili](http://www.xicidaili.com/nn)

## 示例
```go
package main  
  
import (  
   "github.com/hkspirt/proxy"  
 "github.com/hkspirt/proxy/util" "time"  
  _ "github.com/hkspirt/proxy"  
)  
  
func main() {  
   util.GoChan(func(cstop <-chan bool) {  
      ticker := time.NewTicker(time.Second)  
      for {  
         select {  
         case <-cstop:  
            return  
         case <-ticker.C:  
            util.LogInfo("get one:%s", proxy.Proxy.GetOne())  
         }  
      }  
   })  
   util.WaitForSystemExit()  
}
```

参考：[芒果词源助手](https://github.com/lonng/etym)
