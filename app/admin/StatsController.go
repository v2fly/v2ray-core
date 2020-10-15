// +build !confonly

package admin

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sort"
	"strings"
)

func init() {
	RegisterController("stats", &StatsController{})
}

type StatsController struct {
	admin *Server
}

type MapSortByKey struct {
	maps    []map[string]interface{}
	sortKey string
	order string
}

func (m *MapSortByKey) Len() int {
	return len(m.maps)
}
func (m *MapSortByKey) less(i, j int) bool {
	v1 := m.maps[i][m.sortKey]
	v2 := m.maps[j][m.sortKey]
	switch v1.(type) {
	default:
		return false
	case int:
		return v1.(int) <= v2.(int)
	case string:
		return v1.(string) <= v2.(string)
	}
}
func (m *MapSortByKey) Less(i, j int) bool {
	result := m.less(i, j)
	if strings.ToLower(m.order) == "desc" {
		return !result
	}
	return result
}
func (m *MapSortByKey) Swap(i, j int) {
	m.maps[i], m.maps[j] = m.maps[j], m.maps[i]
}

func (ctl *StatsController) InitRouter(admin *Server, httpRouter gin.IRouter) {
	ctl.admin = admin
	httpRouter.GET("/stats", ctl.GetStats)
	httpRouter.POST("/stats/reset", ctl.PutStats)
}
func (ctl *StatsController) PutStats(gCtx *gin.Context) {
	name := gCtx.PostForm("name")
	names := strings.Split(name, ",")
	for _, s := range names {
		ctl.admin.Rm.ResetCounter(s)
	}
	gCtx.JSON(200, gin.H{"status":"ok"})
}
func (ctl *StatsController) GetStats(gCtx *gin.Context) {
	var tableParams TableQuery
	gCtx.ShouldBind(&tableParams)

	rows := make([]map[string]interface{}, 0, 10)
	total := 0
	ctl.admin.Rm.Visit(func(s string, counter CounterRate) bool {
		if tableParams.Search!="" && strings.Index(s, tableParams.Search)==-1 {
			return true
		}
		statsInfo := make(map[string]interface{})
		statsInfo["name"] = s
		statsInfo["value"] = counter.Value()
		statsInfo["rate"] = counter.Rate()
		rows = append(rows, statsInfo)
		total += 1
		return true
	})
	sort.Sort(&MapSortByKey{rows,tableParams.Sort, tableParams.Order})
	gCtx.PureJSON(http.StatusOK, gin.H{
		"total": total,
		"rows":  rows,
	})
}
