/* Анимации для главной страницы: сменяющиеся слова в герое, въезд карточек */

(function () {
    // --- Сменяющиеся слова в hero ---
    var words = ['бизнеса', 'магазина', 'мероприятия', 'фасада'];
    var el = document.getElementById('heroRotatingWord');
    if (el) {
        var idx = 0;
        function nextWord() {
            el.classList.add('hero-rotating-word--out');
            setTimeout(function () {
                idx = (idx + 1) % words.length;
                el.textContent = words[idx];
                el.classList.remove('hero-rotating-word--out');
            }, 350);
        }
        setInterval(nextWord, 2800);
    }

    // --- Стаггер-анимация карточек услуг ---
    var cards = document.querySelectorAll('.anim-card');
    if (!cards.length) return;

    var observer = new IntersectionObserver(function (entries, obs) {
        entries.forEach(function (entry) {
            if (entry.isIntersecting) {
                entry.target.classList.add('anim-card--visible');
                obs.unobserve(entry.target);
            }
        });
    }, { threshold: 0.15 });

    cards.forEach(function (card) {
        observer.observe(card);
    });
})();
