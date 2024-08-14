# HTTPSink

HTTPSink is a lightweight, in-memory HTTP request logger designed for simplicity and ease of use. Unlike many request bins that require databases, Elasticsearch, Redis, or other complex dependencies, HTTPSink aims to be as simple as possible while still providing useful functionality for debugging HTTP requests.

## Features

- Log HTTP requests to /bin/* endpoints
- View logged requests in real-time via WebSocket
- Simple web interface to view and clear logs
- In-memory storage for quick setup and tear down
- Configurable host, port, and maximum number of stored requests

## Purpose

HTTPSink serves a single purpose: to log HTTP requests for debugging. It's perfect for:

- Inspecting browser requests
- Debugging incoming webhooks
- Testing API integrations
- Quick and dirty request analysis

## Screenshot

![HTTPSink Screenshot](https://screenshots-damon.s3.amazonaws.com/screenshot-240814-66b9a9.png)

## Installation

1. Ensure you have Go installed on your system.
2. Clone this repository:
   ```
   git clone https://github.com/yourusername/httpsink.git
   ```
3. Navigate to the project directory:
   ```
   cd httpsink
   ```
4. Build the project:
   ```
   go build
   ```

## Usage

Run HTTPSink with default settings:

```
./httpsink
```

This will start the server on localhost:8000.

### Command-line Options

- `-host`: Host to listen on (default: "localhost")
- `-port`: Port to listen on (default: 8000)
- `-max`: Maximum number of requests to keep in buffer (default: 10)

Example with custom settings:

```
./httpsink -host 0.0.0.0 -port 9000 -max 50
```

## How It Works

1. HTTPSink listens for requests on the `/bin/*` endpoints.
2. When a request is received, it logs the request details in memory.
3. The web interface at `/logs` displays the logged requests.
4. Real-time updates are pushed to the web interface via WebSocket.
5. Logs can be cleared using the "Clear Logs" button on the web interface.

## SSL/TLS Support

While HTTPSink has basic SSL support, in typical use cases, TLS termination is handled by a reverse proxy (such as Traefik running in a Kubernetes cluster) sitting in front of HTTPSink. This approach offloads the SSL/TLS processing and allows for more flexible and robust security configurations.

## Limitations and Security Considerations

HTTPSink is designed for simplicity and ease of use, not for production environments. Be aware of the following limitations:

- All data is stored in memory and will be lost when the server is stopped.
- There's no authentication or authorization mechanism.
- The server may be vulnerable to DOS attacks if flooded with requests.
- The code has not undergone rigorous security auditing.

**Use HTTPSink for development and debugging purposes only. Do not use it to handle sensitive data or in production environments.**

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is open source and available under the [MIT License](LICENSE).

