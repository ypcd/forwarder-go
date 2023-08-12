package main

import (
	"bytes"
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/ypcd/gforwarder/gstunnellib"
)

func srcTOdst_rw(src net.Conn, dst net.Conn, gctx gstunnellib.GsContext) {
	defer gstunnellib.Panic_Recover_GSCtx(g_Logger, gctx)

	defer src.Close()
	defer dst.Close()

	nt_read := gstunnellib.NewNetTimeImpName("read")
	nt_write := gstunnellib.NewNetTimeImpName("write")
	var timer1, timew1 time.Time

	var wbuf []byte
	var rbuf []byte = make([]byte, g_net_read_size)

	var wlent, rlent int64 = 0, 0

	ChangeCryKey_Total := 0

	defer func() {
		g_RuntimeStatistics.AddServerTotalNetData_recv(int(rlent))
		g_RuntimeStatistics.AddSrcTotalNetData_send(int(wlent))
		g_log_List.GSNetIOLen.Printf("[%d] gorou exit.\n\t%s\t%s\tpack  trlen:%d  twlen:%d  ChangeCryKey_total:%d\n\t%s\t%s",
			gctx.GetGsId(),
			gstunnellib.GetNetConnAddrString("src", src),
			gstunnellib.GetNetConnAddrString("dst", dst),
			rlent, wlent, ChangeCryKey_Total,
			nt_read.PrintString(),
			nt_write.PrintString(),
		)

		if g_Values.GetDebug() {

			//g_Logger.Println("goPackTotal:", goPackTotal)

			//	g_Logger.Println("RecoTime_p_r All: ", recot_p_r.StringAll())
			//	g_Logger.Println("RecoTime_p_w All: ", recot_p_w.StringAll())
		}
	}()

	for {
		//	recot_p_r.Run()
		src.SetReadDeadline(time.Now().Add(g_networkTimeout))
		timer1 = time.Now()
		rlen, err := src.Read(rbuf)
		rlent += int64(rlen)
		//	recot_p_r.Run()
		if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) ||
			errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrDeadlineExceeded) {
			checkError_info_GsCtx(err, gctx)
			return
		} else {
			checkError_panic_GsCtx(err, gctx)
		}
		nt_read.Add(time.Now().Sub(timer1))
		/*
			if tmr_out.Run() {
				g_Logger.Printf("Error: [%d] Time out, func exit.\n", gctx.GetGsId())
				return
			}
		*/
		if rlen == 0 {
			g_Logger.Println("Error: src.read() rlen==0 func exit.")
			return
		}

		//outf.Write(buf[:rlen])
		//tmr_out.Boot()
		//rbuf = buf

		if rlen > 0 {
			if gstunnellib.G_RunTime_Debug {
				gstunnellib.G_RunTimeDebugInfo1.AddPackingPackSizeList("server_srcToDstP_st_packing len", rlen)
			}
			wbuf = rbuf[:rlen]

			if len(wbuf) <= 0 {
				g_Logger.Println("Error: gspack.packing is error.")
				return
			}
			dst.SetWriteDeadline(time.Now().Add(g_networkTimeout))
			timew1 = time.Now()
			wlen, err := io.Copy(dst, bytes.NewBuffer(wbuf))
			wlent += int64(wlen)
			if err != nil {
				if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) ||
					errors.Is(err, os.ErrDeadlineExceeded) || strings.Contains(err.Error(), "reset") {
					checkError_info_GsCtx(err, gctx)
					return
				} else {
					checkError_NoExit_GsCtx(err, gctx)
					return
				}
			}
			nt_write.Add(time.Since(timew1))
			//tmr_out.Boot()
		}

	}
}
