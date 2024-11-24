document.addEventListener("DOMContentLoaded", () => {
    // Получаем элемент с отображением даты
    const dobDisplay = document.getElementById("dob-display");

    // Проверяем, найден ли элемент
    if (!dobDisplay) {
        console.error("Element with id 'dob-display' not found.");
        return; // Останавливаем выполнение, если элемент не найден
    }

    // Функция для вычисления возраста
    function calculateAge(dateString) {
        const birthDate = new Date(dateString);
        if (isNaN(birthDate)) {
            console.error("Invalid date format:", dateString);
            return NaN;
        }
        const today = new Date();
        let age = today.getFullYear() - birthDate.getFullYear();
        const monthDiff = today.getMonth() - birthDate.getMonth();
        const dayDiff = today.getDate() - birthDate.getDate();

        // Проверяем, был ли день рождения уже в этом году
        if (monthDiff < 0 || (monthDiff === 0 && dayDiff < 0)) {
            age--;
        }
        return age;
    }

    // Функция для обновления отображения даты и возраста
    function updateDisplay() {
        const rawDate = dobDisplay.innerText.trim(); // Получаем исходную строку даты
        let isoDate;

        // Попробуем извлечь только часть с датой (YYYY-MM-DD)
        if (rawDate.includes(" ")) {
            // Если дата с временем, берём первую часть
            isoDate = rawDate.split(" ")[0];
        } else {
            // Если уже форматированная дата, оставляем её как есть
            isoDate = rawDate;
        }

        // Проверяем корректность даты
        const age = calculateAge(isoDate);
        if (!isNaN(age)) {
            dobDisplay.innerText = `${isoDate} (${age} years old)`; // Обновляем отображение
        } else {
            dobDisplay.innerText = "Invalid date"; // Некорректный формат
        }
    }

    // Инициализация обновления
    updateDisplay();
});
