package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
}

func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
	}
	// 启动这个User的goroutine，专门用于监听自己的Channel消息
	// server只需要往User的Channel写入消息即可 异步写入、收取，通过goroutine通信
	go user.ListenMessage()
	return user
}

func (this *User) ListenMessage() {
	// go启动
	// 监听User的Channel（收件箱），一旦有消息，写回给客户端
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
