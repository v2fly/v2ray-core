// +build !confonly

package admin

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"v2ray.com/core"
	"v2ray.com/core/app/stats"
	"v2ray.com/core/common"
	"v2ray.com/core/common/errors"
	"v2ray.com/core/common/log"
	"v2ray.com/core/features/outbound"
	"v2ray.com/core/features/policy"
	"v2ray.com/core/features/routing"
	featureStats "v2ray.com/core/features/stats"
)

var (
	controllers = make(map[string]Controller)
	// 启动v2ray的配置文件名
	ConfigFileName string
)

func RegisterController(controllerName string, controller Controller) {
	controllers[controllerName] = controller
}

type Controller interface {
	InitRouter(server *Server, httpRouter gin.IRouter)
}

type Server struct {
	srv      *http.Server
	config   *Config
	ohm      outbound.Manager
	router   routing.Router
	sm       *stats.Manager
	Rm       *RateManager
	Instance *core.Instance
}

// NewCommander creates a new Commander based on the given config.
func NewAdminServer(ctx context.Context, config *Config) (*Server, error) {
	g := &Server{
		config: config,
	}
	g.Instance = core.MustFromContext(ctx)
	if s, ok := ctx.Value("config_file").(string); ok {
		ConfigFileName = s
	}

	common.Must(core.RequireFeatures(ctx, func(om outbound.Manager, router routing.Router, pm policy.Manager, sm featureStats.Manager) {
		g.ohm = om
		g.router = router
		if appStats, ok := sm.(*stats.Manager); ok {
			g.sm = appStats
			g.Rm = &RateManager{sm: appStats, counters: make(map[string]*CounterRate)}
		}
	}))
	return g, nil
}

// Type implements common.HasType.
func (admin *Server) Type() interface{} {
	return (*Server)(nil)
}

// Start implements common.Runnable.
func (admin *Server) Start() error {
	gin.DisableConsoleColor()
	httpEngine := gin.Default()
	var httpRouter gin.IRouter = httpEngine
	if admin.config.ContextPath != "" && admin.config.ContextPath != "/" {
		httpRouter = httpRouter.Group(admin.config.ContextPath)
	}
	// public 指向静态文件
	httpRouter.Static("/public", admin.config.PublicPath)
	httpRouter.GET("/", func(c *gin.Context) {
		c.Redirect(302, admin.config.ContextPath+"/public/index.html")
	})
	httpRouter = httpRouter.Group("/api")
	httpRouter.OPTIONS("/*any", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "POST, PUT, DELETE, GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,XFILENAME,XFILECATEGORY,XFILESIZE,x-requested-with,Authorization")
		c.Status(200)
	})
	httpRouter.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "POST, PUT, DELETE, GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,XFILENAME,XFILECATEGORY,XFILESIZE,x-requested-with,Authorization")
	})
	if admin.config.Accounts != nil && len(admin.config.Accounts) > 0 {
		accounts := gin.Accounts{}
		for _, account := range admin.config.Accounts {
			accounts[account.UserName] = account.Password
		}
		httpRouter.Use(gin.BasicAuth(accounts))
	}

	for s, controller := range controllers {
		errors.New("try to init admin controller ", s).WriteToLog()
		// api指向handler
		controller.InitRouter(admin, httpRouter)
	}
	admin.srv = &http.Server{
		Addr:    admin.config.Addr,
		Handler: httpEngine,
	}
	log.Warn("start admin http server %s", admin.config.Addr)
	go func() {
		if err := admin.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("listen: %s", err)
		}
	}()
	if admin.Rm != nil {
		admin.Rm.Start()
	}

	return nil
}
func (admin *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := admin.srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown: %s", err)
	}
	if admin.Rm != nil {
		admin.Rm.Close()
	}
	return nil
}
func (admin *Server) Reload(ctx context.Context, cfg interface{}) error {

	go func() {
		log.Info("尝试关闭admin")
		admin.Close()
		log.Info("关闭admin功能成功")
		admin.config = cfg.(*Config)
		common.Must(core.RequireFeatures(ctx, func(om outbound.Manager, router routing.Router, pm policy.Manager, sm featureStats.Manager) {
			admin.ohm = om
			admin.router = router
			if appStats, ok := sm.(*stats.Manager); ok {
				admin.sm = appStats
				admin.Rm = &RateManager{sm: appStats, counters: make(map[string]*CounterRate)}
			}
			admin.Start()
			log.Info("重启成功")
		}))
	}()

	return nil
}

func (admin *Server) GetInitConfig() interface{} {
	return admin.config
}
