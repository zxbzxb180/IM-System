package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	// 用户在线列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播
	MsgChan chan string
}

// 创建一个server
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: map[string]*User{},
		MsgChan:   make(chan string),
	}

	return server
}

// 从服务端Chan读取消息，广播到每个用户的Chan
func (s *Server) ListenBoardCast() {
	for {
		msg := <-s.MsgChan
		// 广播消息
		s.mapLock.RLock()
		for _, i := range s.OnlineMap {
			i.Chan <- msg
		}
		s.mapLock.RUnlock()
	}
}

// 将消息发送至服务端Chan
func (s *Server) BoardCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ": " + msg
	s.MsgChan <- sendMsg
}

func (s *Server) Handle(conn net.Conn) {
	user := NewUser(conn)

	// 用户加入OnlineMap
	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock()

	// 广播用户上线消息
	s.BoardCast(user, "用户已上线")

}

// 启动服务器
func (s *Server) Start() {
	// socket listen
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	defer listen.Close()

	go s.ListenBoardCast()

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		go s.Handle(conn)

	}

}
