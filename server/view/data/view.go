package data

import (
	"KubePot/core/dbUtil"
	"KubePot/core/models"
	"KubePot/error"
	"KubePot/utils/conf"
	"KubePot/utils/log"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func Html(c *gin.Context) {
	attackCity := conf.Get("admin", "attack_city")
	c.HTML(http.StatusOK, "data.html", gin.H{
		"dataAttack": attackCity,
	})
}

type StatsResult struct {
	Name string
	Sum  int64
}

func GetChina(c *gin.Context) {
	var results []StatsResult
	err := dbUtil.GORM().Model(&models.KubePotInfo{}).
		Select("region as name, count(1) as sum").
		Where("country = ?", "中国").
		Group("region").
		Order("sum desc").
		Limit(8).
		Scan(&results).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "统计攻击地区失败", err)
	}

	var regionList []map[string]string
	for _, r := range results {
		regionMap := make(map[string]string)
		regionMap["name"] = r.Name
		regionMap["value"] = strconv.FormatInt(r.Sum, 10)
		regionList = append(regionList, regionMap)
	}

	data := map[string]interface{}{
		"regionList": regionList,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": data,
	})
}

func GetCountry(c *gin.Context) {
	var results []StatsResult
	err := dbUtil.GORM().Model(&models.KubePotInfo{}).
		Select("country as name, count(1) as sum").
		Where("country != ?", "").
		Group("country").
		Order("sum desc").
		Limit(8).
		Scan(&results).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "统计攻击地区失败", err)
	}

	var regionList []map[string]string
	for _, r := range results {
		regionMap := make(map[string]string)
		regionMap["name"] = r.Name
		regionMap["value"] = strconv.FormatInt(r.Sum, 10)
		regionList = append(regionList, regionMap)
	}

	data := map[string]interface{}{
		"regionList": regionList,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": data,
	})
}

func GetIp(c *gin.Context) {
	var results []StatsResult
	err := dbUtil.GORM().Model(&models.KubePotInfo{}).
		Select("ip as name, count(1) as sum").
		Where("ip != ?", "").
		Group("ip").
		Order("sum desc").
		Limit(10).
		Scan(&results).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "统计攻击IP失败", err)
	}

	var ipList []map[string]string
	for _, r := range results {
		ipMap := make(map[string]string)
		ipMap["name"] = r.Name
		ipMap["value"] = strconv.FormatInt(r.Sum, 10)
		ipList = append(ipList, ipMap)
	}

	data := map[string]interface{}{
		"ipList": ipList,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": data,
	})
}

func GetType(c *gin.Context) {
	var results []StatsResult
	err := dbUtil.GORM().Model(&models.KubePotInfo{}).
		Select("type as name, count(1) as sum").
		Where("type != ?", "").
		Group("type").
		Order("sum desc").
		Limit(10).
		Scan(&results).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "统计攻击类型失败", err)
	}

	var typeList []map[string]string
	for _, r := range results {
		typeMap := make(map[string]string)
		typeMap["name"] = r.Name
		typeMap["value"] = strconv.FormatInt(r.Sum, 10)
		typeList = append(typeList, typeMap)
	}

	data := map[string]interface{}{
		"typeList": typeList,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": data,
	})
}

func GetNewInfo(c *gin.Context) {
	var result []models.KubePotInfo
	err := dbUtil.GORM().Order("id desc").Limit(20).Find(&result).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "获取最新数据流失败", err)
	}

	data := map[string]interface{}{
		"result": result,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": data,
	})
}

func GetAccountInfo(c *gin.Context) {
	var results []StatsResult
	sql := "select account as name, count(0) as sum from Kubepot_passwd GROUP BY account ORDER BY sum desc"
	err := dbUtil.GORM().Raw(sql).Scan(&results).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "查询SQL失败", err)
	}

	var resultMap []map[string]string
	for _, r := range results {
		rMap := make(map[string]string)
		rMap["name"] = r.Name
		rMap["value"] = strconv.FormatInt(r.Sum, 10)
		resultMap = append(resultMap, rMap)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": resultMap,
	})
}

func GetPasswdInfo(c *gin.Context) {
	var results []StatsResult
	sql := "select password as name, count(0) as sum from Kubepot_passwd GROUP BY password ORDER BY sum desc"
	err := dbUtil.GORM().Raw(sql).Scan(&results).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "查询SQL失败", err)
	}

	var resultMap []map[string]string
	for _, r := range results {
		rMap := make(map[string]string)
		rMap["name"] = r.Name
		rMap["value"] = strconv.FormatInt(r.Sum, 10)
		resultMap = append(resultMap, rMap)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": resultMap,
	})
}

func GetWordInfo(c *gin.Context) {
	var results []StatsResult
	sql := "select region as name, count(1) as sum from Kubepot_info GROUP BY region"
	err := dbUtil.GORM().Raw(sql).Scan(&results).Error

	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "查询SQL失败", err)
	}

	var resultMap []map[string]string
	for _, r := range results {
		rMap := make(map[string]string)
		rMap["name"] = r.Name
		rMap["value"] = strconv.FormatInt(r.Sum, 10)
		resultMap = append(resultMap, rMap)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": resultMap,
	})
}

// 往下是 Web Socket 代码

// 存储全部客户端连接
var connClient = make(map[*websocket.Conn]bool)

// 去除跨域限制
var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 客户端连接
func Ws(c *gin.Context) {
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		// 创建 WebSocket 失败
		return
	}

	connClient[ws] = true

	defer ws.Close()

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			// 客户端断开
			connClient[ws] = false
			break
		}
	}
}

// 发送消息
func Send(data map[string]interface{}) {
	for k, v := range connClient {
		if v {
			err := k.WriteJSON(data)
			if err != nil {
				fmt.Println(err)
				break
			}
		}
	}
}

// 生成数据 JSON
func MakeDataJson(typex string, data map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"type": typex,
		"data": data,
	}

	return result
}
