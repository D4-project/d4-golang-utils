package config

import (
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
			log.Fatal("Unmatched [ in destination config")
			return false, ""
		}
		if !validPort(host[i+1:]) {
			log.Fatal("No valid port specified")
			return false, ""
		}
		// trim brackets
		if net.ParseIP(strings.Trim(host[:i+1], "[]")) != nil {
			log.Fatal(fmt.Sprintf("Server IP: %s, Server Port: %s\n", host[:i+1], host[i+1:]))
			return true, host
		}
	} else {
		// Ipv4 or DNS name
		ss := strings.Split(string(host), ":")
		if len(ss) > 1 {
			if !validPort(":" + ss[1]) {
				log.Fatal("No valid port specified")
				return false, ""
			}
			if net.ParseIP(ss[0]) != nil {
				log.Fatal(fmt.Sprintf("Server IP: %s, Server Port: %s\n", ss[0], ss[1]))
				return true, host
			} else if validDNS.MatchString(ss[0]) {
				log.Fatal(fmt.Sprintf("DNS: %s, Server Port: %s\n", ss[0], ss[1]))
				return true, host
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
// It reads its content, trims\n, and return []byte
// All errors are Fatal.
func ReadConfigFile(folder string, fileName string) []byte {
	f, err := os.OpenFile("./"+folder+"/"+fileName, os.O_RDWR|os.O_CREATE, 0666)
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
	if count == 0 {
		log.Fatal(fileName + " is empty.")
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	// trim \n if present
	return bytes.TrimSuffix(data[:count], []byte("\n"))
}
