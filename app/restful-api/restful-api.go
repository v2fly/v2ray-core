package restful_api

import (
	"github.com/gin-gonic/gin"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/transport/internet"
	"net/http"
	"strings"
)

type StatsUser struct {
	uuid  string `form:"uuid" binging:"required_without=email,uuid4"`
	email string `form:"email" binging:"required_without=uuid,email"`
}

func (r *restfulService) statsUser(c *gin.Context) {
	var statsUser StatsUser
	if err := c.BindQuery(&statsUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{
		"uplink":   1,
		"downlink": 1,
	})
}

type Stats struct {
	tag string `form:"tag" binging:"required,alpha,min=1,max=255"`
}

type StatsBound struct { // Better name?
	Uplink   int64 `json:"uplink"`
	Downlink int64 `json:"downlink"`
}

type StatsResponse struct {
	Inbound  StatsBound `json:"inbound"`
	Outbound StatsBound `json:"outbound"`
}

func (r *restfulService) statsRequest(c *gin.Context) {
	var stats Stats
	if err := c.BindQuery(&stats); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	response := StatsResponse{
		Inbound: StatsBound{
			Uplink:   1,
			Downlink: 1,
		},
		Outbound: StatsBound{
			Uplink:   1,
			Downlink: 1,
		}}

	c.JSON(http.StatusOK, response)
}

func (r *restfulService) loggerReboot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func (r *restfulService) TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(auth, prefix) {
			c.JSON(http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		auth = strings.TrimPrefix(auth, prefix)
		if auth != r.config.AuthToken { // tip: Bearer: token123
			c.JSON(http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		c.Next()
	}
}

func (r *restfulService) start() error {
	r.Engine = gin.New()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	v1 := r.Group("/v1")
	v1.Use(r.TokenAuthMiddleware())
	{
		v1.GET("/stats/user", r.statsUser)
		v1.GET("/stats", r.statsRequest)
		v1.POST("/logger/reboot", r.loggerReboot)
	}

	var listener net.Listener
	var err error
	address := net.ParseAddress(r.config.ListenAddr)

	switch {
	case address.Family().IsIP():
		listener, err = internet.ListenSystem(r.ctx, &net.TCPAddr{IP: address.IP(), Port: int(r.config.ListenPort)}, nil)
	case strings.EqualFold(address.Domain(), "localhost"):
		listener, err = internet.ListenSystem(r.ctx, &net.TCPAddr{IP: net.IP{127, 0, 0, 1}, Port: int(r.config.ListenPort)}, nil)
	default:
		return newError("restful api cannot listen on the address: ", address)
	}
	if err != nil {
		return newError("restful api cannot listen on the port ", r.config.ListenPort).Base(err)
	}

	r.listener = listener
	go func() {
		if err := r.RunListener(listener); err != nil {
			newError("unable to serve restful api").WriteToLog()
		}
	}()

	return nil
}
