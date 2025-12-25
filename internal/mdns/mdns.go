// Copyright (c) 2025 JoeGlenn1213
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package mdns provides mDNS service discovery for LGH
package mdns

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/hashicorp/mdns"
)

const (
	// ServiceName is the mDNS service name
	ServiceName = "_lgh._tcp"
	// ServiceDomain is the default domain
	ServiceDomain = "local."
)

// Service represents an mDNS service for LGH
type Service struct {
	server   *mdns.Server
	port     int
	hostname string
}

// NewService creates a new mDNS service
func NewService(port int) (*Service, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "lgh-server"
	}

	// Clean hostname for mDNS
	hostname = strings.ReplaceAll(hostname, ".", "-")

	return &Service{
		port:     port,
		hostname: hostname,
	}, nil
}

// Start starts the mDNS service advertisement
func (s *Service) Start() error {
	// Get local IP address
	ip := getOutboundIP()
	if ip == nil {
		return fmt.Errorf("failed to determine local IP address")
	}

	// Create mDNS service
	info := []string{
		"LGH LocalGitHub Service",
		"version=1.0.0",
	}

	service, err := mdns.NewMDNSService(
		s.hostname,
		ServiceName,
		ServiceDomain,
		"",
		s.port,
		[]net.IP{ip},
		info,
	)
	if err != nil {
		return fmt.Errorf("failed to create mDNS service: %w", err)
	}

	// Create and start server
	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		return fmt.Errorf("failed to start mDNS server: %w", err)
	}

	s.server = server
	return nil
}

// Stop stops the mDNS service
func (s *Service) Stop() error {
	if s.server != nil {
		return s.server.Shutdown()
	}
	return nil
}

// GetServiceURL returns the mDNS URL for the service
func (s *Service) GetServiceURL() string {
	return fmt.Sprintf("http://%s.local:%d", s.hostname, s.port)
}

// DiscoveryResult represents a discovered LGH service
type DiscoveryResult struct {
	Name string
	Host string
	Port int
	IP   net.IP
	Info []string
	URL  string
}

// Discover performs mDNS discovery for LGH services
func Discover(_ int) ([]DiscoveryResult, error) {
	results := []DiscoveryResult{}

	entryCh := make(chan *mdns.ServiceEntry, 10)

	// Start discovery
	go func() {
		_ = mdns.Lookup(ServiceName, entryCh)
		close(entryCh)
	}()

	// Collect results
	for entry := range entryCh {
		result := DiscoveryResult{
			Name: entry.Name,
			Host: entry.Host,
			Port: entry.Port,
			IP:   entry.AddrV4,
			Info: entry.InfoFields,
		}

		if entry.AddrV4 != nil {
			result.URL = fmt.Sprintf("http://%s:%d", entry.AddrV4.String(), entry.Port)
		} else if entry.Host != "" {
			result.URL = fmt.Sprintf("http://%s:%d", entry.Host, entry.Port)
		}

		results = append(results, result)
	}

	return results, nil
}

// getOutboundIP gets the preferred outbound IP of this machine
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// Fallback: try to find a non-loopback IP
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return nil
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP
				}
			}
		}
		return nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}
