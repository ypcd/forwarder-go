/*
*
*Open source agreement:
*   The project is based on the GPLv3 protocol.
*
 */
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	_ "net/http/pprof"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/ypcd/gforwarder/gstunnellib"
)

const g_version string = gstunnellib.G_Version

//gstunnellib.Version

var g_Values = gstunnellib.NewGlobalValuesImp()

//var p = gstunnellib.Nullprint
//var pf = gstunnellib.Nullprintf

var fpnull = os.DevNull

var key string
var bindAddr = flag.String("b", "DefaultArg", "'-b 127.0.0.1:8080'. Local bind addr.")

var serverAddr = flag.String("s", "DefaultArg", "'-s 10.0.0.1:80'. Remote server addr.")

var g_gsconfig *gstunnellib.GsConfig
var g_gsconfig_path = flag.String("c", "config.server.json", "The gstunnel config file path.")

var gsversion = flag.String("v", g_version, "GST Version: "+g_version)

var g_Logger *log.Logger

const g_net_read_size = 4 * 1024
const netPUn_chan_cache_size = 64

var g_Mt_model bool = true

// Default 6s
var g_tmr_display_time time.Duration

// Default 60s
var g_tmr_changekey_time time.Duration

// Default 60s
var g_networkTimeout time.Duration

var g_RuntimeStatistics gstunnellib.Runtime_statistics

var g_log_List gstunnellib.Logger_List

var init_status = false

var g_gid = gstunnellib.NewGIdImp()

var g_IOModel string

var g_connHandleGoNum int = 64

var g_init_status = false

var g_app_close atomic.Bool

var g_ants_pool *ants.Pool

//var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
//var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func init_server_run() {
	if init_status {
		panic(errors.New("The init func is error"))
	} else {
		init_status = true
	}
	/*
		for _, v := range os.Args {
			if v == "-g" {
			}
		}
	*/
	flag.Parse()

	g_RuntimeStatistics = gstunnellib.NewRuntimeStatistics()

	g_log_List.GenLogger = gstunnellib.NewLoggerFileAndStdOut("gstunnel_server.log")
	g_log_List.GSIpLogger = gstunnellib.NewLoggerFileAndStdOut("access.log")
	g_log_List.GSNetIOLen = gstunnellib.NewLoggerFileAndLog("net_io_len.log", g_log_List.GenLogger.Writer())
	g_log_List.GSIpLogger.Println("Gstunnel client access ip list:")

	g_Logger = g_log_List.GenLogger

	g_Logger.Println("gstunnel server.")

	g_Logger.Println("VER:", g_version)

	g_gsconfig = gstunnellib.CreateGsconfig(*g_gsconfig_path)

	g_Values.SetDebug(g_gsconfig.Debug)

	g_Mt_model = g_gsconfig.Mt_model

	g_IOModel = g_gsconfig.IOModel

	g_tmr_display_time = time.Second * time.Duration(g_gsconfig.Tmr_display_time)
	g_tmr_changekey_time = time.Second * time.Duration(g_gsconfig.Tmr_changekey_time)
	g_networkTimeout = time.Second * time.Duration(g_gsconfig.NetworkTimeout)

	key = g_gsconfig.Key

	g_Logger.Println("debug:", g_Values.GetDebug())

	g_Logger.Println("IOModel:", g_IOModel)
	g_Logger.Println("g_Mt_model:", g_Mt_model)
	g_Logger.Println("tmr_display_time:", g_tmr_display_time)
	g_Logger.Println("g_tmr_changekey_time:", g_tmr_changekey_time)
	g_Logger.Println("g_networkTimeout:", g_networkTimeout)

	g_Logger.Println("info_protobuf:", gstunnellib.G_Info_protobuf)

	if g_Values.GetDebug() {
		go func() {
			g_Logger.Fatalln("http server: ", http.ListenAndServe("localhost:3030", nil))
		}()
		g_Logger.Println("pprof server listen: localhost:3030")
	}
	//g_Values.GetDebug() = false

	//go gstunnellib.RunGRuntimeStatistics_print(g_Logger, g_RuntimeStatistics)

}

