package webui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cronschedulesv1 "github.com/z4ck404/cronjob-scale-down-operator/api/v1"
)

type Server struct {
	client              client.Client
	router              *mux.Router
	allowCrossNamespace bool
}

type CronJobStatus struct {
	Name              string         `json:"name"`
	Namespace         string         `json:"namespace"`
	TargetRef         *TargetRefInfo `json:"targetRef,omitempty"`
	ScaleDownSchedule string         `json:"scaleDownSchedule,omitempty"`
	ScaleUpSchedule   string         `json:"scaleUpSchedule,omitempty"`
	CleanupSchedule   string         `json:"cleanupSchedule,omitempty"`
	TimeZone          string         `json:"timeZone"`
	LastScaleDownTime *time.Time     `json:"lastScaleDownTime,omitempty"`
	LastScaleUpTime   *time.Time     `json:"lastScaleUpTime,omitempty"`
	LastCleanupTime   *time.Time     `json:"lastCleanupTime,omitempty"`
	CurrentReplicas   int32          `json:"currentReplicas"`
	TargetStatus      *TargetStatus  `json:"targetStatus,omitempty"`
	IsCleanupOnly     bool           `json:"isCleanupOnly"`
}

type TargetRefInfo struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Kind       string `json:"kind"`
	ApiVersion string `json:"apiVersion"`
}

type TargetStatus struct {
	Ready             bool       `json:"ready"`
	DesiredReplicas   int32      `json:"desiredReplicas"`
	AvailableReplicas int32      `json:"availableReplicas"`
	ReadyReplicas     int32      `json:"readyReplicas"`
	LastUpdateTime    *time.Time `json:"lastUpdateTime,omitempty"`
}

// NewServer creates the dashboard server. When allowCrossNamespace is false, target workload status is
// only read for targets in the CronJobScaleDown's own namespace, matching the controller's policy.
func NewServer(client client.Client, allowCrossNamespace bool) *Server {
	s := &Server{
		client:              client,
		router:              mux.NewRouter(),
		allowCrossNamespace: allowCrossNamespace,
	}
	s.setupRoutes()
	return s
}

// contentSecurityPolicy restricts the dashboard to its own scripts and the pinned CDN assets,
// so that any markup injected through CronJobScaleDown fields cannot execute script.
const contentSecurityPolicy = "default-src 'none'; " +
	"script-src 'self' https://cdn.jsdelivr.net; " +
	"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com https://fonts.googleapis.com; " +
	"font-src https://cdnjs.cloudflare.com https://fonts.gstatic.com; " +
	"img-src 'self' data:; " +
	"connect-src 'self'; " +
	"base-uri 'none'; form-action 'none'; frame-ancestors 'none'"

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", contentSecurityPolicy)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) setupRoutes() {
	s.router.Use(securityHeaders)

	// API routes
	api := s.router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/cronjobs", s.getCronJobs).Methods("GET")
	api.HandleFunc("/cronjobs/{namespace}/{name}", s.getCronJob).Methods("GET")

	// Static files and UI
	staticDir := "./web/static/"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		// Try absolute path from root (container environment)
		staticDir = "/web/static/"
		if _, err := os.Stat(staticDir); os.IsNotExist(err) {
			panic("Static files directory not found in ./web/static/ or /web/static/")
		}
	}

	absStaticDir, err := filepath.Abs(staticDir)
	if err != nil {
		panic(fmt.Sprintf("Failed to determine absolute path for static files: %v", err))
	}
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(absStaticDir))))
	s.router.HandleFunc("/", s.serveUI).Methods("GET")
}

