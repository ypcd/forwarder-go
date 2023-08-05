package main

import (
	"bytes"
	randc "crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ypcd/gforwarder/gstunnellib"
	"github.com/ypcd/gforwarder/gstunnellib/gstestpipe"
)

var logger_test *log.Logger = g_Logger
var g_test_info string
var g_test_hander_count atomic.Int64

func GetRDBytes_local(byteLen int) []byte {
	data := make([]byte, byteLen)
	_, err := randc.Reader.Read(data)
	if err != nil {
		panic(err)
	}
	return data
}

type testSocketEchoHandlerObj struct {
	testReadTimeOut  time.Duration
	testCacheSize    int
	testCacheSizeMiB float64
	wg               *sync.WaitGroup
}

func test_server_socket_echo_handler(obj *testSocketEchoHandlerObj, server net.Conn) {
	defer func() {
		if obj.wg != nil {
			obj.wg.Done()
		}
	}()

	testReadTimeOut := obj.testReadTimeOut
	testCacheSize := obj.testCacheSize
	//testCacheSizeMiB := obj.testCacheSizeMiB

	SendData := GetRDBytes_local(testCacheSize)
	//SendData := []byte("123456")
	rbuf := make([]byte, 0, len(SendData))
	//rbuff := bytes.Buffer{}
	//logger_test.Println("testCacheSize[MiB]:", testCacheSizeMiB)

	//logger_test.Println("inTest_server_NetPipe data transfer start.")
	//t1 := time.Now()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func(server net.Conn) {
		defer wg.Done()
		buf := make([]byte, g_net_read_size)

		for {
			server.SetReadDeadline(time.Now().Add(testReadTimeOut))
			re, err := server.Read(buf)
			//t.Logf("server read len: %d", re)
			if errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrDeadlineExceeded) || errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				checkError_info(err)
				return
			} else {
				gstunnellib.CheckError_panic(err)
			}

			rbuf = append(rbuf, buf[:re]...)

			if len(rbuf) == len(SendData) {
				return
			}
		}

	}(server)

	//_, err := gstunnellib.NetConnWriteAll(server, SendData)
	_, err := gstunnellib.NetConnWriteAll(server, SendData)

	checkError(err)
	//time.Sleep(time.Second * 6)
	wg.Wait()
	server.Close()
	//time.Sleep(time.Second * 60)

	if !bytes.Equal(SendData, rbuf) {
		panic("Error: SendData != rbuf.")
	}
	g_test_hander_count.Add(1)
	//logger_test.Println("[inTest_server_NetPipe] end.")
	//t2 := time.Now()
	//logger_test.Println("[pipe run time(sec)]:", t2.Sub(t1).Seconds())
	//logger_test.Println("[pipe run MiB/sec]:", float64(testCacheSizeMiB)/(t2.Sub(t1).Seconds()))

	//time.Sleep(time.Hour * 60)
}

func inTest_server_socket(t *testing.T, mt_mode bool, mbNum int) {

	var t1, t2 time.Time

	logger_test.Println("[inTest_server_NetPipe] start.")
	t1 = time.Now()
	raws := gstestpipe.NewRawServerSocket_RandAddr()
	defer raws.Close()
	raws.Run()

	listenAddr := gstestpipe.GetRandAddr()
	gsc := gstestpipe.NewRawClientSocketEcho(listenAddr)
	defer gsc.Close()

	g_Mt_model = mt_mode
	g_Values.SetDebug(true)

	handleobj := testSocketEchoHandlerObj{
		testReadTimeOut:  time.Second * 1,
		testCacheSize:    mbNum * 1024 * 1024,
		testCacheSizeMiB: float64(mbNum)}
	//g_networkTimeout = handleobj.testReadTimeOut

	outlistener := &net.TCPListener{}

	go run_pipe_test_listen(listenAddr, raws.GetServerAddr(), outlistener)
	defer outlistener.Close()

	time.Sleep(time.Millisecond * 10)
	gsc.Run()

	server := <-raws.GetConnList()
	test_server_socket_echo_handler(&handleobj, server)
	t2 = time.Now()
	logger_test.Println("[pipe run time(sec)]:", t2.Sub(t1).Seconds())
	logger_test.Println("[pipe run MiB/sec]:", float64(handleobj.testCacheSizeMiB)/(t2.Sub(t1).Seconds()))
}

