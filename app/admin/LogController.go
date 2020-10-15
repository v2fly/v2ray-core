// +build !confonly

package admin

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"v2ray.com/core/common/log"
)

func init() {
	RegisterController("log", &LogController{})
}

type ConfigLog struct {
	AccessLog string `json:"access"`
	ErrorLog  string `json:"error"`
	LogLevel  string `json:"loglevel"`
}
type V2rayConfig struct {
	Log *ConfigLog `json:"log"`
}
type LogController struct {
	admin *Server
}

func (c *V2rayConfig) getAccessLog() string {
	if c.Log != nil {
		return c.Log.AccessLog
	}
	return ""
}
func (c *V2rayConfig) getErrorLog() string {
	if c.Log != nil {
		return c.Log.ErrorLog
	}
	return ""
}

func (ctl *LogController) InitRouter(admin *Server, httpRouter gin.IRouter) {
	ctl.admin = admin
	httpRouter.GET("/log", ctl.ReadLog)
}
func (ctl *LogController) ReadLog(gCtx *gin.Context) {
	from, err := strconv.ParseInt(gCtx.DefaultQuery("from", "-1"), 0, 64)
	logType := gCtx.DefaultQuery("logType", "access")
	if err != nil {
		log.Error("转化请求参数from失败")
		gCtx.Status(500)
		gCtx.Writer.WriteString("转化请求参数from失败")
		return
	}
	to := from + 4096
	configFilePath := getConfigFilePath()

	if configBytes, err := ioutil.ReadFile(configFilePath); err == nil {
		c := &V2rayConfig{}
		err = json.Unmarshal(configBytes, c)
		if err != nil {
			gCtx.Status(500)
			gCtx.Writer.WriteString("解析v2ray日志配置信息失败")
			return
		}
		var logFile string
		if "access" == logType {
			logFile = c.getAccessLog()
		} else {
			logFile = c.getErrorLog()
		}
		log.Info("获取到%s日志文件为:%s", logType, logFile)
		if logFile == "" {
			gCtx.Status(204)
			gCtx.Writer.WriteString("没有设置访问日志")
			return
		}
		content, lastPos, err := readFile(logFile, from, to)
		if err != nil {
			gCtx.Status(500)
			gCtx.Writer.WriteString(err.Error())
			return
		}
		gCtx.JSON(200, gin.H{"lastPos": lastPos, "content": content})
		return
	}
	gCtx.Status(500)
	gCtx.Writer.WriteString("解析v2ray参数配置失败")
	return

}
func readFile(fileName string, from, to int64) (string, int64, error) {
	f, err := os.Open(fileName)

	defer f.Close()
	if err != nil {
		return "", 0, err
	}
	stat, err := f.Stat()
	if err != nil {
		return "", 0, err
	}
	if from == -1 {
		to = stat.Size()
		// 默认读取4k
		from = to - 4096
		if from < 0 {
			from = 0
		}
	}
	if from > stat.Size() {
		from = stat.Size()
	}
	if to > stat.Size() {
		to = stat.Size()
	}
	readBytes := make([]byte, to-from, to-from)
	totalNum := 0
	num := 0
	for err == nil {
		num, err = f.ReadAt(readBytes[totalNum:], from+int64(totalNum))
		totalNum += num
		if int64(totalNum) == (to-from) || err == io.EOF {
			// return string(readBytes), from+int64(totalNum), nil
			err = nil
			break
		}
	}
	if err != nil {
		return "", 0, nil
	}
	firstLineEndIdx := -1
	if from > 0 {
		firstLineEndIdx = bytes.IndexByte(readBytes, byte('\n'))
	}

	// 最尾巴的\n 留到下一次再读取
	lastLineEndIdx := bytes.LastIndexByte(readBytes, byte('\n'))

	return string(readBytes[firstLineEndIdx+1 : lastLineEndIdx+1]), from + int64(lastLineEndIdx-1), nil
}
