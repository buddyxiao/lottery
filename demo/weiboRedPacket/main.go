package main

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// 当前有效红包列表，uint32是红包唯一ID，[]uint是红包里面随机分到的金额（单位分）
var packageList = make(map[uint32][]uint)

func main() {
	app := newApp()
	app.Run(iris.Addr(":8080"))
}

type lotteryController struct {
	Ctx iris.Context
}

// Get 返回全部红包地址
func (l lotteryController) Get() map[uint32][2]int {
	rs := make(map[uint32][2]int)
	for id, list := range packageList {
		var money int
		for _, v := range list {
			money += int(v)
		}
		rs[id] = [2]int{len(list), money}
	}
	return rs
}

// GetSet 发红包
// GET http://localhost:8080/set?uid=1&money=100&num=100
func (l *lotteryController) GetSet() string {
	uid, errUid := l.Ctx.URLParamInt("uid")
	money, errMoney := l.Ctx.URLParamFloat64("money")
	num, errNum := l.Ctx.URLParamInt("num")
	if errUid != nil || errMoney != nil || errNum != nil {
		return fmt.Sprintf("参数格式异常，errUid=%s, errMoney=%s, errNum=%s\n", errUid, errMoney, errNum)
	}
	moneyTotal := int(money * 100) // 分钱为单位
	if uid < 1 || moneyTotal < num || num < 1 {
		return fmt.Sprintf("参数数值异常，uid=%d, money=%d, num=%d\n", uid, money, num)
	}
	// 金额分配算法
	leftMoney := moneyTotal
	leftNum := num
	// 分配的随机数
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	rMax := 0.55              // 随机分配最大比例
	list := make([]uint, num) // 分配的红包
	// 大循环开始，只要还有没分配的名额，继续分配
	for leftNum > 0 {
		if leftNum == 1 {
			// 最后一个名额，把剩余的全部给它
			list[num-1] = uint(leftMoney)
			break
		}
		// 剩下的最多只能分配到1分钱时，不用再随机
		if leftMoney == leftNum {
			for i := num - leftNum; i < num; i++ {
				list[i] = 1
			}
			break
		}
		// 每次对剩余金额的1%-55%随机，最小1，最大就是剩余金额55%（需要给剩余的名额留下1分钱的生存空间）
		rMoney := int(float64(leftMoney-leftNum) * rMax)
		m := r.Intn(rMoney)
		if m < 1 {
			m = 1
		}
		list[num-leftNum] = uint(m)
		leftMoney -= m
		leftNum--
	}
	// 最后再来一个红包的唯一ID
	id := r.Uint32()
	packageList[id] = list
	// 返回抢红包的URL
	return fmt.Sprintf("/get?id=%d&uid=%d&num=%d\n", id, uid, num)
}

// GetGet 抢红包
// GET http://localhost:8080/get?id=1&uid=1
func (l *lotteryController) GetGet() string {
	uid, errUid := l.Ctx.URLParamInt("uid")
	id, errId := l.Ctx.URLParamInt("id")
	if errUid != nil || errId != nil {
		return fmt.Sprintf("参数格式异常，errUid=%s, errId=%s\n", errUid, errId)
	}
	if uid < 1 || id < 1 {
		return fmt.Sprintf("参数数值异常，uid=%d, id=%d\n", uid, id)
	}
	list, ok := packageList[uint32(id)]
	if !ok || len(list) < 1 {
		return fmt.Sprintf("红包不存在,id=%d\n", id)
	}
	// 分配的随机数
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// 从红包金额中随机得到一个
	i := r.Intn(len(list))
	money := list[i]
	// 更新红包列表中的信息
	if len(list) > 1 {
		if i == len(list)-1 {
			packageList[uint32(id)] = list[:i]
		} else if i == 0 {
			packageList[uint32(id)] = list[1:]
		} else {
			packageList[uint32(id)] = append(list[:i], list[i+1:]...)
		}
	} else {
		delete(packageList, uint32(id))
	}
	return fmt.Sprintf("恭喜你抢到一个红包，金额为:%d\n", money)
}

func newApp() *iris.Application {
	app := iris.New()
	mvc.New(app.Party("/")).Handle(new(lotteryController))
	// 初始化日志信息
	initLog()
	return app
}

var logger *log.Logger

// 初始化日志
func initLog() {
	dir, _ := os.Getwd()
	dir, _ = filepath.Abs(dir)
	filename := filepath.Join(dir, "log/lottery_demo.log")
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("初始化Log失败,err:", err)
		os.Exit(-1)
	}
	logger = log.New(file, "", log.Ldate|log.Lmicroseconds)
}