func (s *Server) getCronJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := log.FromContext(ctx)

	var cronJobList cronschedulesv1.CronJobScaleDownList
	if err := s.client.List(ctx, &cronJobList); err != nil {
		log.Error(err, "Failed to list CronJobScaleDown resources")
		http.Error(w, "Failed to list cron jobs", http.StatusInternalServerError)
		return
	}

	statuses := make([]CronJobStatus, 0, len(cronJobList.Items))
	for _, cronJob := range cronJobList.Items {
		status, err := s.buildCronJobStatus(ctx, &cronJob)
		if err != nil {
			log.Error(err, "Failed to build status for cron job", "name", cronJob.Name, "namespace", cronJob.Namespace)
			continue
		}
		statuses = append(statuses, *status)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(statuses); err != nil {
		log.Error(err, "Failed to encode response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) getCronJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	name := vars["name"]

	ctx := r.Context()
	log := log.FromContext(ctx)

	var cronJob cronschedulesv1.CronJobScaleDown
	if err := s.client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &cronJob); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "Cron job not found", http.StatusNotFound)
			return
		}
		log.Error(err, "Failed to get CronJobScaleDown", "name", name, "namespace", namespace)
		http.Error(w, "Failed to get cron job", http.StatusInternalServerError)
		return
	}

	status, err := s.buildCronJobStatus(ctx, &cronJob)
	if err != nil {
		log.Error(err, "Failed to build status for cron job", "name", name, "namespace", namespace)
		http.Error(w, "Failed to build status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Error(err, "Failed to encode response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) buildCronJobStatus(ctx context.Context, cronJob *cronschedulesv1.CronJobScaleDown) (*CronJobStatus, error) { //nolint:unparam // error return kept for future extensibility
	log := log.FromContext(ctx)

	status := &CronJobStatus{
		Name:              cronJob.Name,
		Namespace:         cronJob.Namespace,
		ScaleDownSchedule: cronJob.Spec.ScaleDownSchedule,
		ScaleUpSchedule:   cronJob.Spec.ScaleUpSchedule,
		CleanupSchedule:   cronJob.Spec.CleanupSchedule,
		TimeZone:          cronJob.Spec.TimeZone,
		CurrentReplicas:   cronJob.Status.CurrentReplicas,
		IsCleanupOnly:     cronJob.Spec.TargetRef == nil && cronJob.Spec.CleanupSchedule != "",
	}

	// Handle TargetRef only if it exists (not for cleanup-only resources)
	if cronJob.Spec.TargetRef != nil {
		status.TargetRef = &TargetRefInfo{
			Name:       cronJob.Spec.TargetRef.Name,
			Namespace:  cronJob.Spec.TargetRef.Namespace,
			Kind:       cronJob.Spec.TargetRef.Kind,
			ApiVersion: cronJob.Spec.TargetRef.ApiVersion,
		}
	}

	if !cronJob.Status.LastScaleDownTime.IsZero() {
		status.LastScaleDownTime = &cronJob.Status.LastScaleDownTime.Time
	}
	if !cronJob.Status.LastScaleUpTime.IsZero() {
		status.LastScaleUpTime = &cronJob.Status.LastScaleUpTime.Time
	}
	if !cronJob.Status.LastCleanupTime.IsZero() {
		status.LastCleanupTime = &cronJob.Status.LastCleanupTime.Time
	}

	// Get target resource status only for scaling resources within the allowed namespace scope
	if cronJob.Spec.TargetRef != nil && (s.allowCrossNamespace || cronJob.Spec.TargetRef.Namespace == cronJob.Namespace) {
		targetStatus, err := s.getTargetStatus(ctx, *cronJob.Spec.TargetRef)
		if err != nil {
			// Log as info level for missing target resources (expected during normal operation)
			// This allows the web UI to show partial status even when target resources don't exist
			if apierrors.IsNotFound(err) {
				log.V(1).Info("Target resource not found, skipping target status", "targetRef", cronJob.Spec.TargetRef)
			} else {
				log.Error(err, "Failed to get target status, skipping target status", "targetRef", cronJob.Spec.TargetRef)
			}
		} else {
			status.TargetStatus = targetStatus
		}
	}

	return status, nil
}

func (s *Server) getTargetStatus(ctx context.Context, targetRef cronschedulesv1.TargetRef) (*TargetStatus, error) {
	switch targetRef.Kind {
	case "Deployment":
		return s.getDeploymentStatus(ctx, targetRef)
	case "StatefulSet":
		return s.getStatefulSetStatus(ctx, targetRef)
	default:
		return nil, fmt.Errorf("unsupported target kind: %s", targetRef.Kind)
	}
}

func (s *Server) getDeploymentStatus(ctx context.Context, targetRef cronschedulesv1.TargetRef) (*TargetStatus, error) {
	var deployment appsv1.Deployment
	if err := s.client.Get(ctx, types.NamespacedName{Name: targetRef.Name, Namespace: targetRef.Namespace}, &deployment); err != nil {
		return &TargetStatus{Ready: false}, err
	}

	status := &TargetStatus{
		Ready:             deployment.Status.ReadyReplicas == deployment.Status.Replicas,
		DesiredReplicas:   ptr.Deref(deployment.Spec.Replicas, 0),
		AvailableReplicas: deployment.Status.AvailableReplicas,
		ReadyReplicas:     deployment.Status.ReadyReplicas,
	}

	for _, condition := range deployment.Status.Conditions {
		if condition.Type == appsv1.DeploymentProgressing {
			status.LastUpdateTime = &condition.LastUpdateTime.Time
			break
		}
	}

	return status, nil
}

func (s *Server) getStatefulSetStatus(ctx context.Context, targetRef cronschedulesv1.TargetRef) (*TargetStatus, error) {
	var statefulset appsv1.StatefulSet
	if err := s.client.Get(ctx, types.NamespacedName{Name: targetRef.Name, Namespace: targetRef.Namespace}, &statefulset); err != nil {
		return &TargetStatus{Ready: false}, err
	}

	status := &TargetStatus{
		Ready:             statefulset.Status.ReadyReplicas == statefulset.Status.Replicas,
		DesiredReplicas:   ptr.Deref(statefulset.Spec.Replicas, 0),
		AvailableReplicas: statefulset.Status.Replicas,
		ReadyReplicas:     statefulset.Status.ReadyReplicas,
	}

	if statefulset.Status.ObservedGeneration > 0 {
		// Use creation timestamp as a fallback for last update time
		status.LastUpdateTime = &statefulset.CreationTimestamp.Time
	}

	return status, nil
}

func (s *Server) serveUI(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CronJob Scale Down Operator - Dashboard</title>
    <link rel="icon" type="image/svg+xml" href="/static/favicon.svg">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-1BmE4kWBq78iYhFldvKuhfTAU6auU8tT94WrHftjDbrCEXSU1oBoqyl2QvZ6jIW3" crossorigin="anonymous">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" rel="stylesheet" integrity="sha384-3B6NwesSXE7YJlcLI9RpRqGf2p/EgVH8BgoKTaUrmKNDkHPStTQ3EyoYjCGXaOTS" crossorigin="anonymous">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
    <link href="/static/styles.css" rel="stylesheet">
</head>
<body>
    <nav class="navbar navbar-expand-lg">
        <div class="container-fluid">
            <a href="/" class="navbar-brand text-decoration-none">
                <img src="/static/logo.png" alt="Logo" class="logo">
                <div class="navbar-brand-text">
                    <div class="navbar-brand-title">CronJob Scale Down Operator</div>
                    <div class="navbar-brand-subtitle">Kubernetes Resource Scheduler</div>
                </div>
            </a>
            <div class="d-flex align-items-center gap-3">
                <span class="navbar-text">
                    <i class="fas fa-sync-alt"></i> Last updated: <span id="last-updated">Never</span>
                    <span class="auto-refresh-indicator">(Auto-refresh: 30s)</span>
                </span>
                <button class="btn btn-outline-primary btn-sm" data-action="refresh">
                    <i class="fas fa-sync-alt me-1"></i> Refresh
                </button>
            </div>
        </div>
    </nav>

    <div class="container-fluid">
        <div id="loading" class="text-center d-none">
            <div class="spinner-border loading-pulse" role="status">
                <span class="visually-hidden">Loading...</span>
            </div>
            <p class="mt-3 text-muted">Loading CronJob resources...</p>
        </div>

        <div id="error-alert" class="alert alert-danger d-none" role="alert">
            <i class="fas fa-exclamation-triangle me-2"></i>
            <span id="error-message"></span>
        </div>

        <div id="cronjobs-container" class="row">
            <!-- CronJob cards will be inserted here -->
        </div>
    </div>

    <button class="btn btn-primary refresh-btn" data-action="refresh" title="Refresh Dashboard">
        <i class="fas fa-sync-alt"></i>
    </button>

    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js" integrity="sha384-ka7Sk0Gln4gmtz2MlQnikT1wXgYsOg+OMhuP+IlRH9sENBO0LRn5q+8nbTov4+1p" crossorigin="anonymous"></script>
    <script src="/static/dashboard.js"></script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(html)); err != nil {
		log := log.Log.WithName("webui")
		log.Error(err, "Failed to write response")
	}
}

func (s *Server) Start(addr string) error {
	log := log.Log.WithName("webui")
	log.Info("Starting web UI server", "address", addr)
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Error(err, "Failed to start web UI server", "address", addr, "possible causes", "port conflict, insufficient permissions, or invalid address")
	}
	return err
}
