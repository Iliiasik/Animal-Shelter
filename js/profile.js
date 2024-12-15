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
                <input type="number" id="age_years" name="age_years"><br>
                <label for="age_months">Age (Months):</label>
                <input type="number" id="age_months" name="age_months"><br>
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
                    <button type="submit" class="animal-submit" >Submit</button>
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
            customClass: {
                confirmButton: 'custom-close-button', // Назначаем класс для кнопки
            },
            //ТУТ НУЖНА ДОРАБОТКА. ЭТА ЧАСТЬ КОДА БЫЛА НАПИСАНА В ПОПЫТКЕ ПЕРЕМЕСТИТЬ ВОПРОСИК ВНЕ ФОРМЫ
            // backdrop:'<div class="question-icon-container">\n' +
            //         '    <span class="question-icon" title="Field requirements">&#x3F;</span>\n' +
            //         '</div>',
            didOpen: setupFormInteractions,
            willClose: resetFormState,
        });
    };
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
    const applyFieldRestrictions = () => {
        const ageYearsInput = document.getElementById('age_years');
        const ageMonthsInput = document.getElementById('age_months');
        const weightInput = document.getElementById('weight');

        if (ageYearsInput) {
            ageYearsInput.addEventListener('input', () => {
                let value = parseInt(ageYearsInput.value, 10);
                if (isNaN(value) || value < 0) value = 0;
                if (value > 50) value = 50;
                ageYearsInput.value = value;

                // Проверка для месяцев при изменении значения года
                if (ageMonthsInput) {
                    let monthsValue = parseInt(ageMonthsInput.value, 10);
                    if (value === 0 && monthsValue === 0) {
                        ageMonthsInput.value = 1; // Устанавливаем минимум 1, если год равен 0
                    }
                }
            });
        }

        if (ageMonthsInput) {
            ageMonthsInput.addEventListener('input', () => {
                let value = parseInt(ageMonthsInput.value, 10);
                const ageYearsValue = parseInt(ageYearsInput?.value || 0, 10);

                if (isNaN(value) || value < 0) value = 0;

                // Если значение больше 11, берем последнюю цифру
                if (value > 11) {
                    value = parseInt(value.toString().slice(-1), 10);
                }

                // Разрешить 0 только если год больше 0
                if (value === 0 && ageYearsValue === 0) {
                    value = 1;
                }

                ageMonthsInput.value = value;
            });
        }

        if (weightInput) {
            weightInput.addEventListener('input', () => {
                let value = parseFloat(weightInput.value);
                if (isNaN(value) || value < 0) value = 0;
                if (value > 150) value = 150;
                weightInput.value = value;
            });
        }
    };


    const setupFormInteractions = () => {
        const form = document.querySelector('.swal2-container #animalForm');
        const descriptionField = document.querySelector('#description');


        restoreFormState(form);

        applyFieldRestrictions();

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
    const handleFormSubmitResponse = (response, result) => {
        if (response.ok && result.status === "ok") {
            Swal.fire({
                icon: "success",
                title: "Success",
                text: result.message
            });
        } else {
            Swal.fire({
                icon: "error",
                title: "Error",
                text: result.message || "An unexpected error occurred.",
            }).then(() => {
                openFormWithState();
            });
        }
    };
    const openFormWithState = () => {
        Swal.fire({
            html: getFormHtml(), // Функция для генерации HTML формы
            confirmButtonText: 'Close',
            showCloseButton: false,
            focusConfirm: false,
            didOpen: () => {
                const form = document.querySelector('.swal2-container #animalForm');
                if (form) {
                    restoreFormState(form); // Восстанавливаем состояние формы
                    setupFormInteractions(); // Устанавливаем события для формы
                }
            },
            willClose: resetFormState, // Сброс состояния при закрытии
        });
    };

    const validateFields = (form) => {

        const validationRules = {
            name: {
                maxLength: 15,
                minLength: 2,
                message: "Name should not exceed 15 characters."
            },
            species: {
                checkSelected: true,
                message: "Species must be selected."
            },
            breed: {
                maxLength: 50,
                minLength: 3,
                message: "Breed should not exceed 50 characters."
            },
            age_years: {
                max: 50,
                min: 0,
                checkNotEmpty: true,
                message: "Age (Years) should be between 0 and 50."
            },
            age_months: {
                max: 12,
                min: 0,
                checkNotEmpty: true,
                message: "Age (Months) should be between 0 and 12."
            },
            gender: {
                checkSelected: true,
                message: "Gender must be selected."
            },
            description: {
                minLength: 50,
                maxLength: 500,
                message: "Description should be between 50 and 500 characters.",
            },
            location: {
                maxLength: 50,
                minLength: 3,
                message: "Location should not exceed 50 characters."
            },
            color: {
                maxLength: 50,
                minLength: 3,
                message: "Color should not exceed 50 characters."
            },
            weight:{
                checkNotEmpty: true, // Ensure weight is not empty
            },
            // Проверка на наличие хотя бы одного изображения
            images: {
                checkAtLeastOneImage: true,
                message: "At least one image is required."
            }
        };

        const fields = form.querySelectorAll("input, select, textarea");
        let isValid = true;

        fields.forEach((field) => {
            const rule = validationRules[field.name];
            if (!rule) return;

            let fieldIsValid = true;

            if (rule.checkSelected && field.tagName === "SELECT") {
                if (field.value === "" || field.value === null) {
                    fieldIsValid = false;
                }
            }

            if (rule.maxLength && field.value.length > rule.maxLength) {
                fieldIsValid = false;
            }

            if (rule.minLength && field.value.length < rule.minLength) {
                fieldIsValid = false;
            }

            if (rule.checkNotEmpty && field.value.trim() === "") {
                fieldIsValid = false;
                fieldIsValidMessage = rule.messageEmpty;
            }

            if (rule.min !== undefined || rule.max !== undefined) {
                const value = parseInt(field.value);
                if (isNaN(value) || (rule.min !== undefined && value < rule.min) || (rule.max !== undefined && value > rule.max)) {
                    fieldIsValid = false;
                }
            }


            if (rule.checkAtLeastOneImage) {
                const imageInputs = form.querySelectorAll('input[type="file"]');
                const hasImage = Array.from(imageInputs).some(input => input.files.length > 0);
                if (!hasImage) {
                    fieldIsValid = false;
                    isValid = false;
                    const warningElement = document.querySelector(`#${field.name}-warning`);
                    if (warningElement) {
                        warningElement.classList.add('invalid-warning');
                        warningElement.innerText = rule.message;
                    }
                }
            }

            // Custom validation for age_years and age_months
            if (field.name === 'age_years' || field.name === 'age_months') {
                const ageYearsField = form.querySelector('input[name="age_years"]');
                const ageMonthsField = form.querySelector('input[name="age_months"]');

                const ageYears = parseInt(ageYearsField.value) || 0;
                const ageMonths = parseInt(ageMonthsField.value) || 0;

                if (ageYears === 0 && ageMonths === 0) {
                    fieldIsValid = false;
                    isValid = false;
                    const warningElement = document.querySelector(`#age_months-warning`);
                    if (warningElement) {
                        warningElement.classList.add('invalid-warning');
                        warningElement.innerText = "If age is 0 years, months must be at least 1.";
                    }
                }
            }

            const warningElement = document.querySelector(`#${field.name}-warning`);

            if (!fieldIsValid) {
                isValid = false;
                triggerQuestionMark(field); // Вызов функции для иконки
                highlightInvalidField(field); // Подсветка рамки
                if (warningElement) {
                    warningElement.classList.add('invalid-warning');
                    warningElement.innerText = rule.message || fieldIsValidMessage;
                }
            } else {
                removeHighlight(field); // Убираем подсветку, если поле валидно
                if (warningElement) {
                    warningElement.classList.remove('invalid-warning');
                }
            }
        });

        return isValid;
    }

    const highlightInvalidField = (field) => {
        field.classList.add("invalid-blink");
        setTimeout(() => {
            field.classList.remove("invalid-blink");
        }, 5000); // Мигание будет длиться 5 секунд
    };

        const removeHighlight = (field) => {
        field.classList.remove("invalid-blink");
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

    const showFieldRequirements = (descriptionField) => {
        Swal.fire({
            icon: "info",
            title: "Field Requirements",
            html: `
        <ul style="text-align: left;">
            <li id="name-warning" class="field-warning"><strong>Name:</strong> Should be at least 2 characters long.</li>
            <li id="breed-warning" class="field-warning"><strong>Breed:</strong> Specify the breed of the animal.</li>
            <li id="age-warning" class="field-warning"><strong>Age:</strong> Enter the age in years and months.</li>
            <li id="description-warning" class="field-warning"><strong>Description:</strong> Provide a brief description of the animal.<br>Description should be between 50 and 500 characters.</li>
            <li id="location-warning" class="field-warning"><strong>Location:</strong> Specify the location where the animal is found.</li>
            <li id="weight-warning" class="field-warning"><strong>Weight:</strong> Enter the weight in kilograms.</li>
            <li id="color-warning" class="field-warning"><strong>Color:</strong> Specify the color of the animal.</li>
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
document.addEventListener('DOMContentLoaded', function () {
    // Обработчик клика на контейнер животного
    const animalElements = document.querySelectorAll('.animal');

    animalElements.forEach(animal => {
        animal.addEventListener('click', function () {
            const animalId = this.querySelector('.delete-button').getAttribute('data-animal-id');
            if (animalId) {
                // Редирект на страницу информации о животном
                window.location.href = `/animal_information?id=${animalId}`;
            }
        });
    });

    // Обработчик клика для кнопки удаления
    const deleteButtons = document.querySelectorAll('.remove-button');

    deleteButtons.forEach(button => {
        button.addEventListener('click', function (event) {
            event.stopPropagation(); // Предотвращаем всплытие события клика на родительский контейнер
            const animalId = this.getAttribute('data-animal-id');

            Swal.fire({
                title: 'Are you sure?',
                text: "You won't be able to revert this!",
                icon: 'warning',
                showCancelButton: true,
                confirmButtonColor: '#d33',
                cancelButtonColor: '#3085d6',
                confirmButtonText: 'Yes, delete it!'
            }).then((result) => {
                if (result.isConfirmed) {
                    // Отправка запроса на удаление
                    fetch(`/animals/delete`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify({ id: parseInt(animalId, 10) }),
                    })
                        .then(response => response.json())
                        .then(data => {
                            if (data.status === 'ok') {
                                Swal.fire(
                                    'Deleted!',
                                    'Your animal has been deleted.',
                                    'success'
                                );
                                // Удаляем элемент из DOM
                                this.closest('li').remove();
                            } else {
                                Swal.fire(
                                    'Error!',
                                    data.message || 'Something went wrong.',
                                    'error'
                                );
                            }
                        })
                        .catch(error => {
                            Swal.fire(
                                'Error!',
                                'Failed to delete the animal.',
                                'error'
                            );
                        });
                }
            });
        });
    });
});
document.addEventListener('DOMContentLoaded', function() {
    // Проверьте, если редирект с "already_logged_in"
    const params = new URLSearchParams(window.location.search);
    if (params.has('already_logged_in') && params.get('already_logged_in') === 'true') {
        const Toast = Swal.mixin({
            toast: true,
            position: 'top-end',
            showConfirmButton: false,
            timer: 3000,
            timerProgressBar: true,
            customClass: {
                container: 'custom-toast-container'
            },
            didOpen: (toast) => {
                toast.addEventListener('mouseenter', Swal.stopTimer);
                toast.addEventListener('mouseleave', Swal.resumeTimer);
            }
        });

        Toast.fire({
            icon: 'info',
            title: 'You are already logged in!'
        });

        // Удаляем параметр из URL после показа уведомления
        const url = new URL(window.location);
        url.searchParams.delete('already_logged_in');
        window.history.replaceState({}, document.title, url);
    }
});



