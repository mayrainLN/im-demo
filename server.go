package main

import (
	"fmt"
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

func (this *Server) Handler(conn net.Conn) {
	fmt.Println("tcp链接建立成功")
	// 每个链接创建一个user
	user := NewUser(conn)
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	// 发出广播
	this.BroadCast(user, "已上线")

	select {
	// 目前仅用于阻塞goroutine
	}
}

// 广播用户上线
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
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

func (this *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	defer listener.Close()

	go this.listenBroadCastMsg()

	fmt.Println("server satrt...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			continue
		}

		// 为每个链接单独分配一个携程
		go this.Handler(conn)
	}
}
