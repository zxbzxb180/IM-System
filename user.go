package main

import (
	"fmt"
	"github.com/DanPlayer/randomname"
	"net"
)

type User struct {
	Name   string
	Addr   string
	Chan   chan string
	conn   net.Conn
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   randomname.GenerateName(),
		Addr:   userAddr,
		Chan:   make(chan string),
		conn:   conn,
		server: server,
	}

	// 启动监听user chan的协程
	go user.ListenMessage()

	return user
}

// 用户上线
func (u *User) Online() {
	// 用户加入OnlineMap
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	// 广播用户上线消息
	u.server.BoardCast(u, "用户已上线")
}

// 用户下线
func (u *User) Offline() {
	// 用户移出OnlineMap
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	// 广播用户下线消息
	u.server.BoardCast(u, "用户已下线")
}

// 用户发送消息
func (u *User) SendMessage(msg string) {
	u.server.BoardCast(u, msg)
}

func (u User) ListenMessage() {
	for {
		msg, ok := <-u.Chan
		if !ok {
			fmt.Println("channel closed!")
		}
		_, err := u.conn.Write([]byte(msg + "\n"))
		if err != nil {
			//fmt.Println("conn write err: ", err)
			fmt.Println(fmt.Sprintf("conn write err: %+v", err))
		}
	}

}
