document.addEventListener("DOMContentLoaded", () => {
    const form = document.getElementById("feedback-form");
    const feedbackMessage = document.getElementById("feedback-message");

    form.addEventListener("submit", async (e) => {
        e.preventDefault(); // Предотвращаем стандартную отправку формы

        const formData = new FormData(form);
        try {
            const response = await fetch(form.action, {
                method: form.method,
                body: formData,
            });

            const data = await response.json(); // Получаем JSON от хендлера

            // Очищаем старые сообщения
            feedbackMessage.innerHTML = "";

            // Обрабатываем сообщение
            if (data.success === "true") {
                feedbackMessage.textContent = data.message || "Thank you for your feedback!";
                feedbackMessage.style.color = "green"; // Успешное сообщение
                form.reset(); // Сбрасываем форму
            } else {
                feedbackMessage.textContent = data.message || "Failed to submit feedback.";
                feedbackMessage.style.color = "red"; // Ошибочное сообщение
            }
        } catch (error) {
            feedbackMessage.textContent = "An error occurred. Please try again.";
            feedbackMessage.style.color = "red";
            console.error("Error submitting feedback:", error);
        }
    });
});
