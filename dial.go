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

func Dial(network string, address string) (*EventConn, error) {
	return bindDial(func() (net.Conn, error) {
		return net.Dial(network, address)
	})
}

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
