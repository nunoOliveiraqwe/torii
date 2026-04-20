// ---------- Identity ----------
function loadIdentity() {
    fetch('/api/v1/auth/user', {credentials: 'same-origin'})
        .then(function (resp) {
            if (resp.status === 401) {
                window.location.href = '/ui/login';
                return null;
            }
            if (!resp.ok) throw new Error(resp.statusText);
            return resp.json();
        })
        .then(function (u) {
            if (!u) return;
            var name = u.username || '';
            document.getElementById('sidebar-username').textContent = name;
            document.getElementById('sidebar-avatar').textContent = name.charAt(0).toUpperCase();
        })
        .catch(function (err) {
            console.error('Failed to load identity', err);
        });
}

// ---------- Logout via API ----------
document.getElementById('btn-logout').addEventListener('click', function (e) {
    e.preventDefault();
    fetch('/api/v1/auth/logout', {method: 'POST', credentials: 'same-origin'})
        .then(function () {
            window.location.href = '/ui/login';
        })
        .catch(function () {
            window.location.href = '/ui/login';
        });
});

// ---------- Change Password ----------
document.getElementById('change-password-form').addEventListener('submit', function (e) {
    e.preventDefault();
    var resultEl = document.getElementById('password-result');
    var oldPwd = document.getElementById('old-password').value;
    var newPwd = document.getElementById('new-password').value;
    resultEl.innerHTML = '';
    fetch('/api/v1/auth/user', {
        method: 'POST', credentials: 'same-origin',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({oldPassword: oldPwd, newPassword: newPwd})
    })
        .then(function (resp) {
            if (resp.status === 401) {
                resultEl.innerHTML = '<div class="error-alert">Current password is incorrect.</div>';
                return;
            }
            if (!resp.ok) return resp.text().then(function (t) {
                resultEl.innerHTML = '<div class="error-alert">' + (t || 'Failed to change password.') + '</div>';
            });
            resultEl.innerHTML = '<div class="success-alert">Password updated successfully.</div>';
            document.getElementById('old-password').value = '';
            document.getElementById('new-password').value = '';
        })
        .catch(function () {
            resultEl.innerHTML = '<div class="error-alert">An error occurred. Please try again.</div>';
        });
});
