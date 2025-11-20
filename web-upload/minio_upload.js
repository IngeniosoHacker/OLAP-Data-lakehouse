// Complete JavaScript for MinIO file upload
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