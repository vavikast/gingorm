package main

import (
	"flag"
	"gingorm/controllers"
	"gingorm/helpers"
	"gingorm/models"
	"gingorm/system"
	"github.com/cihub/seelog"
	"github.com/claudiu/gocron"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
)

func main() {
	// 输出了指针类型configFilePath，实际默认获取了值“conf/conf.yaml”
	configFilePath := flag.String("C", "conf/conf.yaml", "config file path")
	// 输出了指针类型logConfigPath，实际默认获取了值“conf/seelog.xml”
	logConfigPath := flag.String("L", "conf/seelog.xml", "log config file path")
	//flag解析， 还没有想通通过flag的意义，后面看完代码再补充。
	flag.Parse()
	//更改默认配置文件，实际取值*logConfigPath就是“config/seelog.xml"
	logger, err := seelog.LoggerFromConfigAsFile(*logConfigPath)
	if err != nil {
		seelog.Critical("err parsing seelog config file", err)
		return
	}
	//替代原有日志输出格式
	seelog.ReplaceLogger(logger)
	defer seelog.Flush()
	//导入配置文件。 解析yaml文件格式的conf.yaml,获取其值，输出到configuration（*Configuration类型）
	/***
	输入的结果为
	configuration=&Configuration{
	github_clientid:  dd91df2447af1906c534
	github_clientsecret: b53fe662cc4014443bc2413d315d6a99c35887a4
	github_authurl: https://github.com/login/oauth/authorize?client_id=%s&scope=user:email&state=%s
	# 与github配置的回调地址一致
	github_redirecturl: http://localhost:8090/oauth2callback
	github_tokenurl: https://github.com/login/oauth/access_token
	session_secret: wblog
	public: static
	addr: :8090
	page_size: 5
	smms_fileserver: https://sm.ms/api/upload
	}
	***/
	if err := system.LoadConfiguration(*configFilePath); err != nil {
		seelog.Critical("err parsing config log file", err)
		return
	}
	//初始化数据库，将db赋值给全局声明DB,延迟数据库关闭.
	db, err := models.InitDB()
	if err != nil {
		seelog.Critical("err open databases", err)
		return
	}
	defer db.Close()

	//设置gin模式
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	//设置输出模板配置
	setTemplate(router)
	//设置session中间件
	setSessions(router)

	//使用shareData（）中间件
	router.Use(SharedData())

	//Periodic tasks
	//每一天执行一次CreateXMLSitemap
	//每7天执行一次backup
	gocron.Every(1).Day().Do(controllers.CreateXMLSitemap)
	gocron.Every(7).Days().Do(controllers.Backup)
	gocron.Start()

	//设置静态资源位置
	//router.Static("/static", filepath.Join(getCurrentDirectory(), "./static"))
	router.Static("/static", "./static")

	//设置访问错误路径状态
	router.NoRoute(controllers.Handle404)
	router.GET("/", controllers.IndexGet)
	router.GET("/index", controllers.IndexGet)
	router.GET("/rss", controllers.RssGet)

	// 默认条件已经设置为true，所以可以下面的操作
	//先是跳转到执行controllers.SignupGet，跳转到signup.html,然后页面form注册成功后跳转到controllers.SigninGet
	//默认开启了，注册即是管理员的身份
	if system.GetConfiguration().SignupEnabled {
		router.GET("/signup", controllers.SignupGet)
		router.POST("/signup", controllers.SignupPost)
	}
	// user signin and logout
	//登录
	//登出，登出后跳转到/signin页面
	//github认证退出
	router.GET("/signin", controllers.SigninGet)
	router.POST("/signin", controllers.SigninPost)
	router.GET("/logout", controllers.LogoutGet)
	router.GET("/oauth2callback", controllers.Oauth2Callback)
	router.GET("/auth/:authType", controllers.AuthGet)

	// captcha 获取验证码
	router.GET("/captcha", controllers.CaptchaGet)

	//访客组路由
	visitor := router.Group("/visitor")
	visitor.Use(AuthRequired())
	{
		//发布评论
		visitor.POST("/new_comment", controllers.CommentPost)
		//删除评论
		visitor.POST("/comment/:id/delete", controllers.CommentDelete)
	}

	// subscriber //访问订阅，激活订阅，取消订阅
	router.GET("/subscribe", controllers.SubscribeGet)
	router.POST("/subscribe", controllers.Subscribe)
	router.GET("/active", controllers.ActiveSubscriber)
	router.GET("/unsubscribe", controllers.UnSubscribe)

	//获取博文信息
	router.GET("/page/:id", controllers.PageGet)
	router.GET("/post/:id", controllers.PostGet)
	router.GET("/tag/:tag", controllers.TagGet)
	router.GET("/archives/:year/:month", controllers.ArchiveGet)

	//获取链接信息
	router.GET("/link/:id", controllers.LinkGet)

	//管理员页面
	authorized := router.Group("/admin")
	//使用认证中间件
	authorized.Use(AdminScopeRequired())
	{
		// index 索引
		authorized.GET("/index", controllers.AdminIndex)

		// image upload 图片上传
		authorized.POST("/upload", controllers.Upload)

		// page 博客管理
		authorized.GET("/page", controllers.PageIndex)
		authorized.GET("/new_page", controllers.PageNew)
		authorized.POST("/new_page", controllers.PageCreate)
		authorized.GET("/page/:id/edit", controllers.PageEdit)
		authorized.POST("/page/:id/edit", controllers.PageUpdate)
		authorized.POST("/page/:id/publish", controllers.PagePublish)
		authorized.POST("/page/:id/delete", controllers.PageDelete)

		// post 博客发布页面
		authorized.GET("/post", controllers.PostIndex)
		authorized.GET("/new_post", controllers.PostNew)
		authorized.POST("/new_post", controllers.PostCreate)
		authorized.GET("/post/:id/edit", controllers.PostEdit)
		authorized.POST("/post/:id/edit", controllers.PostUpdate)
		authorized.POST("/post/:id/publish", controllers.PostPublish)
		authorized.POST("/post/:id/delete", controllers.PostDelete)

		// tag 标签创建
		authorized.POST("/new_tag", controllers.TagCreate)

		//用户管理页面
		authorized.GET("/user", controllers.UserIndex)
		authorized.POST("/user/:id/lock", controllers.UserLock)

		// profile 配置
		authorized.GET("/profile", controllers.ProfileGet)
		authorized.POST("/profile", controllers.ProfileUpdate)
		authorized.POST("/profile/email/bind", controllers.BindEmail)
		authorized.POST("/profile/email/unbind", controllers.UnbindEmail)
		authorized.POST("/profile/github/unbind", controllers.UnbindGithub)

		// subscriber 订阅者
		authorized.GET("/subscriber", controllers.SubscriberIndex)
		authorized.POST("/subscriber", controllers.SubscriberPost)

		// link  链接
		authorized.GET("/link", controllers.LinkIndex)
		authorized.POST("/new_link", controllers.LinkCreate)
		authorized.POST("/link/:id/edit", controllers.LinkUpdate)
		authorized.POST("/link/:id/delete", controllers.LinkDelete)

		// comment 评论
		authorized.POST("/comment/:id", controllers.CommentRead)
		authorized.POST("/read_all", controllers.CommentReadAll)

		// backup  备份
		authorized.POST("/backup", controllers.BackupPost)
		authorized.POST("/restore", controllers.RestorePost)

		// mail 邮件
		authorized.POST("/new_mail", controllers.SendMail)
		authorized.POST("/new_batchmail", controllers.SendBatchMail)
	}

	router.Run(system.GetConfiguration().Addr)

}

