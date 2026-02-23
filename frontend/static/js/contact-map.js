// Передача точек в iframe карты через postMessage
(function() {
  var iframe = document.getElementById('contact-map-frame');
  if (!iframe) return;

  var points = [];
  try {
    points = JSON.parse(iframe.dataset.points || '[]');
  } catch (e) {
    points = [];
  }

  iframe.addEventListener('load', function() {
    iframe.contentWindow.postMessage({ type: 'map-points', points: points }, '*');
  });
})();
