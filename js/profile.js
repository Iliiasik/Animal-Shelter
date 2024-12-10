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


    const addAnimal = document.getElementById('addAnimalBtn');

    if (!addAnimal) {
        console.error("Element with id 'addAnimalBtn' not found.");
        return;
    }

    let formState = {};

    const saveFormState = (form) => {
        const inputs = form.querySelectorAll('input, select, textarea');
        formState = {};
        inputs.forEach((input) => {
            if (input.type === "checkbox") {
                formState[input.name] = input.checked;
            } else if (input.type === "file") {
                formState[input.name] = input.files;
            } else {
                formState[input.name] = input.value;
            }
        });
    };

    const restoreFormState = (form) => {
        const inputs = form.querySelectorAll('input, select, textarea');
        inputs.forEach((input) => {
            if (input.type === "checkbox") {
                input.checked = !!formState[input.name];
            } else if (input.type === "file") {
                // File inputs are read-only, skipping restoration
            } else {
                input.value = formState[input.name] || '';
            }
        });
    };

    const getFormHtml = () => `
        <div class="animal-form-container" style="text-align: left; position: relative;">
            <span class="question-icon" title="Field requirements">&#x3F;</span>
            <h1>Add Animal</h1>
            <form id="animalForm" action="/add-animal" method="POST" enctype="multipart/form-data">
                <!-- Поля формы -->
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
                <textarea id="description" name="description" style="resize: none;" readonly></textarea><br>
                <label for="location">Location:</label>
                <input type="text" id="location" name="location"><br>
                <label for="weight">Weight (kg):</label>
                <input type="number" id="weight" name="weight" step="0.1"><br>
                <label for="color">Color:</label>
                <input type="text" id="color" name="color"><br>
                <label class="checkbox-wrapper-31">
                    <input type="checkbox" id="is_sterilized" name="is_sterilized">
                    <svg viewBox="0 0 35.6 35.6">
                        <circle class="background" cx="17.8" cy="17.8" r="17.8"></circle>
                        <circle class="stroke" cx="17.8" cy="17.8" r="14.37"></circle>
                        <polyline class="check" points="11.78 18.12 15.55 22.23 25.17 12.87"></polyline>
                    </svg>
                    <span>Sterilized</span>
                </label><br>
                <label class="checkbox-wrapper-31">
                    <input type="checkbox" id="has_passport" name="has_passport">
                    <svg viewBox="0 0 35.6 35.6">
                        <circle class="background" cx="17.8" cy="17.8" r="17.8"></circle>
                        <circle class="stroke" cx="17.8" cy="17.8" r="14.37"></circle>
                        <polyline class="check" points="11.78 18.12 15.55 22.23 25.17 12.87"></polyline>
                    </svg>
                    <span>Has Passport</span>
                </label><br>
                <label for="images">Images (up to 4):</label>
                <div class="file-upload-container">
                    <input type="file" id="images" name="images" accept="image/*" multiple style="display:none;">
                    <button type="button" class="upload-btn" onclick="document.getElementById('images').click();">Upload Images</button>
                    <span id="file-chosen">No files chosen</span>
                </div>

                <div class="submit-container">
                    <input type="submit" class="animal-submit" value="Add Animal">
                </div>
            </form>
        </div>`;


    const openForm = (description = '') => {
        Swal.fire({
            html: getFormHtml().replace(
                '<textarea id="description" name="description" style="resize: none;"></textarea>',
                `<textarea id="description" name="description" style="resize: none;">${description}</textarea>`
            ),
            confirmButtonText: 'Close',
            showCloseButton: false,
            focusConfirm: false,
            backdrop:'<div class="question-icon-container">\n' +
                '    <span class="question-icon" title="Field requirements">&#x3F;</span>\n' +
                '</div>',
            didOpen: setupFormInteractions,
            willClose: resetFormState,
        });
    };

    const setupFormInteractions = () => {
        const form = document.querySelector('.swal2-container #animalForm');
        const descriptionField = document.querySelector('#description');

        restoreFormState(form);

        descriptionField.addEventListener('click', (e) => handleDescriptionClick(e, form));

        if (form) form.onsubmit = (e) => handleFormSubmit(e, form);

        // Устанавливаем слушатель для поля загрузки файлов
        const fileInput = document.getElementById('images');
        if (fileInput) {
            fileInput.addEventListener('change', function () {
                const fileChosen = document.getElementById('file-chosen');

                if (!fileChosen) {
                    console.error("Element with id 'file-chosen' not found.");
                    return;
                }

                if (fileInput.files.length === 0) {
                    fileChosen.textContent = 'No files chosen';
                    fileChosen.removeAttribute('title'); // Убираем подсказку
                } else if (fileInput.files.length === 1) {
                    const fileName = fileInput.files[0].name;
                    fileChosen.textContent = fileName.length > 20 ? fileName.substring(0, 13) + '...' : fileName;
                    fileChosen.setAttribute('title', fileName); // Показываем полное имя как подсказку
                } else {
                    const fileCount = fileInput.files.length;
                    fileChosen.textContent = `${fileCount} files chosen`;
                    const fileNames = Array.from(fileInput.files).map(file => file.name).join(', ');
                    fileChosen.setAttribute('title', fileNames); // Список всех файлов в подсказке
                }
            });
        }

        const questionIcon = document.querySelector('.swal2-container .question-icon');
        if (questionIcon) questionIcon.addEventListener('click', () => showFieldRequirements(descriptionField));
    };

    const handleDescriptionClick = async (e, form) => {
        e.preventDefault();
        saveFormState(form);

        const descriptionField = e.target;
        const currentText = descriptionField.value;
        const { value: text } = await Swal.fire({
            input: 'textarea',
            inputLabel: 'Description',
            inputPlaceholder: 'Type your description here...',
            inputAttributes: { 'aria-label': 'Type your description here' },
            inputValue: currentText,
            showCancelButton: true,
        });

        if (text !== undefined) descriptionField.value = text;
        saveFormState(form);

        openForm(descriptionField.value);
    };

    const handleFormSubmit = async (e, form) => {
        e.preventDefault();
        if (!validateFields(form)) {
            console.warn("Validation failed. Check the question mark for more details.");
            return;
        }
        saveFormState(form);

        const formData = new FormData(form);

        try {
            const response = await fetch("/add-animal", { method: "POST", body: formData });
            const result = await response.json();
            handleFormSubmitResponse(response, result);
        } catch (error) {
            console.error("Error:", error);
            Swal.fire({ icon: "error", title: "Error", text: "Failed to communicate with the server." });
        }
    };

    const validateFields = (form) => {
        const validationRules = {
            name: {
                maxLength: 15,
                minLength: 2,
                message: "Name should not exceed 15 characters."
            },
            breed: {
                maxLength: 50,
                message: "Breed should not exceed 50 characters."
            },
            color: {
                maxLength: 50,
                message: "Color should not exceed 50 characters."
            },
            location: {
                maxLength: 50,
                message: "Location should not exceed 50 characters."
            },
            description: {
                minLength: 50,
                maxLength: 500,
                placeholder: "What does your pet like to chew?",
                message: "Description should be between 50 and 500 characters."
            },
            weight: {
                pattern: /^[0-9]+(\.[0-9]+)?$/,
                message: "Weight should be a decimal number using a dot (e.g., 4.5)."
            },
        };

        const fields = form.querySelectorAll("input, textarea");
        let isValid = true;

        fields.forEach((field) => {
            const rule = validationRules[field.name];
            if (!rule) return;

            if (rule.maxLength && field.value.length > rule.maxLength) {
                isValid = false;
                triggerQuestionMark(field);
            }

            if (rule.minLength && field.value.length < rule.minLength) {
                isValid = false;
                triggerQuestionMark(field);
            }

            if (rule.pattern && !rule.pattern.test(field.value)) {
                isValid = false;
                triggerQuestionMark(field);
            }

            if (field.name === "description" && rule.placeholder) {
                field.placeholder = rule.placeholder;
            }
        });

        return isValid;
    };
    const triggerQuestionMark = (field) => {
        const questionIcon = document.querySelector('.swal2-container .question-icon');
        if (questionIcon) {
            // Добавляем класс анимации
            questionIcon.classList.add('shake');

            // Удаляем класс анимации после завершения (0.3s в данном случае)
            setTimeout(() => {
                questionIcon.classList.remove('shake');
            }, 300);
        }
    };



    const handleFormSubmitResponse = (response, result) => {
        if (response.ok && result.status === "ok") {
            Swal.fire({ icon: "success", title: "Success", text: result.message });
        } else {
            Swal.fire({
                icon: "error",
                title: "Error",
                text: result.message || "An unexpected error occurred.",
            });
        }
    };

    const showFieldRequirements = (descriptionField) => {
        Swal.fire({
            icon: "info",
            title: "Field Requirements",
            html: `
                <ul style="text-align: left;">
                    <li><strong>Name:</strong> Should be at least 2 characters long.</li>
                    <li><strong>Breed:</strong> Specify the breed of the animal.</li>
                    <li><strong>Age:</strong> Enter the age in years and months.</li>
                    <li><strong>Description:</strong> Provide a brief description of the animal.<br>Description should be between 50 and 500 characters.</li>
                    <li><strong>Location:</strong> Specify the location where the animal is found.</li>
                    <li><strong>Weight:</strong> Enter the weight in kilograms.</li>
                    <li><strong>Color:</strong> Specify the color of the animal.</li>
                </ul>`,
            confirmButtonText: 'OK',
            willClose: () => openForm(descriptionField.value),
        });
    };

    const resetFormState = () => {
        formState = {};  // Сбрасываем состояние формы
        const form = document.querySelector('.swal2-container #animalForm');
        if (form) form.reset();  // Сбрасываем форму
    };

    addAnimal.addEventListener('click', () => openForm());
});


