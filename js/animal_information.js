document.addEventListener("DOMContentLoaded", function () {
    // Navbar Scroll Hide/Show
    const navbar = document.querySelector('nav');
    let lastScrollTop = 0;

    window.addEventListener('scroll', function () {
        let currentScrollTop = window.pageYOffset || document.documentElement.scrollTop;

        if (currentScrollTop > lastScrollTop) {
            navbar.classList.add('nav-hidden');
        } else {
            navbar.classList.remove('nav-hidden');
        }

        lastScrollTop = currentScrollTop <= 0 ? 0 : currentScrollTop;
    });

    // Slider Logic
    $('.slider').each(function () {
        var $this = $(this);
        var $group = $this.find('.slide_group');
        var $slides = $this.find('.slide');
        var bulletArray = [];
        var currentIndex = 0;
        var timeout;

        function move(newIndex) {
            var animateLeft, slideLeft;

            advance();

            if ($group.is(':animated') || currentIndex === newIndex) {
                return;
            }

            bulletArray[currentIndex].removeClass('active');
            bulletArray[newIndex].addClass('active');

            if (newIndex > currentIndex) {
                slideLeft = '100%';
                animateLeft = '-100%';
            } else {
                slideLeft = '-100%';
                animateLeft = '100%';
            }

            $slides.eq(newIndex).css({
                display: 'block',
                left: slideLeft
            });
            $group.animate({
                left: animateLeft
            }, function () {
                $slides.eq(currentIndex).css({ display: 'none' });
                $slides.eq(newIndex).css({ left: 0 });
                $group.css({ left: 0 });
                currentIndex = newIndex;
            });
        }

        function advance() {
            clearTimeout(timeout);
            timeout = setTimeout(function () {
                if (currentIndex < ($slides.length - 1)) {
                    move(currentIndex + 1);
                } else {
                    move(0);
                }
            }, 10000);
        }

        $('.next_btn').on('click', function () {
            if (currentIndex < ($slides.length - 1)) {
                move(currentIndex + 1);
            } else {
                move(0);
            }
        });

        $('.previous_btn').on('click', function () {
            if (currentIndex !== 0) {
                move(currentIndex - 1);
            } else {
                move($slides.length - 1);
            }
        });

        $.each($slides, function (index) {
            var $button = $('<a class="slide_btn">&bull;</a>');

            if (index === currentIndex) {
                $button.addClass('active');
            }
            $button.on('click', function () {
                move(index);
            }).appendTo('.slide_buttons');
            bulletArray.push($button);
        });

        advance();
    });

    // Weight Conversion
    const weightKg = parseFloat(document.getElementById('weight').dataset.kg);

    if (!isNaN(weightKg)) {
        const weightLb = weightKg * 2.20462;
        document.getElementById('weight').textContent = `${weightKg} kg (${weightLb.toFixed(2)} lb)`;
    }

    // Date Conversion
    const publicationDateElem = document.getElementById("publication-date");
    if (publicationDateElem) {
        const publicationDateStr = publicationDateElem.textContent.trim();
        const publicationDate = new Date(publicationDateStr);
        const now = new Date();
        const timeDiff = Math.floor((now - publicationDate) / (1000 * 60 * 60 * 24));
        let formattedDate; if (timeDiff === 0) {
            formattedDate = "Today";
        } else if (timeDiff === 1) {
            formattedDate = "Yesterday";
        } else if (timeDiff <= 6) {
            formattedDate = `${timeDiff
            } days ago`;
        } else {
            formattedDate = publicationDate.toISOString().split("T")[0];
        } publicationDateElem.textContent = formattedDate;
    }
});

// Increment Views after Delay
let timer;
const animalID = new URLSearchParams(window.location.search).get("id");

if (animalID) {
    timer = setTimeout(() => {
        fetch(`/increment_views?id=${animalID}`, { method: "POST" })
            .then(response => {
                if (!response.ok) {
                    console.error("Failed to increment views");
                }
            })
            .catch(error => console.error("Error:", error));
    }, 180000); // 3 minutes
}

window.addEventListener("beforeunload", () => clearTimeout(timer));
