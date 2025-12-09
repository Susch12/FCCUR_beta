// API Base URL
const API_BASE = '/api';

// Global state
let allPackages = [];
let filteredPackages = [];
let downloadStats = [];

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    loadPackages();
    loadStats();
    setupEventListeners();
});

// Event Listeners
function setupEventListeners() {
    // Search input
    document.getElementById('search-input').addEventListener('input', (e) => {
        filterPackages();
    });

    // Content type filter
    document.getElementById('content-type-filter').addEventListener('change', (e) => {
        const courseFilter = document.getElementById('course-filter');
        if (e.target.value === 'material') {
            courseFilter.style.display = 'block';
            populateCourseFilter();
        } else {
            courseFilter.style.display = 'none';
            // Reset course filter when switching away from materials
            courseFilter.value = '';
        }
        filterPackages();
    });

    // Category filter
    document.getElementById('category-filter').addEventListener('change', (e) => {
        filterPackages();
    });

    // Course filter
    document.getElementById('course-filter').addEventListener('change', (e) => {
        filterPackages();
    });

    // Platform filter
    document.getElementById('platform-filter').addEventListener('change', (e) => {
        filterPackages();
    });

    // Sort filter
    document.getElementById('sort-filter').addEventListener('change', (e) => {
        localStorage.setItem('fccur_sort_preference', e.target.value);
        filterPackages();
    });

    // Content type in upload form
    document.getElementById('package-content-type').addEventListener('change', (e) => {
        const courseNameGroup = document.getElementById('course-name-group');
        if (e.target.value === 'material') {
            courseNameGroup.style.display = 'block';
        } else {
            courseNameGroup.style.display = 'none';
        }
    });

    // Admin toggle
    document.getElementById('admin-toggle').addEventListener('click', toggleAdmin);

    // Upload form
    document.getElementById('upload-form').addEventListener('submit', uploadPackage);

    // Modal close
    document.getElementById('modal-close').addEventListener('click', closeModal);

    // Close modal on outside click
    document.getElementById('modal').addEventListener('click', (e) => {
        if (e.target.id === 'modal') {
            closeModal();
        }
    });
}

// Load packages from API
async function loadPackages() {
    try {
        showLoading();
        const response = await fetch(`${API_BASE}/packages`);

        if (!response.ok) {
            throw new Error('Error loading packages');
        }

        allPackages = await response.json() || [];
        filteredPackages = [...allPackages];

        // Load download stats for sorting
        await loadStatsForSorting();

        // Restore sort preference
        const sortPreference = localStorage.getItem('fccur_sort_preference');
        if (sortPreference) {
            document.getElementById('sort-filter').value = sortPreference;
        }

        renderPackages();
        renderRecentUploads();
        hideLoading();

        // Initialize course filter visibility based on current content-type selection
        initializeCourseFilter();
    } catch (error) {
        showError('Error al cargar paquetes: ' + error.message);
        hideLoading();
    }
}

// Load stats for sorting purposes
async function loadStatsForSorting() {
    try {
        const response = await fetch(`${API_BASE}/stats`);
        if (response.ok) {
            downloadStats = await response.json();
        }
    } catch (error) {
        console.error('Error loading stats:', error);
    }
}

// Render packages to grid
function renderPackages() {
    const grid = document.getElementById('packages-grid');
    const noResults = document.getElementById('no-results');
    grid.innerHTML = '';

    if (filteredPackages.length === 0) {
        grid.style.display = 'none';
        noResults.style.display = 'block';
        return;
    }

    grid.style.display = 'grid';
    noResults.style.display = 'none';

    filteredPackages.forEach(pkg => {
        const card = createPackageCard(pkg);
        grid.appendChild(card);
    });
}

