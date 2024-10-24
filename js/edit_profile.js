let cropper;
const profileImage = document.getElementById('profileImage');
const uploadImage = document.getElementById('uploadImage');
const cropModal = document.getElementById('cropModal');
const cropImage = document.getElementById('cropImage');
const closeModal = document.querySelector('.close');
const cropButton = document.getElementById('cropButton');
const editCropButton = document.getElementById('editCropButton');
let uploadedImageURL = '';
let originalImageFile = null; // Хранение исходного изображения

// Обработка клика по кастомной кнопке загрузки
customUploadButton.addEventListener('click', function () {
    uploadImage.click(); // Триггерим клик на скрытом input для загрузки изображения
});
cropButton.addEventListener('click', function () {
    cropModal.style.display = 'none'; // Закрываем модальное окно
});

// Обработка загрузки изображения
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

closeModal.addEventListener('click', function () {
    cropModal.style.display = 'none'; // Закрываем модальное окно
});

// Обработчик для отправки данных профиля
document.querySelector('.button-save').addEventListener('click', function (e) {
    e.preventDefault(); // Предотвращаем отправку формы

    const formData = new FormData();
    formData.append('firstName', document.getElementById('firstName').value);
    formData.append('lastName', document.getElementById('lastName').value);
    formData.append('bio', document.getElementById('bio').value);
    formData.append('phone', document.getElementById('phone').value);
    formData.append('dob', document.getElementById('dob').value);

    // Проверка, был ли создан cropper (значит, пользователь редактировал изображение)
    if (cropper) {
        cropper.getCroppedCanvas().toBlob(function (blob) {
            formData.append('croppedImage', blob, originalImageFile ? originalImageFile.name : 'cropped_image.jpg');

            // Отправляем форму только после создания блоба
            sendProfileData(formData);
        }, 'image/jpeg');
    } else if (originalImageFile) {
        // Если изображение не обрезано, передаем его как есть
        formData.append('croppedImage', originalImageFile, originalImageFile.name);
        sendProfileData(formData);
    } else {
        // Если изображение не загружено
        sendProfileData(formData);
    }
});

// Функция отправки данных
// Функция отправки данных
function sendProfileData(formData) {
    fetch('/save-profile', {
        method: 'POST',
        body: formData
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json(); // Парсить JSON ответ
        })
        .then(data => {
            if (data.success) {
                window.location.href = '/profile'; // Перенаправляем пользователя на страницу профиля
            } else {
                alert('Error saving profile'); // Если успеха нет, показываем сообщение
            }
        })
        .catch(error => {
            console.error('Error:', error); // Логирование ошибок
            alert('There was an error with your request. Please try again.'); // Сообщение об ошибке
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


// Закрытие модального окна при клике вне его
window.addEventListener('click', function (e) {
    if (e.target === cropModal) {
        cropModal.style.display = 'none';
    }
});
