package client

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/net-byte/gexpose/common/enum"
	"github.com/net-byte/gexpose/common/netutil"
	"github.com/net-byte/gexpose/config"
)

// Start client
func Start(config config.Config) {
	log.Printf("client started \r\nlocal address is %v \r\nserver address is %v \r\nproxy address is %v", config.LocalAddr, config.ServerAddr, config.ProxyAddr)
	for {
		conn, err := net.DialTimeout("tcp", config.ServerAddr, time.Duration(config.Timeout)*time.Second)
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}
		log.Printf("server connected")
		go read(conn, config)
		ping(conn, config)
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
		case enum.CONNECT:
			go proxy(config)
		case enum.CLOSE:
			conn.Close()
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
}

func proxy(config config.Config) {
	proxyConn, proxyErr := net.DialTimeout("tcp", config.ProxyAddr, time.Duration(config.Timeout)*time.Second)
	if proxyErr != nil {
		log.Printf("failed to dial proxy address %v", proxyErr)
		return
	}
	localConn, localErr := net.DialTimeout("tcp", config.LocalAddr, time.Duration(config.Timeout)*time.Second)
	if localErr != nil {
		log.Printf("failed to dial local address %v", localErr)
		proxyConn.Close()
		return
	}
	go netutil.Copy(proxyConn, localConn, config.Key)
	go netutil.Copy(localConn, proxyConn, config.Key)
}
