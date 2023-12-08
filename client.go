package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "server ip")
	flag.IntVar(&serverPort, "port", 8081, "server port")
}

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	status     int
}

// NewClient 创建连接并返回Client对象
func NewClient(serverIp string, serverPort int) *Client {
	//解析命令行
	flag.Parse()
	// 连接
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("cet.Dial err:", err)
		return nil
	}
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		conn:       conn,
		status:     -1,
	}
	return client
}

func (client *Client) menu() bool {
	var flag int
	fmt.Println(">>>键入序号:")
	fmt.Println("1.查看在线用户列表")
	fmt.Println("2.进入私聊模式")
	fmt.Println("3.进入群聊模式")
	fmt.Println("4.更新用户名")
	fmt.Println("5.退出")

	fmt.Scanln(&flag)

	if flag < 1 || flag > 4 {
		fmt.Println(">>>非法序号...")
		return true
	}
	client.status = flag
	return false
}

func (client *Client) speakTo(name, msg string) {
	client.conn.Write([]byte(fmt.Sprintf("/to %s %s\n", name, msg)))
}

func (client *Client) broadcast(msg string) {
	client.conn.Write([]byte(msg + "\n"))
}

func (client *Client) run() {
	for client.status != 0 {
		for client.menu() {

		}
		//printCmdList()
		// 解析命令
		switch client.status {
		case 1:
			client.conn.Write([]byte("/who\n"))
			break
		case 2:
			//var input string
			//fmt.Scanln(&input)
			//client.speakTo()
			break
		case 3:
			var input string
			fmt.Scanln(&input)
			client.conn.Write([]byte(input + "\n"))
			break
		case 4:
			break
		}
	}
}

func (client *Client) receive() {
	//永久阻塞监听
	io.Copy(os.Stdout, client.conn)
}

func main() {
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>服务端连接失败...")
		return
	}
	fmt.Println(">>>服务端连接成功...")
	go client.receive()
	client.run()
	fmt.Println(">>>客户端退出...")

	select {}
}
