// Инициализация Яндекс.Карты на странице контактов с метками из БД
ymaps.ready(function() {
  var mapEl = document.getElementById('contact-map');
  if (!mapEl) return;

  // Фиксируем размер в px до инициализации — иначе при масштабе Windows != 100%
  // Яндекс.Карты инициализируют внутренние слои с неверными размерами
  var mapWidth = mapEl.offsetWidth;
  mapEl.style.width = mapWidth + 'px';

  var map = new ymaps.Map('contact-map', {
    center: [59.938784, 30.315868],
    zoom: 9,
    controls: ['zoomControl', 'fullscreenControl']
  }, {
    balloonAutoPan: true
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

  var placemarks = [];
  points.forEach(function(p) {
    // Собираем тело балуна
    var body = '';
    if (p.description) {
      body += '<div style="margin-bottom:8px;color:#555;font-size:13px;">' + p.description + '</div>';
    }
    if (p.panorama_url) {
      body += '<a href="' + p.panorama_url + '" target="_blank" rel="noopener noreferrer"' +
        ' style="display:inline-block;padding:6px 14px;background:#1a73e8;color:#fff;' +
        'border-radius:6px;font-size:13px;text-decoration:none;font-family:sans-serif;">' +
        '&#9654; Панорама</a>';
    }

    var placemark = new ymaps.Placemark(
      [p.latitude, p.longitude],
      {
        balloonContentHeader: '<strong>' + p.title + '</strong>',
        balloonContentBody: body,
        hintContent: p.title
      },
      {
        preset: 'islands#blueCircleDotIcon'
      }
    );

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
