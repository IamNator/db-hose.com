function showView(viewId) {
    const views = document.querySelectorAll('.view');
    views.forEach(view => view.style.display = 'none');
    document.getElementById(viewId).style.display = 'block';
}

const BASE_URL = 'http://localhost:8080';

async function postData(url = '', data = {}) {
    const response = await fetch(BASE_URL + url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${sessionStorage.getItem('token')}`
        },
        body: JSON.stringify(data)
    });
    return response.json();
}

async function putData(url = '', data = {}) {
    const response = await fetch(BASE_URL + url, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${sessionStorage.getItem('token')}`
        },
        body: JSON.stringify(data)
    });
    return response.json();
}

async function deleteData(url = '') {
    const response = await fetch(BASE_URL + url, {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${sessionStorage.getItem('token')}`
        }
    });
    return response.json();
}

async function getData(url = '') {
    const response = await fetch(BASE_URL + url, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${sessionStorage.getItem('token')}`
        }
    });
    return response.json();
}

// Attach event listeners to forms
document.getElementById('signup-form').addEventListener('submit', function(event) {
    event.preventDefault();
    const data = {
        email: event.target.email.value,
        password: event.target.password.value
    };
    postData('/signup', data).then(data => {
        alert(JSON.stringify(data));
    });
});

document.getElementById('login-form').addEventListener('submit', function(event) {
    event.preventDefault();
    const data = {
        email: event.target.email.value,
        password: event.target.password.value
    };
    postData('/login', data).then(data => {
        if (data.token) {
            sessionStorage.setItem('token', data.token);
        }
        alert(JSON.stringify(data));
    });
});

document.getElementById('logout-form').addEventListener('submit', function(event) {
    event.preventDefault();
    postData('/logout', {}).then(data => {
        sessionStorage.removeItem('token');
        alert(JSON.stringify(data));
    });
});

document.getElementById('delete-account-form').addEventListener('submit', function(event) {
    event.preventDefault();
    postData('/delete', {}).then(data => {
        sessionStorage.removeItem('token');
        alert(JSON.stringify(data));
    });
});

document.getElementById('change-password-form').addEventListener('submit', function(event) {
    event.preventDefault();
    const data = {
        old_password: event.target.old_password.value,
        new_password: event.target.new_password.value
    };
    postData('/change-password', data).then(data => {
        alert(JSON.stringify(data));
    });
});

document.getElementById('store-creds-form').addEventListener('submit', function(event) {
    event.preventDefault();
    const data = {
        email: event.target.email.value,
        key: event.target.key.value,
        user: event.target.user.value,
        password: event.target.password.value,
        host: event.target.host.value,
        port: event.target.port.value,
        dbname: event.target.dbname.value
    };
    postData('/creds/store', data).then(data => {
        alert(JSON.stringify(data));
    });
});

document.getElementById('edit-creds-form').addEventListener('submit', function(event) {
    event.preventDefault();
    const data = {
        email: event.target.email.value,
        key: event.target.key.value,
        user: event.target.user.value,
        password: event.target.password.value,
        host: event.target.host.value,
        port: event.target.port.value,
        dbname: event.target.dbname.value
    };
    putData('/creds/edit', data).then(data => {
        alert(JSON.stringify(data));
    });
});

document.getElementById('delete-creds-form').addEventListener('submit', function(event) {
    event.preventDefault();
    const email = event.target.email.value;
    deleteData(`/creds/delete/${email}`).then(data => {
        alert(JSON.stringify(data));
    });
});

document.getElementById('view-creds-form').addEventListener('submit', function(event) {
    event.preventDefault();
    const email = event.target.email.value;
    getData(`/creds/view/${email}`).then(data => {
        alert(JSON.stringify(data));
    });
});

document.getElementById('backup-form').addEventListener('submit', function(event) {
    event.preventDefault();
    postData('/backup', {}).then(data => {
        alert(JSON.stringify(data));
    });
});

document.getElementById('restore-form').addEventListener('submit', function(event) {
    event.preventDefault();
    postData('/restore', {}).then(data => {
        alert(JSON.stringify(data));
    });
});