// Create package card element
function createPackageCard(pkg, compact = false) {
    const card = document.createElement('div');
    card.className = compact ? 'package-card-compact' : 'package-card';

    // Add class based on content type
    card.classList.add(pkg.content_type === 'material' ? 'material-card' : 'tool-card');

    const size = formatFileSize(pkg.file_size);
    const platform = getPlatformIcon(pkg.platform);
    const category = getCategoryName(pkg.category);
    const contentTypeLabel = pkg.content_type === 'material' ? '📚 Material' : '🛠️ Herramienta';
    const isNew = isPackageNew(pkg.created_at);

    if (compact) {
        card.innerHTML = `
            <img src="/api/thumbnail?id=${pkg.id}" alt="${escapeHtml(pkg.name)}" class="package-thumbnail-compact" loading="lazy">
            <div class="package-header">
                <h4>${escapeHtml(pkg.name)} v${escapeHtml(pkg.version)}</h4>
                ${isNew ? '<span class="badge-new">NUEVO</span>' : ''}
            </div>
            <div class="package-meta-compact">
                <span>${contentTypeLabel}</span>
                <span>📦 ${size}</span>
            </div>
            <button class="btn-info-small" onclick="showPackageInfo(${pkg.id})">Ver detalles</button>
        `;
    } else {
        card.innerHTML = `
            <img src="/api/thumbnail?id=${pkg.id}" alt="${escapeHtml(pkg.name)}" class="package-thumbnail" loading="lazy">
            <div class="package-header">
                <h3>${escapeHtml(pkg.name)} ${isNew ? '<span class="badge-new">NUEVO</span>' : ''}</h3>
                <span class="version">v${escapeHtml(pkg.version)}</span>
            </div>
            <div class="package-body">
                ${pkg.course_name ? `<p class="course-badge">📖 ${escapeHtml(pkg.course_name)}</p>` : ''}
                <p class="description">${escapeHtml(pkg.description) || 'Sin descripción'}</p>
                <div class="package-meta">
                    <span>${contentTypeLabel}</span>
                    ${pkg.content_type === 'tool' ? `<span>${platform} ${escapeHtml(pkg.platform)}</span>` : ''}
                    <span>📦 ${size}</span>
                    <span>📂 ${category}</span>
                </div>
            </div>
            <div class="package-footer">
                <button class="btn-download" onclick="downloadPackage(${pkg.id})">
                    Descargar
                </button>
                <button class="btn-info" onclick="showPackageInfo(${pkg.id})">
                    Info
                </button>
            </div>
        `;
    }

    return card;
}

// Check if package is new (less than 7 days old)
function isPackageNew(createdAt) {
    const packageDate = new Date(createdAt);
    const now = new Date();
    const daysDifference = (now - packageDate) / (1000 * 60 * 60 * 24);
    return daysDifference < 7;
}

// Render recent uploads
function renderRecentUploads() {
    const container = document.getElementById('recent-uploads');
    container.innerHTML = '';

    // Get last 10 packages sorted by date
    const recentPackages = [...allPackages]
        .sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
        .slice(0, 10);

    if (recentPackages.length === 0) {
        container.innerHTML = '<p class="no-recent">No hay paquetes subidos recientemente</p>';
        return;
    }

    recentPackages.forEach(pkg => {
        const card = createPackageCard(pkg, true);
        container.appendChild(card);
    });
}

// Download package
async function downloadPackage(id) {
    try {
        const pkg = allPackages.find(p => p.id === id);
        if (!pkg) return;

        // Create download link
        const a = document.createElement('a');
        a.href = `/download/?id=${id}`;
        a.download = `${pkg.name}-${pkg.version}`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);

        showSuccess('Descarga iniciada');

        // Reload stats after a delay
        setTimeout(loadStats, 1000);
    } catch (error) {
        showError('Error al descargar: ' + error.message);
    }
}

