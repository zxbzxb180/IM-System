package main

import (
	"fmt"
	"github.com/DanPlayer/randomname"
	"net"
	"strings"
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
func (u *User) SendMessage(msg string, tUser *User) {
	fmt.Println(msg)
	if msg == "#who" {
		// 查询当前在线用户的指令
		u.server.mapLock.RLock()
		result := "在线用户: \n"
		for k := range u.server.OnlineMap {
			result = result + k + "\n"
		}
		u.conn.Write([]byte(result))
		u.server.mapLock.RUnlock()
	} else if len(msg) > 7 && msg[:8] == "#rename|" {
		// 重命名
		splitName := strings.Split(msg, "|")[1:]
		newName := strings.Join(splitName, "|")
		if _, ok := u.server.OnlineMap[newName]; ok {
			u.conn.Write([]byte("当前用户名已存在\n"))
		} else {
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.Name = newName
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()
			_, err := u.conn.Write([]byte("您已更新用户名: " + newName + "\n"))
			if err != nil {
				fmt.Println(fmt.Sprintf("conn write err: %+v", err.Error()))
			}
		}
	} else {
		u.server.BoardCast(u, msg)
	}
}

func (u User) ListenMessage() {
	for msg := range u.Chan {
		_, err := u.conn.Write([]byte(msg + "\n"))
		if err != nil {
			//fmt.Println("conn write err: ", err)
			fmt.Println(fmt.Sprintf("conn write err: %+v", err.Error()))
		}
	}
	err := u.conn.Close()
	if err != nil {
		return
	}

}
