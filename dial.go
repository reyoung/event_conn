package event_conn

import (
	"net"
)

func bindDial(fun func() (net.Conn, error)) (ec *EventConn, err error) {
	conn, err := fun()
	if err != nil {
		return
	} else {
		ec = NewEventConn(conn)
		return
	}
}

/*
Helper method for Dial EventConn
*/
func Dial(network string, address string) (*EventConn, error) {
	return bindDial(func() (net.Conn, error) {
		return net.Dial(network, address)
	})
}

/*
Helper method for DialTCP EventConn

Note: In this method, TCP MSS will be setted as READ_BUFFER_CAPACITY
*/
func DialTCP(network string, laddr *net.TCPAddr, raddr *net.TCPAddr) (*EventConn, error) {
	return bindDial(func() (conn net.Conn, err error) {
		tcp_conn, err := net.DialTCP(network, laddr, raddr)
		if err != nil {
			return
		}
		err = tcp_conn.SetReadBuffer(READ_BUFFER_CAPACITY)
		if err != nil {
			return
		}
		return tcp_conn, err
	})
}

/*
Helper method for DialUDP EventConn

Note: In this method, UDP ReadBuffer will be setted as READ_BUFFER_CAPACITY
*/
func DialUDP(network string, laddr, raddr *net.UDPAddr) (*EventConn, error) {
	return bindDial(func() (conn net.Conn, err error) {
		udp_conn, err := net.DialUDP(network, laddr, raddr)
		if err != nil {
			return
		}
		err = udp_conn.SetReadBuffer(READ_BUFFER_CAPACITY)
		if err != nil {
			return
		}
		return udp_conn, err
	})
}