func setTemplate(engine *gin.Engine) {

	funcMap := template.FuncMap{
		"dateFormat": helpers.DateFormat,
		"substring":  helpers.Substring,
		"isOdd":      helpers.IsOdd,
		"isEven":     helpers.IsEven,
		"truncate":   helpers.Truncate,
		"add":        helpers.Add,
		"minus":      helpers.Minus,
		"listtag":    helpers.ListTag,
	}

	engine.SetFuncMap(funcMap)
	//engine.LoadHTMLGlob(filepath.Join(getCurrentDirectory(), "views/**/*"))
	engine.LoadHTMLGlob( "views/**/*")
}

//setSessions initializes sessions & csrf middlewares
func setSessions(router *gin.Engine) {
	config := system.GetConfiguration()
	//https://github.com/gin-gonic/contrib/tree/master/sessions
	store := cookie.NewStore([]byte(config.SessionSecret))
	store.Options(sessions.Options{HttpOnly: true, MaxAge: 7 * 86400, Path: "/"}) //Also set Secure: true if using SSL, you should though
	router.Use(sessions.Sessions("gin-session", store))
	//https://github.com/utrack/gin-csrf
	/*router.Use(csrf.Middleware(csrf.Options{
		Secret: config.SessionSecret,
		ErrorFunc: func(c *gin.Context) {
			c.String(400, "CSRF token mismatch")
			c.Abort()
		},
	}))*/
}

//+++++++++++++ middlewares +++++++++++++++++++++++

//SharedData fills in common data, such as user info, etc...
//获取用户的信息，设置于用户配置
func SharedData() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if uID := session.Get(controllers.SESSION_KEY); uID != nil {
			user, err := models.GetUser(uID)
			if err == nil {
				c.Set(controllers.CONTEXT_USER_KEY, user)
			}
		}
		if system.GetConfiguration().SignupEnabled {
			c.Set("SignupEnabled", true)
		}
		c.Next()
	}
}

//AuthRequired grants access to authenticated users, requires SharedData middleware
func AdminScopeRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if user, _ := c.Get(controllers.CONTEXT_USER_KEY); user != nil {
			if u, ok := user.(*models.User); ok && u.IsAdmin {
				c.Next()
				return
			}
		}
		seelog.Warnf("User not authorized to visit %s", c.Request.RequestURI)
		c.HTML(http.StatusForbidden, "errors/error.html", gin.H{
			"message": "Forbidden!",
		})
		c.Abort()
	}
}

//认证要求中间件，认证不通过直接报错退出，认证成功则能够handle住
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if user, _ := c.Get(controllers.CONTEXT_USER_KEY); user != nil {
			if _, ok := user.(*models.User); ok {
				c.Next()
				return
			}
		}
		seelog.Warnf("User not authorized to visit %s", c.Request.RequestURI)
		c.HTML(http.StatusForbidden, "errors/error.html", gin.H{
			"message": "Forbidden!",
		})
		c.Abort()
	}
}

//func getCurrentDirectory() string {
//	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
//	if err != nil {
//		seelog.Critical(err)
//	}
//	return strings.Replace(dir, "\\", "/", -1)
//}

//func getCurrentDirectory() string {
//	return ""
//}
