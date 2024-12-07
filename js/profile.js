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

document.addEventListener("DOMContentLoaded", () => {
    const addAnimal = document.getElementById('addAnimalBtn');

    if (!addAnimal) {
        console.error("Element with id 'addAnimalBtn' not found.");
        return;
    }

    addAnimal.addEventListener('click', function() {
        let formHtml = `
            <div style="text-align: left; position: relative;">
                <span class="question-icon" title="Field requirements">&#x3F;</span> <!-- Символ вопроса -->
                <h1>Add Animal</h1>
                <form id="animalForm" action="/add-animal" method="POST" enctype="multipart/form-data">
                    <label for="name">Name:</label>
                    <input type="text" id="name" name="name"><br>
        
                    <label for="species">Species:</label>
                    <select id="species" name="species">
                        <option value="Dog">Dog</option>
                        <option value="Cat">Cat</option>
                        <option value="Bird">Bird</option>
                    </select><br>
        
                    <label for="breed">Breed:</label>
                    <input type="text" id="breed" name="breed"><br>
        
                    <label for="age_years">Age (Years):</label>
                    <input type="number" id="age_years" name="age_years" min="0"><br>
        
                    <label for="age_months">Age (Months):</label>
                    <input type="number" id="age_months" name="age_months" min="0" max="11"><br>
        
                    <label for="gender">Gender:</label>
                    <select id="gender" name="gender">
                        <option value="Female">Female</option>
                        <option value="Male">Male</option>
                    </select><br>
        
                    <label for="description">Description:</label>
                    <textarea id="description" name="description" style="resize: none;"></textarea><br>
        
                    <label for="location">Location:</label>
                    <input type="text" id="location" name="location"><br>
        
                    <label for="weight">Weight (kg):</label>
                    <input type="number" id="weight" name="weight" step="0.1"><br>
        
                    <label for="color">Color:</label>
                    <input type="text" id="color" name="color"><br>
        
                    <label for="is_sterilized">Sterilized:</label>
                    <input type="checkbox" id="is_sterilized" name="is_sterilized"><br>
        
                    <label for="has_passport">Has Passport:</label>
                    <input type="checkbox" id="has_passport" name="has_passport"><br>
        
                    <label for="images">Images (up to 4):</label>
                    <input type="file" id="images" name="images" accept="image/*" multiple><br>
        
                    <input type="submit" class="animal-submit" value="Add Animal">
                </form>
            </div>
        `;

        const openForm = (description = '') => {
            Swal.fire({
                html: formHtml.replace(
                    '<textarea id="description" name="description" style="resize: none;"></textarea>',
                    `<textarea id="description" name="description" style="resize: none;">${description}</textarea>`
                ),
                width: '30%',
                confirmButtonText: 'Close',
                showCloseButton: false,
                focusConfirm: false,
                didOpen: () => {
                    const form = document.querySelector('.swal2-container #animalForm');
                    const descriptionField = document.querySelector('#description');

                    // Добавляем обработчик для поля description
                    descriptionField.addEventListener('click', async (e) => {
                        e.preventDefault();
                        const currentText = descriptionField.value;
                        const { value: text } = await Swal.fire({
                            input: 'textarea',
                            inputLabel: 'Description',
                            inputPlaceholder: 'Type your description here...',
                            inputAttributes: {
                                'aria-label': 'Type your description here'
                            },
                            inputValue: currentText,
                            showCancelButton: true
                        });

                        if (text !== undefined) {
                            descriptionField.value = text;
                        }

                        // Открываем форму с обновленным текстом
                        openForm(descriptionField.value);
                    });

                    if (form) {
                        form.onsubmit = async (e) => {
                            e.preventDefault();

                            const formData = new FormData(form);

                            // Массив обязательных полей
                            const requiredFields = ["name", "breed", "age_years","age_months","description","location","weight","color"];
                            let validationPassed = true;

                            for (const field of requiredFields) {
                                const fieldElement = form.querySelector(`[name="${field}"]`);

                                // Если поле пустое, показываем предупреждение
                                if (!formData.get(field) || fieldElement.value.trim() === '') {
                                    Swal.fire({
                                        icon: "warning",
                                        title: "Warning",
                                        text: `The field "${fieldElement.previousElementSibling.innerText}" is required.`,
                                    });
                                    validationPassed = false;
                                    return; // Прерываем дальнейшую обработку
                                }
                            }

                            if (!validationPassed) return;

                            try {
                                const response = await fetch("/add-animal", {
                                    method: "POST",
                                    body: formData
                                });

                                const result = await response.json();

                                if (response.ok && result.status === "ok") {
                                    Swal.fire({
                                        icon: "success",
                                        title: "Success",
                                        text: result.message,
                                    });
                                } else {
                                    Swal.fire({
                                        icon: "error",
                                        title: "Error",
                                        text: result.message || "An unexpected error occurred.",
                                    });
                                }
                            } catch (error) {
                                console.error("Error:", error);
                                Swal.fire({
                                    icon: "error",
                                    title: "Error",
                                    text: "Failed to communicate with the server.",
                                });
                            }
                        };
                    }

                    // Добавляем обработчик для вопросика
                    const questionIcon = document.querySelector('.swal2-container .question-icon');
                    if (questionIcon) {
                        questionIcon.addEventListener('click', () => {
                            Swal.fire({
                                icon: "info",
                                title: "Field Requirements",
                                html: `
                                    <ul style="text-align: left;">
                                        <li><strong>Name:</strong> Should be at least 3 characters long.</li>
                                        <li><strong>Breed:</strong> Specify the breed of the animal.</li>
                                        <li><strong>Age:</strong> Enter the age in years and months.</li>
                                        <li><strong>Description:</strong> Provide a brief description of the animal.</li>
                                        <li><strong>Location:</strong> Specify the location where the animal is found.</li>
                                        <li><strong>Weight:</strong> Enter the weight in kilograms.</li>
                                        <li><strong>Color:</strong> Specify the color of the animal.</li>
                                    </ul>
                                `,
                                confirmButtonText: 'OK',
                                willClose: () => {
                                    openForm(descriptionField.value);
                                }
                            });
                        });
                    }
                },
                willClose: () => {
                    const form = document.querySelector('.swal2-container #animalForm');
                    if (form) {
                        form.reset();
                    }
                },
            });
        }

        openForm();
    });
});


