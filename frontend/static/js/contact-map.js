// Инициализация Яндекс.Карты на странице контактов с метками из БД
ymaps.ready(function() {
  var mapEl = document.getElementById('contact-map');
  if (!mapEl) return;

  var map = new ymaps.Map('contact-map', {
    center: [59.938784, 30.315868],
    zoom: 9,
    controls: ['zoomControl', 'fullscreenControl']
  });

  // Загружаем точки из data-атрибута
  var points = [];
  try {
    points = JSON.parse(mapEl.dataset.points || '[]');
  } catch (e) {
    points = [];
  }

  if (points.length === 0) return;

  // Добавляем метки
  var clusterer = new ymaps.Clusterer({
    preset: 'islands#invertedLightBlueClusterIcons',
    clusterDisableClickZoom: false,
    clusterBalloonContentLayout: 'cluster#balloonCarousel'
  });

  // SVG-иконка 360° для кнопки панорамы
  var panoramaIcon = '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="vertical-align:middle;margin-right:4px;">' +
    '<circle cx="12" cy="12" r="10"/>' +
    '<ellipse cx="12" cy="12" rx="10" ry="4"/>' +
    '<line x1="12" y1="2" x2="12" y2="22"/>' +
    '</svg>';

  var placemarks = [];
  points.forEach(function(p) {
    // Собираем тело балуна
    var body = '';
    if (p.description) {
      body += '<div style="margin-bottom:8px;color:#555;font-size:13px;">' + p.description + '</div>';
    }
    if (p.panorama_url) {
      body += '<a href="' + p.panorama_url + '" target="_blank" rel="noopener" ' +
        'class="map-panorama-btn" ' +
        'style="display:inline-flex;align-items:center;padding:6px 14px;' +
        'background:#1a73e8;color:#fff;border-radius:6px;font-size:13px;' +
        'text-decoration:none;cursor:pointer;position:relative;z-index:1000;">' +
        panoramaIcon + 'Панорама</a>';
    }

    var placemark = new ymaps.Placemark(
      [p.latitude, p.longitude],
      {
        balloonContentHeader: '<span style="font-weight:600;font-size:14px;">' + p.title + '</span>',
        balloonContentBody: body,
        hintContent: p.title
      },
      {
        preset: 'islands#blueCircleDotIcon'
      }
    );

    // Добавляем обработчики hover через события балуна
    placemark.events.add('balloonopen', function() {
      setTimeout(function() {
        var btn = document.querySelector('.map-panorama-btn');
        if (btn) {
          btn.addEventListener('mouseenter', function() {
            this.style.background = '#1557b0';
          });
          btn.addEventListener('mouseleave', function() {
            this.style.background = '#1a73e8';
          });
        }
      }, 100);
    });

    placemarks.push(placemark);
  });

  clusterer.add(placemarks);
  map.geoObjects.add(clusterer);

  // Автоматически подбираем масштаб чтобы все точки были видны
  if (placemarks.length > 1) {
    map.setBounds(clusterer.getBounds(), { checkZoomRange: true, zoomMargin: 40 });
  } else if (placemarks.length === 1) {
    map.setCenter([points[0].latitude, points[0].longitude], 14);
  }
});