// Show package info modal
function showPackageInfo(id) {
    const pkg = allPackages.find(p => p.id === id);
    if (!pkg) return;

    const modalBody = document.getElementById('modal-body');
    const createdDate = new Date(pkg.created_at).toLocaleDateString('es-ES');

    const contentTypeLabel = pkg.content_type === 'material' ? 'Material de Curso' : 'Herramienta/Software';

    // Check if admin mode is active
    const adminSection = document.getElementById('admin-section');
    const isAdminMode = adminSection && adminSection.style.display !== 'none';

    modalBody.innerHTML = `
        <h2>${escapeHtml(pkg.name)} v${escapeHtml(pkg.version)}</h2>
        <div class="info-grid">
            <div class="info-item">
                <strong>Tipo de Contenido:</strong>
                ${contentTypeLabel}
            </div>
            ${pkg.course_name ? `
            <div class="info-item">
                <strong>Curso:</strong>
                ${escapeHtml(pkg.course_name)}
            </div>
            ` : ''}
            <div class="info-item">
                <strong>Categoría:</strong>
                ${getCategoryName(pkg.category)}
            </div>
            <div class="info-item">
                <strong>Plataforma:</strong>
                ${escapeHtml(pkg.platform)}
            </div>
            <div class="info-item">
                <strong>Tamaño:</strong>
                ${formatFileSize(pkg.file_size)}
            </div>
            <div class="info-item">
                <strong>Fecha de carga:</strong>
                ${createdDate}
            </div>
            <div class="info-item">
                <strong>BLAKE3 Hash:</strong>
                <code class="hash">${escapeHtml(pkg.blake3_hash)}</code>
            </div>
            <div class="info-item">
                <strong>SHA256 Hash:</strong>
                <code class="hash">${escapeHtml(pkg.sha256_hash)}</code>
            </div>
            ${pkg.download_url ? `
            <div class="info-item">
                <strong>URL Original:</strong>
                <a href="${escapeHtml(pkg.download_url)}" target="_blank" rel="noopener">${escapeHtml(pkg.download_url)}</a>
            </div>
            ` : ''}
        </div>
        ${pkg.description ? `<p style="margin-top: 1rem; color: var(--gray-700);">${escapeHtml(pkg.description)}</p>` : ''}

        <!-- Checksum Downloads -->
        <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid var(--gray-300);">
            <h3 style="font-size: 1rem; margin-bottom: 0.75rem; color: var(--gray-800);">Archivos de Verificación</h3>
            <div style="display: flex; gap: 0.5rem; flex-wrap: wrap;">
                <a href="${API_BASE}/checksum?id=${pkg.id}&type=sha256" class="btn-checksum" download>
                    Descargar SHA256
                </a>
                <a href="${API_BASE}/checksum?id=${pkg.id}&type=blake3" class="btn-checksum" download>
                    Descargar BLAKE3
                </a>
            </div>
        </div>

        <!-- Archive Contents (if archive) -->
        <div id="archive-section-${pkg.id}" style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid var(--gray-300); display: none;">
            <h3 style="font-size: 1rem; margin-bottom: 0.75rem; color: var(--gray-800);">Contenido del Archivo</h3>
            <button class="btn-info" onclick="loadArchiveContents(${pkg.id})" id="load-archive-btn-${pkg.id}">
                Ver Contenido
            </button>
            <div id="archive-contents-${pkg.id}" style="display: none; margin-top: 1rem;"></div>
        </div>

        ${isAdminMode ? `
        <div style="margin-top: 2rem; padding-top: 1rem; border-top: 1px solid var(--gray-300);">
            <button class="btn-delete" onclick="confirmDeletePackage(${pkg.id}, '${escapeHtml(pkg.name).replace(/'/g, "\\'")}')">
                Eliminar Paquete
            </button>
        </div>
        ` : ''}
    `;

    // Check if file is an archive and show archive section
    const filename = pkg.file_path || '';
    const ext = filename.toLowerCase();
    if (ext.endsWith('.zip') || ext.endsWith('.tar') || ext.endsWith('.tar.gz') || ext.endsWith('.tgz')) {
        const archiveSection = document.getElementById(`archive-section-${pkg.id}`);
        if (archiveSection) {
            archiveSection.style.display = 'block';
        }
    }

    document.getElementById('modal').style.display = 'flex';
}

// Load archive contents
async function loadArchiveContents(id) {
    const button = document.getElementById(`load-archive-btn-${id}`);
    const contentsDiv = document.getElementById(`archive-contents-${id}`);

    try {
        button.disabled = true;
        button.textContent = 'Cargando...';

        const response = await fetch(`${API_BASE}/archive/contents?id=${id}`);
        if (!response.ok) {
            throw new Error('Error al cargar contenido del archivo');
        }

        const data = await response.json();

        // Build HTML for archive contents
        let html = `
            <div class="archive-summary">
                <p><strong>Total de archivos:</strong> ${data.total_files}</p>
                <p><strong>Tamaño total:</strong> ${formatFileSize(data.total_size)}</p>
            </div>
        `;

        // Show README if present
        if (data.readme) {
            html += `
                <div class="readme-section">
                    <h4>README</h4>
                    <pre class="readme-content">${escapeHtml(data.readme)}</pre>
                </div>
            `;
        }

        // Show file list
        html += `
            <div class="file-list-section">
                <h4>Archivos (${data.files.length})</h4>
                <div class="file-list">
        `;

        data.files.forEach(file => {
            const icon = file.is_dir ? '📁' : '📄';
            const sizeStr = file.is_dir ? '' : ` - ${formatFileSize(file.size)}`;
            html += `<div class="file-item">${icon} ${escapeHtml(file.name)}${sizeStr}</div>`;
        });

        html += `
                </div>
            </div>
        `;

        contentsDiv.innerHTML = html;
        contentsDiv.style.display = 'block';
        button.style.display = 'none';

    } catch (error) {
        showError('Error al cargar contenido: ' + error.message);
        button.disabled = false;
        button.textContent = 'Ver Contenido';
    }
}

