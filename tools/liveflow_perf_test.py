#!/usr/bin/env python3
"""
Live Flow WebSocket Performance Test Script

This script tests the Live Flow WebSocket endpoint under load.
It simulates multiple concurrent connections and measures performance.

Usage:
    python3 liveflow_perf_test.py [--host HOST] [--port PORT] [--connections N] [--duration SECONDS]

Requirements:
    pip install websocket-client
"""

import argparse
import json
import time
import threading
import websocket
from datetime import datetime


class LiveFlowPerformanceTest:
    def __init__(self, host, port, token, num_connections, duration):
        self.host = host
        self.port = port
        self.token = token
        self.num_connections = num_connections
        self.duration = duration
        self.connections = []
        self.messages_received = 0
        self.errors = 0
        self.start_time = None
        self.end_time = None
        self.lock = threading.Lock()
        
    def on_message(self, ws, message):
        with self.lock:
            self.messages_received += 1
        
    def on_error(self, ws, error):
        with self.lock:
            self.errors += 1
        print(f"Error: {error}")
        
    def on_close(self, ws, close_status_code, close_msg):
        pass
        
    def on_open(self, ws):
        pass
        
    def create_connection(self):
        url = f"ws://{self.host}:{self.port}/api/live-flow?Authorization={self.token}"
        ws = websocket.WebSocketApp(
            url,
            on_message=self.on_message,
            on_error=self.on_error,
            on_close=self.on_close,
            on_open=self.on_open
        )
        return ws
        
    def run_test(self):
        print(f"Starting performance test with {self.num_connections} connections for {self.duration} seconds")
        self.start_time = time.time()
        
        # Create connections
        threads = []
        for i in range(self.num_connections):
            ws = self.create_connection()
            self.connections.append(ws)
            thread = threading.Thread(target=ws.run_forever)
            thread.daemon = True
            threads.append(thread)
            thread.start()
            
        # Wait for test duration
        time.sleep(self.duration)
        
        # Close connections
        for ws in self.connections:
            ws.close()
            
        self.end_time = time.time()
        
        # Calculate results
        duration = self.end_time - self.start_time
        messages_per_second = self.messages_received / duration if duration > 0 else 0
        
        print("\n=== Performance Test Results ===")
        print(f"Connections: {self.num_connections}")
        print(f"Duration: {duration:.2f} seconds")
        print(f"Messages received: {self.messages_received}")
        print(f"Messages per second: {messages_per_second:.2f}")
        print(f"Errors: {self.errors}")
        print(f"Average latency: {(duration / self.messages_received * 1000):.2f}ms" if self.messages_received > 0 else "N/A")


def main():
    parser = argparse.ArgumentParser(description="Live Flow WebSocket Performance Test")
    parser.add_argument("--host", default="127.0.0.1", help="Server host (default: 127.0.0.1)")
    parser.add_argument("--port", type=int, default=2017, help="Server port (default: 2017)")
    parser.add_argument("--token", required=True, help="JWT authentication token")
    parser.add_argument("--connections", type=int, default=10, help="Number of concurrent connections (default: 10)")
    parser.add_argument("--duration", type=int, default=30, help="Test duration in seconds (default: 30)")
    
    args = parser.parse_args()
    
    test = LiveFlowPerformanceTest(
        args.host,
        args.port,
        args.token,
        args.connections,
        args.duration
    )
    
    test.run_test()


if __name__ == "__main__":
    main()