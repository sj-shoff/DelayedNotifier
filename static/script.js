class NotificationManager {
    constructor() {
        this.baseUrl = '/api/v1';
        this.currentFilter = 'all';
        this.init();
    }

    init() {
        this.bindEvents();
        this.loadNotifications();
        this.setMinDateTime();
    }

    bindEvents() {
        // Форма создания
        document.getElementById('createForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.createNotification();
        });

        // Кнопка обновления
        document.getElementById('refreshBtn').addEventListener('click', () => {
            this.loadNotifications();
        });

        // Фильтр статусов
        document.getElementById('statusFilter').addEventListener('change', (e) => {
            this.currentFilter = e.target.value;
            this.loadNotifications();
        });

        // Модальное окно
        document.querySelector('.close').addEventListener('click', () => {
            this.hideModal();
        });

        document.getElementById('detailsModal').addEventListener('click', (e) => {
            if (e.target.id === 'detailsModal') {
                this.hideModal();
            }
        });
    }

    setMinDateTime() {
        const now = new Date();
        now.setMinutes(now.getMinutes() + 1); // Минимум +1 минута от текущего времени
        document.getElementById('send_at').min = now.toISOString().slice(0, 16);
    }

    async createNotification() {
        const form = document.getElementById('createForm');
        const formData = new FormData(form);
        
        const notificationData = {
            user_id: formData.get('user_id'),
            channel: formData.get('channel'),
            message: formData.get('message'),
            send_at: new Date(formData.get('send_at')).toISOString()
        };

        try {
            const response = await fetch(`${this.baseUrl}/notify`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(notificationData)
            });

            if (!response.ok) {
                const error = await response.text();
                throw new Error(error);
            }

            const result = await response.json();
            this.showMessage('Уведомление успешно создано!', 'success');
            form.reset();
            this.loadNotifications();
            
        } catch (error) {
            this.showMessage(`Ошибка при создании уведомления: ${error.message}`, 'error');
        }
    }

    async loadNotifications() {
        this.showLoading(true);
        this.hideError();

        try {
            const response = await fetch(`${this.baseUrl}/notifications`);
            if (!response.ok) throw new Error('Ошибка загрузки уведомлений');
            
            const notifications = await response.json();
            this.displayNotifications(notifications);
            
        } catch (error) {
            this.showError('Не удалось загрузить уведомления');
        } finally {
            this.showLoading(false);
        }
    }

    displayNotifications(notifications) {
        const container = document.getElementById('notificationsList');
        
        if (!notifications || notifications.length === 0) {
            container.innerHTML = '<div class="notification-card">Уведомлений нет</div>';
            return;
        }

        // Фильтрация по статусу
        const filteredNotifications = this.currentFilter === 'all' 
            ? notifications 
            : notifications.filter(n => n.status === this.currentFilter);

        container.innerHTML = filteredNotifications.map(notification => `
            <div class="notification-card" data-id="${notification.id}">
                <div class="notification-header">
                    <div>
                        <div class="notification-id">ID: ${notification.id}</div>
                        <div class="notification-channel channel-${notification.channel}">
                            ${this.getChannelDisplayName(notification.channel)}
                        </div>
                    </div>
                    <div class="status status-${notification.status}">
                        ${this.getStatusDisplayName(notification.status)}
                    </div>
                </div>
                
                <div class="notification-body">
                    <div class="notification-message">${this.escapeHtml(notification.message)}</div>
                    <div class="notification-meta">
                        <div><strong>Получатель:</strong> ${this.escapeHtml(notification.user_id)}</div>
                        <div><strong>Отправка:</strong> ${this.formatDateTime(notification.send_at)}</div>
                        <div><strong>Попытки:</strong> ${notification.retries}</div>
                        <div><strong>Создано:</strong> ${this.formatDateTime(notification.created_at)}</div>
                    </div>
                </div>
                
                <div class="notification-actions">
                    <button class="btn btn-secondary" onclick="notificationManager.showDetails('${notification.id}')">
                        Детали
                    </button>
                    ${notification.status === 'pending' ? `
                        <button class="btn btn-danger" onclick="notificationManager.cancelNotification('${notification.id}')">
                            Отменить
                        </button>
                    ` : ''}
                </div>
            </div>
        `).join('');
    }

    async showDetails(notificationId) {
        try {
            const response = await fetch(`${this.baseUrl}/notify/${notificationId}`);
            if (!response.ok) throw new Error('Не удалось загрузить детали');
            
            const notification = await response.json();
            this.displayModal(notification);
            
        } catch (error) {
            this.showMessage(`Ошибка при загрузке деталей: ${error.message}`, 'error');
        }
    }

    displayModal(notification) {
        const modalContent = document.getElementById('modalContent');
        modalContent.innerHTML = `
            <div class="detail-item">
                <div class="detail-label">ID</div>
                <div class="detail-value">${notification.id}</div>
            </div>
            <div class="detail-item">
                <div class="detail-label">Получатель</div>
                <div class="detail-value">${this.escapeHtml(notification.user_id)}</div>
            </div>
            <div class="detail-item">
                <div class="detail-label">Канал</div>
                <div class="detail-value">${this.getChannelDisplayName(notification.channel)}</div>
            </div>
            <div class="detail-item">
                <div class="detail-label">Сообщение</div>
                <div class="detail-value">${this.escapeHtml(notification.message)}</div>
            </div>
            <div class="detail-item">
                <div class="detail-label">Статус</div>
                <div class="detail-value status status-${notification.status}">
                    ${this.getStatusDisplayName(notification.status)}
                </div>
            </div>
            <div class="detail-item">
                <div class="detail-label">Время отправки</div>
                <div class="detail-value">${this.formatDateTime(notification.send_at)}</div>
            </div>
            <div class="detail-item">
                <div class="detail-label">Попытки отправки</div>
                <div class="detail-value">${notification.retries}</div>
            </div>
            <div class="detail-item">
                <div class="detail-label">Создано</div>
                <div class="detail-value">${this.formatDateTime(notification.created_at)}</div>
            </div>
            <div class="detail-item">
                <div class="detail-label">Обновлено</div>
                <div class="detail-value">${this.formatDateTime(notification.updated_at)}</div>
            </div>
        `;
        
        this.showModal();
    }

    async cancelNotification(notificationId) {
        if (!confirm('Вы уверены, что хотите отменить это уведомление?')) {
            return;
        }

        try {
            const response = await fetch(`${this.baseUrl}/notify/${notificationId}`, {
                method: 'DELETE'
            });

            if (!response.ok) {
                const error = await response.text();
                throw new Error(error);
            }

            this.showMessage('Уведомление успешно отменено', 'success');
            this.loadNotifications();
            
        } catch (error) {
            this.showMessage(`Ошибка при отмене уведомления: ${error.message}`, 'error');
        }
    }

    // Вспомогательные методы
    getChannelDisplayName(channel) {
        const channels = {
            'email': '📧 Email',
            'telegram': '📱 Telegram'
        };
        return channels[channel] || channel;
    }

    getStatusDisplayName(status) {
        const statuses = {
            'pending': '⏳ Ожидает',
            'sent': '✅ Отправлено',
            'cancelled': '❌ Отменено',
            'failed': '⚠️ Ошибка'
        };
        return statuses[status] || status;
    }

    formatDateTime(dateTimeString) {
        const date = new Date(dateTimeString);
        return date.toLocaleString('ru-RU', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit'
        });
    }

    escapeHtml(unsafe) {
        return unsafe
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;");
    }

    // UI методы
    showLoading(show) {
        document.getElementById('loading').classList.toggle('hidden', !show);
    }

    showError(message) {
        const errorDiv = document.getElementById('errorMessage');
        errorDiv.textContent = message;
        errorDiv.classList.remove('hidden');
    }

    hideError() {
        document.getElementById('errorMessage').classList.add('hidden');
    }

    showMessage(message, type) {
        const messageDiv = document.createElement('div');
        messageDiv.className = `${type === 'success' ? 'success-message' : 'error-message'}`;
        messageDiv.textContent = message;
        
        const container = document.querySelector('.container');
        container.insertBefore(messageDiv, container.firstChild);
        
        setTimeout(() => {
            messageDiv.remove();
        }, 5000);
    }

    showModal() {
        document.getElementById('detailsModal').classList.remove('hidden');
    }

    hideModal() {
        document.getElementById('detailsModal').classList.add('hidden');
    }
}

// Инициализация при загрузке страницы
let notificationManager;

document.addEventListener('DOMContentLoaded', () => {
    notificationManager = new NotificationManager();
    
    // Автообновление каждые 30 секунд
    setInterval(() => {
        notificationManager.loadNotifications();
    }, 30000);
});