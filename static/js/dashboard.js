document.addEventListener('DOMContentLoaded', function () {
    // --- Global Setup ---
    let grid;
    const widgetConfigs = new Map(); // Stores data for each widget (model, fields, etc.)
    let autoRefreshTimer = null;

    // --- Theme Management ---
    const themeToggle = document.getElementById('themeToggle');
    const applyTheme = (theme) => {
        if (theme === 'dark') {
            document.body.classList.add('dark-mode');
            themeToggle.checked = true;
        } else {
            document.body.classList.remove('dark-mode');
            themeToggle.checked = false;
        }
    };
    themeToggle.addEventListener('change', () => {
        const theme = themeToggle.checked ? 'dark' : 'light';
        localStorage.setItem('dashboardTheme', theme);
        applyTheme(theme);
    });
    applyTheme(localStorage.getItem('dashboardTheme') || 'light');

    // --- Panel Management ---
    const applyPanelState = (isHidden) => {
        configPanel.classList.toggle('hidden', isHidden);
    };
    toggleControlsBtn.addEventListener('click', () => {
        const isHidden = !configPanel.classList.contains('hidden');
        localStorage.setItem('configPanelHidden', isHidden);
        applyPanelState(isHidden);
    });
    applyPanelState(localStorage.getItem('configPanelHidden') === 'true');

    // --- Auto-Refresh Setup ---
    const autoRefreshToggle = document.getElementById('autoRefreshToggle');
    const refreshIntervalInput = document.getElementById('refreshInterval');
    const applyAutoRefresh = (enabled) => {
        autoRefreshToggle.checked = enabled;
        refreshIntervalInput.value = localStorage.getItem('refreshInterval') || 30;
        toggleAutoRefresh(enabled);
    };
    autoRefreshToggle.addEventListener('change', () => {
        const enabled = autoRefreshToggle.checked;
        localStorage.setItem('autoRefreshEnabled', enabled);
        localStorage.setItem('refreshInterval', refreshIntervalInput.value);
        toggleAutoRefresh(enabled);
    });
    applyAutoRefresh(localStorage.getItem('autoRefreshEnabled') === 'true');
    refreshIntervalInput.addEventListener('change', () => {
        const intervalSeconds = refreshIntervalInput.value;
        localStorage.setItem('refreshInterval', intervalSeconds);
        if (autoRefreshToggle.checked) {
            toggleAutoRefresh(true);
        }
    });

    function toggleAutoRefresh(enabled) {
        if (autoRefreshTimer) clearInterval(autoRefreshTimer);
        autoRefreshTimer = null;

        if (enabled) {
            const intervalSeconds = localStorage.getItem('refreshInterval') || 30;
            autoRefreshTimer = setInterval(() => {
                widgetConfigs.forEach((config, id) => refreshWidgetData(id));
            }, intervalSeconds * 1000);
        }
    }
    


    // --- Grid and Widget Functions ---
    function initGrid() {
        grid = GridStack.init({
            float: true, cellHeight: 'auto', margin: 10, minRow: 1, acceptWidgets: true, removable: '#trash',
        });
        grid.on('change removed', () => saveLayout());
        grid.on('removed', (event, items) => {
            items.forEach(item => {
                const config = widgetConfigs.get(item.id);
                if (config && config.intervalId) clearInterval(config.intervalId);
                widgetConfigs.delete(item.id);
            });
            saveLayout();
        });
    }

    function saveLayout() {
        const savedWidgets = grid.engine.nodes.map(node => {
            console.log(node)
            const config = widgetConfigs.get(node.id);
            console.log(config)
            if (!config) return null; // Skip if no config found for this widget
            else if (config.model === null || config.model === undefined) {
                console.warn(`Widget ${node.id} has no model defined, skipping save.`);
                return null; // Skip saving this widget
            } else if (config.fields === null || config.fields === undefined) {
                console.warn(`Widget ${node.id} has no fields defined, skipping save.`);
                return null; // Skip saving this widget
            }
            
            return {
                x: node.x, y: node.y, w: node.w, h: node.h, 
                id: config.id,
                title: config ? config.title : config.model,
                model: config ? config.model : '',
                fields: config ? config.fields : [],
                domain: config ? config.domain : [],
                limit: config ? config.limit : 15, 
            };
        });
        fetch('/api/save_layout', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(savedWidgets)
        }).catch(err => console.error('Failed to save layout:', err));
    }

    function loadLayout() {
        fetch('/api/load_layout')
            .then(res => res.json())
            .then(savedWidgets => {
                if (savedWidgets && savedWidgets.length > 0) {
                    grid.batchUpdate();
                    savedWidgets.forEach(widgetData => createWidget(widgetData, false));
                    grid.commit();
                    
                    console.log('Layout loaded:', savedWidgets);
                }
            })
            .catch(err => console.error('Failed to load layout:', err));
    }

    function createWidget(widgetData, isNew = false) {
        const id = widgetData.id || 'widget-' + Date.now(); 
        widgetConfigs.set(id, {
            id: id,
            title: widgetData.title,
            model: widgetData.model,
            fields: widgetData.fields,
            domain: widgetData.domain,
            limit: widgetData.limit || 15,
        });

        console.log('Creating widget:', id, widgetData);

        const widgetEl = grid.addWidget({
            x: widgetData.x, y: widgetData.y,
            w: widgetData.w || 6, h: widgetData.h || 5,
            id: id,
            autoPosition: isNew,
        });

        // Create the inner container for our content
        const contentEl = widgetEl.querySelector('.grid-stack-item-content');
        if (contentEl) {
            const innerContent = document.createElement('div');
            innerContent.className = 'grid-stack-item-content-inner';
            contentEl.appendChild(innerContent);
        }
        
        refreshWidgetData(id);
    }

    function refreshWidgetData(widgetId) {

        console.log('Refreshing widget:', widgetId);

        const config = widgetConfigs.get(widgetId);
        if (!config) return;

        console.log('Widget config:', config);

        // const widgetEl = document.getElementById(widgetId);
        const widgetEl = document.querySelector(`.grid-stack-item[gs-id="${widgetId}"]`);
        if (!widgetEl) return;

        console.log('Widget element found:', widgetEl);
        
        const contentInnerEl = widgetEl.querySelector('.grid-stack-item-content-inner');
        if (contentInnerEl) {
            contentInnerEl.classList.add('loading');
            contentInnerEl.innerHTML = 'Loading...';
        }

        console.log('Fetching data for widget:', config);

        fetch('/api/odoo', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ model: config.model, fields: config.fields, domain: config.domain, limit: config.limit }),
        })
        .then(response => {
            console.log('Response status:', response.status);
            
            if (!response.ok) return response.text().then(text => { throw new Error(text) });
            return response.json();
        })
        .then(data => {
            console.log('Data received for widget:', widgetId, data);

            const records = data.records;
            const totalCount = data.total_count;
            if (!records) throw new Error("No records found in the response.");

            const displayTitle = config.title || config.model;
            let headerHTML = `<div class="widget-header">
                                  <h4 class="widget-title" contenteditable="true" data-widget-id="${widgetId}">${displayTitle}</h4>
                                  <span class="record-count">(${records.length} of ${totalCount})</span>
                                  <div class="widget-actions">
                                      <ion-icon name="settings-outline" class="config-icon" data-widget-id="${widgetId}"></ion-icon>
                                      <ion-icon name="refresh-outline" class="refresh-icon" data-widget-id="${widgetId}"></ion-icon>
                                  </div>
                              </div>`;
            let tableHTML = '<div class="table-container"><table><thead><tr>';
            config.fields.forEach(field => tableHTML += `<th>${(field.charAt(0).toUpperCase() + field.slice(1)).replace(/_/g, ' ')}</th>`);
            tableHTML += '</tr></thead><tbody>';
            records.forEach(item => {
                tableHTML += '<tr>';
                config.fields.forEach(field => {
                    let value = item[field];
                    if (Array.isArray(value) && value.length > 0) value = value[1];
                    else if (value === false || value === null || value === undefined) value = '–';
                    tableHTML += `<td>${value}</td>`;
                });
                tableHTML += '</tr>';
            });
            tableHTML += '</tbody></table></div>';
            if (contentInnerEl) {
                contentInnerEl.innerHTML = headerHTML + tableHTML;
                contentInnerEl.classList.remove('loading');
            }
        })
        .catch(error => {
            console.error('Error refreshing widget:', widgetId, error);
            if (contentInnerEl) {
                const errorHTML = `<div class="widget-header"><h4>${config.model}</h4><ion-icon name="settings-outline" class="config-icon" data-widget-id="${widgetId}"></ion-icon><ion-icon name="refresh-outline" class="refresh-icon" data-widget-id="${widgetId}"></ion-icon></div><div class="error-msg">Failed to load: ${error.message}</div>`;
                contentInnerEl.innerHTML = errorHTML;
                contentInnerEl.classList.remove('loading');
            }
        });
    }

    // --- Modal Logic ---
    const configModal = document.getElementById('configModal');
    const saveConfigBtn = document.getElementById('saveConfigBtn');
    const cancelConfigBtn = document.getElementById('cancelConfigBtn');

    function openConfigModal(widgetId) {
        const config = widgetConfigs.get(widgetId);
        if (!config) return;
        document.getElementById('configWidgetId').value = widgetId;
        document.getElementById('configTitle').value = config.title;
        document.getElementById('configModel').value = config.model;
        document.getElementById('configFields').value = config.fields.join(',');
        document.getElementById('configDomain').value = JSON.stringify(config.domain, null, 2);
        document.getElementById('configLimit').value = config.limit || 15;
        configModal.style.display = 'flex';
    }

    function closeConfigModal() {
        configModal.style.display = 'none';
    }

    saveConfigBtn.addEventListener('click', () => {
        const widgetId = document.getElementById('configWidgetId').value;
        const config = widgetConfigs.get(widgetId);
        if (!config) return;

        try {
            const domain = JSON.parse(document.getElementById('configDomain').value);
            config.title = document.getElementById('configTitle').value;
            config.model = document.getElementById('configModel').value;
            config.fields = document.getElementById('configFields').value.split(',').map(f => f.trim()).filter(f => f);
            config.domain = domain;
            config.limit = parseInt(document.getElementById('configLimit').value, 10) || 15;
            
            refreshWidgetData(widgetId);
            saveLayout();
            closeConfigModal();
        } catch (e) {
            alert('Invalid Domain JSON format. Please check your syntax.');
        }
    });

    cancelConfigBtn.addEventListener('click', closeConfigModal);
    configModal.addEventListener('click', (e) => {
        if (e.target === configModal) closeConfigModal();
    });

    // --- Add Widget Modal Logic ---
    const addWidgetModal = document.getElementById('addWidgetModal');
    // const saveWidgetBtn = document.getElementById('saveWidgetBtn');
    const cancelWidgetBtn = document.getElementById('cancelWidgetBtn');

    function openAddWidgetModal() {
        addWidgetModal.style.display = 'flex';
    }

    function closeAddWidgetModal() {
        addWidgetModal.style.display = 'none';
    }

    // saveWidgetBtn.addEventListener('click', () => {
    //     const title = document.getElementById('titleInput').value;
    //     const model = document.getElementById('modelInput').value;
    //     const fields = document.getElementById('fieldsInput').value.split(',').map(f => f.trim()).filter(f => f);
    //     const domain = JSON.parse(document.getElementById('domainInput').value);
    //     const limit = parseInt(document.getElementById('limitInput').value, 10) || 15;

    //     createWidget({ title, model, fields, domain, limit });
    //     closeAddWidgetModal();
    // });
    cancelWidgetBtn.addEventListener('click', closeAddWidgetModal);
    addWidgetModal.addEventListener('click', (e) => {
        if (e.target === addWidgetModal) closeAddWidgetModal();
    });



    // --- Event Listeners ---
    document.getElementById('fetchButton').addEventListener('click', () => {
        const title = document.getElementById('titleInput').value;
        const model = document.getElementById('modelInput').value;
        const fields = document.getElementById('fieldsInput').value.split(',').map(f => f.trim()).filter(f => f);
        const filterText = document.getElementById('domainInput').value;
        const limit = parseInt(document.getElementById('limitInput').value, 10) || 15;
        let domain = [];

        if (!model || fields.length === 0) {
            alert('Model and Fields are required to create a widget.');
            return;
        }

        try {
            if (filterText) domain = JSON.parse(filterText);
        } catch (e) {
            alert('Invalid Domain format. Must be valid JSON, e.g., [["name", "ilike", "test"]]');
            return;
        }
        
        createWidget({ title, model, fields, domain, limit }, true);
        closeAddWidgetModal();
        saveLayout();
    });

    document.getElementById('addWidgetBtn').addEventListener('click', openAddWidgetModal);

    document.querySelector('.dashboard').addEventListener('click', (e) => {
        if (e.target.classList.contains('refresh-icon')) {
            refreshWidgetData(e.target.dataset.widgetId);
        } else if (e.target.classList.contains('config-icon')) {
            openConfigModal(e.target.dataset.widgetId);
        }
    });

    document.querySelector('.dashboard').addEventListener('blur', (e) => {
        if (e.target.classList.contains('widget-title')) {
            const widgetId = e.target.dataset.widgetId;
            const newTitle = e.target.textContent;
            const config = widgetConfigs.get(widgetId);
            if (config && config.title !== newTitle) {
                config.title = newTitle;
                saveLayout();
            }
        }
    }, true);

    initGrid();
    loadLayout();
});