// Close modal
function closeModal() {
    document.getElementById('modal').style.display = 'none';
}

// Filter packages
function filterPackages() {
    const search = document.getElementById('search-input').value.toLowerCase();
    const contentType = document.getElementById('content-type-filter').value;
    const category = document.getElementById('category-filter').value;
    const course = document.getElementById('course-filter').value;
    const platform = document.getElementById('platform-filter').value;
    const sortBy = document.getElementById('sort-filter').value;

    filteredPackages = allPackages.filter(pkg => {
        const matchesSearch = !search ||
            pkg.name.toLowerCase().includes(search) ||
            (pkg.description && pkg.description.toLowerCase().includes(search)) ||
            (pkg.course_name && pkg.course_name.toLowerCase().includes(search));

        const matchesContentType = !contentType || pkg.content_type === contentType;
        const matchesCategory = !category || pkg.category === category;
        const matchesCourse = !course || pkg.course_name === course;
        const matchesPlatform = !platform || pkg.platform === platform;

        return matchesSearch && matchesContentType && matchesCategory && matchesCourse && matchesPlatform;
    });

    // Apply sorting
    sortPackages(filteredPackages, sortBy);

    renderPackages();
}

// Sort packages based on selected criteria
function sortPackages(packages, sortBy) {
    switch (sortBy) {
        case 'date-desc':
            packages.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
            break;
        case 'date-asc':
            packages.sort((a, b) => new Date(a.created_at) - new Date(b.created_at));
            break;
        case 'name-asc':
            packages.sort((a, b) => a.name.localeCompare(b.name));
            break;
        case 'name-desc':
            packages.sort((a, b) => b.name.localeCompare(a.name));
            break;
        case 'size-desc':
            packages.sort((a, b) => b.file_size - a.file_size);
            break;
        case 'size-asc':
            packages.sort((a, b) => a.file_size - b.file_size);
            break;
        case 'downloads-desc':
            packages.sort((a, b) => {
                const aDownloads = getDownloadCount(a.id);
                const bDownloads = getDownloadCount(b.id);
                return bDownloads - aDownloads;
            });
            break;
        default:
            packages.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
    }
}

// Get download count for a package
function getDownloadCount(packageId) {
    const stat = downloadStats.find(s => s.package_id === packageId);
    return stat ? stat.total_downloads : 0;
}

// Populate course filter with unique course names
function populateCourseFilter() {
    const courseFilter = document.getElementById('course-filter');
    const courses = new Set();

    allPackages.forEach(pkg => {
        if (pkg.content_type === 'material' && pkg.course_name) {
            courses.add(pkg.course_name);
        }
    });

    // Keep the "all" option and add unique courses
    courseFilter.innerHTML = '<option value="">Todos los cursos</option>';
    Array.from(courses).sort().forEach(course => {
        const option = document.createElement('option');
        option.value = course;
        option.textContent = course;
        courseFilter.appendChild(option);
    });
}

// Initialize course filter on page load
function initializeCourseFilter() {
    const contentTypeFilter = document.getElementById('content-type-filter');
    const courseFilter = document.getElementById('course-filter');

    // Check current content-type filter value
    if (contentTypeFilter.value === 'material') {
        // Show and populate course filter if materials are selected
        courseFilter.style.display = 'block';
        populateCourseFilter();
    } else {
        // Hide course filter for "all" or "tool"
        courseFilter.style.display = 'none';
    }
}

// Calculate BLAKE3 hash of a file
async function calculateFileHash(file) {
    // Initialize BLAKE3
    const hasher = blake3.createHash();

    // Read file in chunks
    const chunkSize = 1024 * 1024; // 1MB chunks
    let offset = 0;

    while (offset < file.size) {
        const chunk = file.slice(offset, offset + chunkSize);
        const arrayBuffer = await chunk.arrayBuffer();
        hasher.update(new Uint8Array(arrayBuffer));
        offset += chunkSize;
    }

    // Get hash as hex string
    const hashBuffer = hasher.digest();
    return Array.from(new Uint8Array(hashBuffer))
        .map(b => b.toString(16).padStart(2, '0'))
        .join('');
}

