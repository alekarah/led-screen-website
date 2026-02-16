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

  var placemarks = [];
  points.forEach(function(p) {
    var placemark = new ymaps.Placemark(
      [p.latitude, p.longitude],
      {
        balloonContentHeader: p.panorama_url
          ? '<a href="' + p.panorama_url + '" target="_blank" rel="noopener" style="color:#1a73e8;">' + p.title + '</a>'
          : p.title,
        balloonContentBody: (p.description || '') +
          (p.panorama_url ? '<br><a href="' + p.panorama_url + '" target="_blank" rel="noopener" style="color:#1a73e8;">Открыть панораму</a>' : ''),
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
