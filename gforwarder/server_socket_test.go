package main

import (
	"fmt"
	"math"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ypcd/gforwarder/gstunnellib/gsbase"
)

//var g_test_info string

func init() {
	//runtime.SetCPUProfileRate(200)
	//debug.SetGCPercent(50)
	debug.SetMemoryLimit(20 * 1024 * 1024 * 1024)
	init_server_test()
	logger_test = g_Logger
	g_networkTimeout = time.Second * 60 * 2
	//g_IOModel = "rw"
}

func Test_server_socket(t *testing.T) {
	//debug.SetGCPercent(50)
	if gsbase.GetRaceState() {
		inTest_server_socket(t, false, 128)
	} else {
		inTest_server_socket(t, false, 2000)
	}
}

func Test_server_socket_mt_big(t *testing.T) {
	for i := 0; i < 1; i++ {
		vi := int(math.Pow(2, float64(i)))
		g_app_close.Swap(false)
		if gsbase.GetRaceState() {
			inTest_server_socket_mt_gpool(t, false, 2, 128*1024)
		} else {
			inTest_server_socket_mt_gpool(t, true, 1*vi, 4*1024*1024/vi)
		}
		g_app_close.Swap(true)
		time.Sleep(time.Millisecond * 100)
	}
	fmt.Println(g_test_info)
}

func Test_server_socket_mt_kib(t *testing.T) {
	//for i := 1; i < 20; i++ {
	//i := 20
	//vi := int(math.Pow(2, float64(i)))
	g_app_close.Swap(false)
	inTest_server_socket_mt_gpool(t, false, 10000*5, 4)
	g_app_close.Swap(true)
	time.Sleep(time.Millisecond * 1000)
	//}
	fmt.Println(g_test_info)
	//	fmt.Println("g_test_hander_count:", g_test_hander_count.Load())
	//	fmt.Println("g_ants_pool_running_G_Num:", g_ants_pool.Running())
}
