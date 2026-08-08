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

  window.BK = { search: search, snippet: snippet, highlight: highlight };
})();
