package utils

import (
	"io"
	"log"
	"net"
)

type DebugConn struct {
	io.ReadWriteCloser

	readTotal  int
	writeTotal int
	conn       net.Conn
	tag        string
}

func NewDebugConn(c net.Conn, tag string) *DebugConn {
	var d DebugConn
	d.conn = c
	d.tag = tag
	return &d
}

func (d *DebugConn) Read(p []byte) (int, error) {
	n, err := d.conn.Read(p)
	d.readTotal += n
	d.Debug()
	return n, err
}

func (d *DebugConn) Write(p []byte) (int, error) {
	n, err := d.conn.Write(p)
	d.writeTotal += n
	d.Debug()
	return n, err
}

func (d *DebugConn) Close() error {
	return d.conn.Close()
}

func (d *DebugConn) Debug() {
	log.Printf("{Socket] \"%v\" read: %v, write: %v", d.tag, d.readTotal, d.writeTotal)
}
