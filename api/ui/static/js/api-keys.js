// ---------- API Keys ----------
function loadApiKeys() {
    fetch('/api/v1/apiKeys', {credentials: 'same-origin'})
        .then(function (resp) {
            if (resp.status === 401) {
                window.location.href = '/ui/login';
                return null;
            }
            if (!resp.ok) throw new Error(resp.statusText);
            return resp.json();
        })
        .then(function (keys) {
            if (keys === null) return;
            var tbody = document.getElementById('api-keys-body');
            if (!keys || !keys.length) {
                tbody.innerHTML = '<tr><td colspan="5" class="text-center" style="padding:1.5rem;color:var(--pico-muted-color);">No API keys yet.</td></tr>';
                return;
            }
            var html = '';
            keys.forEach(function (k) {
                var scopeList = k.scopes ? Object.keys(k.scopes) : [];
                var scopeBadges = scopeList.length
                    ? scopeList.map(function (s) {
                        return '<span class="badge badge-info" style="margin-right:0.25rem;">' + s + '</span>';
                    }).join('')
                    : '<span style="color:var(--pico-muted-color);font-size:0.8rem;">none</span>';
                var expiresStr = '–';
                if (k.expires && k.expires !== '0001-01-01T00:00:00Z') {
                    var exp = new Date(k.expires);
                    var now = new Date();
                    var daysLeft = Math.ceil((exp - now) / (1000 * 60 * 60 * 24));
                    if (daysLeft < 0) {
                        expiresStr = '<span class="badge badge-danger">Expired</span>';
                    } else {
                        var badgeCls = daysLeft <= 7 ? 'badge-danger' : daysLeft <= 30 ? 'badge-info' : 'badge-success';
                        expiresStr = '<span class="badge ' + badgeCls + '">' + daysLeft + 'd</span> ' + exp.toLocaleDateString();
                    }
                }
                var created = k.created_at ? new Date(k.created_at * 1000).toLocaleDateString() : '–';
                html += '<tr>' +
                    '<td><code>' + k.alias + '</code></td>' +
                    '<td>' + scopeBadges + '</td>' +
                    '<td>' + expiresStr + '</td>' +
                    '<td>' + created + '</td>' +
                    '<td><button class="badge badge-danger proxy-action-btn api-key-delete-btn" data-alias="' + k.alias + '">Delete</button></td>' +
                    '</tr>';
            });
            tbody.innerHTML = html;
            tbody.querySelectorAll('.api-key-delete-btn').forEach(function (btn) {
                btn.addEventListener('click', function () {
                    var alias = this.getAttribute('data-alias');
                    if (!confirm('Delete API key "' + alias + '"?')) return;
                    deleteApiKey(alias);
                });
            });
        })
        .catch(function (err) {
            console.error('Failed to load API keys', err);
            document.getElementById('api-keys-body').innerHTML =
                '<tr><td colspan="5" class="text-center" style="padding:1.5rem;color:var(--pico-muted-color);">Failed to load API keys.</td></tr>';
        });
}

function deleteApiKey(alias) {
    fetch('/api/v1/apiKeys/' + encodeURIComponent(alias), {method: 'DELETE', credentials: 'same-origin'})
        .then(function (resp) {
            if (resp.status === 401) {
                window.location.href = '/ui/login';
                return;
            }
            if (!resp.ok) return resp.text().then(function (t) {
                showToast(t || 'Failed to delete API key', 'error');
            });
            showToast('API key "' + alias + '" deleted.', 'success');
            loadApiKeys();
        })
        .catch(function () {
            showToast('Failed to delete API key.', 'error');
        });
}

document.getElementById('api-key-create-form').addEventListener('submit', function (e) {
    e.preventDefault();
    var resultEl = document.getElementById('api-key-create-result');
    resultEl.innerHTML = '';
    document.getElementById('api-key-created-banner').style.display = 'none';

    var alias = document.getElementById('api-key-alias').value.trim();
    var scopes = [];
    document.querySelectorAll('input[name="api-key-scope"]:checked').forEach(function (cb) {
        scopes.push(cb.value);
    });

    if (!alias) {
        resultEl.innerHTML = '<div class="error-alert">Alias is required.</div>';
        return;
    }
    if (scopes.length === 0) {
        resultEl.innerHTML = '<div class="error-alert">Select at least one scope.</div>';
        return;
    }

    var body = {alias: alias, scopes: scopes};
    var expiryVal = document.getElementById('api-key-expiry').value;
    if (expiryVal) {
        var d = new Date(expiryVal + 'T23:59:59');
        body.ExpiryDate = d.toISOString();
    }

    fetch('/api/v1/apiKeys', {
        method: 'POST', credentials: 'same-origin',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(body)
    })
        .then(function (resp) {
            if (resp.status === 401) {
                window.location.href = '/ui/login';
                return null;
            }
            if (resp.status === 409) {
                resultEl.innerHTML = '<div class="error-alert">An API key with this alias already exists.</div>';
                return null;
            }
            if (!resp.ok) return resp.text().then(function (t) {
                resultEl.innerHTML = '<div class="error-alert">' + (t || 'Failed to create API key.') + '</div>';
                return null;
            });
            return resp.json();
        })
        .then(function (created) {
            if (!created) return;
            // Show the raw key once
            document.getElementById('api-key-created-value').textContent = created.key;
            document.getElementById('api-key-created-banner').style.display = '';
            // Reset form
            document.getElementById('api-key-alias').value = '';
            document.getElementById('api-key-expiry').value = '';
            document.querySelectorAll('input[name="api-key-scope"]').forEach(function (cb) {
                cb.checked = false;
            });
            showToast('API key "' + created.alias + '" created.', 'success');
            loadApiKeys();
        })
        .catch(function () {
            resultEl.innerHTML = '<div class="error-alert">An error occurred. Please try again.</div>';
        });
});

document.getElementById('api-key-copy-btn').addEventListener('click', function () {
    var val = document.getElementById('api-key-created-value').textContent;
    navigator.clipboard.writeText(val).then(function () {
        showToast('Copied to clipboard.', 'success');
    }).catch(function () {
        showToast('Failed to copy. Select and copy manually.', 'error');
    });
});
