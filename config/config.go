package config

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

// IsNet checks whether the string host is a valid
// IPv6 or IPv4 address
func IsNet(host string) (bool, string) {
	// DNS regex
	validDNS := regexp.MustCompile(`^(([a-zA-Z]{1})|([a-zA-Z]{1}[a-zA-Z]{1})|([a-zA-Z]{1}[0-9]{1})|([0-9]{1}[a-zA-Z]{1})|([a-zA-Z0-9][a-zA-Z0-9-_]{1,61}[a-zA-Z0-9]))\.([a-zA-Z]{2,6}|[a-zA-Z0-9-]{2,30}\.[a-zA-Z
 ]{2,3})$`)
	// Check ipv6
	if strings.HasPrefix(host, "[") {
		// Parse an IP-Literal in RFC 3986 and RFC 6874.
		// E.g., "[fe80::1]:80".
		i := strings.LastIndex(host, "]")
		if i < 0 {
			log.Println("Unmatched [ in destination config")
			return false, ""
		}
		if !validPort(host[i+1:]) {
			log.Println("No valid port specified")
			return false, ""
		}
		// trim brackets
		if net.ParseIP(strings.Trim(host[:i+1], "[]")) != nil {
			return true, host
		}
	} else {
		// Ipv4 or DNS name
		ss := strings.Split(host, ":")
		if len(ss) > 1 {
			if !validPort(":" + ss[1]) {
				log.Println("No valid port specified")
				return false, ""
			}
			// if not nil, its a valid IP adress
			if net.ParseIP(ss[0]) != nil {
				return true, host
			}
			// if "localhost", its valid
			if strings.Compare("localhost", ss[0]) == 0 {
				return true, host
			}
			// check against the regex
			if validDNS.MatchString(ss[0]) {
				return true, host
			} else {
				log.Println(fmt.Sprintf("DNS/IP: %s, Server Port: %s", ss[0], ss[1]))
				return false, ""
			}
		}
	}
	return false, host
}

// validPort reports whether port is either an empty string
// or matches /^:\d*$/
func validPort(port string) bool {
	if port == "" {
		return false
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

// ReadConfigFile takes two argument: folder and fileName.
// Create if not exist
// It reads its content, trims\n and \r, and return []byte
// All errors are Fatal.
func ReadConfigFile(folder string, fileName string) []byte {
	f, err := os.OpenFile(folder+"/"+fileName, os.O_RDONLY|os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}
	data := make([]byte, 100)
	count, err := f.Read(data)
	if err != nil {
		if err != io.EOF {
			log.Fatal(err)
		}
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

	// trim \r and \n if present
	r := bytes.TrimSuffix(data[:count], []byte("\n"))
	return bytes.TrimSuffix(r, []byte("\r"))
}

// ReadConfigFileLines takes two argument: folder and fileName.
// Create if not exist
// It reads its content line by line,
// and return [][]byte
// All errors are Fatal.
func ReadConfigFileLines(folder string, fileName string) [][]byte {
	res := [][]byte{}
	f, err := os.OpenFile(folder+"/"+fileName, os.O_RDONLY|os.O_CREATE, 0666)
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        res = append(res, []byte(scanner.Text()))
    }

    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }
    return res
}
