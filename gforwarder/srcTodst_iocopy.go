package main

import (
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/ypcd/gforwarder/gstunnellib"
)

func srcTOdst_copy1(src net.Conn, dst net.Conn, gctx gstunnellib.GsContext) {
	defer gstunnellib.Panic_Recover_GSCtx(g_Logger, gctx)

	defer src.Close()
	defer dst.Close()

	nt_read := gstunnellib.NewNetTimeImpName("read")
	nt_write := gstunnellib.NewNetTimeImpName("write")
	var timew1 time.Time

	//outf, err := os.Create(fp1)
	//checkError_GsCtx(err,gctx)
	//outf2, err := os.Create(fp2)
	//checkError_GsCtx(err,gctx)

	//defer outf.Close()
	//defer outf2.Close()

	var wlent, rlent int64 = 0, 0

	defer func() {
		g_RuntimeStatistics.AddServerTotalNetData_recv(int(rlent))
		g_RuntimeStatistics.AddSrcTotalNetData_send(int(wlent))
		g_log_List.GSNetIOLen.Printf("[%d] gorou exit.\n\t%s\t%s\tpack  trlen:%d  twlen:%d\n\t%s\t%s",
			gctx.GetGsId(),
			gstunnellib.GetNetConnAddrString("src", src),
			gstunnellib.GetNetConnAddrString("dst", dst),
			rlent, wlent,
			nt_read.PrintString(),
			nt_write.PrintString(),
		)
		/*
			if g_Values.GetDebug() {

				//g_Logger.Println("goPackTotal:", goPackTotal)

				//	g_Logger.Println("RecoTime_p_r All: ", recot_p_r.StringAll())
				//	g_Logger.Println("RecoTime_p_w All: ", recot_p_w.StringAll())
			}
		*/
	}()

	err := src.SetReadDeadline(time.Now().Add(g_networkTimeout))
	if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrDeadlineExceeded) {
		checkError_info_GsCtx(err, gctx)
		return
	}
	err = dst.SetWriteDeadline(time.Now().Add(g_networkTimeout))
	if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrDeadlineExceeded) {
		checkError_info_GsCtx(err, gctx)
		return
	}
	timew1 = time.Now()
	wlen, err := io.Copy(dst, src)
	if err == nil {
		err = io.EOF
	}
	wlent += int64(wlen)
	nt_write.Add(time.Since(timew1))
	if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrDeadlineExceeded) {
		checkError_info_GsCtx(err, gctx)
		return
	} else {
		checkError_panic_GsCtx(err, gctx)
	}

}

func srcTOdst_copy(src net.Conn, dst net.Conn, gctx gstunnellib.GsContext) {

	defer src.Close()
	defer dst.Close()

	err := src.SetReadDeadline(time.Now().Add(g_networkTimeout))
	if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrDeadlineExceeded) {
		checkError_info_GsCtx(err, gctx)
		return
	}
	err = dst.SetWriteDeadline(time.Now().Add(g_networkTimeout))
	if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrDeadlineExceeded) {
		checkError_info_GsCtx(err, gctx)
		return
	}
	wlen, err := io.Copy(dst, src)
	if err == nil {
		err = io.EOF
	}
	_ = wlen
	//	wlent += int64(wlen)
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

}
