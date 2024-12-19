function validateAccess() {
    const token = localStorage.getItem("accessToken");
    if (!token) {
        alert("Вы не авторизованы. Пожалуйста, войдите в систему.");
        window.location.href = "/login";
    }
}


function addPage(notebookId) {
    const pageTitle = prompt("Введите название страницы:");
    if (pageTitle) {
        const token = localStorage.getItem('accessToken');
        fetch(`/api/pages/?notebook_id=${notebookId}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({ title: pageTitle, content: "" })
        })
            .then(() => loadPages(notebookId))
            .catch(error => console.error('Error adding page:', error));
    }
}






async function checkAndRefreshToken() {
    const expiryTime = localStorage.getItem('accessTokenExpiry');
    const currentTime = Date.now();

    // Если время до истечения токена менее 2 минут, обновляем токен
    const timeLeft = expiryTime - currentTime;
    if (timeLeft <= 2 * 60 * 1000) {
        await refreshAccessToken();  // Обновляем токен
    }
}