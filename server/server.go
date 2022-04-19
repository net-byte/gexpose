package server

import (
	"io"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/net-byte/gexpose/common/enum"
	"github.com/net-byte/gexpose/common/netutil"
	"github.com/net-byte/gexpose/config"
)

var _clientConn net.Conn
var _connPool = make(map[string]*ConnMapping)
var _lock = sync.Mutex{}
var _notifyIncomingProxyConn = make(chan int)

type ConnMapping struct {
	proxyConn  *net.Conn
	exposeConn *net.Conn
	addTime    int64
	mapped     bool
}

// Start server
func Start(config config.Config) {
	go listenServerAddr(config)
	go listenExposeAddr(config)
	go listenProxyAddr(config)
	go cleanJob()
	forwardJob(config)
}

func listenServerAddr(config config.Config) {
	ln, err := net.Listen("tcp", config.ServerAddr)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("server address is %v", config.ServerAddr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		if _clientConn != nil {
			log.Printf("client already connected")
			conn.Close()
			continue
		}
		_clientConn = conn
		log.Printf("an incoming client connection from %v", _clientConn.RemoteAddr().String())
		go read(_clientConn, config)
		go ping(_clientConn, config)
	}
}

func read(conn net.Conn, config config.Config) {
	defer conn.Close()
	packet := make([]byte, 1024)
	for {
		conn.SetReadDeadline(time.Now().Add(time.Duration(config.Timeout) * time.Second))
		n, err := conn.Read(packet)
		if err != nil || err == io.EOF {
			break
		}
		b := packet[:n]
		switch b[0] {
		case enum.PING:
			conn.Write([]byte{enum.PONG})
		case enum.PONG:
		case enum.CLOSE:
			conn.Close()
		default:
			log.Printf("received an unsupported msg from client")
		}
	}
}

func ping(conn net.Conn, config config.Config) {
	defer conn.Close()
	for {
		conn.SetWriteDeadline(time.Now().Add(time.Duration(config.Timeout) * time.Second))
		_, err := conn.Write([]byte{enum.PING})
		if err != nil {
			break
		}
		time.Sleep(3 * time.Second)
	}
	cleanClient()
}

func listenExposeAddr(config config.Config) {
	ln, err := net.Listen("tcp", config.ExposeAddr)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("expose address is %v", config.ExposeAddr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		addConn(&conn)
		notityClient()
	}
}

func addConn(conn *net.Conn) {
	_lock.Lock()
	defer _lock.Unlock()
	now := time.Now().UnixNano()
	_connPool[strconv.FormatInt(now, 10)] = &ConnMapping{nil, conn, time.Now().Unix(), false}
}

func notityClient() {
	if _clientConn == nil {
		log.Printf("no client connected")
		return
	}
	_clientConn.Write([]byte{enum.CONNECT})
}

func listenProxyAddr(config config.Config) {
	ln, err := net.Listen("tcp", config.ProxyAddr)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("proxy address is %v", config.ProxyAddr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		mappingProxyConn(&conn)
	}
}

func mappingProxyConn(conn *net.Conn) {
	_lock.Lock()
	mapped := false
	for _, mapping := range _connPool {
		if !mapping.mapped && mapping.exposeConn != nil {
			mapping.proxyConn = conn
			mapping.mapped = true
			mapped = true
			break
		}
	}
	if !mapped {
		(*conn).Close()
	}
	_lock.Unlock()
	_notifyIncomingProxyConn <- 0
}

func forwardJob(config config.Config) {
	for {
		select {
		case <-_notifyIncomingProxyConn:
			_lock.Lock()
			for key, mapping := range _connPool {
				if mapping.mapped && mapping.proxyConn != nil && mapping.exposeConn != nil {
					go netutil.Copy(*mapping.exposeConn, *mapping.proxyConn, config.Key)
					go netutil.Copy(*mapping.proxyConn, *mapping.exposeConn, config.Key)
					delete(_connPool, key)
				}
			}
			_lock.Unlock()
		}
	}
}

func cleanJob() {
	for {
		_lock.Lock()
		for key, mapping := range _connPool {
			if !mapping.mapped && mapping.exposeConn != nil {
				if time.Now().Unix()-mapping.addTime > 10 {
					log.Printf("clean the expired conn %v", (*mapping.exposeConn).RemoteAddr().String())
					(*mapping.exposeConn).Close()
					delete(_connPool, key)
				}
			}
		}
		_lock.Unlock()
		time.Sleep(5 * time.Second)
	}
}

func cleanClient() {
	log.Println("client disconnected")
	_clientConn = nil
	for k := range _connPool {
		delete(_connPool, k)
	}
	log.Println("clean all connections")
}
