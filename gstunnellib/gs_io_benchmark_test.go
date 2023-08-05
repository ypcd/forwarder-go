package gstunnellib

import (
	"bytes"
	randc "crypto/rand"
	"io"
	"os"
	"testing"
)

func in_getRDBytes_local(byteLen int) []byte {
	data := make([]byte, byteLen)
	_, err := randc.Reader.Read(data)
	if err != nil {
		panic(err)
	}
	return data
}

func Benchmark_iocopy(b *testing.B) {
	testCacheSize := 16 * 1024 * 1024
	SendData := in_getRDBytes_local(testCacheSize)
	fd, err := os.OpenFile(os.DevNull, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	CheckError_panic(err)
	defer fd.Close()

	for n := 0; n < b.N; n++ {
		_, err = io.Copy(fd, bytes.NewBuffer(SendData))
		CheckError_panic(err)
	}
}

func Benchmark_NetConnWriteAll(b *testing.B) {
	testCacheSize := 16 * 1024 * 1024
	SendData := in_getRDBytes_local(testCacheSize)
	fd, err := os.OpenFile(os.DevNull, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	CheckError_panic(err)
	defer fd.Close()

	for n := 0; n < b.N; n++ {
		_, err = netConnWriteAll_test(fd, SendData)
		CheckError_panic(err)
	}
}
