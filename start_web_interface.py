#!/usr/bin/env python3
import http.server
import socketserver
import os
import webbrowser
from pathlib import Path
import tempfile
import urllib.parse
from email.mime.multipart import MIMEMultipart
from email.mime.base import MIMEBase
from email import encoders
import io

class UploadHTTPRequestHandler(http.server.SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=Path(__file__).parent / "web-upload", **kwargs)
    
    def do_POST(self):
        if self.path == '/upload':
            self.handle_file_upload()
        else:
            # Handle other POST requests with 404
            self.send_error(404, "Not Found")
    
    def handle_file_upload(self):
        try:
            # Get content length
            content_length = int(self.headers['Content-Length'])
            
            # Read the body
            post_data = self.rfile.read(content_length)
            
            # Parse multipart form data manually
            boundary = self.headers['Content-Type'].split('boundary=')[1].encode()
            parts = post_data.split(b'--' + boundary)
            
            uploaded_count = 0
            files = []
            
            for part in parts[1:-1]:  # Skip first and last empty parts
                # Split headers and body
                if b'\r\n\r\n' in part:
                    headers, body = part.split(b'\r\n\r\n', 1)
                    
                    # Parse headers to find Content-Disposition
                    headers_str = headers.decode('utf-8', errors='ignore')
                    
                    if 'filename=' in headers_str.lower():
                        # Extract filename from Content-Disposition header
                        start = headers_str.find('filename="')
                        if start != -1:
                            start += 10  # length of 'filename="'
                            end = headers_str.find('"', start)
                            if end != -1:
                                filename = headers_str[start:end]
                                
                                # Remove trailing \r\n from body
                                if body.endswith(b'\r\n'):
                                    body = body[:-2]
                                
                                files.append((filename, body))
                                uploaded_count += 1
            
            # Process each file
            for filename, file_content in files:
                # In a real implementation, you would upload to MinIO here
                # For now, we'll just save to local directory
                upload_dir = Path(__file__).parent / "uploads"
                upload_dir.mkdir(exist_ok=True)
                
                file_path = upload_dir / filename
                with open(file_path, 'wb') as f:
                    f.write(file_content)
        
            # Send success response
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            response = f'{{"success": true, "message": "Successfully uploaded {uploaded_count} file(s)"}}'
            self.wfile.write(response.encode())
            
        except Exception as e:
            print(f"Upload error: {str(e)}")
            self.send_error(500, f"Upload failed: {str(e)}")

def start_server():
    port = 8080
    
    # Create a custom handler that serves files from web-upload directory
    Handler = UploadHTTPRequestHandler
    
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