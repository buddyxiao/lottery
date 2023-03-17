package main

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

/**
微信摇一摇
/lucky   只有一个抽奖得接口
压力测试：wrk -t10 -c10 -d5 http://localhost:8080/lucky
*/
// 奖品类型

const (
	giftTypeCoin      = iota // 虚拟币
	giftTypeCoupon           // 不同卷
	giftTypeCouponFix        // 相同卷
	giftTypeRealSmall        // 实物小奖
	giftTypeRealLarge        // 实物大奖
)

type gift struct {
	id       int      // 奖品ID
	name     string   // 奖品名称
	pic      string   // 照片链接
	link     string   // 链接
	gtype    int      // 奖品类型
	data     string   // 奖品的数据（特定的配置信息，如：虚拟币面值，固定优惠券的编码）
	datalist []string // 奖品数据集合（特定的配置信息，如：不同的优惠券的编码）
	total    int      // 总数，0 不限量
	left     int      // 剩余数
	inuse    bool     // 是否使用中
	rate     int      // 中奖概率，万分之N,0-10000
	rateMin  int      // 大于等于，中奖的最小号码,0-10000
	rateMax  int      // 小于，中奖的最大号码,0-10000
}

// 最大的中奖号码
const rateMax = 10000

var logger *log.Logger

// 奖品列表
var giftList []*gift
var mux sync.RWMutex

func main() {
	app := newApp()
	app.Run(iris.Addr(":8080"))
}

type lotteryController struct {
	Ctx iris.Context
}

// Get 奖品数量的信息
func (l lotteryController) Get() string {
	count := 0
	total := 0
	for _, data := range giftList {
		if data.inuse && (data.total == 0 || (data.total > 0 && data.left > 0)) {
			count++
			total += data.left
		}
	}
	return fmt.Sprintf("当前有效奖品种类数量: %d，限量奖品总数量=%d\n", count, total)
}

func (l *lotteryController) GetLucky() map[string]any {
	mux.Lock()
	defer mux.Unlock()
	code := rand.New(rand.NewSource(time.Now().UnixMilli())).Intn(rateMax)
	var (
		ok     bool
		result map[string]any
	)
	result = make(map[string]any)
	result["success"] = ok
	for _, data := range giftList {
		if !data.inuse || (data.total > 0 && data.left <= 0) {
			continue
		}
		if data.rateMin <= code && data.rateMax > code {
			// 中奖了，抽奖编码在奖品中奖编码范围内
			// 开始发奖
			sendData := ""
			switch data.gtype {
			case giftTypeCoin:
				ok, sendData = sendCoin(data)
			case giftTypeCoupon:
				ok, sendData = sendCoupon(data)
			case giftTypeCouponFix:
				ok, sendData = sendCouponFix(data)
			case giftTypeRealSmall:
				ok, sendData = sendRealSmall(data)
			case giftTypeRealLarge:
				ok, sendData = sendRealLarge(data)
			}
			if ok {
				// 中奖后，成功得到奖品（发奖成功）
				// 生成中奖纪录
				saveLuckyData(code, data.id, data.name, data.link, sendData, data.left)
				result["success"] = ok
				result["id"] = data.id
				result["name"] = data.name
				result["link"] = data.link
				result["data"] = sendData
				break
			}
		}
	}
	return result
}

func saveLuckyData(code int, id int, name string, link string, data string, left int) {
	logger.Printf("lucky, code=%d, gift=%d, name=%s, link=%s, data=%s, left=%d ", code, id, name, link, data, left)
}

// 发奖，实物大
func sendRealLarge(data *gift) (bool, string) {
	return sendRealSmall(data)
}

// 发奖，实物小
func sendRealSmall(data *gift) (bool, string) {
	return sendCouponFix(data)
}

// 发奖，优惠券（固定值）
func sendCouponFix(data *gift) (bool, string) {
	if data.total == 0 {
		// 数量无限
		return true, data.data
	} else if data.left > 0 {
		data.left = data.left - 1
		return true, data.data
	} else {
		return false, "奖品已发完"
	}
}

// 发奖，优惠券（不同值
func sendCoupon(data *gift) (bool, string) {
	if data.left > 0 {
		// 还有剩余的奖品
		left := data.left - 1
		data.left = left
		return true, data.datalist[left]
	} else {
		return false, "奖品已发完"
	}
}

// 发奖，虚拟币
func sendCoin(data *gift) (bool, string) {
	if data.total == 0 {
		// 数量无限
		return true, data.data
	} else if data.total > 0 && data.left > 0 {
		// 还有剩余
		data.left = data.left - 1
		return true, data.data
	}
	return false, "奖品已发完"
}

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

// 初始化奖品列表
func initGift() {
	giftList = make([]*gift, 5)
	// 1 实物大奖
	g1 := gift{
		id:      1,
		name:    "手机N7",
		pic:     "",
		link:    "",
		gtype:   giftTypeRealLarge,
		data:    "",
		total:   20000,
		left:    20000,
		inuse:   true,
		rate:    10000,
		rateMin: 0,
		rateMax: 0,
	}
	giftList[0] = &g1
	// 2 实物小奖
	g2 := gift{
		id:      2,
		name:    "安全充电 黑色",
		pic:     "",
		link:    "",
		gtype:   giftTypeRealSmall,
		data:    "",
		total:   5,
		left:    5,
		inuse:   false,
		rate:    100,
		rateMin: 0,
		rateMax: 0,
	}
	giftList[1] = &g2
	// 3 虚拟券，相同的编码
	g3 := gift{
		id:      3,
		name:    "商城满2000元减50元优惠券",
		pic:     "",
		link:    "",
		gtype:   giftTypeCouponFix,
		data:    "mall-coupon-2018",
		total:   50,
		left:    50,
		rate:    5000,
		inuse:   false,
		rateMin: 0,
		rateMax: 0,
	}
	giftList[2] = &g3
	// 4 虚拟券，不相同的编码
	g4 := gift{
		id:       4,
		name:     "商城无门槛直降50元优惠券",
		pic:      "",
		link:     "",
		gtype:    giftTypeCoupon,
		data:     "",
		datalist: []string{"c01", "c02", "c03", "c04", "c05"},
		total:    5,
		left:     5,
		inuse:    false,
		rate:     2000,
		rateMin:  0,
		rateMax:  0,
	}
	giftList[3] = &g4
	// 5 虚拟币
	g5 := gift{
		id:      5,
		name:    "社区10个金币",
		pic:     "",
		link:    "",
		gtype:   giftTypeCoin,
		data:    "10",
		total:   5,
		left:    5,
		inuse:   false,
		rate:    5000,
		rateMin: 0,
		rateMax: 0,
	}
	giftList[4] = &g5
	// 整理奖品数据，把rateMin,rateMax根据rate进行编排
	rateStart := 0
	for _, data := range giftList {
		if !data.inuse {
			continue
		}
		data.rateMin = rateStart
		data.rateMax = data.rateMin + data.rate
		if data.rateMax >= rateMax {
			// 号码达到最大值，分配的范围重头再来
			data.rateMax = rateMax
			rateStart = 0
		} else {
			rateStart += data.rate
		}
	}
	fmt.Printf("giftlist=%v\n", giftList)
}

func newApp() *iris.Application {
	app := iris.New()
	mvc.New(app.Party("/")).Handle(new(lotteryController))
	// 初始化日志信息
	initLog()
	// 初始化奖品信息
	initGift()
	return app
}
