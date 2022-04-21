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

type Server struct {
	config             config.Config
	clientConn         net.Conn
	connPool           sync.Map
	notifyNewProxyConn chan int
}

type ConnMapping struct {
	proxyConn  *net.Conn
	exposeConn *net.Conn
	addTime    int64
	mapped     bool
}

// Start server
func Start(config config.Config) {
	log.Println("server started")
	s := &Server{config: config, notifyNewProxyConn: make(chan int)}
	go s.listenServerAddr()
	go s.listenExposeAddr()
	go s.listenProxyAddr()
	go s.cleanJob()
	s.forwardJob()
}

func (s *Server) listenServerAddr() {
	ln, err := net.Listen("tcp", s.config.ServerAddr)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("server address is %v", s.config.ServerAddr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		if s.clientConn != nil {
			log.Printf("client already connected")
			conn.Close()
			continue
		}
		s.clientConn = conn
		log.Printf("a new client connection from %v", s.clientConn.RemoteAddr().String())
		go s.read(s.clientConn)
		go s.ping(s.clientConn)
	}
}

func (s *Server) read(conn net.Conn) {
	defer conn.Close()
	packet := make([]byte, 1024)
	for {
		conn.SetReadDeadline(time.Now().Add(time.Duration(s.config.Timeout) * time.Second))
		n, err := conn.Read(packet)
		if err != nil || err == io.EOF {
			break
		}
		b := packet[:n]
		switch b[0] {
		case enum.PING:
			conn.Write([]byte{enum.PONG})
		case enum.CLOSE:
			conn.Close()
		}
	}
}

func (s *Server) ping(conn net.Conn) {
	defer conn.Close()
	for {
		conn.SetWriteDeadline(time.Now().Add(time.Duration(s.config.Timeout) * time.Second))
		_, err := conn.Write([]byte{enum.PING})
		if err != nil {
			break
		}
		time.Sleep(3 * time.Second)
	}
	s.cleanClient()
}

func (s *Server) listenExposeAddr() {
	ln, err := net.Listen("tcp", s.config.ExposeAddr)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("expose address is %v", s.config.ExposeAddr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		if s.clientConn == nil {
			conn.Close()
			continue
		}
		s.addConn(&conn)
		s.notityClient()
	}
}

func (s *Server) addConn(conn *net.Conn) {
	key := strconv.FormatInt(time.Now().UnixNano(), 10)
	s.connPool.Store(key, &ConnMapping{nil, conn, time.Now().Unix(), false})
}

func (s *Server) notityClient() {
	s.clientConn.Write([]byte{enum.CONNECT})
}

func (s *Server) listenProxyAddr() {
	ln, err := net.Listen("tcp", s.config.ProxyAddr)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("proxy address is %v", s.config.ProxyAddr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		s.mappingProxyConn(&conn)
	}
}

func (s *Server) mappingProxyConn(conn *net.Conn) {
	mapped := false
	s.connPool.Range(func(k, v interface{}) bool {
		mapping := v.(*ConnMapping)
		if !mapping.mapped && mapping.exposeConn != nil {
			mapping.proxyConn = conn
			mapping.mapped = true
			mapped = true
			return false
		}
		return true
	})
	if !mapped {
		(*conn).Close()
		return
	}
	s.notifyNewProxyConn <- 0
}

func (s *Server) forwardJob() {
	for {
		select {
		case <-s.notifyNewProxyConn:
			s.connPool.Range(func(k, v interface{}) bool {
				mapping := v.(*ConnMapping)
				if mapping.mapped && mapping.proxyConn != nil && mapping.exposeConn != nil {
					go netutil.Copy(*mapping.exposeConn, *mapping.proxyConn, s.config.Key)
					go netutil.Copy(*mapping.proxyConn, *mapping.exposeConn, s.config.Key)
					s.connPool.Delete(k)
				}
				return true
			})
		}
	}
}

func (s *Server) cleanJob() {
	for {
		s.connPool.Range(func(k, v interface{}) bool {
			mapping := v.(*ConnMapping)
			if !mapping.mapped && mapping.exposeConn != nil {
				if time.Now().Unix()-mapping.addTime > int64(s.config.Timeout) {
					log.Printf("clean the expired conn %v", (*mapping.exposeConn).RemoteAddr().String())
					(*mapping.exposeConn).Close()
					s.connPool.Delete(k)
				}
			}
			return true
		})
		time.Sleep(10 * time.Second)
	}
}

func (s *Server) cleanClient() {
	log.Println("client disconnected")
	s.clientConn = nil
	s.connPool.Range(func(k, v interface{}) bool {
		s.connPool.Delete(k)
		return true
	})
	log.Println("clean conn pool")
}
