const profileImage = document.getElementById('profileImage');
const uploadImage = document.getElementById('uploadImage');


//иконка профиля
let uploadedImageURL = '';
let originalImageFile = null; // Хранение исходного изображения
const removeProfileImageButton = document.getElementById('removeProfileImage');
//фон профиля
const uploadBg =document.getElementById('uploadBg')
let originalBgImageFile = null; // Хранение исходного изображения
const removeBgButton = document.getElementById('removeBg');

let removeProfileImageFlag = false;
let removeBackgroundFlag = false;


const customUploadButton=document.getElementById('customUploadButton')
customUploadButton.addEventListener('click', function () {
    uploadImage.click(); // Триггерим клик на скрытом input для загрузки изображения
});
uploadImage.addEventListener('change', function (e) {
    const file = e.target.files[0];
    if (file) {
        originalImageFile = file; // Сохраняем оригинальный файл
        const reader = new FileReader();
        reader.onload = function (e) {
            uploadedImageURL = e.target.result;
            profileImage.src = uploadedImageURL; // Показать загруженное изображение
        };
        reader.readAsDataURL(file);
    }
});

const customBgButton = document.getElementById('customBgButton')
customBgButton.addEventListener('click', function () {
    uploadBg.click(); // Триггерим клик на скрытом input для загрузки фона
});


uploadBg.addEventListener('change', function (e) {
    const file = e.target.files[0];
    if (file) {
        originalBgImageFile = file; // Сохраняем оригинальный файл для фонового изображения
        const reader = new FileReader();
        reader.onload = function (event) {
            document.querySelector('.profile-background').style.backgroundImage = `url(${event.target.result})`;
        };
        reader.readAsDataURL(file);
    }
});

removeProfileImageButton.addEventListener('click', function () {
    profileImage.src = "system_images/default_profile.jpg"; // Устанавливаем изображение по умолчанию
    originalImageFile = null;
    removeProfileImageFlag = true; // Устанавливаем флаг для удаления изображения
    console.log('removeProfileImageFlag set to true');
});

removeBgButton.addEventListener('click', function () {
    document.querySelector('.profile-background').style.backgroundImage = "url('system_images/default_bg.jpg')";
    originalBgImageFile = null;
    removeBackgroundFlag = true; // Устанавливаем флаг для удаления фона
    console.log('removeBackgroundFlag set to true');
});


// Открываем модальное окно обрезки по нажатию кнопки "Edit / Crop Image"
let cropper;
const cropImage = document.getElementById('cropImage');
const editCropButton = document.getElementById('editCropButton');
const cropButton = document.getElementById('cropButton');
const closeModal = document.querySelector('.close');
const cropModal = document.getElementById('cropModal');
editCropButton.addEventListener('click', function () {
    const currentProfileImageSrc = profileImage.src;

    if (currentProfileImageSrc) {
        cropImage.src = currentProfileImageSrc;
        cropModal.style.display = 'block';

        if (cropper) {
            cropper.destroy();
        }

        cropper = new Cropper(cropImage, {
            aspectRatio: 1,
            viewMode: 1,
            autoCropArea: 1
        });
    } else {
        alert("No profile image available to crop.");
    }
});
cropButton.addEventListener('click', function () {
    if (cropper) {
        const croppedCanvas = cropper.getCroppedCanvas();
        const croppedImageDataUrl = croppedCanvas.toDataURL(); // Получаем URL обрезанного изображения

        profileImage.src = croppedImageDataUrl;

        console.log('Cropped Image Data URL:', croppedImageDataUrl);
        cropModal.style.display = 'none';
    }
});
cropButton.addEventListener('click', function () {
    cropModal.style.display = 'none'; // Закрываем модальное окно
});
closeModal.addEventListener('click', function () {
    cropModal.style.display = 'none';
    if (cropper) {
        cropper.destroy();
        cropper = null;
    }
});


