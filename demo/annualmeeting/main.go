package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

var userList []string
var mux sync.RWMutex

type lotteryController struct {
}

func newRouter() *gin.Engine {
	engine := gin.Default()
	var l = new(lotteryController)
	engine.GET("/", l.Get)
	engine.POST("/import", l.Import)
	engine.GET("/lucky", l.Lucky)
	engine.GET("/clear", l.Clear)
	return engine
}

func main() {
	app := newRouter()
	userList = make([]string, 0)
	app.Run(":8080")
}

func (l *lotteryController) Get(c *gin.Context) {
	mux.RLock()
	defer mux.RUnlock()
	count := len(userList)
	c.JSON(http.StatusOK, gin.H{
		"msg": fmt.Sprintf("当前总共参与抽奖的用户数: %d", count),
	})
}

func (l *lotteryController) Import(c *gin.Context) {
	users, ok := c.GetPostForm("users")
	if ok {
		importUserList := strings.Split(users, ",")
		originUserCount := len(userList)
		mux.Lock()
		defer mux.Unlock()
		for _, u := range importUserList {
			u = strings.TrimSpace(u)
			if len(u) > 0 {
				userList = append(userList, u)
			}
		}
		imporedUserCount := len(userList)
		c.JSON(http.StatusOK, gin.H{
			"msg": fmt.Sprintf("当前总共参与抽奖的用户数: %d, 成功导入用户数: %d", imporedUserCount, imporedUserCount-originUserCount),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg": "输入有误",
	})
}

func (l *lotteryController) Lucky(c *gin.Context) {
	count := len(userList)
	mux.Lock()
	defer mux.Unlock()
	if count > 1 {
		rand.Seed(time.Now().UnixMilli())
		index := rand.Int31n(int32(count))
		user := userList[index]
		userList = append(userList[0:index], userList[index+1:]...)
		c.JSON(http.StatusOK, gin.H{
			"msg": fmt.Sprintf("当前中奖用户: %s, 剩余用户数: %d", user, count-1),
		})
	} else if count == 1 {
		user := userList[0]
		c.JSON(http.StatusOK, gin.H{
			"msg": fmt.Sprintf("当前中奖用户: %s, 剩余用户数: %d", user, count-1),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg": "已经没有参与用户，请先通过 /import 导入用户",
		})
	}
}

func (l *lotteryController) Clear(c *gin.Context) {
	mux.Lock()
	defer mux.Unlock()
	userList = make([]string, 0)
}
