document.addEventListener("DOMContentLoaded", () => {
    const form = document.getElementById("feedback-form");
    const modal = document.getElementById("modal");
    const modalMessage = document.getElementById("modal-message");
    const modalImage = document.getElementById("modal-image"); // Добавлено: элемент для изображения
    const closeModal = document.getElementById("close-modal");

    // Функция для отображения модального окна
    const showModal = (message, success = true) => {
        modalMessage.textContent = message;
        modalMessage.style.color = success ? "green" : "red";

        // Устанавливаем изображение в зависимости от успеха или ошибки
        modalImage.src = success ? "./system_images/thanks.jpg" : "./system_images/I-REFUSE.jpg";

        modal.style.display = "block";
    };

    // Скрытие модального окна при клике на "X"
    closeModal.addEventListener("click", () => {
        modal.style.display = "none";
        window.location.href = "/"; // Переадресация на главную страницу
    });

    // Закрытие модального окна при клике вне его
    window.addEventListener("click", (e) => {
        if (e.target === modal) {
            modal.style.display = "none";
            window.location.href = "/"; // Переадресация на главную страницу
        }
    });

    // Обработчик отправки формы
    form.addEventListener("submit", async (e) => {
        e.preventDefault(); // Предотвращаем стандартную отправку формы

        const formData = new FormData(form);
        try {
            const response = await fetch(form.action, {
                method: form.method,
                body: formData,
            });

            const data = await response.json(); // Получаем JSON от хендлера

            // Отображаем модальное окно
            if (data.success === "true") {
                showModal(data.message || "Thank you for your feedback!", true);
                form.reset(); // Сбрасываем форму
            } else {
                showModal(data.message || "Failed to submit feedback.", false);
            }
        } catch (error) {
            showModal("An error occurred. Please try again.", false);
            console.error("Error submitting feedback:", error);
        }
    });
});
