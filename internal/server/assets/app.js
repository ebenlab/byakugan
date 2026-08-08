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
  var $title = document.getElementById('bk-title');
  var $lede = document.getElementById('bk-lede');
  var $foot = document.getElementById('bk-foot');

  // "/payments/" -> "payments"; "/" -> null (show everything).
  var scope = decodeURIComponent(location.pathname).replace(/^\/|\/$/g, '') || null;
  var index = { projects: [] };

  function projectLabel(name) { return name === '' ? 'Top-level documents' : name; }
  function plural(n, word) { return n + ' ' + word + (n === 1 ? '' : 's'); }

  // "payments/adr-001-postgres.html" -> "adr-001-postgres" (the TOC ref).
  function pageRef(p, projName) {
    var ref = p.path;
    if (projName && ref.indexOf(projName + '/') === 0) ref = ref.slice(projName.length + 1);
    return ref.replace(/\.html?$/i, '');
  }

  function render() {
    var all = index.projects || [];
    var projects = all;
    if (scope) {
      projects = all.filter(function (p) { return p.name === scope; });
      document.title = scope + ' · Byakugan';
    }

    // Document header: the folder (or project) is the sheet's title.
    var rootName = (index.root || '').replace(/[\\/]+$/, '').split(/[\\/]/).pop();
    $title.textContent = scope || rootName || 'Documents';
    if (index.pageCount) {
      var scopedPages = scope
        ? projects.reduce(function (n, p) { return n + p.pages.length; }, 0)
        : index.pageCount;
      $lede.textContent = scope
        ? plural(scopedPages, 'page') + ' in this project — indexed from ' +
          (rootName || 'this folder') + ' and reloaded on change.'
        : plural(index.pageCount, 'page') + ' across ' + plural(all.length, 'project') +
          ', indexed straight from the folder and reloaded on change.';
    }
    $meta.textContent = index.pageCount
      ? plural(index.pageCount, 'page') + ' · ' + plural(all.length, 'project')
      : '';
    if (index.root) {
      var scanned = BK.timeAgo(index.generatedAt);
      $foot.textContent = 'byakugan — ' + index.root +
        (scanned ? ' · scanned ' + scanned : '') + ' · watching for changes';
    }

    $projects.innerHTML = '';
    $empty.hidden = projects.length > 0;

    projects.forEach(function (proj) {
      var card = document.createElement('section');
      card.className = 'bk-card';
      var shown = proj.pages.slice(0, 8);
      var ago = BK.timeAgo(proj.updatedAt);
      card.innerHTML =
        '<h2><a href="/' + encodeURI(proj.name) + (proj.name ? '/' : '') + '">' +
        BK.highlight(projectLabel(proj.name), '') + '</a>' +
        '<span class="bk-count">' + plural(proj.pages.length, 'page') + '</span></h2>' +
        (ago ? '<div class="bk-card-updated">Updated ' + ago + '</div>' : '') +
        '<ol class="bk-toc">' + shown.map(function (p) {
          return '<li><a href="/' + encodeURI(p.path) + '">' +
            '<span class="bk-toc-title">' + BK.highlight(p.title, '') + '</span>' +
            '<i class="bk-leader"></i>' +
            '<span class="bk-toc-ref">' + BK.highlight(pageRef(p, proj.name), '') + '</span>' +
            '</a></li>';
        }).join('') + '</ol>' +
        (proj.pages.length > shown.length
          ? '<span class="bk-more">+ ' + (proj.pages.length - shown.length) + ' more</span>' : '');
      $projects.appendChild(card);
    });
  }

  // The query lives in the URL (?q=…) so the browser back button restores
  // search state: replaceState while typing keeps history clean, and a
  // return visit (back from a result, direct link, bfcache) re-reads it.
  function queryFromURL() {
    var m = /[?&]q=([^&]*)/.exec(location.search);
    return m ? decodeURIComponent(m[1].replace(/\+/g, ' ')) : '';
  }
  function syncURL(q) {
    var url = location.pathname + (q ? '?q=' + encodeURIComponent(q) : '');
    try { history.replaceState(history.state, '', url); } catch (_) { /* file: etc. */ }
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

  $search.value = queryFromURL();
  $search.addEventListener('input', function () {
    renderSearch();
    syncURL($search.value.trim());
  });
  // Restore state when the page is revived from bfcache or history moves.
  window.addEventListener('pageshow', function (e) {
    if (e.persisted) { $search.value = queryFromURL(); renderSearch(); }
  });
  window.addEventListener('popstate', function () {
    $search.value = queryFromURL();
    renderSearch();
  });
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
