/*
This EventConn would make net.Conn asynchronise. When you create a new net.Conn,
and pass it into event_conn.NewEventConn(), then you will get a *EventConn Object,
which all things can be done by using chan.

For example, when you want to recieve some message from a server, you can just

		conn, _ := net.Dial("tcp", "pop3.126.com:110")
		cli := NewEventConn(conn)
		msg := <- cli.Recv

You can also using cli.Recv in a goroutine, so the Recv will not blocked, And
when you want to send some message, you can just push some []byte into cli.Send.

There are some helper method in event_conn.DialXXX. It's almost just a wrapper for
net.DialXXX.
*/
package event_conn

import (
	"net"
	"time"
)

// Error's Type used by EventConnError
type ErrorType uint8

// Send Error Type
const SEND_ERROR = ErrorType(0)

// Read Error Type
const READ_ERROR = ErrorType(1)

// The Read Buffer Capacity, default set to Tcp's MSS
const READ_BUFFER_CAPACITY = 1440

// The Error Type Passing By EventConn.Errors chan
type EventConnError struct {
	RawError error     // The Error caused by net.Conn
	Type     ErrorType // The Error happend in which phase. (SEND_ERROR or READ_ERROR)
}

type EventConn struct {
	Conn     net.Conn            // The wrapped net.Conn
	Quit     chan bool           // When net.Conn is Closed, a true will be sended to Quit
	Recv     chan []byte         // Recieve channel
	Send     chan []byte         // Send channel, message send to this channel may be asynchronized.
	Errors   chan EventConnError // The happend errors
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

// The Send Loop, just send by EventConn.Conn
func (ec *EventConn) send_loop() {
	for {
		send_buf := <-ec.Send
		_, err := ec.Conn.Write(send_buf)
		if err != nil {
			ec.Errors <- EventConnError{err, SEND_ERROR}
		}
	}
}

// The Default Recieve Loop, it will merge messages which are sended at same time.
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
				if len(retv) != 0 {
					ec.Recv <- append(retv, buffer[0:sz]...)
				} else {
					ec.Recv <- buffer[0:sz]
				}
			}
		}
	}
}

/*
New EventConn. In this method, a recieve loop and a send loop will be started by goroutine

Note: You'd better use event_conn.DialXXX to create new EventConn. In that helper method,
TCP and UDP's MSS will be setted.
*/
func NewEventConn(conn net.Conn) *EventConn {
	retv := &EventConn{Conn: conn}
	retv.init()

	return retv
}

/*
Method for close EventConn
*/
func (ec *EventConn) Close() (err error) {
	err = ec.Conn.Close()
	ec.Quit <- true
	ec.isClosed = true
	return
}

/*
Custom Check Message Complete while recieve.
Note: This method is blocked, if you need an asynchronize version, try RecieveChanUntil
*/
func (ec *EventConn) RecieveUntil(checker func([]byte) bool) []byte {
	msg := make([]byte, 0)
	for !checker(msg) {
		tmp := <-ec.Recv
		msg = append(msg, tmp...)
	}
	return msg
}

/*
Custom Check Message Complete while recieve.
*/
func (ec *EventConn) RecieveChanUntil(checker func([]byte) bool) chan []byte {
	retv := make(chan []byte)
	go func() {
		msg := ec.RecieveUntil(checker)
		retv <- msg
	}()
	return retv
}
