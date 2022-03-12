package tritonhttp

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

type Request struct {
	Method string // e.g. "GET"
	URL    string // e.g. "/path/to/a/file"
	Proto  string // e.g. "HTTP/1.1"

	// Header stores misc headers excluding "Host" and "Connection",
	// which are stored in special fields below.
	// Header keys are case-incensitive, and should be stored
	// in the canonical format in this map.
	Header map[string]string

	Host  string // determine from the "Host" header
	Close bool   // determine from the "Connection" header
}

// ReadRequest tries to read the next valid request from br.
//
// If it succeeds, it returns the valid request read. In this case,
// bytesReceived should be true, and err should be nil.
//
// If an error occurs during the reading, it returns the error,
// and a nil request. In this case, bytesReceived indicates whether or not
// some bytes are received before the error occurs. This is useful to determine
// the timeout with partial request received condition.
func ReadRequest(br *bufio.Reader) (req *Request, bytesReceived bool, err error) {
	//panic("todo")
	bytesReceived = true
	req = &Request{}

	line, err := ReadLine(br)
	if len(line) == 0 {
		bytesReceived = false
	}
	if err != nil {
		return nil, bytesReceived, err
	}
	// Read start line
	fields, err := parseInitialRequestLine(line)
	if err != nil {
		return nil, bytesReceived, err
	}
	if err := validateStartLine(fields); err != nil {
		return nil, bytesReceived, err
	}

	req.Method = fields[0]
	req.URL = fields[1]
	req.Proto = fields[2]
	fmt.Println(line)

	req.Header = make(map[string]string)
	// Read headers
	for {
		line, err := ReadLine(br)
		if len(line) == 0 {
			bytesReceived = false
		}
		if err != nil {
			return nil, bytesReceived, err
		}
		if line == "" {
			break
		}
		fields, err = parseHeaderRequestLine(line)
		if err != nil {
			return nil, bytesReceived, err
		}
		err = validateHeaderLine(fields)
		if err != nil {
			return nil, bytesReceived, err
		}
		key := CanonicalHeaderKey(fields[0])
		val := fields[1]
		// Check required headers
		if strings.Compare(key, "Host") == 0 {
			if len(val) == 0 {
				return nil, bytesReceived, fmt.Errorf("required header field %q not found", "Host")
			}
			req.Host = val
		} else if strings.Compare(key, "Connection") == 0 { // Handle special headers
			if strings.Compare(val, "close") == 0 {
				req.Close = true
			}
		} else {
			req.Header[key] = val
		}
		fmt.Println(line)
	}

	return req, true, nil
}

func parseInitialRequestLine(line string) ([]string, error) {
	fields := strings.Split(line, " ")

	if len(fields) != 3 {
		return nil, fmt.Errorf("could not parse the initial request line, got fields %q, len: %d", fields, len(fields))
	}

	return fields, nil
}

func parseHeaderRequestLine(line string) ([]string, error) {
	fields := strings.SplitN(line, ":", 2)

	if len(fields) != 2 {
		return nil, fmt.Errorf("could not parse the request line, got fields %q, len: %d", fields, len(fields))
	}
	fields[1] = strings.TrimSpace(fields[1])
	return fields, nil
}

func validateStartLine(line []string) error {
	//TODO
	//check method
	if strings.Compare(line[0], "GET") != 0 {
		return fmt.Errorf("invalid method %q", line[0])
	}

	//check protocal
	if strings.Compare(line[2], "HTTP/1.1") != 0 {
		return fmt.Errorf("invalid protocal %q", line[2])
	}

	return nil
}

func validateHeaderLine(line []string) error {
	//check key is alphanumeric or hyphen
	isAlpNumHyp := regexp.MustCompile(`^[a-zA-Z0-9-]*$`).MatchString
	if !isAlpNumHyp(line[0]) {
		return fmt.Errorf("header Key is not alphanumeric or hypen, got %q", line[0])
	}
	//check if value start with space or end with CRLF
	if strings.Compare(line[1], " ") == 0 || strings.Contains(line[1], "\r\n") {
		return fmt.Errorf("header value is invalid, got %q", line[1])
	}

	return nil
}
