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
        case 'sources':
            showSourceModal();
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

// ========== 书源管理 ==========
let sources = [];

// ========== 净化规则管理 ==========
let cleanupRules = [];

async function loadCleanupRules() {
    try {
        const response = await apiRequest('/cleanup-rules');
        cleanupRules = response.data || [];
        renderRuleList();
    } catch (error) {
        console.error('加载净化规则失败:', error);
        cleanupRules = [];
        renderRuleList();
    }
}

function renderRuleList() {
    const list = $('#ruleList');
    if (!list) return;

    if (cleanupRules.length === 0) {
        list.innerHTML = '<div class="empty-rules">暂无净化规则<br>点击上方按钮添加规则</div>';
        return;
    }

    const typeLabels = {
        replace: '替换',
        remove: '删除',
        regex: '正则'
    };

    list.innerHTML = cleanupRules.map(rule => `
        <div class="rule-item" data-id="${rule.id}">
            <div class="rule-item-header">
                <span class="rule-item-name">${rule.name}</span>
                <span class="rule-item-type">${typeLabels[rule.ruleType] || rule.ruleType}</span>
            </div>
            <div class="rule-item-pattern">${escapeHtml(rule.pattern)}</div>
            <div class="rule-item-footer">
                <span class="rule-item-priority">优先级: ${rule.priority || 0}</span>
                <div class="rule-item-actions">
                    <label class="toggle-switch small">
                        <input type="checkbox" ${rule.enabled ? 'checked' : ''} onchange="toggleRule('${rule.id}', this.checked)">
                        <span class="toggle-slider"></span>
                    </label>
                    <button class="btn-small" onclick="editRule('${rule.id}')">编辑</button>
                    <button class="btn-small" onclick="deleteRule('${rule.id}')">删除</button>
                </div>
            </div>
        </div>
    `).join('');
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function showRuleForm(rule = null) {
    const modal = $('#ruleFormModal');
    const title = $('#ruleFormTitle');
    const form = $('#ruleForm');
    const idInput = $('#ruleId');
    const nameInput = $('#ruleName');
    const typeInput = $('#ruleType');
    const patternInput = $('#rulePattern');
    const replacementInput = $('#ruleReplacement');
    const priorityInput = $('#rulePriority');
    const enabledInput = $('#ruleEnabled');

    if (!modal) return;

    if (rule) {
        title.textContent = '编辑净化规则';
        idInput.value = rule.id || '';
        nameInput.value = rule.name || '';
        typeInput.value = rule.ruleType || 'replace';
        patternInput.value = rule.pattern || '';
        replacementInput.value = rule.replacement || '';
        priorityInput.value = rule.priority || 0;
        enabledInput.checked = rule.enabled !== false;
    } else {
        title.textContent = '添加净化规则';
        form.reset();
        idInput.value = '';
        priorityInput.value = 0;
        enabledInput.checked = true;
    }

    modal.classList.add('show');
}

function hideRuleForm() {
    const modal = $('#ruleFormModal');
    if (modal) modal.classList.remove('show');
}

window.editRule = function(id) {
    const rule = cleanupRules.find(r => r.id == id || r.id === id);
    if (rule) {
        showRuleForm(rule);
    }
};

async function saveRule(e) {
    e.preventDefault();

    const id = $('#ruleId').value;
    const ruleData = {
        name: $('#ruleName').value.trim(),
        ruleType: $('#ruleType').value,
        pattern: $('#rulePattern').value.trim(),
        replacement: $('#ruleReplacement').value,
        priority: parseInt($('#rulePriority').value) || 0,
        enabled: $('#ruleEnabled').checked
    };

    if (!ruleData.name || !ruleData.pattern) {
        showToast('请填写规则名称和匹配模式');
        return;
    }

    if (ruleData.ruleType === 'regex') {
        try {
            new RegExp(ruleData.pattern);
        } catch (e) {
            showToast('正则表达式格式不正确');
            return;
        }
    }

    try {
        if (id) {
            await apiRequest(`/cleanup-rules/${id}`, {
                method: 'PUT',
                body: JSON.stringify(ruleData)
            });
            showToast('更新成功');
        } else {
            await apiRequest('/cleanup-rules', {
                method: 'POST',
                body: JSON.stringify(ruleData)
            });
            showToast('添加成功');
        }
        hideRuleForm();
        loadCleanupRules();
    } catch (error) {
        console.error('保存规则失败:', error);
        showToast(error.message || '保存失败');
    }
}

window.deleteRule = async function(id) {
    if (!confirm('确定要删除这个规则吗？')) return;

    try {
        await apiRequest(`/cleanup-rules/${id}`, { method: 'DELETE' });
        showToast('删除成功');
        loadCleanupRules();
    } catch (error) {
        console.error('删除规则失败:', error);
        showToast(error.message || '删除失败');
    }
};

window.toggleRule = async function(id, enabled) {
    try {
        await apiRequest(`/cleanup-rules/${id}`, {
            method: 'PUT',
            body: JSON.stringify({ enabled: enabled })
        });
        showToast(enabled ? '规则已启用' : '规则已禁用');
        loadCleanupRules();
    } catch (error) {
        console.error('切换规则失败:', error);
        showToast(error.message || '操作失败');
        loadCleanupRules();
    }
};

function initCleanupRules() {
    const addRuleBtn = $('#addRuleBtn');
    const closeRuleForm = $('#closeRuleForm');
    const cancelRuleForm = $('#cancelRuleForm');
    const ruleForm = $('#ruleForm');

    if (addRuleBtn) {
        addRuleBtn.addEventListener('click', () => showRuleForm());
    }

    if (closeRuleForm) {
        closeRuleForm.addEventListener('click', hideRuleForm);
    }

    if (cancelRuleForm) {
        cancelRuleForm.addEventListener('click', hideRuleForm);
    }

    if (ruleForm) {
        ruleForm.addEventListener('submit', saveRule);
    }

    const ruleFormModal = $('#ruleFormModal');
    if (ruleFormModal) {
        ruleFormModal.addEventListener('click', (e) => {
            if (e.target === ruleFormModal) hideRuleForm();
        });
    }

    const tabs = $$('.source-tab');
    tabs.forEach(tab => {
        tab.addEventListener('click', () => {
            const tabName = tab.dataset.tab;
            tabs.forEach(t => t.classList.remove('active'));
            tab.classList.add('active');

            $('#tabSources').style.display = tabName === 'sources' ? 'block' : 'none';
            $('#tabRules').style.display = tabName === 'rules' ? 'block' : 'none';

            if (tabName === 'rules') {
                loadCleanupRules();
            }
        });
    });
}

async function loadSources() {
    try {
        const response = await apiRequest('/sources');
        sources = response.data || [];
        renderSourceList();
    } catch (error) {
        console.error('加载书源失败:', error);
        sources = [];
        renderSourceList();
    }
}

function renderSourceList() {
    const list = $('#sourceList');
    if (!list) return;

    if (sources.length === 0) {
        list.innerHTML = '<div class="empty-sources">暂无书源<br>点击上方按钮添加书源</div>';
        return;
    }

    list.innerHTML = sources.map(source => `
        <div class="source-item" data-id="${source.id}">
            <div class="source-item-header">
                <span class="source-item-name">${source.name}</span>
                <span class="source-item-status ${source.enabled ? 'enabled' : 'disabled'}">${source.enabled ? '启用' : '禁用'}</span>
            </div>
            <div class="source-item-url">${source.urlTemplate || ''}</div>
            <div class="source-item-actions">
                <button class="btn-small" onclick="editSource('${source.id}')">编辑</button>
                <button class="btn-small" onclick="testSource('${source.id}')">测试</button>
                <button class="btn-small" onclick="deleteSource('${source.id}')">删除</button>
            </div>
        </div>
    `).join('');
}

function showSourceModal() {
    const modal = $('#sourceModal');
    if (modal) {
        modal.classList.add('show');
        loadSources();
    }
}

function hideSourceModal() {
    const modal = $('#sourceModal');
    if (modal) modal.classList.remove('show');
}

function showSourceForm(source = null) {
    const modal = $('#sourceFormModal');
    const title = $('#sourceFormTitle');
    const form = $('#sourceForm');
    const idInput = $('#sourceId');
    const nameInput = $('#sourceName');
    const urlTemplateInput = $('#sourceUrlTemplate');
    const encodingInput = $('#sourceEncoding');
    const bookNameRuleInput = $('#bookNameRule');
    const authorRuleInput = $('#authorRule');
    const chapterListRuleInput = $('#chapterListRule');
    const chapterUrlRuleInput = $('#chapterUrlRule');
    const contentRuleInput = $('#contentRule');
    const enabledInput = $('#sourceEnabled');

    if (!modal) return;

    if (source) {
        title.textContent = '编辑书源';
        idInput.value = source.id || '';
        nameInput.value = source.name || '';
        urlTemplateInput.value = source.urlTemplate || '';
        encodingInput.value = source.encoding || 'utf-8';
        bookNameRuleInput.value = source.bookNameRule || '';
        authorRuleInput.value = source.authorRule || '';
        chapterListRuleInput.value = source.chapterListRule || '';
        chapterUrlRuleInput.value = source.chapterUrlRule || '';
        contentRuleInput.value = source.contentRule || '';
        enabledInput.checked = source.enabled !== false;
    } else {
        title.textContent = '添加书源';
        form.reset();
        idInput.value = '';
        enabledInput.checked = true;
    }

    modal.classList.add('show');
}

function hideSourceForm() {
    const modal = $('#sourceFormModal');
    if (modal) modal.classList.remove('show');
}

window.editSource = function(id) {
    const source = sources.find(s => s.id == id || s.id === id);
    if (source) {
        showSourceForm(source);
    }
};

async function saveSource(e) {
    e.preventDefault();
    
    const id = $('#sourceId').value;
    const sourceData = {
        name: $('#sourceName').value.trim(),
        urlTemplate: $('#sourceUrlTemplate').value.trim(),
        encoding: $('#sourceEncoding').value,
        bookNameRule: $('#bookNameRule').value.trim(),
        authorRule: $('#authorRule').value.trim(),
        chapterListRule: $('#chapterListRule').value.trim(),
        chapterUrlRule: $('#chapterUrlRule').value.trim(),
        contentRule: $('#contentRule').value.trim(),
        enabled: $('#sourceEnabled').checked
    };

    // 验证必填字段
    if (!sourceData.name || !sourceData.urlTemplate || !sourceData.bookNameRule || 
        !sourceData.chapterListRule || !sourceData.chapterUrlRule || !sourceData.contentRule) {
        showToast('请填写所有必填项');
        return;
    }

    try {
        if (id) {
            await apiRequest(`/sources/${id}`, {
                method: 'PUT',
                body: JSON.stringify(sourceData)
            });
            showToast('更新成功');
        } else {
            await apiRequest('/sources', {
                method: 'POST',
                body: JSON.stringify(sourceData)
            });
            showToast('添加成功');
        }
        hideSourceForm();
        loadSources();
    } catch (error) {
        console.error('保存书源失败:', error);
        showToast(error.message || '保存失败');
    }
}

window.deleteSource = async function(id) {
    if (!confirm('确定要删除这个书源吗？')) return;
    
    try {
        await apiRequest(`/sources/${id}`, { method: 'DELETE' });
        showToast('删除成功');
        loadSources();
    } catch (error) {
        console.error('删除书源失败:', error);
        showToast(error.message || '删除失败');
    }
};

function showSourceTestModal() {
    const modal = $('#sourceTestModal');
    if (modal) modal.classList.add('show');
}

function hideSourceTestModal() {
    const modal = $('#sourceTestModal');
    if (modal) modal.classList.remove('show');
}

window.testSource = async function(id) {
    showSourceTestModal();
    const body = $('#sourceTestBody');
    if (body) body.innerHTML = '<div class="loading">测试中...</div>';

    try {
        const response = await apiRequest(`/sources/test`, {
            method: 'POST',
            body: JSON.stringify({ sourceId: id })
        });
        
        if (body) {
            const data = response.data || {};
            body.innerHTML = `
                <div class="test-result-item ${data.success ? 'success' : 'error'}">
                    <div class="test-result-label">状态</div>
                    <div class="test-result-value">${data.success ? '✅ 测试成功' : '❌ 测试失败'}</div>
                </div>
                ${data.bookName ? `<div class="test-result-item success">
                    <div class="test-result-label">书名</div>
                    <div class="test-result-value">${data.bookName}</div>
                </div>` : ''}
                ${data.author ? `<div class="test-result-item success">
                    <div class="test-result-label">作者</div>
                    <div class="test-result-value">${data.author}</div>
                </div>` : ''}
                ${data.chapterCount ? `<div class="test-result-item success">
                    <div class="test-result-label">章节数</div>
                    <div class="test-result-value">${data.chapterCount}</div>
                </div>` : ''}
                ${data.error ? `<div class="test-result-item error">
                    <div class="test-result-label">错误信息</div>
                    <div class="test-result-value">${data.error}</div>
                </div>` : ''}
            `;
        }
    } catch (error) {
        console.error('测试书源失败:', error);
        if (body) {
            body.innerHTML = `
                <div class="test-result-item error">
                    <div class="test-result-label">错误</div>
                    <div class="test-result-value">${error.message || '测试失败'}</div>
                </div>
            `;
        }
    }
};

function initSourceManagement() {
    const addSourceBtn = $('#addSourceBtn');
    const closeSourceModal = $('#closeSourceModal');
    const closeSourceForm = $('#closeSourceForm');
    const closeSourceTest = $('#closeSourceTest');
    const cancelSourceForm = $('#cancelSourceForm');
    const sourceForm = $('#sourceForm');

    if (addSourceBtn) {
        addSourceBtn.addEventListener('click', () => showSourceForm());
    }

    if (closeSourceModal) {
        closeSourceModal.addEventListener('click', hideSourceModal);
    }

    if (closeSourceForm) {
        closeSourceForm.addEventListener('click', hideSourceForm);
    }

    if (closeSourceTest) {
        closeSourceTest.addEventListener('click', hideSourceTestModal);
    }

    if (cancelSourceForm) {
        cancelSourceForm.addEventListener('click', hideSourceForm);
    }

    if (sourceForm) {
        sourceForm.addEventListener('submit', saveSource);
    }

    // 点击模态框外部关闭
    const sourceModal = $('#sourceModal');
    const sourceFormModal = $('#sourceFormModal');
    const sourceTestModal = $('#sourceTestModal');

    if (sourceModal) {
        sourceModal.addEventListener('click', (e) => {
            if (e.target === sourceModal) hideSourceModal();
        });
    }

    if (sourceFormModal) {
        sourceFormModal.addEventListener('click', (e) => {
            if (e.target === sourceFormModal) hideSourceForm();
        });
    }

    if (sourceTestModal) {
        sourceTestModal.addEventListener('click', (e) => {
            if (e.target === sourceTestModal) hideSourceTestModal();
        });
    }
}

// ========== Web 导入 ==========
let webImportSources = [];

function showWebImportModal() {
    const modal = $('#webImportModal');
    const form = $('#webImportForm');
    const progress = $('#webImportProgress');
    
    if (modal) {
        modal.classList.add('show');
        form.style.display = 'block';
        progress.style.display = 'none';
        loadSourcesForImport();
    }
}

function hideWebImportModal() {
    const modal = $('#webImportModal');
    const importUrl = $('#importUrl');
    if (modal) modal.classList.remove('show');
    if (importUrl) importUrl.value = '';
}

async function loadSourcesForImport() {
    try {
        const response = await apiRequest('/sources');
        webImportSources = response.data || [];
        renderSourceSelect();
    } catch (error) {
        console.error('加载书源失败:', error);
        webImportSources = [];
        renderSourceSelect();
    }
}

function renderSourceSelect() {
    const select = $('#importSource');
    if (!select) return;

    const options = webImportSources
        .filter(s => s.enabled)
        .map(s => `<option value="${s.id}">${s.name}</option>`)
        .join('');

    select.innerHTML = `<option value="">-- 选择书源 (可选) --</option>${options}`;
}

async function handleWebImport(e) {
    e.preventDefault();
    
    const url = $('#importUrl').value.trim();
    const sourceId = $('#importSource').value;
    
    if (!url) {
        showToast('请输入书籍页面 URL');
        return;
    }

    const form = $('#webImportForm');
    const progress = $('#webImportProgress');
    const status = $('#importStatus');
    
    form.style.display = 'none';
    progress.style.display = 'block';
    status.textContent = '正在获取页面内容...';

    try {
        const requestData = { url: url };
        if (sourceId) {
            requestData.source_id = sourceId;
        }

        const response = await apiRequest('/import/url', {
            method: 'POST',
            body: JSON.stringify(requestData)
        });

        status.textContent = '正在解析内容...';
        
        if (response.data && response.data.id) {
            showToast('导入成功');
            hideWebImportModal();
            loadBooks();
        } else {
            throw new Error('导入结果异常');
        }
    } catch (error) {
        console.error('导入失败:', error);
        status.textContent = `导入失败: ${error.message}`;
        showToast(error.message || '导入失败');
        
        setTimeout(() => {
            form.style.display = 'block';
            progress.style.display = 'none';
        }, 2000);
    }
}

function initWebImport() {
    const btnWebImport = $('#btnWebImport');
    const closeBtn = $('#closeWebImport');
    const cancelBtn = $('#cancelWebImport');
    const form = $('#webImportForm');
    const modal = $('#webImportModal');

    if (btnWebImport) {
        btnWebImport.addEventListener('click', showWebImportModal);
    }

    if (closeBtn) {
        closeBtn.addEventListener('click', hideWebImportModal);
    }

    if (cancelBtn) {
        cancelBtn.addEventListener('click', hideWebImportModal);
    }

    if (form) {
        form.addEventListener('submit', handleWebImport);
    }

    if (modal) {
        modal.addEventListener('click', (e) => {
            if (e.target === modal) hideWebImportModal();
        });
    }
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

    // 初始化书源管理
    initSourceManagement();

    // 初始化净化规则管理
    initCleanupRules();

    // 初始化 Web 导入
    initWebImport();

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
