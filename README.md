# Chaos Proxy

> *"Break production... before production breaks you"* - Some wise engineer, probably

![Disaster Girl Meme](https://upload.wikimedia.org/wikipedia/en/1/11/Disaster_Girl.jpg)

A delightfully devious reverse proxy that weaponizes Murphy's Law for your testing pleasure. Because why wait for production to break when you can break it yourself in style?

## Table of Contents

- [What Is This Madness?](#-what-is-this-madness)
- [Features](#-features)
- [Getting Started](#-getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [Configuration](#%EF%B8%8F-configuration)
  - [Basic Configuration](#basic-configuration)
  - [Chaos Configuration](#chaos-configuration)
  - [Example Configurations](#example-configurations)
- [Usage](#-usage)
  - [Running the Proxy](#running-the-proxy)
  - [Making Requests](#making-requests)
  - [Watching the Chaos](#watching-the-chaos)
- [FAQ](#-faq)
- [Known Issues & Limitations](#%EF%B8%8F-known-issues--limitations)
- [Development](#%EF%B8%8F-development)
  - [Running Tests](#running-tests)
  - [Adding New Chaos Strategies](#adding-new-chaos-strategies)
- [Configuration Reference](#-configuration-reference)
- [Contributing](#-contributing)
- [License](#-license)
- [Acknowledgments](#-acknowledgments)
- [Further Reading](#-further-reading)

## 🤔 What Is This Madness?

Chaos Proxy is a Go-based reverse proxy that deliberately sabotages your HTTP requests in creative and configurable ways. Because who doesn't love it when they're APIs break?

Think of it as a chaos engineering tool that sits between your application and its upstream services, randomly introducing failures, delays, and corrupted data.

## ⚡ Features

### Request Dropping
Drop requests into the void like they never existed. No response, no error, just... nothing. You know that feeling when someone leaves you on read for three days? That's what this does to your API calls. Perfect for seeing how your application handles getting ghosted.

### Error Injection
Return any HTTP error code you want. 500s, 503s, 418 (I'm a teapot)... the possibilities are endless.

### Latency Injection
Add artificial delays to requests because apparently the internet isn't slow enough already. In other words, PTCL simulator for the Pakistani devs. You can choose between:
- **Fixed latency**: Consistent slowness for predictable testing
- **Random latency**: Variable delays within a range for that authentic "why is this so slow sometimes?" experience

### Response Corruption
Chaos Proxy will randomly corrupt your responses using one of four strategies:

1. **Random Byte Corruption**: Flips some bits and corrupts random bytes in the response
2. **JSON Corruption**: Missing brackets, extra commas, colons replaced with equals signs.
3. **Truncation**: Cuts the response in half because who needs complete data anyway?
4. **Content-Length Mismatch**: Tell clients to expect 50 bytes, send 100.

### Hot Reload
Configuration changes are picked up automatically. Tweak your chaos parameters on the fly without restarting.

## 🚀 Getting Started

### Prerequisites

- Go 1.25.5 or later
- An upstream service to test against
- A willingness to break things

### Installation

```bash
# Clone the repository
git clone https://github.com/khizar-sudo/chaos-proxy.git
cd chaos-proxy

# Edit config.yaml with your upstream URL and chaos settings (see Configuration section below)

# Build the binary
go build -o chaos-proxy ./cmd/chaos-proxy

# Run it
./chaos-proxy
```

## ⚙️ Configuration

Chaos Proxy uses a YAML configuration file at the root called `config.yaml`.

### Basic Configuration

```yaml
listen: ":8080"                                    # Port to listen on
upstream: "https://jsonplaceholder.typicode.com"   # Where to proxy requests
```

### Chaos Configuration

All rate values are percentages (0-100).

```yaml
chaos:
  # Error Rate: Percentage of requests that will return an error
  error_rate: 10        # 10% of requests will fail
  error_code: 503       # HTTP status code to return (default: 500)
  
  # Drop Rate: Percentage of requests to completely ignore
  drop_rate: 5          # 5% of requests vanish into the ether
  
  # Latency: Add delays to requests
  latency: "200ms"      # Fixed delay of 200ms
  
  # OR use random latency (commented out by default)
  # latency_min: "0ms"
  # latency_max: "1000ms"  # Random delay between 0-1000ms
  
  # Corrupt Rate: Percentage of responses to corrupt
  corrupt_rate: 15      # 15% of responses will be corrupted
```

### Example Configurations

**Gentle Mode (for testing environments)**

```yaml
chaos:
  error_rate: 2
  error_code: 503
  drop_rate: 1
  latency: "100ms"
  corrupt_rate: 2
```

**Chaos Mode (when you're feeling dangerous)**

```yaml
chaos:
  error_rate: 30
  error_code: 500
  drop_rate: 10
  latency_min: "0ms"
  latency_max: "5s"
  corrupt_rate: 25
```

**Doomsday Mode (I dare you to use this in production)**

```yaml
chaos:
  error_rate: 80
  error_code: 500
  drop_rate: 50
  latency_min: "5s"
  latency_max: "30s"
  corrupt_rate: 90
```

## 🎮 Usage

### Running the Proxy

```bash
# If you installed with go install
chaos-proxy

# If you built from source
./chaos-proxy

# Make sure you have a config.yaml in the current directory
```

You should see output like:

```
==============================================================================
INFO starting server listen=:8080 upstream=<your_server_url>
Chaos configuration
- Error rate: 10%
- Error code: 503
- Drop rate: 5%
- Fixed latency: 200ms
- Corrupt rate: 15%
```

Beautiful! You can probably tell I don't have a single artistic bone in my body. Anyways...

### Making Requests

Point your application at the proxy instead of the upstream service:

```bash
# Direct request (boring)
curl <your_boring_server_url_here>

# Through chaos proxy (exciting!)
curl http://localhost:<listen_port>/your-route
```

### Watching the Chaos

The proxy logs what it's doing to each request, so you can see the chaos in action:

```
[2024-01-19 10:30:45] GET /posts/1 [200] 234ms

[CHAOS] Injecting latency: 500ms
[PROXY] GET /your-route

[CHAOS] Dropping request (no response)
[PROXY] GET /your-route

[CHAOS] Injecting error: 503
[PROXY] GET /your-route

[CHAOS] Corrupting response
[CHAOS] Strategy: JSON Corruption | 543 bytes -> 542 bytes
[PROXY] GET /your-route
```

## 💬 FAQ

**Q: Why would I want to intentionally break my service?**  
A: It's better to find issues yourself in a controlled environment than to have them show up in production at 3 AM.

**Q: Can I use this in production?**  
A: Technically yes, but please don't unless you hate your job... or your life.

**Q: What happens if I set all rates to 100%?**  
A: Your service becomes a very expensive random number generator. How fun!

**Q: Does this work with HTTPS?**  
A: Yes. The proxy handles HTTPS upstream services. It terminates TLS on the proxy side though.

**Q: Can I contribute my own chaos strategies?**  
A: Absolutely. The more creative ways to break things, the merrier.

**Q: My application is working perfectly with chaos enabled. What's wrong?**  
A: Congratulations, either your application is incredibly resilient, or the chaos rates are too low. Try the Doomsday configuration if you want a real challenge.

## ⚠️ Known Issues & Limitations

- The proxy doesn't support WebSockets (yet)
- Very large response bodies might cause memory issues during corruption
- If `latency_min > latency_max`, the proxy will panic. This is a feature, not a bug. Read the documentation, pls.

## 🛠️ Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Adding New Chaos Strategies

Want to add a new type of chaos? Here's the general approach:

1. Add new fields to `ChaosConfig` in `internal/chaos/types.go`
2. Update the decision logic in `internal/chaos/engine.go`
3. Implement the chaos behavior in `internal/middleware/middleware.go`
4. Add configuration options to the YAML schema
5. Write tests (unlike production outages, these are optional - but please write them)

## 📋 Configuration Reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `listen` | string | `:8080` | Port to listen on |
| `upstream` | string | *required* | Upstream service URL |
| `chaos.error_rate` | float | `0` | Percentage of requests to return errors (0-100) |
| `chaos.error_code` | int | `500` | HTTP status code for error responses |
| `chaos.drop_rate` | float | `0` | Percentage of requests to drop (0-100) |
| `chaos.latency` | string | `""` | Fixed latency (e.g., "200ms", "1s") |
| `chaos.latency_min` | string | `""` | Minimum random latency |
| `chaos.latency_max` | string | `""` | Maximum random latency |
| `chaos.corrupt_rate` | float | `0` | Percentage of responses to corrupt (0-100) |

## 🤝 Contributing

Found a bug? Want to add a new chaos strategy? Contributions are welcome.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/even-more-chaos`)
3. Commit your changes (`git commit -m 'feat: some chaos abc'`)
4. Push to the branch (`git push origin feature/even-more-chaos`)
5. Open a Pull Request

Please include tests for new features. We may inject chaos into your code, but we're not barbarians.

## 📜 License

This project is provided "as is" without warranty of any kind. Use at your own risk. Not responsible for any production outages, sleepless nights, or existential crises caused by this software.

## 🙏 Acknowledgments

- Inspired by [Chaos Monkey](https://netflix.github.io/chaosmonkey/) and other chaos engineering tools
- Built with [Go](https://golang.org/) because I wanted to
- Thanks to Murphy's Law for the original concept
- Shoutout to all the developers who will use this to make their applications more resilient

## 📚 Further Reading

- [Principles of Chaos Engineering](https://principlesofchaos.org/)
- [Why you should break things on purpose](https://www.youtube.com/watch?v=dQw4w9WgXcQ)
- [Chaos Engineering: Building Confidence in System Behavior Through Experiments](https://www.oreilly.com/library/view/chaos-engineering/9781491988459/)

