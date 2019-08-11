package main

import (
	"encoding/json"
	"github.com/jacobsa/go-serial/serial"
	"io"
	"io/ioutil"
	"log"
	"net"
)

type config struct {
	TcpAddr            string    `json:"tcpAddr"`
	SerialPortName     string `json:"serialPortName"`
	SerialPortBaudRate uint    `json:"serialPortBaudRate"`
}

var port io.ReadWriteCloser = nil

var isConnecting bool = false;

var socketConn net.Conn = nil


var sysConfig  = &config{}

func connHandler(c net.Conn) {
	if c == nil {
		log.Panic("socket error")
	}

	buf := make([]byte, 4096)
	//循环读取网络数据流
	for {
		//网络数据流读入 buffer
		cnt, err := c.Read(buf)
		//读取错误 关闭 socket 连接
		if cnt == 0 || err != nil {
			_ = c.Close()
			break
		}
		//把网络接收到的数据从串口写出
		if port != nil {
			port.Write(buf[0:cnt])
		}
	}
	isConnecting = false;
	socketConn = nil
	log.Printf("%v closed\n", c.RemoteAddr())
}

//开启serverSocket
func ServerSocket() {
	//监听端口
	server, err := net.Listen("tcp", sysConfig.TcpAddr)


	if err != nil {
		log.Println("start socket server error")
		return
	}

	log.Printf("tcp listening at:%s",sysConfig.TcpAddr)


	for {
		//接收来自 client 的连接,会阻塞
		conn, err := server.Accept()

		if err != nil {
			log.Println("link error")
		}

		if !isConnecting {
			isConnecting = true;
			socketConn = conn;
			connHandler(conn);
		} else {
			_ = conn.Close();
		}

	}

}

func main() {


	configBytes, _ := ioutil.ReadFile("config.json")

	getConfigErr := json.Unmarshal(configBytes, sysConfig)

	if getConfigErr != nil {
		log.Fatalf("parse config.json error")
		return
	}



	go ServerSocket();


	options := serial.OpenOptions{
		PortName:        sysConfig.SerialPortName,
		BaudRate:        sysConfig.SerialPortBaudRate,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 1,
	}

	var err error = nil

	// Open the port.

	port, err = serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	log.Printf("serial %s at %d opened",sysConfig.SerialPortName,sysConfig.SerialPortBaudRate)

	// Make sure to close it later.
	defer port.Close()

	for {
		buf := make([]byte, 4096)
		cnt, _ := port.Read(buf)
		//把串口收到数据从网络写出
		if socketConn != nil {
			socketConn.Write(buf[0:cnt])
		}
	}

}
