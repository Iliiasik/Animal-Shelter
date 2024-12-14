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
            didClose:()=>{
                window.location.href = "/";
            }
        });

        Toast.fire({
            icon: 'success',
            title: 'Signed in successfully'
        });
    }
});
document.addEventListener('DOMContentLoaded', function() {
    // Проверьте, если редирект с успешного логина
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
            didClose:()=>{
                window.location.href = "/";
            }
        });

        Toast.fire({
            icon: 'info',
            title: 'Logged out '
        });
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