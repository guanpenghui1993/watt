package middleware

import (
	"fmt"
	"strings"
	"watt/pkg/services"
	"watt/pkg/utils"

	"github.com/gin-gonic/gin"
)

// 全局检测token有效性
func CheckLogin() gin.HandlerFunc {

	return func(c *gin.Context) {

		code := utils.SUCCESS

		msg := ""

		token := c.GetHeader(utils.Setting.Jwt.Header)

		if token == "" {

			code = utils.HEADER_ERROR
			msg = "缺少令牌信息"

		} else {

			uid, err := utils.Parse(token)

			if err != nil {
				code = utils.ERROR
				msg = "令牌已失效"
			} else {
				if list := services.UserService.Info(uid); list.ID <= 0 {
					msg = "用户不存在或已禁用"
					code = utils.ERROR
				}
			}
		}

		if code != utils.SUCCESS {
			c.JSON(200, utils.Response{code, msg, nil})
			c.Abort()
			return
		}

		c.Next()
	}
}

// 全局异常信息
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				utils.Error(r)
				if utils.Setting.Common.Debug {
					c.JSON(200, utils.Response{utils.SERVER_ERROR, fmt.Sprintf("%v", r), nil})
				} else {
					c.JSON(200, utils.Response{utils.SERVER_ERROR, "服务器异常，请稍后再试", nil})
				}
				c.Abort()
				return
			}

			if c.FullPath() == "" { // 排除404
				c.JSON(200, utils.Response{utils.ERROR, "404 Not Found", nil})
				c.Abort()
				return
			}
		}()
		c.Next()
	}
}

// 跨域设置
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", utils.Setting.Cross.Domain)
		c.Writer.Header().Set("Access-Control-Allow-Headers", utils.Setting.Cross.Header)
		c.Writer.Header().Set("Access-Control-Allow-Methods", utils.Setting.Cross.Method)

		if c.Request.Method == "OPTIONS" {
			c.Abort()
			return
		}
		c.Next()
	}
}

// 当前用户操作权限（权限菜单路由权限）
func Access() gin.HandlerFunc {

	return func(c *gin.Context) {

		uid, _ := utils.Parse(c.GetHeader(utils.Setting.Jwt.Header))

		userList := services.UserService.Info(uid)

		if userList.Roleid == utils.SUPER_ROLE {
			c.Next()
			return
		}

		routerList, forbid := services.AccessService.RouterData(userList.Roleid), false

		if len(routerList) > 0 {
			for _, val := range routerList {

				if strings.ToLower(strings.Trim(c.FullPath(), "/")) == val {
					forbid = true
					break
				}
			}
		}

		if !forbid {
			c.JSON(200, utils.Response{utils.ERROR, "暂无访问权限，请联系管理员", nil})
			c.Abort()
			return
		}

		fmt.Println(routerList, c.FullPath())
		c.Next()
	}
}
