let cropper;
const customUploadButton=document.getElementById('customUploadButton')
const profileImage = document.getElementById('profileImage');
const uploadImage = document.getElementById('uploadImage');
const cropModal = document.getElementById('cropModal');
const cropImage = document.getElementById('cropImage');
const closeModal = document.querySelector('.close');
const cropButton = document.getElementById('cropButton');
const editCropButton = document.getElementById('editCropButton');
const customBgButton = document.getElementById('customBgButton')
//иконка профиля
let uploadedImageURL = '';
let originalImageFile = null; // Хранение исходного изображения
const removeProfileImageButton = document.getElementById('removeProfileImage');
//фон профиля
const uploadBg =document.getElementById('uploadBg')
let originalBgImageFile = null; // Хранение исходного изображения
const removeBgButton = document.getElementById('removeBg');

let removeProfileImageFlag = false; // Флаг для удаления изображения профиля
let removeBackgroundFlag = false;   // Флаг для удаления фонового изображения

// Обработка клика по кастомной кнопке загрузки иконки
customUploadButton.addEventListener('click', function () {
    uploadImage.click(); // Триггерим клик на скрытом input для загрузки изображения
});
cropButton.addEventListener('click', function () {
    cropModal.style.display = 'none'; // Закрываем модальное окно
});

// Обработка загрузки иконки
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

// Обработчик клика на кнопке редактирования фонового изображения
customBgButton.addEventListener('click', function () {
    uploadBg.click(); // Триггерим клик на скрытом input для загрузки фона
});

// Обработка загрузки фонового изображения
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
// Обработка удаления изображения профиля
removeProfileImageButton.addEventListener('click', function () {
    profileImage.src = "system_images/default_profile.jpg"; // Устанавливаем изображение по умолчанию
    originalImageFile = null;
    removeProfileImageFlag = true; // Устанавливаем флаг для удаления изображения
    console.log('removeProfileImageFlag set to true');
});
// Обработка удаления фонового изображения
removeBgButton.addEventListener('click', function () {
    document.querySelector('.profile-background').style.backgroundImage = "url('system_images/default_bg.jpg')";
    originalBgImageFile = null;
    removeBackgroundFlag = true; // Устанавливаем флаг для удаления фона
    console.log('removeBackgroundFlag set to true');
});
// Открываем модальное окно обрезки по нажатию кнопки "Edit / Crop Image"
editCropButton.addEventListener('click', function () {
    // Проверяем, есть ли текущее изображение
    const currentProfileImageSrc = profileImage.src;

    if (currentProfileImageSrc) {
        cropImage.src = currentProfileImageSrc; // Устанавливаем текущее изображение в модалку
        cropModal.style.display = 'block'; // Открываем модальное окно

        if (cropper) {
            cropper.destroy(); // Уничтожаем предыдущий экземпляр cropper
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

// Сохраняем обрезанное изображение при нажатии на кнопку "Crop & Save"
cropButton.addEventListener('click', function () {
    if (cropper) {
        const croppedCanvas = cropper.getCroppedCanvas();
        const croppedImageDataUrl = croppedCanvas.toDataURL(); // Получаем URL обрезанного изображения

        // Обновляем иконку профиля новым обрезанным изображением
        profileImage.src = croppedImageDataUrl;
        // Сохраняем или отправляем обрезанное изображение на сервер
        console.log('Cropped Image Data URL:', croppedImageDataUrl);
        cropModal.style.display = 'none';
    }
});

// Закрытие модального окна только при нажатии на крестик
closeModal.addEventListener('click', function () {
    cropModal.style.display = 'none'; // Закрываем модальное окно

    // Уничтожаем cropper и сбрасываем состояние
    if (cropper) {
        cropper.destroy();
        cropper = null;
    }
});
window.addEventListener('click', function (e) {
    if (e.target === cropModal) {
        cropModal.style.display = 'none';
        if (cropper) {
            cropper.destroy(); // Уничтожаем экземпляр cropper, отменяя все изменения
            cropper = null;
        }
    }
});
// Обработчик для отправки данных профиля
document.querySelector('.button-save').addEventListener('click', function (e) {
    e.preventDefault();

    const formData = new FormData();
    formData.append('firstName', document.getElementById('firstName').value);
    formData.append('lastName', document.getElementById('lastName').value);
    formData.append('bio', document.getElementById('bio').value);
    formData.append('phone', document.getElementById('phone').value);
    formData.append('dob', document.getElementById('dob').value);

    // Обработка фонового изображения
    if (originalBgImageFile) {
        formData.append('backgroundImage', originalBgImageFile, originalBgImageFile.name);
    } else if (removeBackgroundFlag) {
        formData.append('removeBackgroundImage', 'true');
    }

    // Обработка изображения профиля
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


