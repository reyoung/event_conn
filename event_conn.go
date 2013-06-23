package event_conn

import (
	"log"
	"net"
	"time"
)

type ErrorType uint8

const SEND_ERROR = ErrorType(0)
const READ_ERROR = ErrorType(1)
const READ_BUFFER_CAPACITY = 1440

type EventConnError struct {
	RawError error
	Type     ErrorType
}

type EventConn struct {
	Conn     net.Conn
	Quit     chan bool
	Recv     chan []byte
	Send     chan []byte
	Errors   chan EventConnError
	isClosed bool
}

func (ec *EventConn) init() {
	ec.Quit = make(chan bool, 1)
	ec.Recv = make(chan []byte)
	ec.Send = make(chan []byte, 10)
	ec.Errors = make(chan EventConnError, 10)
	ec.isClosed = false
	go ec.recv_loop()
	go ec.send_loop()
}

func (ec *EventConn) send_loop() {
	for {
		send_buf := <-ec.Send
		_, err := ec.Conn.Write(send_buf)
		if err != nil {
			ec.Errors <- EventConnError{err, SEND_ERROR}
		}
	}
}

func (ec *EventConn) recv_loop() {
	ec.Conn.SetReadDeadline(time.Now().Add(5e8))
	const capacity = READ_BUFFER_CAPACITY
	retv := make([]byte, 0)
	for !ec.isClosed {
		buffer := make([]byte, capacity)
		sz, err := ec.Conn.Read(buffer)
		if err != nil && !err.(*net.OpError).Timeout() {
			ec.Errors <- EventConnError{err, READ_ERROR}
			break
		} else {
			if sz == 0 {
				if len(retv) != 0 {
					ec.Recv <- retv
					retv = make([]byte, 0)
				}
			} else if sz == capacity {
				retv = append(retv, buffer...)
			} else {
				log.Print(sz, " ", capacity, " \n")
				if len(retv) != 0 {
					ec.Recv <- append(retv, buffer[0:sz]...)
				} else {
					ec.Recv <- buffer[0:sz]
				}
			}
		}
	}
}

func NewEventConn(conn net.Conn) *EventConn {
	retv := &EventConn{Conn: conn}
	retv.init()

	return retv
}

func (ec *EventConn) Close() (err error) {
	err = ec.Conn.Close()
	ec.Quit <- true
	ec.isClosed = true
	return
}
