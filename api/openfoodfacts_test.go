package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewOFFClient(t *testing.T) {
	c := NewOFFClient()
	if c == nil || c.httpClient == nil {
		t.Error("expected non-nil OFFClient with httpClient")
	}
}

func TestParseShelfLife(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{name: "empty string", input: "", want: 0},
		{name: "no match", input: "best before end", want: 0},
		{name: "days", input: "30 days", want: 30},
		{name: "day singular", input: "1 day", want: 1},
		{name: "months", input: "6 months", want: 180},
		{name: "month singular", input: "1 month", want: 30},
		{name: "years", input: "2 years", want: 730},
		{name: "year singular", input: "1 year", want: 365},
		{name: "uppercase input", input: "12 MONTHS", want: 360},
		{name: "leading text", input: "best before: 3 months", want: 90},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseShelfLife(tc.input)
			if got != tc.want {
				t.Errorf("parseShelfLife(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}

// newOFFTestClient creates an OFFClient that redirects all requests to srv.
func newOFFTestClient(srv *httptest.Server) *OFFClient {
	host := strings.TrimPrefix(srv.URL, "http://")
	return &OFFClient{httpClient: &http.Client{
		Transport: &redirectTransport{targetHost: host},
	}}
}

func TestOFFLookup(t *testing.T) {
	makeOFFResponse := func(status int, name, categories, expirationDate string) []byte {
		payload := map[string]interface{}{
			"status": status,
			"product": map[string]string{
				"product_name":    name,
				"categories":      categories,
				"expiration_date": expirationDate,
			},
		}
		data, _ := json.Marshal(payload)
		return data
	}

	t.Run("product not found (status 0)", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(makeOFFResponse(0, "", "", ""))
		}))
		defer srv.Close()

		got, err := newOFFTestClient(srv).Lookup("000000000000")
		if err != nil {
			t.Fatal(err)
		}
		if got != nil {
			t.Errorf("expected nil for status=0, got %+v", got)
		}
	})

	t.Run("product found with name only", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(makeOFFResponse(1, "Organic Oats", "", ""))
		}))
		defer srv.Close()

		got, err := newOFFTestClient(srv).Lookup("012345678905")
		if err != nil {
			t.Fatal(err)
		}
		if got == nil {
			t.Fatal("expected non-nil product")
		}
		if got.Name != "Organic Oats" {
			t.Errorf("Name = %q, want %q", got.Name, "Organic Oats")
		}
		if got.ShelfLifeDays != nil {
			t.Errorf("expected nil ShelfLifeDays, got %d", *got.ShelfLifeDays)
		}
	})

	t.Run("product found with expiration date", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(makeOFFResponse(1, "Canned Beans", "", "24 months"))
		}))
		defer srv.Close()

		got, err := newOFFTestClient(srv).Lookup("012345678905")
		if err != nil {
			t.Fatal(err)
		}
		if got == nil {
			t.Fatal("expected non-nil product")
		}
		if got.ShelfLifeDays == nil {
			t.Fatal("expected non-nil ShelfLifeDays")
		}
		if *got.ShelfLifeDays != 720 {
			t.Errorf("ShelfLifeDays = %d, want 720", *got.ShelfLifeDays)
		}
	})

	t.Run("product found with category-based shelf life fallback", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(makeOFFResponse(1, "Premium Pasta", "pasta, grains", ""))
		}))
		defer srv.Close()

		got, err := newOFFTestClient(srv).Lookup("012345678905")
		if err != nil {
			t.Fatal(err)
		}
		if got == nil {
			t.Fatal("expected non-nil product")
		}
		if got.ShelfLifeDays == nil {
			t.Fatal("expected non-nil ShelfLifeDays from category fallback")
		}
		if *got.ShelfLifeDays != 365 {
			t.Errorf("ShelfLifeDays = %d, want 365", *got.ShelfLifeDays)
		}
	})

	t.Run("invalid JSON response returns error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Return plain text that isn't valid JSON
			w.Write([]byte("not json"))
		}))
		defer srv.Close()

		_, err := newOFFTestClient(srv).Lookup("000000000000")
		if err == nil {
			t.Error("expected error for non-JSON response, got nil")
		}
	})
}

// redirectTransport rewrites every request's host to the test server's host:port.
type redirectTransport struct {
	targetHost string // host:port only, e.g. "127.0.0.1:12345"
}

func (rt *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	cloned.URL.Scheme = "http"
	cloned.URL.Host = rt.targetHost
	return http.DefaultTransport.RoundTrip(cloned)
}
