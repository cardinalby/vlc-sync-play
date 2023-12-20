package osutil

import "net"

const localHost = "127.0.0.1"

func GetFreePort() (host string, port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", localHost+":0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			port := l.Addr().(*net.TCPAddr).Port
			err = l.Close()
			return localHost, port, err
		}
	}
	return
}
