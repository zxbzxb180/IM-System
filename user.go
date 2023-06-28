package main

import (
	"fmt"
	"github.com/DanPlayer/randomname"
	"net"
)

type User struct {
	Name string
	Addr string
	Chan chan string
	conn net.Conn
}

func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: randomname.GenerateName(),
		Addr: userAddr,
		Chan: make(chan string),
		conn: conn,
	}

	// 启动监听user chan的协程
	go user.ListenMessage()

	return user
}

func (u User) ListenMessage() {
	for {
		msg, ok := <-u.Chan
		if !ok {
			fmt.Println("channel closed!")
		}
		u.conn.Write([]byte(msg + "\n"))
	}

}