func init_server_test() {
	if g_init_status {
		panic(errors.New("the init func is error"))
	} else {
		g_init_status = true
	}

	g_RuntimeStatistics = gstunnellib.NewRuntimeStatistics()

	g_log_List.GenLogger = gstunnellib.NewLoggerFileAndStdOut("gstunnel_server.log")
	g_log_List.GSIpLogger = gstunnellib.NewLoggerFileAndStdOut("access.log")
	g_log_List.GSNetIOLen = gstunnellib.NewLoggerFileAndLog("net_io_len.log", g_log_List.GenLogger.Writer())
	g_log_List.GSIpLogger.Println("Gstunnel client access ip list:")

	g_Logger = g_log_List.GenLogger

	g_Logger.Println("gstunnel server.")

	g_Logger.Println("VER:", g_version)

	g_gsconfig = gstunnellib.CreateGsconfig(*g_gsconfig_path)

	g_Values.SetDebug(g_gsconfig.Debug)

	g_Mt_model = g_gsconfig.Mt_model
	g_IOModel = g_gsconfig.IOModel

	g_tmr_display_time = time.Second * time.Duration(g_gsconfig.Tmr_display_time)
	g_tmr_changekey_time = time.Second * time.Duration(g_gsconfig.Tmr_changekey_time)
	g_networkTimeout = time.Second * time.Duration(g_gsconfig.NetworkTimeout)

	var err error
	g_ants_pool, err = ants.NewPool(0, ants.WithExpiryDuration(time.Second*60))
	checkError_panic(err)
	//defer g_ants_pool.Release()
	gstunnellib.G_ants_pool = g_ants_pool
	//g_key = g_gsconfig.Key

	g_Logger.Println("debug:", g_Values.GetDebug())

	g_Logger.Println("IOModel:", g_IOModel)
	g_Logger.Println("g_Mt_model:", g_Mt_model)
	g_Logger.Println("g_tmr_display_time:", g_tmr_display_time)
	g_Logger.Println("g_tmr_changekey_time:", g_tmr_changekey_time)
	g_Logger.Println("g_networkTimeout:", g_networkTimeout)

	g_Logger.Println("info_protobuf:", gstunnellib.G_Info_protobuf)

	if g_Values.GetDebug() {
		go func() {
			g_Logger.Fatalln("http server: ", http.ListenAndServe("localhost:3030", nil))
		}()
		g_Logger.Println("pprof server listen: localhost:3030")
	}
	//g_Values.GetDebug() = false

	//go gstunnellib.RunGRuntimeStatistics_print(g_Logger, g_RuntimeStatistics)

}

func main() {
	init_server_run()
	for {
		run()
	}
}

func run_old() {
	//defer gstunnellib.Panic_Recover_GSCtx(g_Logger, gctx)

	var lstnaddr, connaddr string

	lstnaddr = g_gsconfig.Listen
	connaddr = g_gsconfig.GetServers()[0]

	//lstnaddr = *bindAddr
	//connaddr = *serverAddr

	//lstnaddr = ":1234"
	//connaddr = "127.0.0.1:8080"

	g_Logger.Println("Listen_Addr:", lstnaddr)
	g_Logger.Println("Server_Addr:", connaddr)
	g_Logger.Println("Begin......")

	service := lstnaddr
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		acc, err := listener.Accept()
		if err != nil {
			checkError_NoExit(err)
			continue
		}
		g_log_List.GSIpLogger.Printf("ip: %s\n", acc.RemoteAddr().String())

		service := connaddr

		connServiceError_count := 0
		const maxTryConnService int = 5000
		var dst net.Conn
		for {
			dst, err = net.Dial("tcp", service)
			checkError_NoExit(err)
			connServiceError_count += 1
			if err == nil {
				break
			}
			if connServiceError_count > maxTryConnService {
				checkError(
					errors.New(
						fmt.Sprintf("connService_count > maxTryConnService(%d)", maxTryConnService),
					),
				)
			}
			//g_Logger.Println("conn.")
		}

		gctx := gstunnellib.NewGsContextImp(g_gid.GetId())
		go srcTOdstUn(acc, dst, gctx)
		go srcTOdstP(dst, acc, gctx)
		g_Logger.Printf("go [%d].\n", gctx.GetGsId())
	}
}