// Check for duplicate file by hash
async function checkDuplicate(hash) {
    const response = await fetch(`${API_BASE}/check-duplicate?hash=${hash}`);
    if (!response.ok) {
        throw new Error('Error checking for duplicates');
    }
    return await response.json();
}

// Upload package
async function uploadPackage(e) {
    e.preventDefault();

    const form = e.target;
    const formData = new FormData(form);
    const fileInput = form.querySelector('input[type="file"]');
    const file = fileInput.files[0];

    if (!file) {
        showError('Por favor seleccione un archivo');
        return;
    }

    const progressDiv = document.getElementById('upload-progress');
    const progressFill = document.getElementById('progress-fill');
    const progressPercentage = document.getElementById('progress-percentage');

    try {
        // Show progress for hash calculation
        form.style.display = 'none';
        progressDiv.style.display = 'block';
        progressPercentage.textContent = 'Calculando hash...';

        // Calculate file hash
        const hash = await calculateFileHash(file);

        // Check for duplicates
        const duplicateCheck = await checkDuplicate(hash);

        if (duplicateCheck.duplicate) {
            // Hide progress, show form
            form.style.display = 'block';
            progressDiv.style.display = 'none';
            progressFill.style.width = '0%';
            progressPercentage.textContent = '0%';

            // Show duplicate warning
            const shouldContinue = await showDuplicateWarning(duplicateCheck.package);
            if (!shouldContinue) {
                return; // User cancelled
            }
        }

        // Show upload progress
        form.style.display = 'none';
        progressDiv.style.display = 'block';
        progressFill.style.width = '0%';
        progressPercentage.textContent = '0%';

        const xhr = new XMLHttpRequest();

        // Track upload progress
        xhr.upload.addEventListener('progress', (e) => {
            if (e.lengthComputable) {
                const percentComplete = (e.loaded / e.total) * 100;
                progressFill.style.width = percentComplete + '%';
                progressPercentage.textContent = Math.round(percentComplete) + '%';
            }
        });

        // Handle completion
        xhr.addEventListener('load', async () => {
            if (xhr.status === 201) {
                showSuccess('Paquete subido exitosamente');
                form.reset();
                form.style.display = 'block';
                progressDiv.style.display = 'none';
                progressFill.style.width = '0%';
                progressPercentage.textContent = '0%';
                await loadPackages();
                await loadStats();
            } else {
                throw new Error('Error al subir paquete');
            }
        });

        xhr.addEventListener('error', () => {
            throw new Error('Error de red al subir paquete');
        });

        xhr.open('POST', `${API_BASE}/upload`);
        xhr.send(formData);

    } catch (error) {
        form.style.display = 'block';
        progressDiv.style.display = 'none';
        showError('Error al subir paquete: ' + error.message);
    }
}

// Show duplicate warning dialog
async function showDuplicateWarning(existingPackage) {
    const contentTypeLabel = existingPackage.content_type === 'material' ? 'Material' : 'Herramienta';
    const createdDate = new Date(existingPackage.created_at).toLocaleDateString('es-ES');

    const message = `ADVERTENCIA: Ya existe un archivo idéntico en el sistema

Paquete existente:
Nombre: ${existingPackage.name}
Versión: ${existingPackage.version}
Tipo: ${contentTypeLabel}
${existingPackage.course_name ? `Curso: ${existingPackage.course_name}\n` : ''}Fecha de carga: ${createdDate}
Tamaño: ${formatFileSize(existingPackage.file_size)}

Opciones:
1. "Cancelar" - No subir el archivo duplicado
2. "Aceptar" - Subir de todas formas (creará una entrada duplicada)

¿Desea continuar con la subida?`;

    return confirm(message);
}

// Load statistics
async function loadStats() {
    try {
        const response = await fetch(`${API_BASE}/stats`);
        if (!response.ok) return;

        const stats = await response.json();
        renderStats(stats);
    } catch (error) {
        console.error('Error loading stats:', error);
    }
}

