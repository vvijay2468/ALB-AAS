from http.server import BaseHTTPRequestHandler, HTTPServer
import time
import sys

PORT = int(sys.argv[1])
DELAY = float(sys.argv[2]) if len(sys.argv) > 2 else 0

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        time.sleep(DELAY)
        self.send_response(200)
        self.end_headers()
        self.wfile.write(f"backend {PORT}\n".encode())

HTTPServer(("", PORT), Handler).serve_forever()
