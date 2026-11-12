// CronJob Scale Down Operator Web UI JavaScript

class CronJobDashboard {
    constructor() {
        this.autoRefreshInterval = null;
        this.refreshIntervalMs = 30000; // 30 seconds
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.refreshData();
        this.startAutoRefresh();
    }

    setupEventListeners() {
        // Refresh button click
        document.addEventListener('click', (e) => {
            if (e.target.closest('[data-action="refresh"]')) {
                this.refreshData();
            }
        });

        // Cleanup on page unload
        window.addEventListener('beforeunload', () => {
            this.stopAutoRefresh();
        });
    }

    startAutoRefresh() {
        this.autoRefreshInterval = setInterval(() => {
            this.refreshData();
        }, this.refreshIntervalMs);
    }

    stopAutoRefresh() {
        if (this.autoRefreshInterval) {
            clearInterval(this.autoRefreshInterval);
            this.autoRefreshInterval = null;
        }
    }

    // Escapes API-provided values before they are interpolated into HTML.
    // CronJobScaleDown spec fields are user-controlled and must never be rendered as markup.
    escapeHtml(value) {
        if (value === null || value === undefined) return '';
        return String(value)
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;');
    }

    toCount(value) {
        const n = Number(value);
        return Number.isFinite(n) && n >= 0 ? Math.floor(n) : 0;
    }

    formatDateTime(dateString) {
        if (!dateString) return 'Never';
        const date = new Date(dateString);
        if (isNaN(date.getTime())) return 'Never';
        return this.escapeHtml(date.toLocaleString());
    }

    getStatusIcon(ready) {
        return ready 
            ? '<i class="fas fa-check-circle status-ready"></i>' 
            : '<i class="fas fa-times-circle status-not-ready"></i>';
    }

    getStatusBadge(ready, currentReplicas) {
        if (currentReplicas === 0) {
            return `<span class="status-badge" style="background: var(--warning-light); color: var(--warning-color); border: 1px solid var(--warning-color);">
                <i class="fas fa-pause-circle"></i> Scaled Down
            </span>`;
        }
        
        const badgeClass = ready ? 'ready' : 'not-ready';
        const icon = ready ? 'check-circle' : 'times-circle';
        const text = ready ? 'Ready' : 'Not Ready';
        
        return `<span class="status-badge ${badgeClass}">
            <i class="fas fa-${icon}"></i> ${text}
        </span>`;
    }

    // Shown when the API omits targetStatus (target missing, unreadable, or outside the
    // allowed namespace scope) so it is not mistaken for a scaled-down workload.
    getUnavailableBadge() {
        return `<span class="status-badge not-ready">
            <i class="fas fa-question-circle"></i> Status Unavailable
        </span>`;
    }

    createUnavailableReplicaInfo() {
        return `
            <div class="replica-info">
                <span class="replica-text">–</span>
                <small class="text-muted">Target status unavailable</small>
            </div>
        `;
    }

    createReplicaBar(ready, desired) {
        ready = this.toCount(ready);
        desired = this.toCount(desired);
        if (desired === 0) {
            return `
                <div class="replica-info">
                    <span class="replica-text">0/0</span>
                    <div class="replica-bar">
                        <div class="replica-bar-fill" style="width: 0%"></div>
                    </div>
                    <small class="text-muted">Scaled down</small>
                </div>
            `;
        }

        const percentage = Math.min(100, desired > 0 ? (ready / desired) * 100 : 0);
        let fillClass = '';
        let statusText = '';
        
        if (percentage === 100) {
            fillClass = '';
            statusText = 'All ready';
        } else if (percentage >= 50) {
            fillClass = 'warning';
            statusText = 'Partially ready';
        } else {
            fillClass = 'danger';
            statusText = 'Not ready';
        }

        return `
            <div class="replica-info">
                <span class="replica-text">${ready}/${desired}</span>
                <div class="replica-bar">
                    <div class="replica-bar-fill ${fillClass}" style="width: ${percentage}%"></div>
                </div>
                <small class="text-muted">${statusText}</small>
            </div>
        `;
    }

