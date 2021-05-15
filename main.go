package main

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/fatih/color"
)

var (
	bytesMode bool
	listen    string
	target    string
)

func main() {
	flag.BoolVar(&bytesMode, "bytes", false, "Run logger in bytes mode")
	flag.StringVar(&listen, "listen", "", "listen bind")
	flag.StringVar(&target, "target", "", "destination")
	flag.Parse()

	// check
	for _, ok := range []bool{
		len(listen) > 0,
		len(target) > 0,
	} {
		if !ok {
			flag.PrintDefaults()
			return
		}
	}

	listener, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	logger := log.New(color.Output, fmt.Sprintf("%s ", conn.RemoteAddr().String()), log.LstdFlags)
	target, err := net.Dial("tcp", target)
	if err != nil {
		log.Print(err)
		return
	}
	defer target.Close()

	// colorize
	colorIncoming := color.New(color.BgGreen)
	colorOutgoing := color.New(color.BgRed)
	// mode
	var scanner func(logger *log.Logger, r io.Reader, w io.Writer, clr *color.Color)
	if bytesMode {
		scanner = ScanBytesLog
	} else {
		scanner = ScanTextLog
	}
	// main
	go scanner(logger, conn, target, colorIncoming)
	scanner(logger, target, conn, colorOutgoing)
}

func AcceptOnce(addr string) (net.Conn, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer listener.Close()
	conn, err := listener.Accept()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func ScanTextLog(logger *log.Logger, r io.Reader, w io.Writer, clr *color.Color) {
	reader := bufio.NewReader(r)
	for {
		raw, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}
		log.Print(clr.Sprint(strings.TrimSpace(string(raw))))
		w.Write(raw)
	}
}

func ScanBytesLog(logger *log.Logger, r io.Reader, w io.Writer, clr *color.Color) {
	buf := make([]byte, 128)
	for {
		n, err := r.Read(buf)
		if err != nil {
			return
		}
		logger.Print(clr.Sprint(hex.EncodeToString(buf[:n])))
		w.Write(buf[:n])
	}
}
