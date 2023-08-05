package gstestpipe

import (
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// no echo
type RawServerSocket struct {
	connList        chan net.Conn
	close           atomic.Bool
	listening       bool
	serverAddr      string
	connListIn      []net.Conn
	lock_connListIn sync.Mutex
	listen_close    atomic.Bool
	listener        net.Listener
	lock_listener   sync.Mutex
}

func NewRawServerSocket_RandAddr() *RawServerSocket {
	return &RawServerSocket{connList: make(chan net.Conn, 10240),
		serverAddr: GetRandAddr()}
}

func (ss *RawServerSocket) listen() {
	if ss.listening {
		return
	}
	wgrun := sync.WaitGroup{}
	wgrun.Add(1)
	go func() {
		defer ss.listen_close.Swap(true)
		defer close(ss.connList)

		lst, err := net.Listen("tcp4", ss.serverAddr)
		checkError_exit(err)
		defer lst.Close()
		ss.lock_listener.Lock()
		ss.listener = lst
		ss.lock_listener.Unlock()
		wg := sync.WaitGroup{}
		for i := 0; i < 32; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for !ss.close.Load() {
					conna, err := lst.Accept()
					if errors.Is(err, net.ErrClosed) {
						return
					}
					checkError_exit(err)
					ss.lock_connListIn.Lock()
					ss.connListIn = append(ss.connListIn, conna)
					ss.lock_connListIn.Unlock()
					ss.connList <- conna
				}
			}()
		}
		wgrun.Done()
		wg.Wait()
	}()
	ss.listening = true
	wgrun.Wait()
}

func (ss *RawServerSocket) Run() { ss.listen() }

func (ss *RawServerSocket) Close() {
	ss.close.Swap(true)
	ss.lock_listener.Lock()
	ss.listener.Close()
	ss.lock_listener.Unlock()

	for conn := range ss.connList {
		conn.Close()
	}

	for {
		if ss.listen_close.Load() {
			ss.lock_connListIn.Lock()
			defer ss.lock_connListIn.Unlock()
			for _, v := range ss.connListIn {
				v.Close()
			}
			return
		}
		time.Sleep(time.Millisecond)
	}
}

func (ss *RawServerSocket) GetConnList() chan net.Conn { return ss.connList }
func (ss *RawServerSocket) GetServerAddr() string      { return ss.serverAddr }

type RawServerSocketEcho struct {
	RawServerSocket
	echo_running bool
}

func NewRawServerSocketEcho_RandAddr() *RawServerSocketEcho {
	return &RawServerSocketEcho{RawServerSocket: RawServerSocket{connList: make(chan net.Conn, 10240),
		serverAddr: GetRandAddr()}}
}

func (ss *RawServerSocketEcho) echoHandler() {
	if ss.echo_running {
		return
	}

	go func() {
		for !ss.close.Load() {
			for conn1 := range ss.connList {
				go func(conn net.Conn) {

					_, err := io.Copy(conn, conn)
					if errors.Is(err, net.ErrClosed) {
						checkError_info(err)
						return
					}
					checkError_exit(err)

				}(conn1)
			}
		}
	}()
	ss.echo_running = true
}

func (ss *RawServerSocketEcho) Run() {
	ss.listen()
	ss.echoHandler()
}
