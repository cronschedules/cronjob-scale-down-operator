package webui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	cronschedulesv1 "github.com/z4ck404/cronjob-scale-down-operator/api/v1"
)

func newTestServer(t *testing.T, allowCrossNamespace bool) *Server {
	t.Helper()

	// NewServer resolves ./web/static relative to the working directory (repository root).
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir("../.."); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = cronschedulesv1.AddToScheme(scheme)

	cr := &cronschedulesv1.CronJobScaleDown{
		ObjectMeta: metav1.ObjectMeta{Name: "scaler", Namespace: "tenant-a"},
		Spec: cronschedulesv1.CronJobScaleDownSpec{
			TargetRef: &cronschedulesv1.TargetRef{Name: "app", Namespace: "victim", Kind: "Deployment", ApiVersion: "apps/v1"},
			TimeZone:  "UTC",
		},
	}
	victim := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "victim"},
		Spec:       appsv1.DeploymentSpec{Replicas: ptr.To[int32](5)},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cr, victim).Build()
	return NewServer(c, allowCrossNamespace)
}

func TestSecurityHeaders(t *testing.T) {
	s := newTestServer(t, false)

	for _, path := range []string{"/", "/api/v1/cronjobs", "/static/dashboard.js"} {
		rec := httptest.NewRecorder()
		s.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

		csp := rec.Header().Get("Content-Security-Policy")
		if !strings.Contains(csp, "script-src 'self' https://cdn.jsdelivr.net;") || !strings.Contains(csp, "frame-ancestors 'none'") {
			t.Errorf("%s: unexpected Content-Security-Policy %q", path, csp)
		}
		if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Errorf("%s: X-Content-Type-Options = %q, want nosniff", path, got)
		}
		if got := rec.Header().Get("X-Frame-Options"); got != "DENY" {
			t.Errorf("%s: X-Frame-Options = %q, want DENY", path, got)
		}
	}
}

func TestGetCronJobNotFoundDoesNotLeakAPIError(t *testing.T) {
	s := newTestServer(t, false)

	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/cronjobs/tenant-a/missing", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if body := rec.Body.String(); strings.Contains(body, "cronjobscaledowns") {
		t.Errorf("response leaks API server error: %q", body)
	}
}

func TestCrossNamespaceTargetStatusIsScoped(t *testing.T) {
	tests := []struct {
		name                string
		allowCrossNamespace bool
		wantTargetStatus    bool
	}{
		{name: "cross-namespace target status hidden by default", allowCrossNamespace: false, wantTargetStatus: false},
		{name: "cross-namespace target status shown when allowed", allowCrossNamespace: true, wantTargetStatus: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(t, tt.allowCrossNamespace)

			rec := httptest.NewRecorder()
			s.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/cronjobs", nil))
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}

			var statuses []CronJobStatus
			if err := json.Unmarshal(rec.Body.Bytes(), &statuses); err != nil {
				t.Fatal(err)
			}
			if len(statuses) != 1 {
				t.Fatalf("got %d statuses, want 1", len(statuses))
			}
			if got := statuses[0].TargetStatus != nil; got != tt.wantTargetStatus {
				t.Errorf("targetStatus present = %v, want %v", got, tt.wantTargetStatus)
			}
		})
	}
}
