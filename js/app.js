/**
 * InkRead - 墨水屏阅读器前端逻辑
 */

// ========== 全局状态 ==========
const state = {
    theme: localStorage.getItem('inkread-theme') || 'light',
    books: [],
    currentBook: null,
    currentPage: 1,
    totalPages: 1,
    fontSize: parseInt(localStorage.getItem('inkread-fontsize') || '18'),
    brightness: 100
};

// ========== API 基础路径 ==========
const API_BASE = '/api';

// ========== 工具函数 ==========
function $(selector) {
    return document.querySelector(selector);
}

function $$(selector) {
    return document.querySelectorAll(selector);
}

async function apiRequest(endpoint, options = {}) {
    const url = `${API_BASE}${endpoint}`;
    const config = {
        headers: {
            'Content-Type': 'application/json',
            ...options.headers
        },
        ...options
    };
    
    try {
        const response = await fetch(url, config);
        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.message || '请求失败');
        }
        
        return data;
    } catch (error) {
        console.error('API Error:', error);
        throw error;
    }
}

function showToast(message) {
    const toast = document.createElement('div');
    toast.className = 'toast';
    toast.textContent = message;
    toast.style.cssText = `
        position: fixed;
        bottom: 80px;
        left: 50%;
        transform: translateX(-50%);
        background: var(--accent);
        color: var(--bg-primary);
        padding: 12px 24px;
        border-radius: 8px;
        font-size: 0.9rem;
        z-index: 1000;
    `;
    document.body.appendChild(toast);
    setTimeout(() => toast.remove(), 2000);
}

// ========== 主题管理 ==========
function initTheme() {
    document.documentElement.setAttribute('data-theme', state.theme);
    updateThemeIcon();
}

function toggleTheme() {
    state.theme = state.theme === 'light' ? 'dark' : 'light';
    document.documentElement.setAttribute('data-theme', state.theme);
    localStorage.setItem('inkread-theme', state.theme);
    updateThemeIcon();
}

function updateThemeIcon() {
    const icon = state.theme === 'light' ? '🌓' : '🌙';
    $$('.theme-icon, #themeToggleReader').forEach(el => {
        el.textContent = el.id === 'themeToggleReader' ? `${icon} 切换主题` : icon;
    });
}

// ========== 书架页面 ==========
async function loadBooks() {
    try {
        const response = await apiRequest('/books');
        state.books = response.data?.books || response.data || [];
        renderBookshelf();
    } catch (error) {
        console.error('加载书籍失败:', error);
        state.books = [];
        renderBookshelf();
    }
}

function renderBookshelf() {
    const shelf = $('#bookshelf');
    if (!shelf) return;

    if (state.books.length === 0) {
        shelf.innerHTML = `
            <div class="empty-state">
                <div class="empty-icon">📚</div>
                <p class="empty-text">书架空空如也<br>快上传一本书开始阅读吧</p>
            </div>
        `;
        return;
    }

    shelf.innerHTML = state.books.map((book) => `
        <div class="book-card" data-id="${book.id}">
            <div class="book-cover">📖</div>
            <div class="book-title" title="${book.title}">${book.title}</div>
            <div class="book-author">${book.author || '未知作者'}</div>
        </div>
    `).join('');

    // 绑定点击事件
    shelf.querySelectorAll('.book-card').forEach(card => {
        card.addEventListener('click', () => {
            const bookId = card.dataset.id;
            const book = state.books.find(b => b.id == bookId);
            if (book) openBook(book);
        });
    });
}

function openBook(book) {
    state.currentBook = book;
    state.currentPage = 1;
    window.location.href = `reader.html?id=${book.id}`;
}

// ========== 文件上传 ==========
function initUpload() {
    const uploadArea = $('#uploadArea');
    const fileInput = $('#fileInput');
    
    if (!uploadArea || !fileInput) return;

    uploadArea.addEventListener('click', () => fileInput.click());
    
    uploadArea.addEventListener('dragover', (e) => {
        e.preventDefault();
        uploadArea.classList.add('drag-over');
    });
    
    uploadArea.addEventListener('dragleave', () => {
        uploadArea.classList.remove('drag-over');
    });
    
    uploadArea.addEventListener('drop', (e) => {
        e.preventDefault();
        uploadArea.classList.remove('drag-over');
        const files = e.dataTransfer.files;
        if (files.length > 0) {
            handleFileUpload(files[0]);
        }
    });

    fileInput.addEventListener('change', (e) => {
        if (e.target.files.length > 0) {
            handleFileUpload(e.target.files[0]);
        }
    });
}

