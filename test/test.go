//Copyright (C) Mr.Pungle

package main

import (
	"fmt"
	"github.com/pungle/dawn/logging"
	"github.com/pungle/dawn/web"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var msgChans map[string]chan string

func TestIndex(ctx *web.HttpContext) {
	resp := ctx.Response
	resp.Write([]byte("Hello world~"))
}

func TestSession(ctx *web.HttpContext) {
	session := ctx.Session()
	var count int
	if session != nil {
		data, _ := session.Get("count")
		count = int(data.(float64))
	} else {
		session = ctx.NewSession()
	}

	session.Set("count", count+1)
	ctx.SaveSession()
	ctx.Response.Write([]byte(fmt.Sprintf("hello: %d", count)))
}

func TestReceiveMsg(ctx *web.HttpContext) {
	req := ctx.Request
	resp := ctx.Response
	querys := req.URL.Query()
	id := querys.Get("id")
	channel, ok := msgChans[id]
	if !ok {
		channel = make(chan string, 10)
		msgChans[id] = channel
	}
	select {
	case <-time.After(time.Second * 50):
		resp.Write([]byte("timeout"))
	case msg := <-channel:
		resp.Write([]byte(fmt.Sprintf("Get: %s", msg)))
	}
}

func TestSendMsg(ctx *web.HttpContext) {
	req := ctx.Request
	resp := ctx.Response
	querys := req.URL.Query()
	id := querys.Get("id")
	msg := querys.Get("msg")
	channel, ok := msgChans[id]
	if !ok {
		channel = make(chan string, 10)
		msgChans[id] = channel
	}
	channel <- msg
	resp.Write([]byte("ok"))
}

func RegexpUrlTest(ctx *web.HttpContext) {
	resp := ctx.Response
	data := fmt.Sprintf("id=%s, name=%s, age=%s", ctx.GetVar("id"), ctx.GetVar("name"), ctx.GetVar("age"))
	resp.Write([]byte(data))
}

func main() {
	runtime.GOMAXPROCS(2)
	driver := web.NewRedisSessionDriver("tcp", ":6379", 100, 1000, 60*time.Second)
	sessionCtx := web.NewSessionContext(
		driver,
		"sid",      // cookie sessionid
		"test.com", // cookie domain
		time.Duration(3600*24*365)*time.Second, //cookie Expire
		"/",   // cookie Path
		true,  // cookie httponly
		false, // cookie secure
		time.Duration(3600*24*7)*time.Second, // session age(server)
	)

	//errfile, _ := os.OpenFile("/data/logs/error.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	logging.StdOpen()
	//logging.SetHandler(errfile)
	logging.SetFlags(logging.F_DATE | logging.F_LEVEL | logging.F_SHORT_FILE)
	logging.SetLevel(logging.L_ERROR)

	f, _ := os.OpenFile("/data/logs/access.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	handler := logging.NewBuffHandler(f, 1024)
	config := web.NewConfig(":80", web.DEFAULT_LOG_FLAG, web.DEFAULT_LOG_LEVEL, false, "", "")
	server := web.NewServer(config, sessionCtx, handler)

	server.AddHandler("/", TestIndex)
	server.AddHandler("/counter", TestSession)
	server.AddHandler("/get", TestReceiveMsg)
	server.AddHandler("/set", TestSendMsg)
	server.AddHandler("^/test/{id :[0-9]+}$/article/{name: [a-zA-Z]+}$/page/{age: [0-9]{2}}$/", RegexpUrlTest)

	msgChans = make(map[string]chan string)

	go server.Start()
	waitForExit(server)
}

func waitForExit(server *web.HttpServer) {
	// 参考: http://en.wikipedia.org/wiki/Unix_signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGTSTP, syscall.SIGQUIT)
	println("waiting...")
	s := <-c
	println("Got signal: ", s.String())
	logging.Std().Close()
	server.Close()
	println("Terminated")
}
