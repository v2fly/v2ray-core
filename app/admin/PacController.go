// +build !confonly

package admin

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unsafe"
	"v2ray.com/core/common/log"
)

const GfwlistUrl = "https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt"

func init() {
	RegisterController("pac", &PacController{})
}

type PacController struct {
	admin *Server
}
type PacConfig struct {
	Proxy    string `json:"proxy"`
	UserRule string `json:"userRule"`
	GfwProxy string `json:"gfwProxy"`
}

func (ctl *PacController) InitRouter(admin *Server, httpRouter gin.IRouter) {
	ctl.admin = admin
	httpRouter.POST("/pac/gfwlist/download", ctl.UpdatePac)
	httpRouter.POST("/pac/geodat/download", ctl.DownloadGeoDat)
	httpRouter.POST("/pac/save", ctl.SavePac)
	httpRouter.GET("/pac", ctl.GetPac)
	httpRouter.GET("/pac/config", ctl.GetPacConfig)
}
func (ctl *PacController) GetPac(gCtx *gin.Context) {

	pacFile := getPacV2rayFile()
	if !fileExists(pacFile) {
		gCtx.Status(404)
		gCtx.Writer.WriteString("还没有生成pac文件")
		return
	}
	pacContent, err := ioutil.ReadFile(pacFile)
	if err != nil {
		gCtx.Status(500)
		gCtx.Writer.WriteString("读取pac文件失败" + err.Error())
		return
	}
	gCtx.Header("Content-Type", "application/x-ns-proxy-autoconfig; charset=utf-8")
	gCtx.Status(200)
	gCtx.Writer.Write(pacContent)
}
func (ctl *PacController) GetPacConfig(gCtx *gin.Context) {

	pacConfigFile := getPacConfigFile()
	if !fileExists(pacConfigFile) {
		log.Warn("pac配置文件%s不存在，返回空配置", pacConfigFile)
		gCtx.JSON(200, &PacConfig{})
		return
	}
	configContent, err := ioutil.ReadFile(pacConfigFile)
	if err != nil {
		log.Warn("读取pac配置文件%s失败，返回空配置", pacConfigFile)
		gCtx.JSON(200, &PacConfig{})
		return
	}
	gCtx.Header("Content-Type", "application/json; charset=utf-8")
	gCtx.Status(200)
	gCtx.Writer.Write(configContent)
}
func generateV2rayPac(proxy, userRule, gfwProxy string) error {
	pacGfwFile := getPacGfwFile()
	if !fileExists(pacGfwFile) {
		if err := downloadGfwPac(gfwProxy); err != nil {
			return err
		}
	}
	gfwContent, _ := ioutil.ReadFile(pacGfwFile)

	gfwContent = bytes.ReplaceAll(gfwContent, Str2Bytes("__PROXY__"), Str2Bytes(proxy))
	if userRule != "" {
		userRule = regexp.MustCompile("[,;\r\n]+").ReplaceAllString(userRule, "\",\n  \"")

		userRule = "var rules = [\n  \"" + userRule + "\","
		gfwContent = bytes.ReplaceAll(gfwContent, Str2Bytes("var rules = ["), Str2Bytes(userRule))
	}
	err := ioutil.WriteFile(getPacV2rayFile(), gfwContent, 0644)
	return err
}
func DownloadFile(gfwProxy, downloadUrl, localFileName string) error {
	var client *http.Client
	if gfwProxy == "" {
		client = http.DefaultClient
	} else {
		log.Info("使用代理%s下载%s", gfwProxy, downloadUrl)
		transport := &http.Transport{Proxy: func(_ *http.Request) (*url.URL, error) {
			return url.Parse(gfwProxy)
		}}
		client = &http.Client{Transport: transport}
	}
	resp, err := client.Get(downloadUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	localFile, err := os.Create(localFileName + ".tmp")
	defer localFile.Close()
	buf := make([]byte, 4096, 4096)
	var readNum int
	for {
		readNum, err = resp.Body.Read(buf[0:])
		if err != nil && err != io.EOF {
			return err
		}
		if readNum > 0 {
			if _, writeErr := localFile.Write(buf[0:readNum]); writeErr != nil {
				return writeErr
			}
		}
		if err == io.EOF {
			break
		}

	}
	localFile.Close()
	return os.Rename(localFileName+".tmp", localFileName)
}
func downloadGfwPac(gfwProxy string) error {
	var client *http.Client
	if gfwProxy == "" {
		client = http.DefaultClient
	} else {
		log.Info("使用代理%s下载gfwlist:%s", gfwProxy, GfwlistUrl)
		transport := &http.Transport{Proxy: func(_ *http.Request) (*url.URL, error) {
			return url.Parse(gfwProxy)
		}}
		client = &http.Client{Transport: transport}
	}

	resp, err := client.Get(GfwlistUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	gfwListContent, err := ioutil.ReadAll(resp.Body)
	// bytes 转化为string的高效方法
	gfwListBytes, err := base64.StdEncoding.DecodeString(*(*string)(unsafe.Pointer(&gfwListContent)))
	gwfScanner := bufio.NewScanner(bytes.NewReader(gfwListBytes[0:]))
	lines := make([]string, 0, 4096)
	lastModifyLine := ""
	// 跳过第一行 [AutoProxy 0.2.9]
	gwfScanner.Scan()
	for gwfScanner.Scan() {
		line := gwfScanner.Text()
		if strings.Contains(line, "! Last Modified:") {
			lastModifyLine = "// GFWList" + line[1:] + "\n"
		}
		// 跳过空行和注释行
		if line == "" || line[0] == '!' {
			continue
		}
		lines = append(lines, line)
	}
	rules, err := json.MarshalIndent(lines, "", "  ")
	rulesJson := *(*string)(unsafe.Pointer(&rules))
	tplContent, err := ioutil.ReadFile(getTemplateFile())
	if err != nil {
		return err
	}
	pacString := strings.Replace(*(*string)(unsafe.Pointer(&tplContent)), "__RULES__", rulesJson, 1)
	err = ioutil.WriteFile(getPacGfwFile(), Str2Bytes(lastModifyLine+pacString), 0644)
	return err
}
func (ctl *PacController) UpdatePac(gCtx *gin.Context) {
	config := PacConfig{}
	gCtx.ShouldBindJSON(&config)

	err := downloadGfwPac(config.GfwProxy)
	if err != nil {
		gCtx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if config.Proxy != "" {
		saveConfig(&config)
		generateV2rayPac(config.Proxy, config.UserRule, config.GfwProxy)
	}
	gCtx.JSON(200, gin.H{"msg": "下载gfwlist成功"})
	return

}
func (ctl *PacController) DownloadGeoDat(gCtx *gin.Context) {
	config := PacConfig{}
	gCtx.ShouldBindJSON(&config)

	err := DownloadFile(config.GfwProxy, "https://github.com/v2fly/geoip/raw/release/geoip.dat",
		GetExecutableDir()+"/geoip.dat")
	if err != nil {
		gCtx.JSON(500, gin.H{"error": "geoip.dat下载出错"+err.Error()})
		return
	}
	err = DownloadFile(config.GfwProxy, "https://github.com/v2fly/domain-list-community/raw/release/dlc.dat",
		GetExecutableDir()+"/geosite.dat")
	if err != nil {
		gCtx.JSON(500, gin.H{"error": "geosite.dat下载出错"+err.Error()})
		return
	}

	gCtx.JSON(200, gin.H{"msg": "下载geoip.data,geosite.dat成功"})
	return

}
func saveConfig(config *PacConfig) error {
	configJson, _ := json.MarshalIndent(config, "", "  ")
	configFileName := getPacConfigFile()
	log.Info("写入pac自自定义信息到文件%s", configFileName)
	ioutil.WriteFile(configFileName, configJson, 0644)
	return ioutil.WriteFile(configFileName, configJson, 0644)
}
func (ctl *PacController) SavePac(gCtx *gin.Context) {
	config := PacConfig{}
	gCtx.BindJSON(&config)
	saveConfig(&config)
	err := generateV2rayPac(config.Proxy, config.UserRule, config.GfwProxy)
	if err != nil {
		gCtx.Status(500)
		gCtx.Writer.WriteString(err.Error())
		return
	}
	gCtx.JSON(200, gin.H{"msg": "生成pac成功"})
	return

}
func GetExecutableDir() string {
	exec, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exec)
}
func getTemplateFile() string {
	return GetExecutableDir() + "/pac_template.js"
}
func getPacGfwFile() string {
	return GetExecutableDir() + "/pac_gfw.js"
}
func getPacV2rayFile() string {
	return GetExecutableDir() + "/pac_v2ray.js"
}
func getPacConfigFile() string {
	return GetExecutableDir() + "/pac_config.json"
}

func Str2Bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func Bytes2Str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
