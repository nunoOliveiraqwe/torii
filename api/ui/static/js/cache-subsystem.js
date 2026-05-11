// ---------- Cache Subsystem ----------
var cacheSubsystemInterval = null;
var latestCacheSubsystemSnapshots = null;

function loadCacheSubsystem() {
    fetch('/api/v1/cache/snapshots', {credentials: 'same-origin'})
        .then(function (r) {
            if (!r.ok) throw new Error('HTTP ' + r.status);
            return r.json();
        })
        .then(function (caches) {
            latestCacheSubsystemSnapshots = caches;
            renderCacheSubsystem(caches);
        })
        .catch(function (err) {
            var container = document.getElementById('cache-subsystem-container');
            if (!container) return;
            container.innerHTML = '<div class="error-empty">Failed to load cache data: ' + err.message + '</div>';
        });
}

function handleCacheSnapshots(caches) {
    latestCacheSubsystemSnapshots = caches;
    if (isCacheSubsystemPageActive()) {
        renderCacheSubsystem(caches);
    }
}

function renderLatestCacheSubsystem() {
    if (latestCacheSubsystemSnapshots) {
        renderCacheSubsystem(latestCacheSubsystemSnapshots);
        return;
    }
    loadCacheSubsystem();
}

function isCacheSubsystemPageActive() {
    var page = document.querySelector('.page[data-page="http-caches"]');
    return page && page.style.display !== 'none';
}

function renderCacheSubsystem(caches) {
    var container = document.getElementById('cache-subsystem-container');
    if (!container) return;
    if (!caches || caches.length === 0) {
        container.innerHTML = '<div class="error-empty">No caches registered</div>';
        return;
    }

    var html = '';
    for (var i = 0; i < caches.length; i++) {
        var c = caches[i];
        var usagePct = c.max_entries > 0 ? ((c.current_entries / c.max_entries) * 100).toFixed(1) : '0.0';
        var rates = c.rates || {};
        var hits = numValue(rates.hits, c.hits);
        var misses = numValue(rates.misses, c.misses);
        var insertions = numValue(rates.insertion_total, c.insertion_total);
        var m1Rate = numValue(rates.m1_rate, c.m1_rate);
        var totalReads = hits + misses;
        var hitPct = totalReads > 0 ? ((hits / totalReads) * 100).toFixed(1) : '0.0';
        var entries = c.entries || [];

        html += '<section class="cache-panel">';
        html += '<div class="cache-panel-header">';
        html += '<div>';
        html += '<h4>' + escapeHtml(c.name) + '</h4>';
        html += '<div class="cache-meta-row">';
        if (c.owner) html += '<span>' + escapeHtml(c.owner) + '</span>';
        if (c.purpose) html += '<span>' + escapeHtml(c.purpose) + '</span>';
        if (c.scope) html += '<span>' + escapeHtml(c.scope) + '</span>';
        html += '</div>';
        html += '</div>';
        html += '<span class="cache-entry-count">' + fmtNum(c.current_entries) + ' / ' + fmtNum(c.max_entries) + ' entries (' + usagePct + '%)</span>';
        html += '</div>';

        html += '<div class="cache-capacity-bar">';
        html += '<div class="cache-capacity-fill" style="width:' + Math.min(parseFloat(usagePct), 100) + '%;"></div>';
        html += '</div>';

        html += '<div class="cache-metric-grid">';
        html += '<div class="cache-metric"><strong>' + fmtNum(hits) + '</strong><span>Hits</span></div>';
        html += '<div class="cache-metric"><strong>' + fmtNum(misses) + '</strong><span>Misses</span></div>';
        html += '<div class="cache-metric"><strong>' + hitPct + '%</strong><span>Hit Rate</span></div>';
        html += '<div class="cache-metric"><strong>' + fmtRate(m1Rate) + '</strong><span>Insert/s 1m</span></div>';
        html += '<div class="cache-metric"><strong>' + fmtNum(insertions) + '</strong><span>Insertions</span></div>';
        html += '</div>';

        html += '<details open class="cache-entry-details">';
        html += '<summary>Entries (' + entries.length + ')</summary>';
        if (entries.length > 0) {
            html += '<div class="cache-entry-list">';
            for (var j = 0; j < entries.length; j++) {
                html += renderCacheEntry(entries[j]);
            }
            html += '</div>';
        } else {
            html += '<p class="cache-empty">No entries cached</p>';
        }
        html += '</details>';

        html += '</section>';
    }

    container.innerHTML = html;
}

function renderCacheEntry(entry) {
    var disposition = entry.disposition || 'unknown';
    var dispositionClass = disposition.replace(/[^a-z0-9_-]/gi, '').toLowerCase() || 'unknown';
    var html = '<article class="cache-entry">';
    html += '<div class="cache-entry-main">';
    html += '<code class="cache-key-tag">' + escapeHtml(entry.key || '') + '</code>';
    html += '<span class="cache-disposition cache-disposition-' + dispositionClass + '">' + escapeHtml(disposition) + '</span>';
    html += '</div>';
    if (entry.summary) {
        html += '<div class="cache-entry-summary">' + escapeHtml(entry.summary) + '</div>';
    }
    html += renderEntryFields(entry.fields);
    html += '</article>';
    return html;
}

function renderEntryFields(fields) {
    if (!fields) return '';
    var keys = Object.keys(fields).sort();
    if (keys.length === 0) return '';
    var html = '<dl class="cache-entry-fields">';
    for (var i = 0; i < keys.length; i++) {
        html += '<div><dt>' + escapeHtml(keys[i]) + '</dt><dd>' + escapeHtml(String(fields[keys[i]])) + '</dd></div>';
    }
    html += '</dl>';
    return html;
}

function numValue(primary, fallback) {
    if (typeof primary === 'number') return primary;
    if (typeof fallback === 'number') return fallback;
    return 0;
}

function fmtRate(value) {
    return value.toFixed(2);
}

function escapeHtml(str) {
    var div = document.createElement('div');
    div.appendChild(document.createTextNode(str));
    return div.innerHTML;
}
