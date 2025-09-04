package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestTrustedSubnetMiddleware_EmptyCIDR_Allows(t *testing.T) {
	mw := TrustedSubnetMiddleware("")
	req := httptest.NewRequest(http.MethodPost, "/updates/", nil)
	req.Header.Set("X-Real-IP", "203.0.113.10")
	rr := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestTrustedSubnetMiddleware_ValidCIDR_AllowInside(t *testing.T) {
	mw := TrustedSubnetMiddleware("10.0.0.0/8")
	req := httptest.NewRequest(http.MethodPost, "/updates/", nil)
	req.Header.Set("X-Real-IP", "10.1.2.3")
	rr := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestTrustedSubnetMiddleware_ValidCIDR_RejectOutside(t *testing.T) {
	mw := TrustedSubnetMiddleware("192.168.0.0/16")
	req := httptest.NewRequest(http.MethodPost, "/updates/", nil)
	req.Header.Set("X-Real-IP", "10.1.2.3")
	rr := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestTrustedSubnetMiddleware_MissingOrBadIP_Rejects(t *testing.T) {
	mw := TrustedSubnetMiddleware("10.0.0.0/8")
	// Missing header
	req1 := httptest.NewRequest(http.MethodPost, "/updates/", nil)
	rr1 := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusForbidden {
		t.Fatalf("missing header status=%d", rr1.Code)
	}
	// Bad IP
	req2 := httptest.NewRequest(http.MethodPost, "/updates/", nil)
	req2.Header.Set("X-Real-IP", "not-an-ip")
	rr2 := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusForbidden {
		t.Fatalf("bad ip status=%d", rr2.Code)
	}
}
