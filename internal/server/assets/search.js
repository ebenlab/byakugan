// Shared search core: scoring + rendering helpers, used by both the landing
// page (app.js) and the injected overlay (inject.js).
(function () {
  'use strict';

  // score ranks a page against space-separated query tokens. Every token
  // must appear somewhere; matches in the title weigh most, then headings,
  // then body text. Returns 0 when the page does not match.
  function score(page, tokens) {
    var title = (page.title || '').toLowerCase();
    var path = (page.path || '').toLowerCase();
    var headings = (page.headings || []).join(' ').toLowerCase();
    var text = (page.text || '').toLowerCase();
    var total = 0;
    for (var i = 0; i < tokens.length; i++) {
      var t = tokens[i], s = 0;
      if (title.indexOf(t) >= 0) s += title.indexOf(t) === 0 ? 8 : 5;
      if (path.indexOf(t) >= 0) s += 3;
      if (headings.indexOf(t) >= 0) s += 3;
      if (text.indexOf(t) >= 0) s += 1;
      if (s === 0) return 0;
      total += s;
    }
    return total;
  }

  // search returns the top `limit` pages of the index matching `query`.
  function search(index, query, limit) {
    var tokens = query.toLowerCase().split(/\s+/).filter(Boolean);
    if (!tokens.length) return [];
    var hits = [];
    (index.projects || []).forEach(function (proj) {
      (proj.pages || []).forEach(function (page) {
        var s = score(page, tokens);
        if (s > 0) hits.push({ page: page, score: s });
      });
    });
    hits.sort(function (a, b) { return b.score - a.score; });
    return hits.slice(0, limit || 20).map(function (h) { return h.page; });
  }

  // snippet extracts ~26 words of context around the first token match.
  function snippet(page, query) {
    var text = page.text || '';
    var token = query.toLowerCase().split(/\s+/).filter(Boolean)[0] || '';
    var at = text.toLowerCase().indexOf(token);
    if (at < 0) return text.slice(0, 160);
    var start = Math.max(0, at - 70);
    return (start > 0 ? '…' : '') + text.slice(start, at + 90) + '…';
  }

  // highlight wraps query tokens in <mark>, HTML-escaping everything else.
  function highlight(str, query) {
    var esc = String(str).replace(/[&<>"']/g, function (c) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
    });
    var tokens = query.toLowerCase().split(/\s+/).filter(Boolean)
      .map(function (t) { return t.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'); });
    if (!tokens.length) return esc;
    return esc.replace(new RegExp('(' + tokens.join('|') + ')', 'gi'), '<mark>$1</mark>');
  }

  // timeAgo renders an ISO timestamp as a short relative phrase ("3h ago").
  function timeAgo(iso) {
    var t = Date.parse(iso || '');
    if (isNaN(t)) return '';
    var s = Math.max(0, (Date.now() - t) / 1000);
    if (s < 60) return 'just now';
    if (s < 3600) return Math.floor(s / 60) + 'm ago';
    if (s < 86400) return Math.floor(s / 3600) + 'h ago';
    if (s < 86400 * 30) return Math.floor(s / 86400) + 'd ago';
    return new Date(t).toLocaleDateString();
  }

  // ---- Theme: three-state (auto → light → dark), persisted in localStorage.
  // An explicit choice is written as data-bk-theme on <html>; app.css scopes
  // every color variable to Byakugan roots, so host documents are unaffected.
  var THEME_KEY = 'bk-theme';
  var THEME_ORDER = ['auto', 'light', 'dark'];
  var THEME_GLYPH = { auto: '◐', light: '☀', dark: '☾' };
  var themeButtons = [];

  function themeGet() {
    try {
      var t = localStorage.getItem(THEME_KEY);
      return t === 'light' || t === 'dark' ? t : 'auto';
    } catch (_) { return 'auto'; }
  }

  function themeApply(t) {
    var el = document.documentElement;
    if (t === 'light' || t === 'dark') el.setAttribute('data-bk-theme', t);
    else el.removeAttribute('data-bk-theme');
  }

  function themeSet(t) {
    try { localStorage.setItem(THEME_KEY, t); } catch (_) { /* private mode */ }
    themeApply(t);
    themeButtons.forEach(paintThemeButton);
  }

  function paintThemeButton(btn) {
    var t = themeGet();
    btn.textContent = THEME_GLYPH[t];
    btn.setAttribute('data-bk-mode', t);
    btn.setAttribute('aria-label', 'Theme: ' + t);
    btn.title = 'Theme: ' + t + ' (click to change)';
  }

  // themeInit wires a toggle button and normalizes the attribute on <html>.
  function themeInit(btn) {
    themeApply(themeGet());
    if (!btn) return;
    themeButtons.push(btn);
    paintThemeButton(btn);
    btn.addEventListener('click', function () {
      var next = THEME_ORDER[(THEME_ORDER.indexOf(themeGet()) + 1) % THEME_ORDER.length];
      themeSet(next);
    });
  }

  window.BK = {
    search: search, snippet: snippet, highlight: highlight,
    timeAgo: timeAgo, themeInit: themeInit, themeGet: themeGet
  };
})();
