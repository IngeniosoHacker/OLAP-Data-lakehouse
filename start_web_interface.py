#!/usr/bin/env python3
import http.server
import socketserver
import os
import webbrowser
from pathlib import Path

def start_server():
    # Change to the web-upload directory
    web_dir = Path(__file__).parent / "web-upload"
    os.chdir(web_dir)
    
    port = 8000
    
    # Create a custom handler that serves files from web-upload directory
    Handler = http.server.SimpleHTTPRequestHandler
    
    with socketserver.TCPServer(("", port), Handler) as httpd:
        print(f"Web interface started at http://localhost:{port}/minimal_upload.html")
        print("Press Ctrl+C to stop the server")
        
        # Try to automatically open in browser
        try:
            webbrowser.open(f"http://localhost:{port}/minimal_upload.html")
        except:
            print("Could not automatically open browser. Please navigate to the URL above.")
        
        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            print("\nServer stopped.")

if __name__ == "__main__":
    start_server()