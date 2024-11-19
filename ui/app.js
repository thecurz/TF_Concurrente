document.addEventListener('DOMContentLoaded', () => {
    // Crear la conexión WebSocket al servidor Go
    const socket = new WebSocket('ws://localhost:8080/ws');
    const categoryInput = document.getElementById('category');
    const sendButton = document.getElementById('send');
    const recommendationList = document.getElementById('recommendation-list');

    // Mensaje que indica que la conexión WebSocket está establecida
    socket.onopen = () => {
        console.log('WebSocket connection established.');
    };

    // Manejar mensajes recibidos desde el servidor
    socket.onmessage = (event) => {
        console.log('Message received:', event.data);
        const data = JSON.parse(event.data);
        recommendationList.innerHTML = '';
        data.recommendations.forEach((recommendation) => {
            const li = document.createElement('li');
            li.textContent = recommendation;
            recommendationList.appendChild(li);
        });
    };

    // Manejar errores de WebSocket
    socket.onerror = (error) => {
        console.error('WebSocket error:', error);
    };

    // Enviar la categoría ingresada al servidor
    sendButton.addEventListener('click', () => {
        const category = categoryInput.value.trim();
        if (category) {
            socket.send(JSON.stringify({ category }));
        } else {
            alert('Please enter a category.');
        }
    });
});