package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	//消息广播的Channel
	Message chan string
}

func NewServer(ip string, port int) *Server {
	return &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: map[string]*User{},
		Message:   make(chan string),
	}
}

func (s *Server) Handler(conn net.Conn) {
	fmt.Println("tcp链接建立成功")
	// 每个链接创建一个user
	user := NewUser(conn, s)
	// 广播上线
	user.Online()

	// 读取用户发送的消息
	go func() {
		for {
			buf := make([]byte, 4096)
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			msg := string(buf[:n-1]) // 去掉用户输入的换行符
			user.DoMessage(msg)
		}
	}()

	select {
	// 目前仅用于阻塞goroutine
	}
}

// BroadCast 使用某个user的身份发出广播
func (s *Server) BroadCast(user *User, msg string) {
	//TODO 防止将广播消息发送给自己 msg需要消息头[from] 如果from==user 直接return
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- sendMsg
}

func (s *Server) listenBroadCastMsg() {
	for {
		msg := <-s.Message
		// 广播消息给所有用户
		s.mapLock.Lock()
		for _, user := range s.OnlineMap {
			user.C <- msg
		}
		s.mapLock.Unlock()
	}
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	defer listener.Close()

	go s.listenBroadCastMsg()

	fmt.Println("server start...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			continue
		}

		// 为每个链接单独分配一个携程
		go s.Handler(conn)
	}
}