    createCronJobCard(cronJob) {
        const targetStatus = cronJob.targetStatus;
        const isCleanupOnly = !cronJob.targetRef;
        const statusBadge = isCleanupOnly 
            ? '<span class="badge status-cleanup">Cleanup Only</span>'
            : targetStatus
                ? this.getStatusBadge(targetStatus.ready, cronJob.currentReplicas)
                : this.getUnavailableBadge();
        const replicaBar = isCleanupOnly 
            ? '<div class="info-item"><span class="info-value">No target resource</span></div>'
            : targetStatus
                ? this.createReplicaBar(targetStatus.readyReplicas, targetStatus.desiredReplicas)
                : this.createUnavailableReplicaInfo();
        const esc = (v) => this.escapeHtml(v);
        
        return `
            <div class="col-md-6 col-lg-4 mb-4">
                <div class="card h-100">
                    <div class="card-header">
                        <h5 class="card-title mb-0">
                            <i class="fas fa-cog"></i> ${esc(cronJob.name)}
                        </h5>
                        <small>${esc(cronJob.namespace)}</small>
                    </div>
                    <div class="card-body">
                        <div class="mb-3">
                            <h6 class="section-title">
                                <i class="fas fa-bullseye"></i> Target Resource
                            </h6>
                            ${isCleanupOnly ? 
                                '<div class="info-item"><span class="info-value text-muted">Cleanup-only mode - no target resource</span></div>' :
                                `<div class="info-item">
                                    <span class="resource-kind-badge">${esc(cronJob.targetRef.kind)}</span>
                                </div>
                                <div class="info-item">
                                    <span class="info-label">Name:</span>
                                    <span class="info-value">${esc(cronJob.targetRef.name)}</span>
                                </div>
                                <div class="info-item">
                                    <span class="info-label">Namespace:</span>
                                    <span class="info-value">${esc(cronJob.targetRef.namespace)}</span>
                                </div>`
                            }
                        </div>
                        
                        <hr class="section-divider">
                        
                        <div class="mb-3">
                            <h6 class="section-title">
                                <i class="fas fa-server"></i> Current Status
                            </h6>
                            <div class="mb-2">${statusBadge}</div>
                            <div class="replica-container">
                                <div class="info-item mb-2">
                                    <span class="info-label">Replicas:</span>
                                </div>
                                ${replicaBar}
                            </div>
                        </div>
                        
                        <hr class="section-divider">
                        
                        <div class="mb-3">
                            <h6 class="section-title">
                                <i class="fas fa-clock"></i> Schedules
                            </h6>
                            ${isCleanupOnly ? 
                                `<div class="info-item">
                                    <span class="info-label">Cleanup:</span>
                                    ${cronJob.cleanupSchedule ? `<span class="cron-schedule">${esc(cronJob.cleanupSchedule)}</span>` : '<span class="text-muted">Not set</span>'}
                                </div>` :
                                `<div class="info-item">
                                    <span class="info-label">Scale Down:</span>
                                    ${cronJob.scaleDownSchedule ? `<span class="cron-schedule">${esc(cronJob.scaleDownSchedule)}</span>` : '<span class="text-muted">Not set</span>'}
                                </div>
                                <div class="info-item">
                                    <span class="info-label">Scale Up:</span>
                                    ${cronJob.scaleUpSchedule ? `<span class="cron-schedule">${esc(cronJob.scaleUpSchedule)}</span>` : '<span class="text-muted">Not set</span>'}
                                </div>`
                            }
                            <div class="info-item">
                                <span class="info-label">Timezone:</span>
                                <span class="info-value">${esc(cronJob.timeZone)}</span>
                            </div>
                        </div>
                        
                        <hr class="section-divider">
                        
                        <div class="mb-0">
                            <h6 class="section-title">
                                <i class="fas fa-history"></i> Last Actions
                            </h6>
                            ${isCleanupOnly ? 
                                `<div class="info-item">
                                    <span class="info-label">Last Cleanup:</span>
                                    <div class="last-action-time">${this.formatDateTime(cronJob.lastCleanupTime)}</div>
                                </div>` :
                                `<div class="info-item">
                                    <span class="info-label">Scale Down:</span>
                                    <div class="last-action-time">${this.formatDateTime(cronJob.lastScaleDownTime)}</div>
                                </div>
                                <div class="info-item">
                                    <span class="info-label">Scale Up:</span>
                                    <div class="last-action-time">${this.formatDateTime(cronJob.lastScaleUpTime)}</div>
                                </div>`
                            }
                        </div>
                    </div>
                </div>
            </div>
        `;
    }

    async fetchCronJobs() {
        try {
            const response = await fetch('/api/v1/cronjobs');
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return await response.json();
        } catch (error) {
            throw new Error(`Failed to fetch cron jobs: ${error.message}`);
        }
    }

    async refreshData() {
        const loadingEl = document.getElementById('loading');
        const errorEl = document.getElementById('error-alert');
        const containerEl = document.getElementById('cronjobs-container');
        const lastUpdatedEl = document.getElementById('last-updated');

        // Show loading state
        loadingEl.classList.remove('d-none');
        errorEl.classList.add('d-none');
        containerEl.style.opacity = '0.6';

        try {
            const cronJobs = await this.fetchCronJobs();
            
            // Clear existing content
            containerEl.innerHTML = '';
            
            if (cronJobs.length === 0) {
                containerEl.innerHTML = `
                    <div class="col-12">
                        <div class="empty-state">
                            <div class="empty-state-icon">
                                <i class="fas fa-clock"></i>
                            </div>
                            <h4>No CronJobScaleDown Resources Found</h4>
                            <p>Create some CronJobScaleDown resources to see them here.</p>
                            <small class="text-muted">
                                Use <code>kubectl apply -f examples/</code> to create sample resources.
                            </small>
                        </div>
                    </div>
                `;
            } else {
                // Check if the fetched data has changed
                if (!this.cachedCronJobs || JSON.stringify(this.cachedCronJobs) !== JSON.stringify(cronJobs)) {
                    // Sort cronJobs by namespace then name for consistent ordering
                    cronJobs.sort((a, b) => {
                        if (a.namespace !== b.namespace) {
                            return a.namespace.localeCompare(b.namespace);
                        }
                        return a.name.localeCompare(b.name);
                    });
                    // Cache the sorted data
                    this.cachedCronJobs = cronJobs;
                }

                this.cachedCronJobs.forEach(cronJob => {
                    containerEl.innerHTML += this.createCronJobCard(cronJob);
                });
            }
            
            // Update last updated time
            lastUpdatedEl.textContent = new Date().toLocaleString();
            
        } catch (error) {
            console.error('Error refreshing data:', error);
            document.getElementById('error-message').textContent = error.message;
            errorEl.classList.remove('d-none');
        } finally {
            loadingEl.classList.add('d-none');
            containerEl.style.opacity = '1';
        }
    }
}

// Initialize the dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.cronJobDashboard = new CronJobDashboard();
});

// Global refresh function for backward compatibility
function refreshData() {
    if (window.cronJobDashboard) {
        window.cronJobDashboard.refreshData();
    }
}
