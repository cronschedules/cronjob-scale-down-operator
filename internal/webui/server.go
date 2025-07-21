package webui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cronschedulesv1 "github.com/z4ck404/cronjob-scale-down-operator/api/v1"
)

type Server struct {
	client client.Client
	router *mux.Router
}

type CronJobStatus struct {
	Name              string        `json:"name"`
	Namespace         string        `json:"namespace"`
	TargetRef         TargetRefInfo `json:"targetRef"`
	ScaleDownSchedule string        `json:"scaleDownSchedule"`
	ScaleUpSchedule   string        `json:"scaleUpSchedule"`
	TimeZone          string        `json:"timeZone"`
	LastScaleDownTime *time.Time    `json:"lastScaleDownTime,omitempty"`
	LastScaleUpTime   *time.Time    `json:"lastScaleUpTime,omitempty"`
	CurrentReplicas   int32         `json:"currentReplicas"`
	TargetStatus      TargetStatus  `json:"targetStatus"`
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

func NewServer(client client.Client) *Server {
	s := &Server{
		client: client,
		router: mux.NewRouter(),
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// API routes
	api := s.router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/cronjobs", s.getCronJobs).Methods("GET")
	api.HandleFunc("/cronjobs/{namespace}/{name}", s.getCronJob).Methods("GET")

	// Static files and UI
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))
	s.router.HandleFunc("/", s.serveUI).Methods("GET")
}

func (s *Server) getCronJobs(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	log := log.FromContext(ctx)

	var cronJobList cronschedulesv1.CronJobScaleDownList
	if err := s.client.List(ctx, &cronJobList); err != nil {
		log.Error(err, "Failed to list CronJobScaleDown resources")
		http.Error(w, fmt.Sprintf("Failed to list cron jobs: %v", err), http.StatusInternalServerError)
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

	ctx := context.Background()
	log := log.FromContext(ctx)

	var cronJob cronschedulesv1.CronJobScaleDown
	if err := s.client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &cronJob); err != nil {
		log.Error(err, "Failed to get CronJobScaleDown", "name", name, "namespace", namespace)
		http.Error(w, fmt.Sprintf("Failed to get cron job: %v", err), http.StatusNotFound)
		return
	}

	status, err := s.buildCronJobStatus(ctx, &cronJob)
	if err != nil {
		log.Error(err, "Failed to build status for cron job", "name", name, "namespace", namespace)
		http.Error(w, fmt.Sprintf("Failed to build status: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Error(err, "Failed to encode response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) buildCronJobStatus(ctx context.Context, cronJob *cronschedulesv1.CronJobScaleDown) (*CronJobStatus, error) {
	status := &CronJobStatus{
		Name:              cronJob.Name,
		Namespace:         cronJob.Namespace,
		ScaleDownSchedule: cronJob.Spec.ScaleDownSchedule,
		ScaleUpSchedule:   cronJob.Spec.ScaleUpSchedule,
		TimeZone:          cronJob.Spec.TimeZone,
		CurrentReplicas:   cronJob.Status.CurrentReplicas,
		TargetRef: TargetRefInfo{
			Name:       cronJob.Spec.TargetRef.Name,
			Namespace:  cronJob.Spec.TargetRef.Namespace,
			Kind:       cronJob.Spec.TargetRef.Kind,
			ApiVersion: cronJob.Spec.TargetRef.ApiVersion,
		},
	}

	if !cronJob.Status.LastScaleDownTime.IsZero() {
		status.LastScaleDownTime = &cronJob.Status.LastScaleDownTime.Time
	}
	if !cronJob.Status.LastScaleUpTime.IsZero() {
		status.LastScaleUpTime = &cronJob.Status.LastScaleUpTime.Time
	}

	// Get target resource status
	targetStatus, err := s.getTargetStatus(ctx, cronJob.Spec.TargetRef)
	if err != nil {
		return nil, fmt.Errorf("failed to get target status: %v", err)
	}
	status.TargetStatus = *targetStatus

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
		DesiredReplicas:   *deployment.Spec.Replicas,
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
		DesiredReplicas:   *statefulset.Spec.Replicas,
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
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" rel="stylesheet">
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

    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js"></script>
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
	err := http.ListenAndServe(addr, s.router)
	if err != nil {
		log.Error(err, "Failed to start web UI server", "address", addr, "possible causes", "port conflict, insufficient permissions, or invalid address")
	}
	return err
}
