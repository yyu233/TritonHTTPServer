package tritonhttp

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Response struct {
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.1"

	// Header stores all headers to write to the response.
	// Header keys are case-incensitive, and should be stored
	// in the canonical format in this map.
	Header map[string]string

	// Request is the valid request that leads to this response.
	// It could be nil for responses not resulting from a valid request.
	Request *Request

	// FilePath is the local path to the file to serve.
	// It could be "", which means there is no file to serve.
	FilePath string
}

// Write writes the res to the w.
func (res *Response) Write(w io.Writer) error {
	if err := res.WriteStatusLine(w); err != nil {
		return err
	}
	if err := res.WriteSortedHeaders(w); err != nil {
		return err
	}
	if res.StatusCode == 200 {
		if err := res.WriteBody(w); err != nil {
			return err
		}
	}

	return nil
}

// WriteStatusLine writes the status line of res to w, including the ending "\r\n".
// For example, it could write "HTTP/1.1 200 OK\r\n".
func (res *Response) WriteStatusLine(w io.Writer) error {
	//panic("todo")
	statusLine := ""
	if res.StatusCode == 200 {
		statusLine = fmt.Sprintf("%s %d %s\r\n", res.Proto, res.StatusCode, "OK")
	}
	if res.StatusCode == 400 {
		statusLine = fmt.Sprintf("%s %d %s\r\n", res.Proto, res.StatusCode, "Bad Request")
	}
	if res.StatusCode == 404 {
		statusLine = fmt.Sprintf("%s %d %s\r\n", res.Proto, res.StatusCode, "Not Found")
	}
	_, err := w.Write([]byte(statusLine))

	return err
}

// WriteSortedHeaders writes the headers of res to w, including the ending "\r\n".
// For example, it could write "Connection: close\r\nDate: foobar\r\n\r\n".
// For HTTP, there is no need to write headers in any particular order.
// TritonHTTP requires to write in sorted order for the ease of testing.
func (res *Response) WriteSortedHeaders(w io.Writer) error {
	//panic("todo")
	buf := make([]string, 0, len(res.Header))
	for k := range res.Header {
		buf = append(buf, k)
	}
	quickSort(buf, 0, len(buf)-1)
	for i := range buf {
		k := buf[i]
		header := fmt.Sprintf("%s: %s\r\n", k, res.Header[k])
		_, err := w.Write([]byte(header))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	return nil
}

// WriteBody writes res' file content as the response body to w.
// It doesn't write anything if there is no file to serve.
func (res *Response) WriteBody(w io.Writer) error {
	//panic("todo")
	if len(res.FilePath) != 0 {
		fp := res.FilePath
		dir, err := os.Stat(res.FilePath)
		if err != nil {
			return err
		}
		//request a directory, return index.html
		if dir.IsDir() {
			fp = fp + "index.html"
		}
		f, err := os.Open(fp)
		if err != nil {
			return err
		}
		defer f.Close()
		reader := bufio.NewReader(f)
		reader = bufio.NewReaderSize(reader, 5000)
		buf := make([]byte, 16)
		for {
			n, err := reader.Read(buf)
			if err != nil {
				if err != io.EOF {
					return err
				}
				break
			}
			_, err = w.Write(buf[0:n])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func partition(arr []string, low, high int) ([]string, int) {
	piv := arr[high]
	i := low
	for j := low; j < high; j++ {
		if strings.Compare(arr[j], piv) < 0 {
			arr[i], arr[j] = arr[j], arr[i]
			i++
		}
	}
	arr[i], arr[high] = arr[high], arr[i]

	return arr, i
}

func quickSort(arr []string, low, high int) []string {
	if low < high {
		arr, p := partition(arr, low, high)
		arr = quickSort(arr, low, p-1)
		arr = quickSort(arr, p+1, high)
	}

	return arr
}
