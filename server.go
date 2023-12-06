package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
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

	// 监听用户是否活跃的Channel
	isLive := make(chan bool)

	// 新建goroutine，读取Socket中的消息
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
			// 去掉用户输入的换行符
			msg := string(buf[:n-1])
			// 处理用户发送的消息
			user.DoMessage(msg)
			// 用户活跃
			isLive <- true
		}
	}()

	// 循环检测是否挂机
	for {
		select { //select会按照随机顺序遍历所有（如果有机会的话）的 case 表达式，只要有一个信道有接收到数据，那么 select 就结束
		case <-isLive:
			//当前用户是活跃的，应该重置定时器
			//不做任何事情，直接进入下一轮循环，重新求一轮值，下一个定时器会重新启动
		case <-time.After(time.Second * 30):
			//已经超时
			user.sendMsg("[系统] 长时间未操作,已被强制下线\n")
			close(user.C)
			conn.Close()
			return
		}

	}
}

// BroadCast 使用某个user的身份发出广播
func (s *Server) BroadCast(user *User, msg string) {
	//TODO 防止将广播消息发送给自己 msg需要消息头[from] 如果from==user 直接return
	//"[" + user.Addr + "]" +
	sendMsg := "[公聊] " + user.Name + ":" + msg
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

		// 为每个链接单独分配一个携程，用于Handler处理链接消息
		go s.Handler(conn)
	}
}
