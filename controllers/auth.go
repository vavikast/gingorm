package controllers

import (
	"fmt"
	"net/http"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gingorm/helpers"
	"gingorm/system"
)


//认证写入seession并认证跳转
func AuthGet(c *gin.Context) {
	authType := c.Param("authType") //获取authType字段
	//设置session_id
	session := sessions.Default(c)
	//生成随机的uuid
	uuid := helpers.UUID()
	session.Delete(SESSION_GITHUB_STATE)
	session.Set(SESSION_GITHUB_STATE, uuid)
	session.Save()

	authurl := "/signin"
	switch authType {
	case "github":
		authurl = fmt.Sprintf(system.GetConfiguration().GithubAuthUrl, system.GetConfiguration().GithubClientId, uuid)
	case "weibo":
	case "qq":
	case "wechat":
	case "oschina":
	default:
	}
	c.Redirect(http.StatusFound, authurl)
}
