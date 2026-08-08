// Injected into every served HTML document. Adds a floating "Byakugan"
// button that opens a navigation drawer (project tree + search + prev/next)
// and subscribes to live-reload events. Deliberately self-contained: loads
// its own CSS and the shared search core, touches nothing in the host page.
(function () {
  'use strict';
  if (window.__byakugan) return; // idempotent if a page is served twice
  window.__byakugan = true;

  // Apply a saved explicit theme immediately so Byakugan UI paints right on
  // first frame. Only bk-scoped variables react to this attribute; the host
  // document's own colors and styles are untouched (see app.css).
  try {
    var saved = localStorage.getItem('bk-theme');
    if (saved === 'light' || saved === 'dark') {
      document.documentElement.setAttribute('data-bk-theme', saved);
    }
  } catch (_) { /* storage unavailable: stay on auto */ }

  function load(tag, attrs) {
    var el = document.createElement(tag);
    for (var k in attrs) el.setAttribute(k, attrs[k]);
    document.head.appendChild(el);
    return el;
  }
  load('link', { rel: 'stylesheet', href: '/__byakugan/app.css' });

  var script = load('script', { src: '/__byakugan/search.js' });
  script.onload = init;

  function init() {
    var current = decodeURIComponent(location.pathname).replace(/^\//, '');

    var fab = document.createElement('button');
    fab.id = 'bk-fab';
    fab.innerHTML = '<span class="bk-eye">👁️</span> Byakugan';
    fab.title = 'Navigation and search (b)';
    document.body.appendChild(fab);

    var drawer = document.createElement('div');
    drawer.id = 'bk-drawer';
    drawer.innerHTML =
      '<div class="bk-drawer-head"><a href="/"><span class="bk-eye">👁️</span> Byakugan</a>' +
      '<button class="bk-theme-btn" type="button"></button>' +
      '<button class="bk-drawer-close" title="Close (Esc)">×</button></div>' +
      '<div class="bk-drawer-search"><input type="search" placeholder="Search docs…" autocomplete="off"></div>' +
      '<div class="bk-tree"></div>' +
      '<div class="bk-drawer-foot">' +
      '<a id="bk-prev" aria-disabled="true">← Prev</a>' +
      '<a id="bk-next" aria-disabled="true">Next →</a></div>';
    document.body.appendChild(drawer);

    var $input = drawer.querySelector('input');
    var $tree = drawer.querySelector('.bk-tree');
    var $prev = drawer.querySelector('#bk-prev');
    var $next = drawer.querySelector('#bk-next');
    var index = { projects: [] };

    function open() { drawer.classList.add('bk-open'); $input.focus(); }
    function close() { drawer.classList.remove('bk-open'); }
    fab.addEventListener('click', open);
    drawer.querySelector('.bk-drawer-close').addEventListener('click', close);
    BK.themeInit(drawer.querySelector('.bk-theme-btn'));
    document.addEventListener('keydown', function (e) {
      var typing = /^(input|textarea|select)$/i.test((document.activeElement || {}).tagName || '');
      if (e.key === 'Escape') close();
      else if (e.key === 'b' && !typing && !e.metaKey && !e.ctrlKey && !e.altKey) {
        e.preventDefault();
        drawer.classList.contains('bk-open') ? close() : open();
      }
    });

    function renderTree(filter) {
      var html = '';
      (index.projects || []).forEach(function (proj) {
        var pages = filter ? BK.search({ projects: [proj] }, filter, 100) : proj.pages;
        if (!pages.length) return;
        html += '<div class="bk-tree-project"><span>' +
          (proj.name === '' ? 'Top level' : BK.highlight(proj.name, filter || '')) +
          '</span><span class="bk-tree-count">' + pages.length + '</span></div>';
        pages.forEach(function (p) {
          html += '<a href="/' + encodeURI(p.path) + '"' +
            (p.path === current ? ' class="bk-current"' : '') + '>' +
            BK.highlight(p.title, filter || '') + '</a>';
        });
      });
      $tree.innerHTML = html || '<div class="bk-tree-project">No matches</div>';
    }

    function renderPrevNext() {
      var flat = [];
      (index.projects || []).forEach(function (proj) {
        proj.pages.forEach(function (p) { flat.push(p); });
      });
      var at = flat.findIndex(function (p) { return p.path === current; });
      if (at > 0) {
        $prev.href = '/' + encodeURI(flat[at - 1].path);
        $prev.removeAttribute('aria-disabled');
        $prev.textContent = '← ' + flat[at - 1].title.slice(0, 22);
      }
      if (at >= 0 && at < flat.length - 1) {
        $next.href = '/' + encodeURI(flat[at + 1].path);
        $next.removeAttribute('aria-disabled');
        $next.textContent = flat[at + 1].title.slice(0, 22) + ' →';
      }
    }

    $input.addEventListener('input', function () { renderTree($input.value.trim()); });

    fetch('/api/index.json')
      .then(function (r) { return r.json(); })
      .then(function (data) { index = data; renderTree(''); renderPrevNext(); });

    try {
      new EventSource('/events').addEventListener('reload', function () { location.reload(); });
    } catch (_) { /* SSE unsupported: manual refresh still works */ }
  }
})();
