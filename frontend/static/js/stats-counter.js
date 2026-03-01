// Анимация счётчиков в секции статистики
// Запускается один раз при попадании .stats-section в viewport
document.addEventListener('DOMContentLoaded', function() {
  var section = document.querySelector('.stats-section');
  if (!section) return;

  function animateCounters() {
    section.querySelectorAll('.stat-count').forEach(function(el) {
      var target = parseInt(el.dataset.count, 10);
      if (!target) return;
      var duration = 1500;
      var start = null;
      function step(ts) {
        if (!start) start = ts;
        var progress = Math.min((ts - start) / duration, 1);
        var ease = 1 - (1 - progress) * (1 - progress); // easeOutQuad
        el.textContent = Math.floor(ease * target);
        if (progress < 1) requestAnimationFrame(step);
        else el.textContent = target;
      }
      requestAnimationFrame(step);
    });
  }

  if (!('IntersectionObserver' in window)) {
    animateCounters();
    return;
  }

  var observer = new IntersectionObserver(function(entries) {
    if (entries[0].isIntersecting) {
      animateCounters();
      observer.unobserve(section);
    }
  }, { threshold: 0.3 });

  observer.observe(section);
});
