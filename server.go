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

func (this *Server) Handler(conn net.Conn) {
	fmt.Println("tcp链接建立成功")
	// 每个链接创建一个user
	user := NewUser(conn)
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	// 广播上线
	this.BroadCast(user, "上线")

	// 读取用户发送的消息
	go func() {
		for {
			buf := make([]byte, 4096)
			n, err := conn.Read(buf)
			if n == 0 {
				this.BroadCast(user, "下线")
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			msg := string(buf[:n-1]) // 去掉用户输入的换行符
			this.BroadCast(user, msg)
		}
	}()

	select {
	// 目前仅用于阻塞goroutine
	}
}

// 使用某个user的身份发出广播
func (this *Server) BroadCast(user *User, msg string) {
	//TODO 防止将广播消息发送给自己 msg需要消息头[from] 如果from==user 直接return
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
