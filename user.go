package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	// 启动这个User的goroutine，专门用于监听自己的Channel消息
	// server只需要往User的Channel写入消息即可 异步写入、收取，通过goroutine通信
	go user.ListenMessage()
	return user
}

func (u *User) ListenMessage() {
	// go启动
	// 监听User的Channel（收件箱），一旦有消息，写回给客户端
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}

func (u *User) Online() {
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	u.server.BroadCast(u, "上线了")
}

func (u *User) Offline() {
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	u.server.BroadCast(u, "下线了")
}

func (u *User) sendMsg(msg string) {
	u.conn.Write([]byte(msg))
}

// DoMessage 用户处理自己发出的消息
func (u *User) DoMessage(msg string) {
	if msg == "/who" {
		onlineList := "[系统] 当前用户在线列表:\n"
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineList += ("[" + user.Addr + "]" + user.Name + "\n")
		}
		u.server.mapLock.Unlock()
		u.sendMsg(onlineList)
	} else if len(msg) > 8 && msg[:8] == "/rename " {
		// /rename
		newName := msg[8:]
		if _, ok := u.server.OnlineMap[newName]; ok {
			u.sendMsg("[系统] 用户名已经被使用\n")
		} else {
			onlineMap := u.server.OnlineMap
			u.server.mapLock.Lock()
			delete(onlineMap, u.Name)
			u.Name = newName
			onlineMap[msg[8:]] = u
			u.server.mapLock.Unlock()
			u.sendMsg("[系统] 用户名已经被修改为:" + msg[8:] + "\n")
		}
	} else if len(msg) > 5 && msg[:4] == "/to " {
		// /to u x
		receiverName := strings.Split(msg, " ")[1]
		sendMsg := msg[5+len(receiverName):]

		if receiver, ok := u.server.OnlineMap[receiverName]; ok {
			receiver.sendMsg("[私聊] " + u.Name + ":" + sendMsg + "\n")
		} else {
			u.sendMsg("[系统] 找不到此用户\n")
		}
	} else {
		u.server.BroadCast(u, msg)
	}
}
