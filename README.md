# dawn

## Abstract
dawn一个轻量级HTTP框架，基于原生http框架进行了简单的封装，目前它能实现了URL匹配（精确匹配「=」，前缀匹配「～」和正则表达式匹配「\^」），另外实现了Session模块和Logging模块。
## Usage
安装dawn只需要使用`go get`然后使用`import`把所需要的包导入到代码中即可使用:

	import (
		"github.com/pungle/dawn/web"
		"github.com/pungle/dawn/logging"
	)
`web.NewConfig`函数返回`web.HttpConfig`对象，`web.HttpServer`需要使用`web.HttpConfig`对象才能创建，`web.HttpConfig`记录了`web.HttpServer`的全部配置信息。
如需响应指定URL可通过`server.AddHandler`方法注册对应的`web.Handler`，`web.Handler`要求只有一个`web.HttpContext`参数的函数。

	type Handler func(ctx *HttpContext)
`web.HttpContext`包涵了当次HTTP响应过程的上下文件内容，包括Request, Response和Session，其中`ctx.Request = *http.Request`，`ctx.Response = http.ResponseWriter`

	package main

	import (
		"github.com/pungle/dawn/web"
	)

	func TestIndex(ctx *web.HttpContext) {
		req := ctx.Request
		resp := ctx.Response
		resp.Write([]byte("hello world"))
	}

	func main() {
		config := web.NewConfig(":80", web.DEFAULT_LOG_FLAG, web.DEFAULT_LOG_LEVEL, false, "", "")
		server := web.NewServer(config, nil, nil)
		server.AddHandler("/", TestIndex)
		server.ListenAndServe()
	}
dawn提供了三种URI响应方式「固定URI映射(mapping)，前缀优先(prefix)和模板匹配(match)」，其优先级为：`mapping > prefix > match`，如果使用了`match`方法响应，你可以通过`ctx.GetVar`方法得到之前你在URI定义时所写的变量名字，如：`{id: [0-9]+}`将可以通过`ctx.GetVar("id")`得到相应的内容，匹配方式由冒号后的正则表达式所决定。

	package main

	import (
		"github.com/pungle/dawn/web"
	)

	func MappingUrlTest(ctx *web.HttpContext) {
		resp := ctx.Response
		resp.Write([]byte("Hi mapping handler"))
	}

	func PrefixUrlTest(ctx *web.HttpContext) {
		resp := ctx.Response
		resp.Write([]byte("Hi prefix handler"))
	}

	func RegexpUrlTest(ctx *web.HttpContext) {
		resp := ctx.Response
		id := ctx.GetVar("id")
		name := ctx.GetVar("name")
		data := fmt.Sprintf("Hi regexp, vars = [id=%s, name=%s, age=%s]", id, name)
		resp.Write([]byte(data))
	}

	func main() {
		config := web.NewConfig(":80", web.DEFAULT_LOG_FLAG, web.DEFAULT_LOG_LEVEL, false, "", "")
		server := web.NewServer(config, nil, nil)
		server.AddHandler("article/2345/name/mapping_test", MappingUrlTest)
		server.AddHandler("~/article/2345/name", PrefixUrlTest)
		server.AddHandler("^/article/{id :\d+}$/name/{name:[a-z]+}$", RegexpUrlTest)
		server.ListenAndServe()
	}

dawn还提供了session的支持但这不是必选项，用户可以根据需要来加入session。session的配置需要通过构造一个`web.SessionContext`对象来创建，`web.SessionContext`包涵了session的相关配置信息。其中`driver`参数可以使用我们提供的`web.NewRedisSessionDriver`，如果你需要使用别的存储方式你也可以自己实现一个｀driver｀，只要符合以下接口即可：

	type SessionDriver interface {
		Get(string) (interface{}, error)
		Set(string, interface{}, time.Duration) error
	}
需要获取session实例可以通过`ctx.Session()`获得，如果取得`session`为`nil`那就代表当前没有有效的`session`信息，如果需要生成新`session`可以通过`ctx.NewSession()`方法生成并替换当前`session`，`session`生成后不会马上保存，当需要保存时可以使用`ctx.SaveSession()`把当前`session`保存到`driver`指定的存储器中并生成`cookie`

	package main

	import (
		"github.com/pungle/dawn/web"
	)

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
		ctx.Response.Write([]byte(fmt.Sprintf("count: %d", count)))
	}

	func main() {
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
		config := web.NewConfig(":80", web.DEFAULT_LOG_FLAG, web.DEFAULT_LOG_LEVEL, false, "", "")
		server := web.NewServer(config, sessionCtx, nil)
		server.AddHandler("/counter", TestSession)
		server.ListenAndServe()
	}
> 注意：如果没有给`web.HttpServer`配置`web.SessionContext`参数，以上操作均会抛出异常`web.ErrSessionNotSetup`，使用`session`前请确保配置是否正确，以免带来不必要的问题。


待续.....
