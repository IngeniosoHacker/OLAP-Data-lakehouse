// MinIO upload functionality
function uploadFiles() {
    if (selectedFiles.length === 0) {
        showStatus('No files selected for upload', 'error');
        return;
    }

    const minioEndpoint = document.getElementById('minioEndpoint').value;
    const accessKey = document.getElementById('minioAccessKey').value;
    const secretKey = document.getElementById('minioSecretKey').value;
    const bucketName = document.getElementById('bucketName').value;

    if (!minioEndpoint || !accessKey || !secretKey || !bucketName) {
        showStatus('Please fill in all MinIO configuration fields', 'error');
        return;
    }

    // Show progress bar
    progressBar.style.display = 'block';
    statusMessage.style.display = 'none';
    
    // Process each file
    let completed = 0;
    const total = selectedFiles.length;

    selectedFiles.forEach((file, index) => {
        // Update file status to uploading
        updateFileStatus(index, 'Uploading...', 'uploading');
        
        // Create form data for file upload
        const formData = new FormData();
        formData.append('file', file);
        
        // Note: Direct browser upload to MinIO would require CORS configuration
        // In a real implementation, you would need a backend API to handle the upload
        // For demo purposes, we'll simulate the upload
        
        // Simulate upload progress
        simulateUpload(file, index, () => {
            completed++;
            const percent = Math.floor((completed / total) * 100);
            progressFill.style.width = percent + '%';
            
            updateFileStatus(index, 'Uploaded', 'success');
            
            if (completed === total) {
                showStatus('All files uploaded successfully!', 'success');
                // Reset the form
                selectedFiles = [];
                updateFileList();
            }
        });
    });
}

function simulateUpload(file, index, callback) {
    // Simulate upload delay
    setTimeout(() => {
        // In a real implementation, you would make an API call here
        // Example API call:
        /*
        fetch('/api/upload', {
            method: 'POST',
            body: formData
        })
        .then(response => response.json())
        .then(data => {
            console.log('Upload successful:', data);
            callback();
        })
        .catch(error => {
            console.error('Upload error:', error);
            updateFileStatus(index, 'Upload failed', 'error');
            showStatus('Error uploading file: ' + file.name, 'error');
        });
        */
        
        // For demo purposes
        callback();
    }, 2000); // Simulate 2 second upload time
}

function updateFileStatus(index, status, type) {
    const fileItems = document.querySelectorAll('.file-item');
    if (fileItems[index]) {
        const statusElement = fileItems[index].querySelector('.file-status');
        if (statusElement) {
            statusElement.textContent = status;
            statusElement.className = `file-status status-${type}`;
        }
    }
}

function showStatus(message, type) {
    statusMessage.textContent = message;
    statusMessage.className = `status-message status-${type}-message`;
    statusMessage.style.display = 'block';
    
    // Auto-hide success messages after 5 seconds
    if (type === 'success') {
        setTimeout(() => {
            statusMessage.style.display = 'none';
        }, 5000);
    }
}

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    // DOM elements
    const uploadArea = document.getElementById('uploadArea');
    const fileInput = document.getElementById('fileInput');
    const browseBtn = document.getElementById('browseBtn');
    const uploadBtn = document.getElementById('uploadBtn');
    const fileList = document.getElementById('fileList');
    const progressBar = document.getElementById('progressBar');
    const progressFill = document.getElementById('progressFill');
    const statusMessage = document.getElementById('statusMessage');
    
    let selectedFiles = [];

    // Event listeners
    browseBtn.addEventListener('click', () => {
        fileInput.click();
    });

    fileInput.addEventListener('change', handleFileSelect);

    uploadArea.addEventListener('dragover', (e) => {
        e.preventDefault();
        uploadArea.classList.add('dragover');
    });

    uploadArea.addEventListener('dragleave', () => {
        uploadArea.classList.remove('dragover');
    });

    uploadArea.addEventListener('drop', (e) => {
        e.preventDefault();
        uploadArea.classList.remove('dragover');
        handleDroppedFiles(e.dataTransfer.files);
    });

    uploadBtn.addEventListener('click', uploadFiles);

    function handleFileSelect(e) {
        handleDroppedFiles(e.target.files);
    }

    function handleDroppedFiles(files) {
        // Convert FileList to Array
        const filesArray = Array.from(files);
        
        // Add new files to the selectedFiles array
        selectedFiles = [...selectedFiles, ...filesArray];
        
        // Remove duplicate files (based on name and size)
        selectedFiles = selectedFiles.filter((file, index, self) =>
            index === self.findIndex(f => f.name === file.name && f.size === file.size)
        );
        
        // Update the file list display
        updateFileList();
    }

    function updateFileList() {
        fileList.innerHTML = '';
        
        if (selectedFiles.length === 0) {
            uploadBtn.disabled = true;
            return;
        }
        
        uploadBtn.disabled = false;
        
        selectedFiles.forEach((file, index) => {
            const fileItem = document.createElement('div');
            fileItem.className = 'file-item';
            
            // Determine file icon based on extension
            let icon = '📄'; // default
            const ext = file.name.split('.').pop().toLowerCase();
            if (['csv'].includes(ext)) icon = '📊';
            else if (['json'].includes(ext)) icon = ' {}