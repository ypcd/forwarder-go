package gstunnellib

import (
	"net"
	"os"
)

func NetConnWriteAll(dst net.Conn, buf []byte) (int64, error) {
	var wlen int = 0

	for {
		wsz, err := dst.Write(buf)
		wlen += wsz
		if err != nil {
			return 0, err
		}
		if wlen == len(buf) {
			return int64(wlen), err
		} else if wlen > len(buf) {
			panic("error wlen>len(buf)")
		}
		buf = buf[wsz:]
	}
}

// 比io.copy快12.9%
func netConnWriteAll_test(dst *os.File, buf []byte) (int, error) {
	wlen := 0

	for {
		wsz, err := dst.Write(buf)
		wlen += wsz
		if err != nil {
			return 0, err
		}
		if wlen == len(buf) {
			return wlen, err
		} else if wlen > len(buf) {
			panic("error wlen>len(buf)")
		}
		buf = buf[wsz:]
	}
}
