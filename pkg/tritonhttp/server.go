package tritonhttp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const TIMEOUT = 5 * time.Second
const PROTO = "HTTP/1.1"
const DATE = "Date"
const LASTMOD = "Last-Modified"
const CONTTYPE = "Content-Type"
const CONTLEN = "Content-Length"
const CONN = "Connection"

type Server struct {
	// Addr specifies the TCP address for the server to listen on,
	// in the form "host:port". It shall be passed to net.Listen()
	// during ListenAndServe().
	Addr string // e.g. ":0"

	// DocRoot specifies the path to the directory to serve static files from.
	DocRoot string
}

// ListenAndServe listens on the TCP network address s.Addr and then
// handles requests on incoming connections.
func (s *Server) ListenAndServe() error {
	//panic("todo")

	// Hint: call HandleConnection
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	fmt.Println("Listening on", ln.Addr())

	defer func() {
		err = ln.Close()
		if err != nil {
			fmt.Println("Error in closing listerner", err)
		}
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		fmt.Println("Accepting connection", conn.RemoteAddr())
		go s.HandleConnection(conn)
	}
}

// HandleConnection reads requests from the accepted conn and handles them.
func (s *Server) HandleConnection(conn net.Conn) {
	//panic("todo")
	res := &Response{}
	br := bufio.NewReader(conn)
	// Hint: use the other methods below

	for {
		// Set timeout
		if err := conn.SetReadDeadline(time.Now().Add(TIMEOUT)); err != nil {
			_ = conn.Close()
			log.Fatal(err)
		}
		// Try to read next request
		req, bytesReceived, err := ReadRequest(br)
		if err != nil {
			// Handle EOF
			if err == io.EOF {
				_ = conn.Close()
				return
			}
			// Handle timeout
			if err, ok := err.(net.Error); ok && err.Timeout() {
				if bytesReceived {
					fmt.Println("timeout, byte received, 400")
					res.HandleBadRequest()
					err := res.Write(conn)
					_ = conn.Close()
					if err != nil {
						log.Fatal(err)
					}
					return
				} else {
					_ = conn.Close()
					return
				}
			}
			fmt.Printf("get err %s\n", err)
			// Handle Bad request
			res.HandleBadRequest()
			err := res.Write(conn)
			_ = conn.Close()
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		// Handle good request
		res = s.HandleGoodRequest(req)
		err = res.Write(conn)
		if err != nil {
			log.Fatal(err)
		}
		// Close conn if requested
		if req.Close {
			_ = conn.Close()
			return
		}
	}
}

// HandleGoodRequest handles the valid req and generates the corresponding res.
func (s *Server) HandleGoodRequest(req *Request) (res *Response) {
	//panic("todo")
	// Hint: use the other methods below
	res = &Response{}
	url := filepath.Clean(req.URL)
	fp := s.DocRoot + url

	if _, err := os.Stat(fp); errors.Is(err, os.ErrNotExist) {
		res.HandleNotFound(req)
	} else if err == nil {
		res.HandleOK(req, fp)
	} else {
		log.Fatal(err)
	}

	return res
}

// HandleOK prepares res to be a 200 OK response
// ready to be written back to client.
func (res *Response) HandleOK(req *Request, path string) {
	//panic("todo")
	res.StatusCode = 200
	res.Proto = PROTO
	res.Header = make(map[string]string)
	res.Header[DATE] = FormatTime(time.Now())

	fileErrorHandler := func(res *Response, req *Request, path string) {
		res.Header[LASTMOD] = ""
		res.Header[CONTTYPE] = ""
		res.Header[CONTLEN] = ""
		if req.Close {
			res.Header[CONN] = "close"
		}
		res.Request = req
		res.FilePath = path
	}

	fh, err := os.Stat(path)
	if err != nil {
		fileErrorHandler(res, req, path)
		return
	}
	if fh.IsDir() {
		path = path + "index.html"
		fh, err = os.Stat(path)
		if err != nil {
			fileErrorHandler(res, req, path)
			return
		}
	}
	res.Header[LASTMOD] = FormatTime(fh.ModTime())
	res.Header[CONTTYPE] = MIMETypeByExtension(filepath.Ext(path))
	res.Header[CONTLEN] = strconv.Itoa(int(fh.Size()))

	if req.Close {
		res.Header[CONN] = "close"
	}
	res.Request = req
	res.FilePath = path
}

// HandleBadRequest prepares res to be a 400 Bad Request response
// ready to be written back to client.
func (res *Response) HandleBadRequest() {
	//panic("todo")
	res.StatusCode = 400
	res.Proto = PROTO
	res.Header = make(map[string]string)
	res.Header[DATE] = FormatTime(time.Now())
	res.Header[CONN] = "close"
}

// HandleNotFound prepares res to be a 404 Not Found response
// ready to be written back to client.
func (res *Response) HandleNotFound(req *Request) {
	//panic("todo")
	res.StatusCode = 404
	res.Proto = PROTO
	res.Header = map[string]string{DATE: FormatTime(time.Now())}
	if req.Close {
		res.Header[CONN] = "close"
	}
}
