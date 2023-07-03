package main

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
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

	// 监听用户是否活跃的channel
	isAlive := make(chan bool)

	go func() {
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
			user.SendMessage(msg, nil)

			isAlive <- true
		}
	}()
	for {
		select {
		case <-isAlive:
			// 什么都不需要做，从isAlive可以读到数据，判断用户在线，继续循环重置select
		case <-time.After(100 * time.Second):
			// 在进入select时，会同时判断多个case是否满足执行条件，达到准备就绪的状态，然后执行准备就绪的case
			// time.After(10 * time.Second)在进入select时就已经执行，并返回了一个channel给select判断是否可读
			// select在10秒后定时器触发，time.After(10 * time.Second)返回的channel可以读到数据了，达到准备就绪状态，就可以执行case里的代码了
			// 已经超时了，所以将用户强制下线，并退出循环
			user.conn.Write([]byte("已超时\n"))
			time.Sleep(1 * time.Second)
			// 关闭用户channel
			close(user.Chan)

			// 退出当前协程
			runtime.Goexit()

		}
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
