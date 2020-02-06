package controllers

import (
	"github.com/dchest/captcha"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

//获取验证码
func CaptchaGet(c *gin.Context) {
	session := sessions.Default(c)
	//设置验证码长度为4位
	captchaId := captcha.NewLen(4)
	session.Delete(SESSION_CAPTCHA)
	session.Set(SESSION_CAPTCHA, captchaId)
	session.Save()
	captcha.WriteImage(c.Writer, captchaId, 100, 40)
}
