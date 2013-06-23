package event_conn

import (
	"log"
	"net"
	"testing"
	"time"
)

func TestInitRecvEventConn(*testing.T) {
	conn, err := net.Dial("tcp", "pop3.126.com:110")
	if err != nil {
		log.Fatal(err)
	}
	cli := NewEventConn(conn)
	go func() {
		buf := <-cli.Recv
		log.Print(string(buf))
		log.Print("outof goroutine")
		cli.Close()
	}()
	time.Sleep(1e9)
	log.Print("Sleep Complete.")
}

func TestSendRecv(*testing.T) {
	conn, err := net.Dial("tcp", "www.baidu.com:80")
	if err != nil {
		log.Fatal(err)
	}
	send_str := "GET / HTTP/1.1\r\nHost:www.baidu.com\r\n\r\n"
	cli := NewEventConn(conn)
	cli.Send <- []byte(send_str)
	go func() {
		buf := <-cli.Recv
		buf2 := <-cli.Recv
		log.Println(string(buf))
		log.Println(string(buf2))
		log.Print("outof goroutine")
		cli.Close()
	}()
	time.Sleep(1e9)
	log.Print("Sleep Complete.")
}
