/**
* Created by GoLand.
* User: link1st
* Date: 2019-08-15
* Time: 18:14
 */

package server

import (
	"fmt"
	"go-stress-testing/model"
	"strings"
	"time"
)

// 接收结果
// 统计的时间都是纳秒，显示的时间 都是毫秒
// concurrent 并发数
func ReceivingResults(concurrent uint64, ch <-chan *model.RequestResults) {

	// 时间
	var (
		processingTime uint64 // 处理总时间
		requestTime    uint64 // 请求总时间
		maxTime        uint64 // 最大时长
		minTime        uint64 // 最小时长
		successNum     uint64 // 成功处理数，code为0
		failureNum     uint64 // 处理失败数，code不为0
	)

	statTime := uint64(time.Now().UnixNano())

	// 错误码/错误个数
	var errCode = make(map[int]int)

	// 每个秒时间输出一次 计算结果
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				endTime := uint64(time.Now().UnixNano())
				requestTime = endTime - statTime
				go calculateData(concurrent, processingTime, requestTime, maxTime, minTime, successNum, failureNum, errCode)
			}
		}
	}()

	header()

	for data := range ch {
		// fmt.Println("处理一条数据", data.Id, data.Time, data.IsSucceed, data.ErrCode)
		processingTime = processingTime + data.Time

		if maxTime <= data.Time {
			maxTime = data.Time
		}

		if minTime == 0 {
			minTime = data.Time
		} else if minTime > data.Time {
			minTime = data.Time
		}

		// 是否请求成功
		if data.IsSucceed == true {
			successNum = successNum + 1
		} else {
			failureNum = failureNum + 1
		}

		// 统计错误码
		if value, ok := errCode[data.ErrCode]; ok {
			errCode[data.ErrCode] = value + 1
		} else {
			errCode[data.ErrCode] = 1
		}
	}

	endTime := uint64(time.Now().UnixNano())
	requestTime = endTime - statTime

	calculateData(concurrent, processingTime, requestTime, maxTime, minTime, successNum, failureNum, errCode)

	fmt.Printf("\n\n")

	fmt.Println("*************************  结果 stat  ****************************")
	fmt.Println("处理协程数量:", concurrent)
	// fmt.Println("处理协程数量:", concurrent, "程序处理总时长:", fmt.Sprintf("%.3f", float64(processingTime/concurrent)/1e9), "秒")
	fmt.Println("请求总数:", successNum+failureNum, "总请求时间:", fmt.Sprintf("%.3f", float64(requestTime)/1e9),
		"秒", "successNum:", successNum, "failureNum:", failureNum)

	fmt.Println("*************************  结果 end   ****************************")

	fmt.Printf("\n\n")
}

// 计算数据
func calculateData(concurrent, processingTime, requestTime, maxTime, minTime, successNum, failureNum uint64, errCode map[int]int) {
	if processingTime == 0 {
		processingTime = 1
	}

	var (
		qps              float64
		averageTime      float64
		maxTimeFloat     float64
		minTimeFloat     float64
		requestTimeFloat float64
	)

	// 平均 每个协程成功数*总协程数据/总耗时 (每秒)
	if processingTime != 0 {
		qps = float64(successNum*1e9*concurrent) / float64(processingTime)
	}

	// 平均时长 总耗时/总请求数/并发数 纳秒=>毫秒
	if successNum != 0 && concurrent != 0 {
		averageTime = float64(processingTime) / float64(successNum*1e6*concurrent)
	}

	// 纳秒=>毫秒
	maxTimeFloat = float64(maxTime) / 1e6
	minTimeFloat = float64(minTime) / 1e6
	requestTimeFloat = float64(requestTime) / 1e9

	// 打印的时长都为毫秒
	// result := fmt.Sprintf("请求总数:%8d|successNum:%8d|failureNum:%8d|qps:%9.3f|maxTime:%9.3f|minTime:%9.3f|平均时长:%9.3f|errCode:%v", successNum+failureNum, successNum, failureNum, qps, maxTimeFloat, minTimeFloat, averageTime, errCode)
	// fmt.Println(result)
	table(successNum, failureNum, errCode, qps, averageTime, maxTimeFloat, minTimeFloat, requestTimeFloat)
}

// 打印表头信息
func header() {
	fmt.Printf("\n\n")
	// 打印的时长都为毫秒
	fmt.Println("─────┬───────┬───────┬───────┬────────┬────────┬────────┬────────┬───────────")
	result := fmt.Sprintf(" 耗时│总请数 │ 成功数│ 失败数│   qps  │最长耗时│最短耗时│平均耗时│  错误码")
	fmt.Println(result)
	// result = fmt.Sprintf("耗时(s)  │总请求数│成功数│失败数│QPS│最长耗时│最短耗时│平均耗时│错误码")
	// fmt.Println(result)
	fmt.Println("─────┼───────┼───────┼───────┼────────┼────────┼────────┼────────┼───────────")

	return
}

func table(successNum, failureNum uint64, errCode map[int]int, qps, averageTime, maxTimeFloat, minTimeFloat, requestTimeFloat float64) {
	// 打印的时长都为毫秒
	result := fmt.Sprintf("%4.0fs│%7d│%7d│%7d│%8.2f│%8.2f│%8.2f│%8.2f│%v", requestTimeFloat, successNum+failureNum, successNum, failureNum, qps, maxTimeFloat, minTimeFloat, averageTime, printMap(errCode))
	fmt.Println(result)

	return
}

func printMap(errCode map[int]int) (mapStr string) {

	var (
		mapArr []string
	)
	for key, value := range errCode {
		mapArr = append(mapArr, fmt.Sprintf("%d:%d", key, value))
	}

	mapStr = strings.Join(mapArr, ";")

	return
}