func inTest_server_socket_mt(t *testing.T, mt_mode bool, connes int, dataSZKiB int) {

	mtNum := connes

	time_run_begin := time.Now()
	logger_test.Println("[inTest_server_socket_mt] start.")

	raws := gstestpipe.NewRawServerSocket_RandAddr()
	defer raws.Close()
	raws.Run()

	listenAddr := gstestpipe.GetRandAddr()
	gsc := gstestpipe.NewRawClientSocketEcho(listenAddr)
	defer gsc.Close()

	g_Mt_model = mt_mode
	g_Values.SetDebug(true)
	wg := sync.WaitGroup{}
	wgrawserver := sync.WaitGroup{}

	handleobj := testSocketEchoHandlerObj{
		testReadTimeOut:  g_networkTimeout,
		testCacheSize:    dataSZKiB * 1024,
		testCacheSizeMiB: float64(dataSZKiB) / float64(1024),
		wg:               &wg}
	//g_networkTimeout = handleobj.testReadTimeOut

	outlistener := &net.TCPListener{}

	go run_pipe_test_listen_pool(listenAddr, raws.GetServerAddr(), outlistener)
	defer outlistener.Close()

	wgrawserver.Add(1)
	go func() {
		defer wgrawserver.Done()
		rawsConnList := raws.GetConnList()
		for i := 0; i < mtNum; i++ {
			server := <-rawsConnList
			wg.Add(1)
			go test_server_socket_echo_handler(&handleobj, server)
		}
	}()

	time.Sleep(time.Millisecond * 10)
	time_connSocket_begin := time.Now()
	for i := 0; i < mtNum; i++ {
		gsc.Run()
	}
	time_connSocket := time.Since(time_connSocket_begin)

	wgrawserver.Wait()
	wg.Wait()
	time_run_end := time.Now()
	os.Stdout.Sync()
	time.Sleep(time.Millisecond * 100)
	logger_test.Println("time_connSocket:", time_connSocket)
	logger_test.Println("time_run:", time_run_end.Sub(time_run_begin))
	logger_test.Println("connNum:", connes, "dataSZKiB:", dataSZKiB, "testCacheSizeMiB:", handleobj.testCacheSizeMiB)
	logger_test.Println("run MiB/s:", handleobj.testCacheSizeMiB*float64(connes)/float64(time_run_end.Sub(time_run_begin).Seconds()))

	g_test_info += "\n"
	g_test_info += fmt.Sprintln("time_connSocket:", time_connSocket)
	g_test_info += fmt.Sprintln("time_run:", time_run_end.Sub(time_run_begin))
	g_test_info += fmt.Sprintln("connNum:", connes, "dataSZKiB:", dataSZKiB, "testCacheSizeMiB:", handleobj.testCacheSizeMiB)
	g_test_info += fmt.Sprintln("run MiB/s:", handleobj.testCacheSizeMiB*float64(connes)/float64(time_run_end.Sub(time_run_begin).Seconds()))
	g_test_info += "\n"
}

func inTest_server_socket_mt_gpool(t *testing.T, mt_mode bool, connes int, dataSZKiB int) {

	mtNum := connes

	time_run_begin := time.Now()
	logger_test.Println("[inTest_server_socket_mt] start.")

	raws := gstestpipe.NewRawServerSocket_RandAddr()
	defer raws.Close()
	raws.Run()

	listenAddr := gstestpipe.GetRandAddr()
	gsc := gstestpipe.NewRawClientSocketEcho(listenAddr)
	defer gsc.Close()

	g_Mt_model = mt_mode
	g_Values.SetDebug(true)
	wg := sync.WaitGroup{}
	wgrawserver := sync.WaitGroup{}

	handleobj := testSocketEchoHandlerObj{
		testReadTimeOut:  g_networkTimeout,
		testCacheSize:    dataSZKiB * 1024,
		testCacheSizeMiB: float64(dataSZKiB) / float64(1024),
		wg:               &wg}
	//g_networkTimeout = handleobj.testReadTimeOut

	outlistener := &net.TCPListener{}

	go run_pipe_test_listen_pool(listenAddr, raws.GetServerAddr(), outlistener)
	defer outlistener.Close()

	wgrawserver.Add(1)
	go func() {
		defer wgrawserver.Done()
		rawsConnList := raws.GetConnList()
		for i := 0; i < mtNum; i++ {
			server := <-rawsConnList
			wg.Add(1)
			g_ants_pool.Submit(func() {
				test_server_socket_echo_handler(&handleobj, server)
			})
		}
	}()

	time.Sleep(time.Millisecond * 10)
	time_connSocket_begin := time.Now()
	for i := 0; i < mtNum; i++ {
		gsc.Run()
	}
	time_connSocket := time.Since(time_connSocket_begin)

	wgrawserver.Wait()
	wg.Wait()
	time_run_end := time.Now()
	os.Stdout.Sync()
	time.Sleep(time.Millisecond * 100)
	logger_test.Println("time_connSocket:", time_connSocket)
	logger_test.Println("time_run:", time_run_end.Sub(time_run_begin))
	logger_test.Println("connNum:", connes, "dataSZKiB:", dataSZKiB, "testCacheSizeMiB:", handleobj.testCacheSizeMiB)
	logger_test.Println("run MiB/s:", handleobj.testCacheSizeMiB*float64(connes)/float64(time_run_end.Sub(time_run_begin).Seconds()))
	logger_test.Println("qps:", float64(connes)/float64(time_run_end.Sub(time_run_begin).Seconds()), "qps(yugu):", float64(connes)/float64(time_run_end.Sub(time_run_begin).Seconds())*2)
	logger_test.Println("g_test_hander_count:", g_test_hander_count.Load())
	logger_test.Println("g_ants_pool_running_G_Num:", g_ants_pool.Running())

	g_test_info += "\n"
	g_test_info += fmt.Sprintln("time_connSocket:", time_connSocket)
	g_test_info += fmt.Sprintln("time_run:", time_run_end.Sub(time_run_begin))
	g_test_info += fmt.Sprintln("connNum:", connes, "dataSZKiB:", dataSZKiB, "testCacheSizeMiB:", handleobj.testCacheSizeMiB)
	g_test_info += fmt.Sprintln("run MiB/s:", handleobj.testCacheSizeMiB*float64(connes)/float64(time_run_end.Sub(time_run_begin).Seconds()))
	g_test_info += fmt.Sprintln("qps:", float64(connes)/float64(time_run_end.Sub(time_run_begin).Seconds()), "qps(yugu):", float64(connes)/float64(time_run_end.Sub(time_run_begin).Seconds())*2)
	g_test_info += fmt.Sprintln("g_test_hander_count:", g_test_hander_count.Load())
	g_test_info += fmt.Sprintln("g_ants_pool_running_G_Num:", g_ants_pool.Running())
	//lookmem := gsmemstats.NewLookHeapMemAndMaxMem()
	//g_test_info += fmt.Sprintf("lookmem:\n\t%s\n\t%s", lookmem.GetMaxHeapMemStatsMiB(), lookmem.GetGCInfo())

	g_test_info += "\n"
}
