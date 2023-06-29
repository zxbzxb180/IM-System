package main

import (
	"fmt"
	"io"
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

//func (s *Server) ReceiveUserMessage(conn net.Conn, user *User)  {
//
//}

func (s *Server) Handle(conn net.Conn) {
	user := NewUser(conn, s)
	user.Online()

	// 接收用户信息
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		// 长度为0, 读到0代表对端关闭了
		if n == 0 {
			user.Offline()
			return
		}
		// 读取出错，且非end of file
		if err != nil && err != io.EOF {
			//fmt.Println("conn write err: ", err)
			fmt.Println(fmt.Sprintf("conn write err: %+v", err))
			return
		}
		// 提取用户消息
		msg := string(buf[:n-1])

		// 广播用户发送的消息
		// TODO user.SendMessage
		s.BoardCast(user, msg)
	}

}

// 启动服务器
func (s *Server) Start() {
	// socket listen
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	defer func(listen net.Listener) {
		err := listen.Close()
		if err != nil {
			fmt.Println("close listener err: ", err)
		}
	}(listen)

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
