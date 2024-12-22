document.addEventListener('DOMContentLoaded', () => {
    const logoutLink = document.querySelector('a[href="/logout"]'); // Найти ссылку логаута

    if (logoutLink) {
        logoutLink.addEventListener('click', (event) => {
            event.preventDefault(); // Отключить стандартное поведение ссылки

            Swal.fire({
                title: 'Are you sure?',
                text: 'This will log you out and redirect you to the home page.',
                icon: 'warning',
                showCancelButton: true,
                confirmButtonColor: '#3085d6',
                cancelButtonColor: '#d33',
                confirmButtonText: 'Yes, log me out',
                cancelButtonText: 'Cancel'
            }).then((result) => {
                if (result.isConfirmed) {
                    window.location.href = '/logout'; // Выполнить переход
                } else {
                    console.log('User canceled logout');
                }
            });
        });
    }
});
// Получаем элементы полей ввода
const breedInput = document.getElementById('breed-search');
const colorInput = document.getElementById('color-search');
const ageYearsInput = document.getElementById('age-years-search');
const ageMonthsInput = document.getElementById('age-months-search');
const genderSelect = document.getElementById('gender-search');

// Функция фильтрации
function applyFilters() {
    const breed = breedInput.value;
    const color = colorInput.value;
    const ageYears = ageYearsInput.value;
    const ageMonths = ageMonthsInput.value;
    const gender = genderSelect.value;

    const url = new URL(window.location.href);

    // Добавляем или удаляем параметры фильтрации
    if (breed) url.searchParams.set('breed', breed);
    else url.searchParams.delete('breed');

    if (color) url.searchParams.set('color', color);
    else url.searchParams.delete('color');

    // Добавляем возраст (годы и месяцы)
    if (ageYears) url.searchParams.set('age_years', ageYears);
    else url.searchParams.delete('age_years');

    if (ageMonths) url.searchParams.set('age_months', ageMonths);
    else url.searchParams.delete('age_months');

    if (gender) url.searchParams.set('gender', gender);
    else url.searchParams.delete('gender');

    window.location.href = url.toString(); // Перенаправляем на новый URL
}


document.addEventListener("DOMContentLoaded", function() {
    const navbar = document.querySelector('nav');
    let lastScrollTop = 0;

    window.addEventListener('scroll', function() {
        let currentScrollTop = window.pageYOffset || document.documentElement.scrollTop;

        if (currentScrollTop > lastScrollTop) {
            // Scrolling down
            navbar.classList.add('nav-hidden');
        } else {
            // Scrolling up
            navbar.classList.remove('nav-hidden');
        }

        lastScrollTop = currentScrollTop <= 0 ? 0 : currentScrollTop; // For Mobile or negative scrolling
    });
    const urlParams = new URLSearchParams(window.location.search);

    // Устанавливаем значения полей из URL-параметров
    breedInput.value = urlParams.get('breed') || '';
    colorInput.value = urlParams.get('color') || '';
    ageYearsInput.value = urlParams.get('age_years') || '';
    ageMonthsInput.value = urlParams.get('age_months') || '';
    genderSelect.value = urlParams.get('gender') || '';
});
function filterAnimals(species) {
    console.log('Filtering animals by species:', species); // Для отладки
    const url = new URL(window.location.href);
    url.searchParams.set('species', species);
    console.log('Redirecting to URL:', url.toString()); // Для отладки
    window.location.href = url.toString();
}