func createToRawServiceConnHandler(chanconnlist <-chan net.Conn, rawServiceAddr string) {

	const maxTryConnService int = 5000
	var err error
	for {

		gstClient, ok := <-chanconnlist
		if !ok {
			checkError_NoExit(errors.New("'gstServer, ok := <-chanconnlist' is not ok"))
			return
		}
		g_log_List.GSIpLogger.Printf("Gstunnel client ip: %s\n", gstClient.RemoteAddr().String())

		connServiceError_count := 0
		var rawService net.Conn
		for {
			rawService, err = net.Dial("tcp", rawServiceAddr)
			checkError_NoExit(err)
			connServiceError_count += 1
			if err == nil {
				break
			}
			if connServiceError_count > maxTryConnService {
				checkError(
					fmt.Errorf("connService_count > maxTryConnService(%d)", maxTryConnService))
			}
			//g_Logger.Println("conn.")
		}

		gctx := gstunnellib.NewGsContextImp(g_gid.GenerateId())
		//g_gstst.GetStatusConnList().Add(gctx.GetGsId(), gstClient, rawService)

		go srcTOdstUn(gstClient, rawService, gctx)
		go srcTOdstP(rawService, gstClient, gctx)
		g_Logger.Printf("go [%d].\n", gctx.GetGsId())
	}
}

func createToRawServiceConnHandler_test(chanconnlist <-chan net.Conn, rawServiceAddr string) {

	const maxTryConnService int = 6
	var err error
	for !g_app_close.Load() {

		gstClient, ok := <-chanconnlist
		if !ok {
			checkError_info(errors.New("'gstServer, ok := <-chanconnlist' is not ok"))
			return
		}
		g_log_List.GSIpLogger.Printf("Gstunnel client ip: %s\n", gstClient.RemoteAddr().String())

		connServiceError_count := 0
		var rawService net.Conn
		for !g_app_close.Load() {
			rawService, err = net.Dial("tcp", rawServiceAddr)
			checkError_NoExit(err)
			connServiceError_count += 1
			if err == nil {
				break
			}
			if connServiceError_count > maxTryConnService {
				checkError_NoExit(
					fmt.Errorf("connService_count > maxTryConnService(%d)", maxTryConnService))
				gstClient.Close()
				break
			}
			//g_Logger.Println("conn.")
		}
		if err != nil {
			continue
		}

		gctx := gstunnellib.NewGsContextImp(g_gid.GenerateId())
		//g_gstst.GetStatusConnList().Add(gctx.GetGsId(), gstClient, rawService)

		go srcTOdstUn(gstClient, rawService, gctx)
		go srcTOdstP(rawService, gstClient, gctx)
		g_Logger.Printf("go [%d].\n", gctx.GetGsId())
	}
}

func createToRawServiceConnHandler_pool(chanconnlist <-chan net.Conn, rawServiceAddr string) {

	const maxTryConnService int = 6
	var err error
	for !g_app_close.Load() {

		gstClient, ok := <-chanconnlist
		if !ok {
			checkError_info(errors.New("'gstServer, ok := <-chanconnlist' is not ok"))
			return
		}
		g_log_List.GSIpLogger.Printf("Gstunnel client ip: %s\n", gstClient.RemoteAddr().String())

		connServiceError_count := 0
		var rawService net.Conn
		for !g_app_close.Load() {
			rawService, err = net.Dial("tcp", rawServiceAddr)
			checkError_NoExit(err)
			connServiceError_count += 1
			if err == nil {
				break
			}
			if connServiceError_count > maxTryConnService {
				checkError_NoExit(
					fmt.Errorf("connService_count > maxTryConnService(%d)", maxTryConnService))
				gstClient.Close()
				break
			}
			//g_Logger.Println("conn.")
		}
		if err != nil {
			continue
		}

		gctx := gstunnellib.NewGsContextImp(g_gid.GenerateId())

		g_ants_pool.Submit(func() {
			srcTOdstUn(gstClient, rawService, gctx)
		})

		g_ants_pool.Submit(func() {
			srcTOdstP(rawService, gstClient, gctx)
		})

		g_Logger.Printf("go [%d].\n", gctx.GetGsId())
	}
}

