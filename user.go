package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string // channel
	conn net.Conn    //

	server *Server
}

// NewUser 创建一个用户
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	// 启动监听当前的user channel消息的goroutine
	go user.ListenMessage()

	return user
}

// Online 用户上线的功能
func (this *User) Online() {
	// 当前用户上线，将用户加入到onlinemap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 广播当前用户上线消息
	this.server.BroadCast(this, "已上线")

}

// Offline 用户下线功能
func (this *User) Offline() {
	// 下线 从map删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	// 广播当前用户下线
	this.server.BroadCast(this, "已下线")
}

// DoMessage 处理消息
func (this *User) DoMessage(msg string) {
	this.server.BroadCast(this, msg)
}

// ListenMessage 监听当前user channel 的方法，一旦有消息 发给客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
