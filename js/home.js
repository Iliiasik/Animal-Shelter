document.addEventListener("DOMContentLoaded", function() {
    const images = document.querySelectorAll('.carousel-image');
    let index = 0;

    function showNextImage() {
        images[index].style.opacity = 0; // Скрываем текущее изображение
        index = (index + 1) % images.length; // Переход к следующему изображению
        images[index].style.opacity = 1; // Показываем следующее изображение
    }

    // Показ следующего изображения сразу после загрузки
    showNextImage();

    setInterval(showNextImage, 6000); // Изменить изображение каждые 3 секунды
});
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
});

document.addEventListener('DOMContentLoaded', function() {
    // Проверьте, если редирект с успешного логина
    const params = new URLSearchParams(window.location.search);
    if (params.has('login') && params.get('login') === 'success') {
        const Toast = Swal.mixin({
            toast: true,
            position: 'top-end',
            showConfirmButton: false,
            timer: 3000,
            timerProgressBar: true,
            customClass: {
                container:'custom-toast-container'
            },
            didOpen: (toast) => {
                toast.addEventListener('mouseenter', Swal.stopTimer);
                toast.addEventListener('mouseleave', Swal.resumeTimer);
            },
        });

        Toast.fire({
            icon: 'success',
            title: 'Signed in successfully'
        });
        // Удалить параметр "login" из URL
        const url = new URL(window.location);
        url.searchParams.delete('login');
        window.history.replaceState({}, document.title, url);
    }
});
document.addEventListener('DOMContentLoaded', function() {
    // Проверьте, если редирект с успешного логаута
    const params = new URLSearchParams(window.location.search);
    if (params.has('logout') && params.get('logout') === 'success') {
        const Toast = Swal.mixin({
            toast: true,
            position: 'top-end',
            showConfirmButton: false,
            timer: 3000,
            timerProgressBar: true,
            customClass: {
                container:'custom-toast-container'
            },
            didOpen: (toast) => {
                toast.addEventListener('mouseenter', Swal.stopTimer);
                toast.addEventListener('mouseleave', Swal.resumeTimer);
            },
        });

        Toast.fire({
            icon: 'info',
            title: 'Logged out '
        });

        // Удалить параметр "logout" из URL
        const url = new URL(window.location);
        url.searchParams.delete('logout');
        window.history.replaceState({}, document.title, url);
    }
});
document.addEventListener('DOMContentLoaded', () => {
    const logoutLink = document.querySelector('a[href="/logout"]'); // Найти ссылку логаута

    if (logoutLink) {
        logoutLink.addEventListener('click', (event) => {
            event.preventDefault(); // Отключить стандартное поведение ссылки

            Swal.fire({
                title: 'Are you sure?',
                text: 'This will log out and remain at the current page.',
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

// Function to handle cookie acceptance
function acceptCookies() {
    document.getElementById('cookie-card').style.display = 'none';
    // You can set a cookie to remember that the user accepted the notice
    document.cookie = "cookieAccepted=true; max-age=31536000; path=/"; // Cookie lasts for 1 year
}

// Check if the cookie is already set
function checkCookie() {
    let cookies = document.cookie.split(';');
    let cookieAccepted = false;

    // Check if cookieAccepted exists
    cookies.forEach(cookie => {
        if (cookie.trim().startsWith("cookieAccepted=")) {
            cookieAccepted = true;
        }
    });

    // If the cookie is not set, show the cookie notice
    if (!cookieAccepted) {
        document.getElementById('cookie-card').style.display = 'block';
    }
}

// Call the checkCookie function when the page loads
window.onload = checkCookie;
