# Go Reverse Proxy Server

A lightweight HTTP reverse proxy server built with Go and the Gin framework. This server provides flexible request forwarding capabilities with support for special header-based routing and configurable target servers.

## Features

- **Reverse Proxy**: Forward all incoming requests to a configurable target server
- **Special Forward Header**: Handle requests with custom "Forward" header for direct transport
- **Header Preservation**: Maintains original request headers when forwarding
- **Catchall Routing**: Accepts requests on any path and forwards them appropriately
- **Built with Gin**: Fast and lightweight HTTP framework

## Prerequisites

- Go 1.19 or higher
- Internet connection (for downloading dependencies)

## Installation & Setup

1. **Clone or download the project files**
2. **Navigate to the project directory**:
   ```bash
   cd /path/to/your/project
   ```

3. **Install dependencies**:
   ```bash
   go mod tidy
   ```

## Configuration

Before running the server, you need to configure the target server URL:

1. Open `main.go`
2. Find the `Reverse` function
3. Update the target URL in this line:
   ```go
   remote, _ := url.Parse("http://xxx.xxx.xxx")
   ```
   
   Replace `"http://xxx.xxx.xxx"` with your actual target server URL, for example:
   ```go
   remote, _ := url.Parse("http://api.example.com")
   // or
   remote, _ := url.Parse("https://backend.myapp.com:8080")
   ```

## Usage

### Starting the Server

Run the server with:
```bash
go run main.go
```

The server will start on port **8888** and display:
```
[GIN-debug] Listening and serving HTTP on :8888
```

### Making Requests

#### Regular Proxy Requests
Send any HTTP request to `http://localhost:8888` and it will be forwarded to your configured target server:

```bash
# GET request
curl http://localhost:8888/api/users

# POST request with data
curl -X POST http://localhost:8888/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com"}'

# Any HTTP method works
curl -X DELETE http://localhost:8888/api/users/123
```

#### Special Forward Header Requests
Include a `Forward: ok` header to use direct HTTP transport instead of the reverse proxy:

```bash
curl -H "Forward: ok" http://localhost:8888/some-endpoint
```

When this header is present with value "ok", the request bypasses the reverse proxy and uses Go's default HTTP transport directly.

## How It Works

### Request Flow

1. **Incoming Request** → Server receives request on port 8888
2. **Forward Middleware Check** → Checks for "Forward: ok" header
   - If present: Uses direct HTTP transport
   - If absent: Continues to reverse proxy
3. **Reverse Proxy** → Forwards request to configured target server
4. **Response** → Returns target server's response to client

### Components

- **ForwardMid**: Middleware that handles special "Forward" header requests
- **Reverse**: Main reverse proxy handler that forwards requests to the target server
- **copyHeader**: Utility function that preserves HTTP headers during forwarding

## Customization

### Changing the Port
To run on a different port, modify the `Run` call in `main.go`:
```go
if err := r.Run(":3000"); err != nil {  // Change to port 3000
    panic(err)
}
```

### Modifying Forward Header Logic
Update the `ForwardMid` function to change the special header behavior:
```go
func ForwardMid(c *gin.Context) {
    // Change header name and value as needed
    if v, ok := c.Request.Header["Custom-Forward"]; ok {
        if v[0] == "enabled" {
            // Your custom logic here
        }
    }
    c.Next()
}
```

### Adding Authentication
You can add authentication middleware:
```go
func main() {
    r := gin.Default()
    
    // Add authentication middleware
    r.Use(AuthMiddleware)
    r.Use(ForwardMid)
    
    r.Any("/*proxyPath", Reverse)
    
    if err := r.Run(":8888"); err != nil {
        panic(err)
    }
}
```

## Development

### Building the Application
```bash
go build -o proxy-server main.go
./proxy-server
```

### Running Tests
```bash
go test ./...
```

## Troubleshooting

### Common Issues

1. **"connection refused" errors**: Ensure your target server URL is correct and the target server is running

2. **"bind: address already in use"**: Port 8888 is already in use. Either:
   - Stop the process using port 8888
   - Change the port in `main.go`

3. **Import errors**: Run `go mod tidy` to ensure all dependencies are installed

### Logs
The server uses Gin's default logging. You'll see:
- Request logs with status codes, response times, and paths
- Error messages for failed requests

## Dependencies

- **Gin** (v1.11.0): HTTP web framework
- **Go Standard Library**: HTTP utilities and networking

## License

This project is provided as-is for educational and development purposes.

## Contributing

Feel free to submit issues and enhancement requests!