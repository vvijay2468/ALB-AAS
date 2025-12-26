# ALB-AAS
# Adaptive Load Balancer (Go)

A transparent, observable HTTP/HTTPS load balancer built **from scratch in Go** to study and evaluate routing strategies under **high-variance traffic**.

This project is explicitly **not** meant to replace production-grade tools like **NGINX** or **AWS ALB**.  
Its purpose is **learning, experimentation, and observability** — exposing routing decisions that are normally hidden.

---

## 🚀 Motivation

Modern load balancers abstract routing behind opaque heuristics.  
That abstraction hides critical system behavior such as:

- How latency degrades **before** failures occur
- How concurrency impacts backend selection
- Why certain backends overload earlier than others
- How adaptive strategies react under stress

### Goal

Build a **minimal but realistic** load balancer that:

- Makes routing decisions explicit
- Allows controlled experiments under load
- Surfaces real-time metrics for analysis

---

## 🧠 Key Features

### Core Load Balancing Strategies

- **Round Robin**
- **Least Connections**
- **Sticky Sessions** (IP-based)
- **Adaptive Strategy**
  - Backend score = `EWMA Latency × Active Connections`

---

### Observability

- Prometheus metrics exposed at `/metrics`
- Grafana-ready dashboards
- Per-backend **EWMA latency tracking**
- Request rate counters
- Latency histograms
- Error counters

---

### Reliability

- Active health checks
- Automatic backend exclusion on failure
- Backend recovery after health restoration
- Graceful degradation under overload

---

### Traffic Control

- Token-bucket rate limiting
- HTTP `429 Too Many Requests`
- Backpressure instead of uncontrolled crashes

---

### Security

- HTTPS termination (TLS)
- Automatic HTTP → HTTPS redirection

---



## ⚙️ Configuration

**`config/config.json`**

```json
{
  "port": "8080",
  "tls_port": "8443",
  "cert_file": "server.crt",
  "key_file": "server.key",
  "strategy": "adaptive",
  "backends": [
    "http://localhost:8001",
    "http://localhost:8002",
    "http://localhost:8003"
  ],
  "health_check_path": "/health",
  "health_check_interval": 5
}
```
Supported Strategies
1) round_robin
2) least_conn
3) sticky
4) EWMA adaptive

▶️ How to Run
```python
1️⃣ Start Backend Servers
bash
Copy code
python3 server.py 8001
python3 server.py 8002
python3 server.py 8003
Each backend responds with its port number.

2️⃣ Start Load Balancer

go run cmd/lb/main.go
Expected output:


HTTP server listening on :8080
HTTPS server listening on :8443

3️⃣ Test Manually
bash
Copy code
curl https://localhost:8443/ -k
Expected output (depends on strategy):


backend 8001
backend 8002
backend 8003

📊 Metrics & Observability
Prometheus
Start Prometheus:


prometheus --config.file=prometheus.yml
Prometheus UI:


http://localhost:9090
Target status:


load_balancer (UP)
Grafana
Add Prometheus data source:

arduino
Copy code
http://localhost:9090
Useful metrics:

lb_http_requests_total

lb_request_duration_seconds

lb_backends_alive

lb_backend_latency_ewma

🔥 Load Testing
Using hey:

bash
Copy code
hey -n 200000 -c 100 https://localhost:8443/ -k
Expected Behavior
Traffic distribution across backends is visible

Adaptive routing shifts traffic dynamically

Rate limiting triggers 429 under overload

Metrics update in real time
```

🧪 What This Project Demonstrates
Adaptive routing outperforms static strategies under variable load

Latency degradation occurs before backend failure

Backpressure is safer than uncontrolled overload

Observability is critical for routing correctness

⚠️ Non-Goals
This project intentionally does not include:

Replacement for NGINX / AWS ALB

Auto-scaling

Kubernetes integration

Clarity and experimentation are prioritized over production completeness.

🧩 Future Extensions (Optional)
Feedback-based autoscaling

Chaos testing (latency & fault injection)

Canary routing

Shadow traffic replay

Per-endpoint routing policies

---