//открытие закрытие боковой панели и модального окна обрезки по нажатии вне окна
const toggleButton = document.getElementById('toggleButton');
const sidebar = document.getElementById('sidebar');
window.addEventListener('click', function (e) {
    if (e.target === cropModal) {
        cropModal.style.display = 'none';
        if (cropper) {
            cropper.destroy();
            cropper = null;
        }
    }

    if (
        !sidebar.contains(e.target) &&
        !toggleButton.contains(e.target) &&
        sidebar.classList.contains('open')
    ) {
        sidebar.classList.remove('open');
    }
});
function toggleSidebar() {
    sidebar.classList.toggle('open');
}
//адаптивный перенос кнопки настроек
const originalContainer = document.getElementById('original-container');
const targetContainer = document.getElementById('target-container');

// Функция для перемещения кнопки
const mediaQuery = window.matchMedia('(max-width: 990px)');

function handleMediaChange(e) {
    if (e.matches) {
        targetContainer.appendChild(toggleButton);
    } else {
        originalContainer.appendChild(toggleButton);
    }
}
mediaQuery.addListener(handleMediaChange);
handleMediaChange(mediaQuery);


// Обработчик для отправки данных профиля
document.querySelector('.button-save').addEventListener('click', function (e) {
    e.preventDefault();

    const formData = new FormData();
    formData.append('firstName', document.getElementById('firstName').value);
    formData.append('lastName', document.getElementById('lastName').value);
    formData.append('bio', document.getElementById('bio').value);
    formData.append('phone', document.getElementById('phone').value);
    formData.append('dob', document.getElementById('dob').value);


    if (originalBgImageFile) {
        formData.append('backgroundImage', originalBgImageFile, originalBgImageFile.name);
    } else if (removeBackgroundFlag) {
        formData.append('removeBackgroundImage', 'true');
    }


    if (cropper) {
        cropper.getCroppedCanvas().toBlob(function (blob) {
            formData.append('croppedImage', blob, originalImageFile ? originalImageFile.name : 'cropped_image.jpg');
            if (removeProfileImageFlag) formData.append('removeProfileImage', 'true');
            sendProfileData(formData);
        }, 'image/jpeg');
    } else if (originalImageFile) {
        formData.append('croppedImage', originalImageFile, originalImageFile.name);
        if (removeProfileImageFlag) formData.append('removeProfileImage', 'true');
        sendProfileData(formData);
    } else {
        if (removeProfileImageFlag) formData.append('removeProfileImage', 'true');
        sendProfileData(formData);
    }
});