async function handleFileUpload(file) {
    const validTypes = ['.epub', '.pdf', '.txt'];
    const ext = '.' + file.name.split('.').pop().toLowerCase();
    
    if (!validTypes.includes(ext)) {
        showToast('不支持的文件格式');
        return;
    }

    const formData = new FormData();
    formData.append('file', file);

    try {
        const response = await fetch(`${API_BASE}/books`, {
            method: 'POST',
            body: formData
        });

        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.message || '上传失败');
        }

        showToast('上传成功');
        loadBooks(); // 刷新书架
    } catch (error) {
        console.error('上传失败:', error);
        showToast(error.message || '上传失败');
    }
}

// ========== 阅读器页面 ==========
async function initReader() {
    const params = new URLSearchParams(window.location.search);
    const bookId = params.get('id');
    
    if (!bookId) {
        window.location.href = 'index.html';
        return;
    }

    try {
        // 获取书籍信息
        const bookResponse = await apiRequest(`/books/${bookId}`);
        state.currentBook = bookResponse.data;
        
        // 获取书籍内容
        const contentResponse = await apiRequest(`/books/${bookId}/content`);
        state.currentBook.content = contentResponse.data?.content || '';
        
        $('#readerTitle').textContent = state.currentBook.title;
        
        // 加载保存的字体大小
        const savedFontSize = localStorage.getItem('inkread-fontsize');
        if (savedFontSize) {
            state.fontSize = parseInt(savedFontSize);
            updateFontSize();
        }

        loadBookContent();
        initReaderControls();
        initAiSummary();
    } catch (error) {
        console.error('加载书籍失败:', error);
        showToast('加载书籍失败');
        window.location.href = 'index.html';
    }
}

function loadBookContent() {
    const content = $('#readerContent');
    if (!content) return;

    const text = state.currentBook.content || '暂无内容';
    const pages = chunkText(text, 800);
    state.totalPages = Math.max(1, pages.length);
    state.currentPage = 1;

    renderPage();
}

function chunkText(text, charsPerPage) {
    const paragraphs = text.split('\n\n');
    const pages = [];
    let currentPage = '';

    for (const para of paragraphs) {
        if (currentPage.length + para.length > charsPerPage && currentPage.length > 0) {
            pages.push(currentPage);
            currentPage = '';
        }
        currentPage += (currentPage ? '\n\n' : '') + para;
    }
    
    if (currentPage) {
        pages.push(currentPage);
    }

    return pages.length > 0 ? pages : [text];
}

function renderPage() {
    const content = $('#readerContent');
    const pageInfo = $('#pageInfo');
    const progressFill = $('#progressFill');
    
    if (!content) return;

    const text = state.currentBook.content || '';
    const pages = chunkText(text, 800);
    const currentText = pages[state.currentPage - 1] || pages[0] || '';

    content.innerHTML = currentText.split('\n\n').map(p => `<p>${p}</p>`).join('');
    
    if (pageInfo) {
        pageInfo.textContent = `第 ${state.currentPage} / ${state.totalPages} 页`;
    }
    
    if (progressFill) {
        const progress = (state.currentPage / state.totalPages) * 100;
        progressFill.style.width = `${progress}%`;
    }
}

function updateFontSize() {
    const content = $('#readerContent');
    if (content) {
        content.style.setProperty('--reader-font-size', `${state.fontSize}px`);
        content.style.fontSize = `${state.fontSize}px`;
    }
    
    const display = $('#fontSizeDisplay');
    if (display) {
        display.textContent = `${state.fontSize}px`;
    }
}

function initReaderControls() {
    const btnBack = $('#btnBack');
    const fontDecrease = $('#fontDecrease');
    const fontIncrease = $('#fontIncrease');
    const brightnessSlider = $('#brightnessSlider');
    const themeToggle = $('#themeToggleReader');

    if (btnBack) {
        btnBack.addEventListener('click', () => {
            window.location.href = 'index.html';
        });
    }

    if (fontDecrease) {
        fontDecrease.addEventListener('click', () => {
            if (state.fontSize > 12) {
                state.fontSize -= 2;
                localStorage.setItem('inkread-fontsize', state.fontSize);
                updateFontSize();
            }
        });
    }

    if (fontIncrease) {
        fontIncrease.addEventListener('click', () => {
            if (state.fontSize < 32) {
                state.fontSize += 2;
                localStorage.setItem('inkread-fontsize', state.fontSize);
                updateFontSize();
            }
        });
    }

    if (brightnessSlider) {
        brightnessSlider.addEventListener('input', (e) => {
            state.brightness = e.target.value;
            document.body.style.filter = `brightness(${state.brightness}%)`;
        });
    }

    if (themeToggle) {
        themeToggle.addEventListener('click', toggleTheme);
    }

    // 点击内容区域切换控制面板 (双击)
    const readerContent = $('#readerContent');
    const readerControls = $('#readerControls');
    
    if (readerContent && readerControls) {
        readerContent.addEventListener('click', (e) => {
            if (e.detail === 2) { // 双击
                readerControls.classList.toggle('show');
            }
        });
    }

    // 页面滑动
    let touchStartX = 0;
    let touchEndX = 0;
    
    if (readerContent) {
        readerContent.addEventListener('touchstart', (e) => {
            touchStartX = e.changedTouches[0].screenX;
        });
        
        readerContent.addEventListener('touchend', (e) => {
            touchEndX = e.changedTouches[0].screenX;
            handleSwipe();
        });
    }

    function handleSwipe() {
        const diff = touchStartX - touchEndX;
        if (Math.abs(diff) > 80) {
            if (diff > 0 && state.currentPage < state.totalPages) {
                state.currentPage++;
                renderPage();
            } else if (diff < 0 && state.currentPage > 1) {
                state.currentPage--;
                renderPage();
            }
        }
    }
}

