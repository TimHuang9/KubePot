package dashboard

import (
	"KubePot/core/dbUtil"
	"KubePot/error"
	"KubePot/utils/conf"
	"KubePot/utils/log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type HourlyStats struct {
	Hour string
	Sum  int
}

func Html(c *gin.Context) {
	var webSum, sshSum, redisSum, mysqlSum, telnetSum, ftpSum, memCacheSum, httpSum, tftpSum, esSum, vncSum int64

	db := dbUtil.GORM()
	db.Table("Kubepot_info").Where("type = ?", "WEB").Count(&webSum)
	db.Table("Kubepot_info").Where("type = ?", "SSH").Count(&sshSum)
	db.Table("Kubepot_info").Where("type = ?", "REDIS").Count(&redisSum)
	db.Table("Kubepot_info").Where("type = ?", "MYSQL").Count(&mysqlSum)
	db.Table("Kubepot_info").Where("type = ?", "TELNET").Count(&telnetSum)
	db.Table("Kubepot_info").Where("type = ?", "FTP").Count(&ftpSum)
	db.Table("Kubepot_info").Where("type = ?", "MEMCACHE").Count(&memCacheSum)
	db.Table("Kubepot_info").Where("type = ?", "HTTP").Count(&httpSum)
	db.Table("Kubepot_info").Where("type = ?", "TFTP").Count(&tftpSum)
	db.Table("Kubepot_info").Where("type = ?", "ES").Count(&esSum)
	db.Table("Kubepot_info").Where("type = ?", "VNC").Count(&vncSum)

	mysqlStatus := conf.Get("mysql", "status")
	redisStatus := conf.Get("redis", "status")
	sshStatus := conf.Get("ssh", "status")
	webStatus := conf.Get("web", "status")
	apiStatus := conf.Get("api", "status")
	telnetStatus := conf.Get("telnet", "status")
	ftpStatus := conf.Get("ftp", "status")
	memCacheStatus := conf.Get("mem_cache", "status")
	httpStatus := conf.Get("http", "status")
	tftpStatus := conf.Get("tftp", "status")
	esStatus := conf.Get("elasticsearch", "status")
	vncStatus := conf.Get("vnc", "status")

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"webSum":         webSum,
		"sshSum":         sshSum,
		"redisSum":       redisSum,
		"mysqlSum":       mysqlSum,
		"telnetSum":      telnetSum,
		"ftpSum":         ftpSum,
		"memCacheSum":    memCacheSum,
		"httpSum":        httpSum,
		"tftpSum":        tftpSum,
		"esSum":          esSum,
		"vncSum":         vncSum,
		"webStatus":      webStatus,
		"sshStatus":      sshStatus,
		"redisStatus":    redisStatus,
		"mysqlStatus":    mysqlStatus,
		"apiStatus":      apiStatus,
		"telnetStatus":   telnetStatus,
		"ftpStatus":      ftpStatus,
		"memCacheStatus": memCacheStatus,
		"httpStatus":     httpStatus,
		"tftpStatus":     tftpStatus,
		"esStatus":       esStatus,
		"vncStatus":      vncStatus,
	})
}

func getHourlyData(attackType string) map[string]interface{} {
	type Result struct {
		Hour string
		Sum  int
	}

	var results []Result
	sql := `
		SELECT
			DATE_FORMAT(create_time,"%H") AS hour,
			sum(1) AS sum
		FROM
			Kubepot_info
		WHERE
			create_time >= (NOW() - INTERVAL 24 HOUR)
		AND type = ?
		GROUP BY
			hour
	`

	err := dbUtil.GORM().Raw(sql, attackType).Scan(&results).Error
	if err != nil {
		log.Pr("KubePot", "127.0.0.1", "查询SQL失败", err)
	}

	resultMap := make(map[string]interface{})
	for _, r := range results {
		if r.Hour != "" {
			resultMap[r.Hour] = r.Sum
		}
	}

	return resultMap
}

func GetFishData(c *gin.Context) {
	webMap := getHourlyData("WEB")
	sshMap := getHourlyData("SSH")
	redisMap := getHourlyData("REDIS")
	mysqlMap := getHourlyData("MYSQL")
	ftpMap := getHourlyData("FTP")
	telnetMap := getHourlyData("TELNET")
	memCacheMap := getHourlyData("MEMCACHE")
	httpMap := getHourlyData("HTTP")
	tftpMap := getHourlyData("TFTP")
	esMap := getHourlyData("ES")
	vncMap := getHourlyData("VNC")
	kubeletMap := getHourlyData("Kubelet")
	dockerMap := getHourlyData("Docker")
	etcdMap := getHourlyData("Etcd")
	apiserverMap := getHourlyData("Apiserver")
	bashMap := getHourlyData("Bash")

	data := map[string]interface{}{
		"web":          webMap,
		"ssh":          sshMap,
		"redis":        redisMap,
		"mysql":        mysqlMap,
		"ftp":          ftpMap,
		"telnet":       telnetMap,
		"memCache":     memCacheMap,
		"httpMap":      httpMap,
		"tftpMap":      tftpMap,
		"vncMap":       vncMap,
		"esMap":        esMap,
		"kubeletMap":   kubeletMap,
		"dockerMap":    dockerMap,
		"etcdMap":      etcdMap,
		"apiserverMap": apiserverMap,
		"bashMap":      bashMap,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": data,
	})
}

type PieStats struct {
	Country string
	Ip      string
	Sum     int64
}

func GetFishPieData(c *gin.Context) {
	db := dbUtil.GORM()

	var regionResults []PieStats
	errRegion := db.Table("Kubepot_info").
		Select("country, count(1) as sum").
		Where("country != ?", "").
		Group("country").
		Order("sum desc").
		Limit(10).
		Scan(&regionResults).Error

	if errRegion != nil {
		log.Pr("KubePot", "127.0.0.1", "统计攻击地区失败", errRegion)
	}

	var regionList []map[string]string
	for _, r := range regionResults {
		regionMap := make(map[string]string)
		if r.Country != "" {
			regionMap["name"] = r.Country
		} else {
			regionMap["name"] = "未知"
		}
		regionMap["value"] = strconv.FormatInt(r.Sum, 10)
		regionList = append(regionList, regionMap)
	}

	var ipResults []PieStats
	errIp := db.Table("Kubepot_info").
		Select("ip, count(1) as sum").
		Where("ip != ?", "").
		Group("ip").
		Order("sum desc").
		Limit(10).
		Scan(&ipResults).Error

	if errIp != nil {
		log.Pr("KubePot", "127.0.0.1", "统计攻击IP失败", errIp)
	}

	var ipList []map[string]string
	for _, r := range ipResults {
		ipMap := make(map[string]string)
		if r.Ip != "" {
			ipMap["name"] = r.Ip
		} else {
			ipMap["name"] = "未知"
		}
		ipMap["value"] = strconv.FormatInt(r.Sum, 10)
		ipList = append(ipList, ipMap)
	}

	data := map[string]interface{}{
		"regionList": regionList,
		"ipList":     ipList,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": error.ErrSuccessCode,
		"msg":  error.ErrSuccessMsg,
		"data": data,
	})
}