func run() {
	//defer gstunnellib.Panic_Recover_GSCtx(g_Logger, gctx)

	var lstnaddr, rawServiceAddr string

	lstnaddr = g_gsconfig.Listen
	rawServiceAddr = g_gsconfig.GetServer_rand()

	g_Logger.Println("Listen_Addr:", lstnaddr)
	g_Logger.Println("Conn_Addr:", rawServiceAddr)
	g_Logger.Println("Begin......")

	tcpAddr, err := net.ResolveTCPAddr("tcp4", lstnaddr)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	//var rawService net.Conn

	chanConnList := make(chan net.Conn, 10240)
	defer close(chanConnList)

	for i := 0; i < g_connHandleGoNum; i++ {
		go createToRawServiceConnHandler(chanConnList, rawServiceAddr)
	}

	for {
		Client, err := listener.Accept()
		if err != nil {
			checkError_NoExit(err)
			continue
		}
		chanConnList <- Client
	}
}

func run_pipe_test_listen(lstnAddr, rawServiceAddr string, outListener *net.TCPListener) {
	//defer gstunnellib.Panic_Recover_GSCtx(g_Logger, gctx)

	//	var lstnaddr, rawServiceAddr string

	//	lstnaddr = g_gsconfig.Listen
	//	rawServiceAddr = g_gsconfig.GetServer_rand()

	g_Logger.Println("Listen_Addr:", lstnAddr)
	g_Logger.Println("Conn_Addr:", rawServiceAddr)
	g_Logger.Println("Begin......")

	tcpAddr, err := net.ResolveTCPAddr("tcp4", lstnAddr)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	*outListener = *listener

	//var rawService net.Conn

	chanConnList := make(chan net.Conn, 10240)
	defer close(chanConnList)

	for i := 0; i < g_connHandleGoNum; i++ {
		go createToRawServiceConnHandler_test(chanConnList, rawServiceAddr)
	}

	wg := sync.WaitGroup{}

	for i := 0; i < g_connHandleGoNum/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for !g_app_close.Load() {
				Client, err := listener.Accept()
				if errors.Is(err, net.ErrClosed) {
					return
				} else if err != nil {
					checkError_NoExit(err)
					continue
				}
				chanConnList <- Client
			}
		}()
	}
	wg.Wait()
}

func run_pipe_test_listen_pool(lstnAddr, rawServiceAddr string, outListener *net.TCPListener) {
	//defer gstunnellib.Panic_Recover_GSCtx(g_Logger, gctx)

	//	var lstnaddr, rawServiceAddr string

	//	lstnaddr = g_gsconfig.Listen
	//	rawServiceAddr = g_gsconfig.GetServer_rand()

	g_Logger.Println("Listen_Addr:", lstnAddr)
	g_Logger.Println("Conn_Addr:", rawServiceAddr)
	g_Logger.Println("Begin......")

	tcpAddr, err := net.ResolveTCPAddr("tcp4", lstnAddr)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	*outListener = *listener

	//var rawService net.Conn

	chanConnList := make(chan net.Conn, 10240)
	defer close(chanConnList)

	for i := 0; i < g_connHandleGoNum; i++ {
		go createToRawServiceConnHandler_pool(chanConnList, rawServiceAddr)
	}

	wg := sync.WaitGroup{}

	for i := 0; i < g_connHandleGoNum/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for !g_app_close.Load() {
				Client, err := listener.Accept()
				if errors.Is(err, net.ErrClosed) {
					return
				} else if err != nil {
					checkError_NoExit(err)
					continue
				}
				chanConnList <- Client
			}
		}()
	}
	wg.Wait()
}

// service to gstunnel client
func srcTOdstP(src net.Conn, dst net.Conn, gctx gstunnellib.GsContext) {
	if strings.EqualFold(g_IOModel, "rw") {
		srcTOdst_rw(src, dst, gctx)
	} else {
		srcTOdst_copy(src, dst, gctx)
	}
}

// gstunnel client to service
func srcTOdstUn(src net.Conn, dst net.Conn, gctx gstunnellib.GsContext) {
	if strings.EqualFold(g_IOModel, "rw") {
		srcTOdst_rw(src, dst, gctx)
	} else {
		srcTOdst_copy(src, dst, gctx)
	}
}
