package main

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"math/rand"
	"time"
)

/**
双色球彩票
1. 即开即得型 http://localhost:8080/
2. 双色球自选 http://localhost:8080/prize
*/

type lotteryController struct {
	ctx iris.Context
}

func main() {
	app := newApp()
	app.Run(iris.Addr(":8080"))
}

func newApp() *iris.Application {
	app := iris.New()
	mvc.New(app.Party("/")).Handle(&lotteryController{})
	return app
}

// Get 即开即得 /
func (l *lotteryController) Get() string {
	var prize string
	code := rand.New(rand.NewSource(time.Now().UnixMilli())).Intn(10)
	switch {
	case code == 1:
		prize = "一等奖"
	case code >= 2 && code <= 3:
		prize = "二等奖"
	case code >= 4 && code <= 6:
		prize = "三等奖"
	default:
		return fmt.Sprintf("尾号为1获得一等奖</br>"+
			"尾号为2或3获得二等奖</br>"+
			"尾号为4、5、6获得三等奖</br>"+
			"尾号为1获得一等奖</br>"+
			"code=%d</br>"+
			"很遗憾，没有获奖", code)

	}
	return fmt.Sprintf("尾号为1获得一等奖</br>"+
		"尾号为2或3获得二等奖</br>"+
		"尾号为4、5、6获得三等奖</br>"+
		"尾号为1获得一等奖</br>"+
		"code=%d</br>"+
		"恭喜你获得: %s", code, prize)
}

// GetPrize 双色球
func (l *lotteryController) GetPrize() string {
	r := rand.New(rand.NewSource(time.Now().UnixMilli()))
	var prize [7]int
	// 6个红色球，1-33
	for i := 0; i < 6; i++ {
		prize[i] = r.Intn(33) + 1
	}
	// 最后一位是蓝色球，1-16
	prize[6] = r.Intn(16) + 1
	return fmt.Sprintf("今日开奖号码是: %v", prize)
}
