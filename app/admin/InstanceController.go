// +build !confonly

package admin

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"os"
	"regexp"
	"time"
	"unsafe"
	"v2ray.com/core"
	"v2ray.com/core/common/cmdarg"
	"v2ray.com/core/common/log"
)

func init() {
	RegisterController("instance", &InstanceController{})
}

type InstanceController struct {
	admin *Server
}

var commentReg = regexp.MustCompile("//.*?[\r]?\n")

func (ctl *InstanceController) InitRouter(admin *Server, httpRouter gin.IRouter) {
	ctl.admin = admin
	httpRouter.GET("/server/reload", ctl.Reload)
	httpRouter.GET("/server/config", ctl.ReadConfig)
	httpRouter.POST("/server/config", ctl.UpdateConfig)
}
func (ctl *InstanceController) ReadConfig(gCtx *gin.Context) {
	configFile := getConfigFilePath()
	if getConfigFilePath() == "" {
		gCtx.Status(500)
		gCtx.Writer.WriteString("没有获取到v2ray启动配置文件。目前配置更新重启只支持使用json配置文件方式")
		return
	}

	ioutil.ReadFile(configFile)
	if bytes, err := ioutil.ReadFile(configFile); err == nil {
		gCtx.Header("Content-Type", "application/json;charset=utf-8")
		gCtx.Header("Access-Control-Allow-Origin", "*")
		s := *(*string)(unsafe.Pointer(&bytes))

		s = commentReg.ReplaceAllString(s, "\n")

		gCtx.Writer.WriteString(s)
	} else {
		gCtx.JSON(500, gin.H{"status": "读取文件失败"})
	}
}
func (ctl *InstanceController) UpdateConfig(gCtx *gin.Context) {
	if getConfigFilePath() == "" {
		gCtx.Status(500)
		gCtx.Writer.WriteString("没有获取到v2ray启动配置文件。目前配置更新重启只支持使用json配置文件方式")
		return
	}
	if reqBytes, err := ioutil.ReadAll(gCtx.Request.Body); err == nil {
		configFilePath := getConfigFilePath()
		tmpFileName := getConfigFilePath() + ".tmp"
		ioutil.WriteFile(tmpFileName, reqBytes, 0600)
		log.Info("写入配置文件成功：%s", tmpFileName)
		_, configErr := core.LoadConfig("json", tmpFileName, cmdarg.Arg{tmpFileName})
		if configErr != nil {
			gCtx.JSON(500, gin.H{"status": configErr.Error()})
			return
		}
		// 时间格式  https://blog.csdn.net/x356982611/article/details/87972400
		// 2006-01-02 15:04:05 年月日 时分秒的值是固定的
		os.Rename(configFilePath, configFilePath + "."+time.Now().Format("20060102150405"))
		os.Rename(tmpFileName, configFilePath)
		gCtx.JSON(200, gin.H{"status": "ok"})
	} else {
		gCtx.JSON(500, gin.H{"status": "读取数据失败"})
	}
}
func (ctl *InstanceController) Reload(gCtx *gin.Context) {
	if getConfigFilePath() == "" {
		gCtx.Status(500)
		gCtx.Writer.WriteString("没有获取到v2ray启动配置文件。目前配置更新重启只支持使用json配置文件方式")
		return
	}
	config, err := core.LoadConfig("json", getConfigFilePath(), cmdarg.Arg{getConfigFilePath()})
	if err != nil {
		log.Warn("failed to read config files: %s", getConfigFilePath())
		gCtx.JSON(500, gin.H{"status": "读取文件失败"})
		return
	}
	ctl.admin.Instance.Reload(config)
	gCtx.JSON(200, gin.H{"status": "ok"})

}

func fileExists(file string) bool {
	info, err := os.Stat(file)
	return err == nil && !info.IsDir()
}
func getConfigFilePath() string {
	return ConfigFileName
}
