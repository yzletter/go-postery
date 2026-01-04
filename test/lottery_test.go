package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	giftdto "github.com/yzletter/go-postery/dto/gift"
)

const url = "http://localhost:8765/api/v1/lottery/lucky" // 压测接口
const P = 200                                            // 模拟 200 个用户，在疯狂抽奖

type Response struct {
	Code int         `json:"code"`           // 业务状态码，0 表示成功，非 0 表示失败
	Msg  string      `json:"msg"`            // 提示信息
	Data giftdto.DTO `json:"data,omitempty"` // 具体数据，失败时可以为空
}

func TestLottery(t *testing.T) {
	hitMap := make(map[string]int, 10) // 每个奖品被抽中的次数
	giftCh := make(chan string, 10000) // 抽中的奖品id放入这个channel
	counterCh := make(chan struct{})   // 判断异步协程是否结束

	//异步统计每个奖品被抽中的次数
	go func() {
		for name := range giftCh {
			hitMap[name]++
		}
		counterCh <- struct{}{} //异步协程结束
	}()

	wg := sync.WaitGroup{}
	wg.Add(P)
	begin := time.Now()
	var totalCall int64    //记录接口总调用次数
	var totalUseTime int64 //接口调用耗时总和
	for i := 0; i < P; i++ {
		go func() {
			defer wg.Done()
			for {
				t1 := time.Now()

				client := &http.Client{}
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					panic(err)
				}
				// 添加 Header
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJVaWQiOjIwMDIwMTE2Njg4MTIzOTg1OTIsIlNTaWQiOiI5OTY4MDdjMi05ZmY4LTQzNmQtODE0MS1hZTdhYjRiZDJjN2QiLCJSb2xlIjowLCJVc2VyQWdlbnQiOiJBcGlmb3gvMS4wLjAgKGh0dHBzOi8vYXBpZm94LmNvbSkiLCJpc3MiOiJnby1wb3N0ZXJ5IiwiZXhwIjoxNzY3NTA4NDM1fQ.MlzRmf33sr91PUkdfU9nw8fuM_0hvBJz3A82pIY69UbvQzi0cXu7lcRRZ1oHYG2jFiK2QxY0d5uqN0p15lTlBQ")
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("User-Agent", "Go-http-client/1.1")

				resp, err := client.Do(req)
				if err != nil {
					panic(err)
				}

				atomic.AddInt64(&totalUseTime, time.Since(t1).Milliseconds())
				atomic.AddInt64(&totalCall, 1) //调用次数加1
				if err != nil {
					fmt.Println(err)
					break
				}
				bs, err := io.ReadAll(resp.Body)
				if err != nil {
					fmt.Println(err)
					break
				}
				resp.Body.Close()
				var v Response
				json.Unmarshal(bs, &v)

				fmt.Println(v)

				if v.Data.ID >= 0 {
					if v.Data.ID == 0 { //如果返回的奖品ID为0，说明已抽完
						break
					}
					giftCh <- v.Data.Name //抽中一个gift，就放入channel
				}
			}
		}()
	}
	wg.Wait()
	close(giftCh)
	<-counterCh // 等hitMap准备好

	totalTime := int64(time.Since(begin).Seconds())
	if totalTime > 0 && totalCall > 0 {
		qps := totalCall / totalTime
		avgTime := totalUseTime / totalCall
		fmt.Printf("QPS %d, avg time %dms\n", qps, avgTime)
		//QPS 1650, avg time 69ms

		total := 0
		for name, count := range hitMap {
			fmt.Printf("%s\t%d\n", name, count)
			total += count
		}
		fmt.Printf("共计%d件商品\n", total)
	}
}

// go test -v ./test -run=^TestLottery$ -count=1
/*
	QPS 2850, avg time 79ms
	论坛定制马克杯  500
	咖啡兑换券      500
	机械键盘        400
	谢谢参与        5000
	VIP 月度会员    600
	论坛纪念徽章    500
	技术书籍兑换券  2000
	积分翻倍卡      1000
	无线鼠标        200
	论坛周边T恤     500
	共计11200件商品
*/
