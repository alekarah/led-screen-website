(() => {
  const steps = document.querySelectorAll('.process-steps .step-number');
  if (!steps.length) return;

  // назначим каждому индекс для сдвига по времени
  steps.forEach((el, idx) => el.style.setProperty('--i', idx));

  const start = () => {
    steps.forEach(el => el.classList.add('is-lit'));
  };

  // запускаем, когда секция попадает в область видимости
  const io = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (!entry.isIntersecting) return;
      start();
      io.disconnect(); // чтобы не дёргать лишний раз
    });
  }, { root: null, threshold: 0.35 });

  // наблюдаем за контейнером с шагами (или за первой цифрой)
  const host = document.querySelector('.process-steps') || steps[0];
  if (host) io.observe(host);
})();