document.addEventListener('DOMContentLoaded', (event) => {
    let dropArea = document.getElementById('drop-area');
    let fileContent = document.getElementById('file-content');

    // Drag and drop
    dropArea.addEventListener('dragover', (e) => {
        e.preventDefault();
    });

    dropArea.addEventListener('drop', (e) => {
        e.preventDefault();
        let file = e.dataTransfer.files[0];
        let reader = new FileReader();
        reader.onload = function(e) {
            fileContent.style.display = 'block';
	    fileContent.style.left = "670px";
	    fileContent.style.top = "40px";
            fileContent.innerText = e.target.result;
        };
        reader.readAsText(file);
    });

    // Draggable div
    fileContent.onmousedown = function(event) {
        let shiftX = event.clientX - fileContent.getBoundingClientRect().left;
        let shiftY = event.clientY - fileContent.getBoundingClientRect().top;

        function moveAt(pageX, pageY) {
            fileContent.style.left = pageX - shiftX + 'px';
            fileContent.style.top = pageY - shiftY + 'px';
        }

        function onMouseMove(event) {
            moveAt(event.pageX, event.pageY);
        }

        document.addEventListener('mousemove', onMouseMove);

        fileContent.onmouseup = function() {
            document.removeEventListener('mousemove', onMouseMove);
            fileContent.onmouseup = null;
        };

    };

    fileContent.ondragstart = function() {
        return false;
    };
});
