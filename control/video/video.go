package video

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/lixiangyuan98/ogateway/conf"
)

// 下发给视频采集端的指令
type command struct {
	method string // 要执行的操作
	dest   string // 目的地址
	port   string // 目的端口
	src    string // 数据源
}

func newCommand(method, dest string, port string, src string) *command {
	return &command{
		method: method,
		dest:   dest,
		port:   port,
		src:    src,
	}
}

func (c *command) getBytes() []byte {
	message := make([]byte, 0)
	message = append(message, []byte(c.method)...)
	message = append(message, []byte(" ")...)
	message = append(message, []byte(c.dest)...)
	message = append(message, []byte(" ")...)
	message = append(message, []byte(c.port)...)
	message = append(message, []byte(" ")...)
	message = append(message, []byte(c.src)...)
	message = append(message, []byte("\r\n\r\n")...)
	return message
}

// 视频采集端上报的数据
type report struct {
	status string   // 采集端状态
	files  []string // 采集端存储的文件
}

func newReport(b []byte) *report {
	r := &report{}
	message := string(b)
	lines := strings.Split(message, "\r\n")
	r.status = lines[0]
	r.files = make([]string, 0)
	for _, line := range lines[1:] {
		r.files = append(r.files, line)
	}
	return r
}

// 视频采集端
type VideoServer struct {
	Host   string
	Status string
	Files  []string
}

func newVideoServer(host, status string, files []string) *VideoServer {
	return &VideoServer{
		Host:   host,
		Status: status,
		Files:  files,
	}
}

type connection struct {
	conn   *net.TCPConn
	timer  *time.Timer
	closed chan struct{}
	mu     sync.Mutex
	client *VideoServer
}

func newConnection() *connection {
	c := &connection{}
	return c
}

func (c *connection) open(conn *net.TCPConn) {
	c.conn = conn
	c.timer = time.NewTimer(conf.GlobalConf.VideoConf.VideoServerTimeout)
	c.closed = make(chan struct{})
	go func() {
		for {
			select {
			case <-c.timer.C:
				c.close()
				return
			case <-c.closed:
				return
			}
		}
	}()
}

func (c *connection) recv(b []byte) {
	c.timer.Reset(conf.GlobalConf.VideoConf.VideoServerTimeout)
	r := newReport(b)
	log.Printf("recv report from %v: %v\n", c.getHost(), r)
	c.mu.Lock()
	c.client = newVideoServer(c.getAddr(), r.status, r.files)
	c.mu.Unlock()
}

func (c *connection) send(method, dest string, port string, src string) error {
	cmd := newCommand(method, dest, port, src)
	b := cmd.getBytes()
	_, err := c.conn.Write(b)
	if err != nil {
		log.Printf("send command to %v error: %v\n", c.getHost(), err)
		return err
	}
	log.Printf("send command to %v: %v\n", c.getHost(), cmd)
	return nil
}

func (c *connection) close() {
	if err := c.conn.Close(); err != nil {
		log.Printf("close conn error: %v\n", err)
	}
	if c.timer.Stop() {
		c.closed <- struct{}{}
	}
	c.mu.Lock()
	c.client.Status = "inactive"
	c.mu.Unlock()
}

func (c *connection) getAddr() string {
	if c.conn == nil {
		return ""
	}
	return strings.Split(c.conn.RemoteAddr().String(), ":")[0]
}

func (c *connection) getHost() string {
	if c.conn == nil {
		return ""
	}
	return c.conn.RemoteAddr().String()
}

func (c *connection) getPort() string {
	if c.conn == nil {
		return ""
	}
	return strings.Split(c.conn.RemoteAddr().String(), ":")[1]
}

type server struct {
	bindAddr *net.TCPAddr
	listener *net.TCPListener
	conns    *sync.Map
}

var svr *server

func Init() {
	svr = &server{}
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:8000")
	if err != nil {
		log.Printf("resolve addr error: %v\n", err)
		return
	}
	svr.bindAddr = addr

	listener, err := net.ListenTCP("tcp", svr.bindAddr)
	if err != nil {
		log.Printf("listen error: %v\n", err)
		return
	}
	svr.listener = listener
	defer func() {
		if err := svr.listener.Close(); err != nil {
			log.Printf("close server error: %v\n", err)
		}
	}()

	conns := &sync.Map{}
	for _, c := range conf.VideoConf {
		conns.Store(c.Host, &connection{
			mu:     sync.Mutex{},
			client: newVideoServer(c.Host, "inactive", make([]string, 0)),
		})
	}
	svr.conns = conns

	for {
		conn, err := svr.listener.AcceptTCP()
		log.Printf("accept connection: %v\n", conn.RemoteAddr().String())
		if err != nil {
			log.Printf("accept connection error: %v\n", err)
			continue
		}
		go svr.recv(conn)
	}
}

func (s *server) getConn(conn *net.TCPConn) *connection {
	c, ok := s.conns.Load(strings.Split(conn.RemoteAddr().String(), ":")[0])
	if !ok {
		log.Printf("unrecognized connection from %v\n", conn.RemoteAddr().String())
		defer conn.Close()
		return nil
	}
	return c.(*connection)
}

func (s *server) recv(conn *net.TCPConn) {
	c := s.getConn(conn)
	if c == nil {
		return
	}
	c.open(conn)
	message := make([]byte, 0)
	buffer := make([]byte, 4096)
	for {
		len, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Printf("closed by remote: %v\n", c.getHost())
				c.close()
			} else {
				log.Printf("close connection: %v\n", c.getHost())
			}
			return
		}
		message = append(message, buffer[:len]...)
		if idx := bytes.Index(message, []byte("\r\n\r\n")); idx != -1 {
			c.recv(message[:idx])
			message = message[idx+4:]
		}
	}
}

// 发送指令到指定采集端
func Send(method, host, src, dest, port string) error {
	c, ok := svr.conns.Load(host)
	if !ok {
		return errors.New("no such host")
	}
	if c.(*connection).conn == nil || c.(*connection).client.Status != "active" {
		return errors.New("host inactive")
	}
	return c.(*connection).send(method, dest, port, src)
}

// 获取指定采集端的相关信息
func GetInfo(hosts []string) []*VideoServer {
	clients := make([]*VideoServer, 0)
	for _, host := range hosts {
		c, ok := svr.conns.Load(host)
		if !ok {
			continue
		}
		clients = append(clients, c.(*connection).client)
	}
	return clients
}
