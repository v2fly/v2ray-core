package restful_api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type StatsUser struct {
	uuid  string `form:"uuid" binging:"required_without=email,uuid4"`
	email string `form:"email" binging:"required_without=uuid,email"`
}

func statsUser(c *gin.Context) {
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

func stats(c *gin.Context) {
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

func loggerReboot(c *gin.Context)  {
	c.JSON(http.StatusOK, gin.H{})
}

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth[6:] != "token123" { // tip: Bearer: token123
			c.JSON(http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		c.Next()
	}
}

func Start() error {
	r := gin.New()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	v1 := r.Group("/v1")
	v1.Use(TokenAuthMiddleware())
	{
		v1.GET("/stats/user", statsUser)
		v1.GET("/stats", stats)
		v1.POST("/logger/reboot", loggerReboot)
	}

	if err := r.Run(":3000"); err != nil {
		return err
	}

	return nil
}
