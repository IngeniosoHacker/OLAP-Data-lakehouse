#!/usr/bin/env python3
import http.server
import socketserver
import os
import cgi
from pathlib import Path
import tempfile

class UploadHTTPRequestHandler(http.server.SimpleHTTPRequestHandler):
    def do_POST(self):
        if self.path == '/upload':
            self.handle_file_upload()
        else:
            self.send_error(404, "Not Found")
    
    def handle_file_upload(self):
        # Parse form data
        form = cgi.FieldStorage(
            fp=self.rfile,
            headers=self.headers,
            environ={'REQUEST_METHOD': 'POST',
                    'CONTENT_TYPE': self.headers['Content-Type']}
        )
        
        # Get files from form data
        uploaded_files = form.getlist('files')
        
        if not uploaded_files:
            self.send_error(400, "No files uploaded")
            return
        
        # Get MinIO configuration
        minio_endpoint = form.getvalue('minioEndpoint', '')
        minio_access_key = form.getvalue('minioAccessKey', '')
        minio_secret_key = form.getvalue('minioSecretKey', '')
        bucket_name = form.getvalue('bucketName', '')
        
        # Process each file
        for file_item in uploaded_files:
            filename = file_item.filename
            file_content = file_item.file.read()
            
            # Save file temporarily
            with tempfile.NamedTemporaryFile(delete=False) as temp_file:
                temp_file.write(file_content)
                temp_file_path = temp_file.name
            
            # In a real implementation, you would upload to MinIO here
            # For now, we'll just save to local directory
            upload_dir = Path("uploads")
            upload_dir.mkdir(exist_ok=True)
            
            file_path = upload_dir / filename
            with open(file_path, 'wb') as f:
                f.write(file_content)
        
        # Send success response
        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        self.wfile.write(b'{"success": true, "message": "Files uploaded successfully"}')

def start_server():
    # Change to the web-upload directory
    web_dir = Path(__file__).parent / "web-upload"
    os.chdir(web_dir)
    
    port = 8080
    
    # Create a custom handler that serves files from web-upload directory
    Handler = UploadHTTPRequestHandler
    
    with socketserver.TCPServer(("", port), Handler) as httpd:
        print(f"Web interface started at http://localhost:{port}/minimal_upload.html")
        print("Press Ctrl+C to stop the server")
        
        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            print("\nServer stopped.")

if __name__ == "__main__":
    start_server()