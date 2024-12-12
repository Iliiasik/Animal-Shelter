document.addEventListener("DOMContentLoaded", () => {
    const form = document.getElementById("feedback-form");

    // Функция для отображения модального окна
    const showModal = (message, success = true) => {
        const imageSrc = success ? "./system_images/thanks.jpg" : "./system_images/I-REFUSE.jpg";

        // Отображаем модальное окно с помощью SweetAlert
        Swal.fire({
            title: message,
            text: success ? "Thank you for your feedback!" : "Failed to submit feedback.",
            icon: success ? "success" : "error",
            imageUrl: imageSrc,
            imageWidth: 400,  // Размер изображения по ширине (можно подстроить под ваш макет)
            imageHeight: 400, // Размер изображения по высоте
            imageAlt: 'Feedback Image',
            showConfirmButton: true,
            confirmButtonText: 'Close',
            willClose: () => {
                window.location.href = "/"; // Переадресация на главную страницу после закрытия
            }
        });
    };

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
