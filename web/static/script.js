async function createNotification(event) {
    event.preventDefault();
    const form = event.target;
    const data = {
        message: form.message.value,
        recipient: form.recipient.value,
        channel: form.channel.value,
        send_at: new Date(form.send_at.value).toISOString()
    };

    try {
        const res = await fetch('/notify', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });
        if (res.ok) {
            loadNotifications();
            form.reset();
        } else {
            alert('Error creating notification');
        }
    } catch (err) {
        alert('Error: ' + err);
    }
}

async function loadNotifications() {
    try {
        const res = await fetch('/notifications');
        const notifs = await res.json();
        const tbody = document.querySelector('#notifsTable tbody');
        tbody.innerHTML = '';
        notifs.forEach(notif => {
            const tr = document.createElement('tr');
            tr.innerHTML = `
                <td>${notif.id}</td>
                <td>${notif.message}</td>
                <td>${notif.recipient}</td>
                <td>${notif.channel}</td>
                <td>${notif.send_at}</td>
                <td>${notif.status}</td>
                <td><button onclick="cancelNotification('${notif.id}')">Cancel</button></td>
            `;
            tbody.appendChild(tr);
        });
    } catch (err) {
        console.error('Error loading notifications', err);
    }
}

async function cancelNotification(id) {
    try {
        const res = await fetch('/notify/' + id, { method: 'DELETE' });
        if (res.ok) {
            loadNotifications();
        } else {
            alert('Error cancelling notification');
        }
    } catch (err) {
        alert('Error: ' + err);
    }
}