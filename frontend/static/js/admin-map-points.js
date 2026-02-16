// Управление точками на карте в админке
document.addEventListener('DOMContentLoaded', function() {
  var createMap = null;
  var editMap = null;

  // Drag & drop сортировка
  var sortableList = document.getElementById('sortable-map-points');
  if (sortableList && typeof Sortable !== 'undefined') {
    Sortable.create(sortableList, {
      handle: '.drag-handle',
      animation: 150,
      onEnd: function() {
        var ids = [];
        sortableList.querySelectorAll('.project-item').forEach(function(item) {
          ids.push(parseInt(item.dataset.pointId));
        });
        fetch('/admin/map-points/sort', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ ids: ids })
        }).then(function(r) { return r.json(); })
          .then(function(data) {
            if (data.success) showAdminMessage('Порядок обновлён');
          });
      }
    });
  }

  // === Мини-карты в модалках ===

  function initMiniMap(containerId, latInput, lngInput) {
    return new Promise(function(resolve) {
      ymaps.ready(function() {
        var lat = parseFloat(latInput.value) || 59.938784;
        var lng = parseFloat(lngInput.value) || 30.315868;

        var map = new ymaps.Map(containerId, {
          center: [lat, lng],
          zoom: 10,
          controls: ['zoomControl', 'searchControl']
        });

        var placemark = new ymaps.Placemark([lat, lng], {}, {
          draggable: true,
          preset: 'islands#redDotIcon'
        });
        map.geoObjects.add(placemark);

        map.events.add('click', function(e) {
          var coords = e.get('coords');
          placemark.geometry.setCoordinates(coords);
          latInput.value = coords[0].toFixed(6);
          lngInput.value = coords[1].toFixed(6);
        });

        placemark.events.add('dragend', function() {
          var coords = placemark.geometry.getCoordinates();
          latInput.value = coords[0].toFixed(6);
          lngInput.value = coords[1].toFixed(6);
        });

        resolve({ map: map, placemark: placemark });
      });
    });
  }

  function setupMapModal(modalId, mapContainerId, latInputId, lngInputId, mapRef) {
    return function() {
      setTimeout(function() {
        var container = document.getElementById(mapContainerId);
        if (mapRef.map) { mapRef.map.destroy(); mapRef.map = null; }
        container.innerHTML = '';
        initMiniMap(mapContainerId,
          document.getElementById(latInputId),
          document.getElementById(lngInputId)
        ).then(function(result) {
          mapRef.map = result.map;
          mapRef.placemark = result.placemark;
        });
      }, 200);
    };
  }

  var createMapRef = { map: null, placemark: null };
  var editMapRef = { map: null, placemark: null };

  // === Создание точки ===

  document.getElementById('openCreateMapPointModal').addEventListener('click', function() {
    document.getElementById('createMapPointForm').reset();
    document.getElementById('createPointActive').checked = true;
    openModal('createMapPointModal');
    setupMapModal('createMapPointModal', 'createPointMap', 'createPointLat', 'createPointLng', createMapRef)();
  });

  document.getElementById('createMapPointForm').addEventListener('submit', function(e) {
    e.preventDefault();
    var form = new FormData(this);
    if (!document.getElementById('createPointActive').checked) {
      form.set('is_active', 'false');
    }
    fetch('/admin/map-points', { method: 'POST', body: form })
      .then(function(r) { return r.json(); })
      .then(function(data) {
        if (data.success) {
          showAdminMessage(data.message);
          closeModal('createMapPointModal');
          location.reload();
        } else {
          showAdminMessage(data.error || 'Ошибка', 'error');
        }
      })
      .catch(function() { showAdminMessage('Ошибка сети', 'error'); });
  });

  // === Редактирование точки ===

  window.editMapPoint = function(id) {
    fetch('/admin/map-points/' + id)
      .then(function(r) { return r.json(); })
      .then(function(data) {
        if (!data.success) { showAdminMessage(data.error || 'Ошибка', 'error'); return; }
        var p = data.map_point;
        document.getElementById('editPointId').value = p.id;
        document.getElementById('editPointTitle').value = p.title;
        document.getElementById('editPointDescription').value = p.description || '';
        document.getElementById('editPointLat').value = p.latitude;
        document.getElementById('editPointLng').value = p.longitude;
        document.getElementById('editPointPanorama').value = p.panorama_url || '';
        document.getElementById('editPointActive').checked = p.is_active;

        openModal('editMapPointModal');

        setTimeout(function() {
          var container = document.getElementById('editPointMap');
          if (editMapRef.map) { editMapRef.map.destroy(); editMapRef.map = null; }
          container.innerHTML = '';
          initMiniMap('editPointMap',
            document.getElementById('editPointLat'),
            document.getElementById('editPointLng')
          ).then(function(result) {
            editMapRef.map = result.map;
            editMapRef.placemark = result.placemark;
            editMapRef.map.setCenter([p.latitude, p.longitude], 14);
            editMapRef.placemark.geometry.setCoordinates([p.latitude, p.longitude]);
          });
        }, 200);
      });
  };

  document.getElementById('editMapPointForm').addEventListener('submit', function(e) {
    e.preventDefault();
    var id = document.getElementById('editPointId').value;
    var form = new FormData(this);
    if (!document.getElementById('editPointActive').checked) {
      form.set('is_active', 'false');
    }
    fetch('/admin/map-points/' + id + '/update', { method: 'POST', body: form })
      .then(function(r) { return r.json(); })
      .then(function(data) {
        if (data.success) {
          showAdminMessage(data.message);
          closeModal('editMapPointModal');
          location.reload();
        } else {
          showAdminMessage(data.error || 'Ошибка', 'error');
        }
      })
      .catch(function() { showAdminMessage('Ошибка сети', 'error'); });
  });

  // === Удаление точки ===

  window.deleteMapPoint = function(id, title) {
    if (!confirm('Удалить точку "' + title + '"?')) return;
    fetch('/admin/map-points/' + id, { method: 'DELETE' })
      .then(function(r) { return r.json(); })
      .then(function(data) {
        if (data.success) { showAdminMessage(data.message); location.reload(); }
        else { showAdminMessage(data.error || 'Ошибка', 'error'); }
      })
      .catch(function() { showAdminMessage('Ошибка сети', 'error'); });
  };

  // === Парсинг координат из ссылки Яндекс.Карт (клиентская сторона) ===

  function parseCoordsFromURL(urlStr) {
    try {
      var u = new URL(urlStr);
      var ll = u.searchParams.get('ll');
      if (!ll) return null;
      var parts = ll.split(',');
      if (parts.length !== 2) return null;
      // Яндекс: ll = долгота,широта
      return { lat: parseFloat(parts[1]), lng: parseFloat(parts[0]) };
    } catch (e) { return null; }
  }

  function extractTitleFromURL(urlStr) {
    try {
      var u = new URL(urlStr);
      var segments = u.pathname.split('/');
      for (var i = 0; i < segments.length; i++) {
        if (segments[i] === 'house' && i + 1 < segments.length) {
          return decodeURIComponent(segments[i + 1]).replace(/_/g, ' ');
        }
      }
    } catch (e) {}
    return '';
  }

  // === Импорт из одной ссылки ===

  document.getElementById('openImportLinkModal').addEventListener('click', function() {
    document.getElementById('importLinkForm').reset();
    openModal('importLinkModal');
  });

  // Авто-заполнение при вставке ссылки
  document.getElementById('importLinkUrl').addEventListener('input', function() {
    var url = this.value.trim();
    if (!url) return;
    var title = extractTitleFromURL(url);
    var titleInput = document.getElementById('importLinkTitle');
    if (title && !titleInput.value) {
      titleInput.value = title;
    }
  });

  document.getElementById('importLinkForm').addEventListener('submit', function(e) {
    e.preventDefault();
    var url = document.getElementById('importLinkUrl').value.trim();
    var title = document.getElementById('importLinkTitle').value.trim();
    var description = document.getElementById('importLinkDescription').value.trim();

    if (!url || !title) {
      showAdminMessage('Заполните ссылку и название', 'error');
      return;
    }

    var coords = parseCoordsFromURL(url);
    if (!coords) {
      showAdminMessage('Не удалось извлечь координаты из ссылки. Проверьте что в URL есть параметр ll=', 'error');
      return;
    }

    var form = new FormData();
    form.set('title', title);
    form.set('description', description);
    form.set('latitude', coords.lat.toString());
    form.set('longitude', coords.lng.toString());
    form.set('panorama_url', url);

    fetch('/admin/map-points', { method: 'POST', body: form })
      .then(function(r) { return r.json(); })
      .then(function(data) {
        if (data.success) {
          showAdminMessage(data.message);
          closeModal('importLinkModal');
          location.reload();
        } else {
          showAdminMessage(data.error || 'Ошибка', 'error');
        }
      })
      .catch(function() { showAdminMessage('Ошибка сети', 'error'); });
  });

  // === Массовый импорт ===

  document.getElementById('openBulkImportModal').addEventListener('click', function() {
    document.getElementById('bulkImportForm').reset();
    openModal('bulkImportModal');
  });

  document.getElementById('bulkImportForm').addEventListener('submit', function(e) {
    e.preventDefault();
    var text = document.getElementById('bulkImportLinks').value.trim();
    if (!text) {
      showAdminMessage('Вставьте ссылки', 'error');
      return;
    }

    var links = text.split('\n').map(function(l) { return l.trim(); }).filter(function(l) { return l.length > 0; });

    if (links.length === 0) {
      showAdminMessage('Нет ссылок для импорта', 'error');
      return;
    }

    fetch('/admin/map-points/bulk-import', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ links: links })
    })
      .then(function(r) { return r.json(); })
      .then(function(data) {
        if (data.success) {
          showAdminMessage(data.message);
          closeModal('bulkImportModal');
          if (data.created > 0) location.reload();
        } else {
          showAdminMessage(data.error || 'Ошибка', 'error');
        }
      })
      .catch(function() { showAdminMessage('Ошибка сети', 'error'); });
  });
});
