package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(ForwardMid)

	// Create a catchall route for all methods including CONNECT
	r.Any("/*proxyPath", func(c *gin.Context) {
		if c.Request.Method == "CONNECT" {
			HandleConnect(c)
		} else {
			Reverse(c)
		}
	})

	log.Printf("Starting reverse proxy server on :8080")
	log.Printf("Target server: https://bdtunnel.com:2096/sub/imzami?format=json")

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}

func HandleConnect(c *gin.Context) {
	// Extract the target host from the request
	targetHost := c.Request.Host
	if targetHost == "" {
		targetHost = c.Request.URL.Host
	}

	// If no host is specified, return 400
	if targetHost == "" {
		c.String(http.StatusBadRequest, "Bad Request: No target host specified")
		return
	}

	log.Printf("CONNECT request to: %s", targetHost)

	// For CONNECT requests, we need to establish a tunnel
	// However, for a simple reverse proxy, we might want to reject CONNECT requests
	// or handle them differently based on your use case

	// Option 1: Reject CONNECT requests (recommended for simple HTTP proxy)
	c.String(http.StatusMethodNotAllowed, "CONNECT method not allowed")

	// Option 2: If you want to support HTTPS tunneling, uncomment below:
	/*
		destConn, err := net.DialTimeout("tcp", targetHost, 10*time.Second)
		if err != nil {
			c.String(http.StatusServiceUnavailable, "Failed to connect to target: %v", err)
			return
		}
		defer destConn.Close()

		c.Writer.WriteHeader(http.StatusOK)
		hijacker, ok := c.Writer.(http.Hijacker)
		if !ok {
			c.String(http.StatusInternalServerError, "Hijacking not supported")
			return
		}

		clientConn, _, err := hijacker.Hijack()
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to hijack connection: %v", err)
			return
		}
		defer clientConn.Close()

		// Start copying data between client and destination
		go io.Copy(destConn, clientConn)
		io.Copy(clientConn, destConn)
	*/
}

func ForwardMid(c *gin.Context) {
	// !!! adapt to your request header set
	if v, ok := c.Request.Header["Forward"]; ok {
		if v[0] == "ok" {
			resp, err := http.DefaultTransport.RoundTrip(c.Request)
			if err != nil {
				http.Error(c.Writer, err.Error(), http.StatusServiceUnavailable)
				c.Abort()
				return
			}
			defer resp.Body.Close()
			copyHeader(c.Writer.Header(), resp.Header)
			c.Writer.WriteHeader(resp.StatusCode)
			io.Copy(c.Writer, resp.Body)
			c.Abort()
			return
		}
	}

	c.Next()
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func Reverse(c *gin.Context) {
	// Check if target server is reachable
	targetURL := "https://bdtunnel.com:2096"
	remote, err := url.Parse(targetURL)
	if err != nil {
		log.Printf("Invalid target URL: %v", err)
		c.String(http.StatusInternalServerError, "Invalid proxy configuration")
		return
	}

	// Test connectivity to target server (HTTPS on port 2096)
	conn, err := net.DialTimeout("tcp", remote.Host, 5*time.Second)
	if err != nil {
		log.Printf("Target server unreachable: %v", err)
		c.String(http.StatusBadGateway, "Target server is currently unavailable")
		return
	}
	conn.Close()

	proxy := httputil.NewSingleHostReverseProxy(remote)

	// Improve error handling
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error: %v", err)
		if strings.Contains(err.Error(), "connection refused") {
			http.Error(w, "Target server connection refused", http.StatusBadGateway)
		} else if strings.Contains(err.Error(), "timeout") {
			http.Error(w, "Target server timeout", http.StatusGatewayTimeout)
		} else {
			http.Error(w, "Proxy error: "+err.Error(), http.StatusBadGateway)
		}
	}

	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Host = remote.Host
		req.URL.Scheme = remote.Scheme

		// For this specific API endpoint, we want to preserve the path or redirect to the specific endpoint
		if c.Request.URL.Path == "/" || c.Request.URL.Path == "" {
			// If requesting root, redirect to the specific API endpoint
			req.URL.Path = "/sub/imzami"
			req.URL.RawQuery = "format=json"
		} else {
			// Otherwise preserve the original path and query
			req.URL.Path = c.Request.URL.Path
			req.URL.RawQuery = c.Request.URL.RawQuery
		}

		// Add forwarded headers for transparency
		req.Header.Set("X-Forwarded-For", c.ClientIP())
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Forwarded-Host", c.Request.Host)
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}
