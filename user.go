package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string // channel
	conn net.Conn    // 连接
}

// NewUser 创建一个用户
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
	}
	// 启动监听当前的user channel消息的goroutine
	go user.ListenMessage()

	return user
}

// ListenMessage 监听当前user channel 的方法，一旦有消息 发给客户端
func (this *User) ListenMessage() {
	msg := <-this.C
	this.conn.Write([]byte(msg + "\n"))
}