// Функция отправки данных FormData`  и логирование перед отправкой
function sendProfileData(formData) {
    saveSettings()
    console.log('FormData being sent:');
    for (let [key, value] of formData.entries()) {
        console.log(`${key}: ${value}`);
    }

    fetch('/save-profile', {
        method: 'POST',
        body: formData
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            if (data.success) {
                window.location.href = '/profile';
            } else {
                alert('Error saving profile');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            alert('There was an error with your request. Please try again.');
        });
}

function saveSettings() {
    const showEmail = document.getElementById("showEmail").checked;
    const showPhone = document.getElementById("showPhone").checked;

    // Логирование перед отправкой запроса
    console.log("Sending settings:", { showEmail, showPhone });

    fetch("/save-visibility-settings", {
        method: "POST",
        headers: {
            "Content-Type": "application/x-www-form-urlencoded"
        },
        body: `showEmail=${showEmail}&showPhone=${showPhone}`
    })
        .then(response => {
            console.log("Response status:", response.status); // Логируем статус ответа

            if (!response.ok) {
                throw new Error("Network response was not ok");
            }
            return response.text(); // Получаем ответ как текст
        })
        .then(text => {
            console.log("Response text:", text); // Логируем текст ответа
            try {
                const data = JSON.parse(text); // Пробуем парсить как JSON
                console.log("Parsed JSON:", data); // Логируем полученные данные

                const confirmationElement = document.getElementById("saveConfirmation");
                if (data.success) {
                    confirmationElement.style.display = "block";
                    confirmationElement.textContent = "Settings saved successfully!";
                    confirmationElement.style.color = "green";
                } else {
                    confirmationElement.style.display = "block";
                    confirmationElement.textContent = "Failed to save settings.";
                    confirmationElement.style.color = "red";
                }
                setTimeout(() => {
                    confirmationElement.style.display = "none";
                }, 3000);
            } catch (e) {
                console.error("Invalid JSON response:", e);
                const confirmationElement = document.getElementById("saveConfirmation");
                confirmationElement.style.display = "block";
                confirmationElement.textContent = "Failed to parse server response.";
                confirmationElement.style.color = "red";
                setTimeout(() => {
                    confirmationElement.style.display = "none";
                }, 3000);
            }
        })
        .catch(error => {
            console.error("Error updating visibility settings:", error);
            const confirmationElement = document.getElementById("saveConfirmation");
            confirmationElement.style.display = "block";
            confirmationElement.textContent = "An error occurred.";
            confirmationElement.style.color = "red";
            setTimeout(() => {
                confirmationElement.style.display = "none";
            }, 3000);
        });
}
document.getElementById('logout').addEventListener('click', function () {
    const confirmLogout = confirm('This action will redirect you to the Home Page and will log you out. ' +
        'Are you sure?');
    if (confirmLogout) {
        // Действие при подтверждении
        window.location.href = '/logout'; // Ссылка на выход
    } else {
        // Действие при отмене (если нужно)
        console.log('Пользователь отменил выход');
    }
});

document.addEventListener("DOMContentLoaded", function() {
    const phoneNumberInput = document.getElementById('phone');

    // Устанавливаем начальное значение "+996 "
    if (!phoneNumberInput.value) {
        phoneNumberInput.value = '+996 ';
    }

    phoneNumberInput.addEventListener('input', function(e) {
        let input = e.target.value.replace(/\D/g, ''); // Удаляем все нецифровые символы

        // Предзаполняем код страны "+996"
        if (!input.startsWith('996')) {
            input = '996' + input;
        }

        // Обрезаем до максимальной длины номера (9 цифр после кода страны)
        input = input.slice(0, 13);

        // Форматируем номер в вид +996 XXX XX XX XX
        let formattedNumber = '+996 ';
        if (input.length > 3) {
            formattedNumber += input.slice(3, 6);  // XXX
        }
        if (input.length > 6) {
            formattedNumber += ' ' + input.slice(6, 8);  // XX
        }
        if (input.length > 8) {
            formattedNumber += ' ' + input.slice(8, 10);  // XX
        }
        if (input.length > 10) {
            formattedNumber += ' ' + input.slice(10, 12);  // XX
        }

        e.target.value = formattedNumber;
    });

    phoneNumberInput.addEventListener('focus', function() {
        // Если поле пустое, подставляем "+996 " при фокусе
        if (phoneNumberInput.value === '') {
            phoneNumberInput.value = '+996 ';
        }
    });

    phoneNumberInput.addEventListener('blur', function() {
        // Если после потери фокуса введен только "+996 ", очищаем поле
        if (phoneNumberInput.value === '+996 ') {
            phoneNumberInput.value = '';
        }
    });
});


document.addEventListener("DOMContentLoaded", function() {
    const dobInput = document.getElementById('dob');

    // Устанавливаем минимальную и максимальную дату
    const today = new Date();
    const minDate = new Date('1900-01-01');
    const maxDate = today;

    // Преобразуем даты в формат YYYY-MM-DD
    const minDateString = minDate.toISOString().split('T')[0];
    const maxDateString = maxDate.toISOString().split('T')[0];

    // Применяем ограничения
    dobInput.setAttribute('min', minDateString);
    dobInput.setAttribute('max', maxDateString);
});
