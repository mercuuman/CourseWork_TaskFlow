function validateAccess() {
    const token = localStorage.getItem("accessToken");
    if (!token) {
        alert("Вы не авторизованы. Пожалуйста, войдите в систему.");
        window.location.href = "/login";
    }
}
