#!/usr/bin/env python3
import http.server
import socketserver
import os
import cgi
from pathlib import Path
import tempfile
import json
import traceback
import subprocess

class UploadHTTPRequestHandler(http.server.SimpleHTTPRequestHandler):
    server_version = "UploadHTTP/1.0"

    def _set_cors(self):
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "POST, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Content-Type")

    def do_OPTIONS(self):
        self.send_response(204, "No Content")
        self._set_cors()
        self.end_headers()

    def do_POST(self):
        if self.path == '/upload':
            self.handle_file_upload()
        else:
            self.send_error(404, "Not Found")
    
    def handle_file_upload(self):
        try:
            form = cgi.FieldStorage(
                fp=self.rfile,
                headers=self.headers,
                environ={
                    'REQUEST_METHOD': 'POST',
                    'CONTENT_TYPE': self.headers.get('Content-Type', '')
                }
            )

            uploaded_files = []
            if form.list:
                for part in form.list:
                    if isinstance(part, cgi.FieldStorage) and getattr(part, "filename", ""):
                        if part.name == "files":
                            uploaded_files.append(part)

            if not uploaded_files:
                self.send_error(400, "No files uploaded or missing form field 'files'")
                return

            base_dir = Path(__file__).parent
            upload_dir = base_dir / "uploads"
            upload_dir.mkdir(exist_ok=True)

            saved_files = []
            ingested = []
            for file_item in uploaded_files:
                filename = file_item.filename or "uploaded_file"
                file_content = file_item.file.read()

                # Save to uploads/
                file_path = upload_dir / filename
                with open(file_path, 'wb') as f:
                    f.write(file_content)
                saved_files.append(str(file_path))

                # Trigger ETL ingestion into star schema via docker compose + etl-go
                try:
                    cmd = [
                        "docker", "compose", "run", "--rm",
                        "-e", "ETL_SOURCE_TYPE=file",
                        "-e", f"ETL_SOURCE_FILE=/app/uploads/{filename}",
                        "-e", "ETL_MODE=star",
                        "-v", f"{upload_dir}:/app/uploads",
                        "etl-go"
                    ]
                    result = subprocess.run(
                        cmd,
                        cwd=base_dir,
                        capture_output=True,
                        text=True,
                        check=True
                    )
                    ingested.append({"file": filename, "status": "ingested", "output": result.stdout})
                except subprocess.CalledProcessError as etl_err:
                    ingested.append({
                        "file": filename,
                        "status": "failed",
                        "error": etl_err.stderr or str(etl_err)
                    })

            self.send_response(200)
            self._set_cors()
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            resp = {
                "success": True,
                "message": f"Saved {len(saved_files)} file(s)",
                "files": saved_files,
                "ingest": ingested
            }
            self.wfile.write(json.dumps(resp).encode("utf-8"))
        except Exception as e:
            print("Upload error:", e)
            traceback.print_exc()
            self.send_response(500)
            self._set_cors()
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            resp = {"success": False, "error": str(e)}
            self.wfile.write(json.dumps(resp).encode("utf-8"))

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
