package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"testing"

	pool "github.com/qydysky/part/pool"
	reqf "github.com/qydysky/part/reqf"
	web "github.com/qydysky/part/web"
)

func main() {

	var (
		addr       = flag.String("addr", "", "服务地址(0.0.0.0:10002)")
		proxy      = flag.String("proxy", "", "http代理地址(http://127.0.0.1:10807)")
		concurrent = flag.Int("concurrent", 10, "并发限制")
		to         = flag.Int("to", 5000, "超时ms")
		wto        = flag.Int("wto", 5000, "读超时ms")
	)
	testing.Init()
	flag.Parse()

	if *addr == "" {
		log.Printf("[MAIN]: must set addr!\n")
		return
	}

	var webpath web.WebPath
	var wlimit web.Limits

	limitItem := web.NewLimitItem(*concurrent)
	limitItem.Cidr("0.0.0.0/0")
	wlimit.AddLimitItem(limitItem)

	reqPool := pool.New[reqf.Req](
		func() *reqf.Req { return reqf.New() },
		func(r *reqf.Req) bool { return r.IsLive() },
		func(r *reqf.Req) *reqf.Req { return r },
		func(r *reqf.Req) *reqf.Req { return r },
		*concurrent,
	)

	webpath.Store(`/`, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		p := strings.Split(r.URL.Path[1:], "_")
		if p[0] != "http" && p[0] != "https" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if _, e := strconv.Atoi(p[2]); e != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if wlimit.AddCount(r) {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		req := reqPool.Get()
		defer reqPool.Put(req)

		if e := req.Reqf(reqf.Rval{
			Url:         p[0] + "://" + p[1] + ":" + p[2] + "/announce?" + r.URL.RawQuery,
			Timeout:     *to,
			WriteLoopTO: *wto,
			Proxy:       *proxy,
		}); e != nil {
			log.Printf("%s -> %v\n", r.URL.Path[1:], e)
			w.WriteHeader(http.StatusServiceUnavailable)
		} else if _, e := w.Write(req.Respon); e != nil {
			log.Printf("%s -> %v\n", r.URL.Path[1:], e)
		}
	})

	ser := web.NewSyncMap(&http.Server{
		Addr: *addr,
	}, &webpath)

	var interrupt = make(chan os.Signal, 1)
	//捕获ctrl+c退出
	signal.Notify(interrupt, os.Interrupt)

	log.Printf("Running!\neg http:t.acg.rip:6699 -> http://192.168.31.101:10002/http_t.acg.rip_6699\n")
	<-interrupt
	log.Printf("Shutdown!\n")
	ser.Shutdown()
}
