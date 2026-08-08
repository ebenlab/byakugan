// Landing page: renders the project grid from /api/index.json, runs live
// search, and reloads on server push. When served for a directory URL
// (anything other than "/"), it scopes itself to that project.
(function () {
  'use strict';

  var $search = document.getElementById('bk-search');
  var $results = document.getElementById('bk-results');
  var $projects = document.getElementById('bk-projects');
  var $empty = document.getElementById('bk-empty');
  var $meta = document.getElementById('bk-meta');

  // "/payments/" -> "payments"; "/" -> null (show everything).
  var scope = decodeURIComponent(location.pathname).replace(/^\/|\/$/g, '') || null;
  var index = { projects: [] };

  function projectLabel(name) { return name === '' ? 'Top-level documents' : name; }

  function render() {
    var projects = index.projects || [];
    if (scope) {
      projects = projects.filter(function (p) { return p.name === scope; });
      document.title = scope + ' · Byakugan';
    }
    $meta.textContent = index.pageCount
      ? index.pageCount + ' page' + (index.pageCount === 1 ? '' : 's') + ' · ' +
        (index.projects || []).length + ' project' + (index.projects.length === 1 ? '' : 's')
      : '';
    $projects.innerHTML = '';
    $empty.hidden = projects.length > 0;

    projects.forEach(function (proj) {
      var card = document.createElement('div');
      card.className = 'bk-card';
      var shown = proj.pages.slice(0, 8);
      var ago = BK.timeAgo(proj.updatedAt);
      card.innerHTML =
        '<h2><a href="/' + encodeURI(proj.name) + (proj.name ? '/' : '') + '">' +
        BK.highlight(projectLabel(proj.name), '') + '</a>' +
        '<span class="bk-count">' + proj.pages.length + ' page' + (proj.pages.length === 1 ? '' : 's') + '</span></h2>' +
        (ago ? '<div class="bk-card-updated">Updated ' + ago + '</div>' : '') +
        '<ul>' + shown.map(function (p) {
          return '<li><a href="/' + encodeURI(p.path) + '">' + BK.highlight(p.title, '') + '</a></li>';
        }).join('') + '</ul>' +
        (proj.pages.length > shown.length
          ? '<span class="bk-more">+ ' + (proj.pages.length - shown.length) + ' more</span>' : '');
      $projects.appendChild(card);
    });
  }

  function renderSearch() {
    var q = $search.value.trim();
    if (!q) {
      $results.hidden = true;
      $projects.hidden = false;
      return;
    }
    var scoped = scope
      ? { projects: (index.projects || []).filter(function (p) { return p.name === scope; }) }
      : index;
    var hits = BK.search(scoped, q, 20);
    $results.innerHTML = hits.length
      ? hits.map(function (p) {
          return '<a class="bk-hit" href="/' + encodeURI(p.path) + '">' +
            '<div class="bk-hit-title">' + BK.highlight(p.title, q) + '</div>' +
            '<div class="bk-hit-path">' + BK.highlight(p.path, q) + '</div>' +
            '<div class="bk-hit-snippet">' + BK.highlight(BK.snippet(p, q), q) + '</div></a>';
        }).join('')
      : '<p class="bk-empty">No matches for “' + BK.highlight(q, '') + '”.</p>';
    $results.hidden = false;
    $projects.hidden = true;
  }

  BK.themeInit(document.getElementById('bk-theme'));

  $search.addEventListener('input', renderSearch);
  document.addEventListener('keydown', function (e) {
    if (e.key === '/' && document.activeElement !== $search) {
      e.preventDefault();
      $search.focus();
    }
    if (e.key === 'Enter' && document.activeElement === $search) {
      var first = $results.querySelector('.bk-hit');
      if (first) location.href = first.getAttribute('href');
    }
  });

  fetch('/api/index.json')
    .then(function (r) { return r.json(); })
    .then(function (data) { index = data; render(); renderSearch(); });

  try {
    new EventSource('/events').addEventListener('reload', function () { location.reload(); });
  } catch (_) { /* SSE unsupported: manual refresh still works */ }
})();