// ========== AI 摘要 ==========
function initAiSummary() {
    const btnAiSummary = $('#btnAiSummary');
    const summaryModal = $('#summaryModal');
    const btnCloseSummary = $('#btnCloseSummary');

    if (btnAiSummary) {
        btnAiSummary.addEventListener('click', async () => {
            if (summaryModal) {
                summaryModal.classList.add('show');
                await fetchAiSummary();
            }
        });
    }

    if (btnCloseSummary) {
        btnCloseSummary.addEventListener('click', () => {
            if (summaryModal) {
                summaryModal.classList.remove('show');
            }
        });
    }

    if (summaryModal) {
        summaryModal.addEventListener('click', (e) => {
            if (e.target === summaryModal) {
                summaryModal.classList.remove('show');
            }
        });
    }
}

async function fetchAiSummary() {
    const summaryLoading = $('#summaryLoading');
    const summaryText = $('#summaryText');
    
    if (!summaryText) return;

    try {
        const response = await apiRequest('/ai/summarize', {
            method: 'POST',
            body: JSON.stringify({ book_id: state.currentBook.id })
        });
        
        if (summaryLoading) summaryLoading.style.display = 'none';
        summaryText.innerHTML = `<p>${response.data?.summary || '暂无摘要内容'}</p>`;
        
    } catch (error) {
        console.error('AI Summary Error:', error);
        
        // 降级处理
        if (summaryLoading) summaryLoading.style.display = 'none';
        const mockSummary = `
            <p><strong>📖 内容概要</strong></p>
            <p>这本书是《${state.currentBook.title || '未知书籍'}》，由${state.currentBook.author || '未知作者'}所著。</p>
            <p>本书内容丰富，涵盖了多个主题，适合阅读学习。</p>
            <p><em>（实际摘要将由AI生成）</em></p>
        `;
        summaryText.innerHTML = mockSummary;
    }
}

// ========== 底部导航 ==========
function initNavigation() {
    const navBtns = $$('.nav-btn');
    
    navBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            const view = btn.dataset.view;
            
            navBtns.forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            
            handleNavigation(view);
        });
    });
}

function handleNavigation(view) {
    switch (view) {
        case 'shelf':
            break;
        case 'summary':
            if (state.currentBook) {
                window.location.href = `reader.html?id=${state.currentBook.id}&view=summary`;
            }
            break;
        case 'settings':
            showSettings();
            break;
    }
}

function showSettings() {
    const settingsHtml = `
        <div class="summary-modal show" id="settingsModal">
            <div class="summary-content">
                <div class="summary-header">
                    <h3>⚙️ 设置</h3>
                    <button class="btn-close" onclick="this.closest('.summary-modal').remove()">×</button>
                </div>
                <div class="summary-body">
                    <div class="settings-section">
                        <div class="settings-item">
                            <span class="settings-label">暗色主题</span>
                            <label class="toggle-switch">
                                <input type="checkbox" ${state.theme === 'dark' ? 'checked' : ''} onchange="toggleTheme(); this.closest('.summary-modal').remove()">
                                <span class="toggle-slider"></span>
                            </label>
                        </div>
                        <div class="settings-item">
                            <span class="settings-label">版本信息</span>
                            <span style="color: var(--text-secondary)">v1.0.0</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `;
    
    document.body.insertAdjacentHTML('beforeend', settingsHtml);
}

// ========== 初始化 ==========
document.addEventListener('DOMContentLoaded', () => {
    // 初始化主题
    initTheme();
    
    // 主题切换按钮
    const themeToggle = $('#themeToggle');
    if (themeToggle) {
        themeToggle.addEventListener('click', toggleTheme);
    }

    // 根据页面类型初始化
    const isReaderPage = window.location.pathname.includes('reader.html');
    
    if (isReaderPage) {
        initReader();
    } else {
        loadBooks();
        initUpload();
        initNavigation();
    }
});