// Render statistics
function renderStats(stats) {
    const grid = document.getElementById('stats-grid');
    grid.innerHTML = '';

    // Total packages
    const totalCard = createStatCard('Total Paquetes', allPackages.length);
    grid.appendChild(totalCard);

    // Total downloads
    const totalDownloads = stats.reduce((sum, s) => sum + s.total_downloads, 0);
    const downloadsCard = createStatCard('Total Descargas', totalDownloads);
    grid.appendChild(downloadsCard);

    // Most downloaded
    if (stats.length > 0 && stats[0].total_downloads > 0) {
        const topPkg = stats[0];
        const topCard = createStatCard(
            'Más Descargado',
            topPkg.package_name,
            `${topPkg.total_downloads} descargas`
        );
        grid.appendChild(topCard);
    }
}

// Create stat card
function createStatCard(title, value, detail = '') {
    const card = document.createElement('div');
    card.className = 'stat-card';

    card.innerHTML = `
        <h4>${escapeHtml(title)}</h4>
        <div class="value">${escapeHtml(String(value))}</div>
        ${detail ? `<p class="detail">${escapeHtml(detail)}</p>` : ''}
    `;

    return card;
}

// Toggle admin section
function toggleAdmin() {
    const section = document.getElementById('admin-section');
    section.style.display = section.style.display === 'none' ? 'block' : 'none';
}

// Show loading
function showLoading() {
    document.getElementById('loading').style.display = 'block';
    document.getElementById('packages-grid').style.display = 'none';
    document.getElementById('no-results').style.display = 'none';
}

// Hide loading
function hideLoading() {
    document.getElementById('loading').style.display = 'none';
}

// Show error
function showError(message) {
    const errorDiv = document.getElementById('error');
    errorDiv.textContent = message;
    errorDiv.style.display = 'block';
    setTimeout(() => {
        errorDiv.style.display = 'none';
    }, 5000);
}

// Show success
function showSuccess(message) {
    alert(message);
}

// Utility Functions

// Format file size
function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
}

// Get platform icon
function getPlatformIcon(platform) {
    const icons = {
        'windows': '🪟',
        'mac': '🍎',
        'linux': '🐧',
        'all': '💻'
    };
    return icons[platform.toLowerCase()] || '💻';
}

// Get category name
function getCategoryName(category) {
    const names = {
        'os': 'Sistema Operativo',
        'compiler': 'Compilador',
        'ide': 'IDE',
        'tool': 'Herramienta',
        'library': 'Biblioteca'
    };
    return names[category] || category;
}

// Escape HTML to prevent XSS
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Confirm package deletion with name validation
function confirmDeletePackage(id, packageName) {
    const userInput = prompt(`Para confirmar la eliminación, escriba el nombre del paquete:\n\n"${packageName}"`);

    if (userInput === null) {
        // User cancelled
        return;
    }

    if (userInput.trim() !== packageName) {
        showError('El nombre del paquete no coincide. Eliminación cancelada.');
        return;
    }

    // Name matches, proceed with deletion
    deletePackage(id);
}

// Delete package (requires authentication)
async function deletePackage(id) {
    try {
        // Get credentials from localStorage or prompt
        let username = localStorage.getItem('fccur_admin_user');
        let password = localStorage.getItem('fccur_admin_pass');

        if (!username || !password) {
            username = prompt('Usuario administrador:');
            if (!username) return;

            password = prompt('Contraseña:');
            if (!password) return;

            // Optionally save credentials for session
            const saveCredentials = confirm('¿Guardar credenciales para esta sesión?');
            if (saveCredentials) {
                localStorage.setItem('fccur_admin_user', username);
                localStorage.setItem('fccur_admin_pass', password);
            }
        }

        // Create Basic Auth header
        const credentials = btoa(`${username}:${password}`);

        const response = await fetch(`${API_BASE}/delete?id=${id}`, {
            method: 'POST',
            headers: {
                'Authorization': `Basic ${credentials}`
            }
        });

        if (response.status === 401) {
            // Clear invalid credentials
            localStorage.removeItem('fccur_admin_user');
            localStorage.removeItem('fccur_admin_pass');
            showError('Credenciales inválidas. Intente de nuevo.');
            return;
        }

        if (!response.ok) {
            throw new Error('Error al eliminar paquete');
        }

        const result = await response.json();
        showSuccess('Paquete eliminado exitosamente');

        // Close modal and reload data
        closeModal();
        await loadPackages();
        await loadStats();

    } catch (error) {
        showError('Error al eliminar paquete: ' + error.message);
    }
}

// Clear admin credentials (can be called from console or added to UI)
function clearAdminCredentials() {
    localStorage.removeItem('fccur_admin_user');
    localStorage.removeItem('fccur_admin_pass');
    showSuccess('Credenciales eliminadas');
}
