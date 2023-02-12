package main

import (
	"net"
	"strings"
)

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

// SendMsg 给当前User对应的客户端发送消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// DoMessage 处理消息
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前的在线用户

		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			this.SendMsg(onlineMsg)

		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 消息格式： rename|zhangsan
		newName := strings.Split(msg, "|")[1]
		// 判断name是否存在
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("当前用户名被使用\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("您已经更新用户名:" + this.Name + "\n")
		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 消息格式 to｜zhangsan｜消息内容

		// 1 获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg("消息格式不正确，请使用\"to|zhangsan|消息内容\"格式。 \n")
			return
		}

		// 2 根据用户名 得到对方的对象User
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("用户名不存在\n")
			return
		}

		// 3 获取消息内容，通过对方的User对象将消息内容发送出去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("无效消息内容，重发\n")
			return
		}
		remoteUser.SendMsg(this.Name + "对您说:" + content)

	} else {
		this.server.BroadCast(this, msg)
	}

}

// ListenMessage 监听当前user channel 的方法，一旦有消息 发给客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
