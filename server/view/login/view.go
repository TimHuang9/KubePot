package login

import (
	"KubePot/core/dbUtil"
	"KubePot/core/models"
	kerr "KubePot/error"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Html(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

func Jump(c *gin.Context) {
	session := sessions.Default(c)
	loginCookie := session.Get("is_login")

	if loginCookie == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": kerr.ErrFailLoginCode,
			"msg":  "未登录",
		})
		c.Abort()
	} else {
		c.Next()
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": kerr.ErrFailLoginCode,
			"msg":  "参数错误",
		})
		return
	}

	var user models.KubePotUser
	db := dbUtil.GORM()
	err := db.Where("username = ?", req.Username).First(&user).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": kerr.ErrFailLoginCode,
			"msg":  "用户名或密码错误",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": kerr.ErrFailLoginCode,
			"msg":  "用户名或密码错误",
		})
		return
	}

	session := sessions.Default(c)
	session.Set("is_login", user.Username)
	session.Set("user_id", user.ID)
	session.Set("time", time.Now().Format("2006-01-02 15:04:05"))
	session.Save()

	c.JSON(http.StatusOK, gin.H{
		"code": kerr.ErrSuccessCode,
		"msg":  kerr.ErrSuccessMsg,
		"data": gin.H{
			"username": user.Username,
		},
	})
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.JSON(http.StatusOK, gin.H{
		"code": kerr.ErrSuccessCode,
		"msg":  "登出成功",
	})
}

func CheckLogin(c *gin.Context) {
	session := sessions.Default(c)
	loginCookie := session.Get("is_login")

	if loginCookie == nil {
		c.JSON(http.StatusOK, gin.H{
			"code": kerr.ErrFailLoginCode,
			"msg":  "未登录",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": kerr.ErrSuccessCode,
			"msg":  "已登录",
			"data": gin.H{
				"username": loginCookie,
			},
		})
	}
}

func InitDefaultUser() error {
	db := dbUtil.GORM()
	if db == nil {
		return nil
	}

	var count int64
	db.Model(&models.KubePotUser{}).Count(&count)
	if count > 0 {
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("kubepot"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()
	user := models.KubePotUser{
		Username:   "admin",
		Password:   string(hashedPassword),
		CreateTime: now,
		UpdateTime: now,
	}

	return db.Create(&user).Error
}
