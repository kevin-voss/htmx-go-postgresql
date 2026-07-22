package request

import "net/http"

// HTMX 4 sends HX-Request-Type to distinguish navigation kinds:
//
//	HX-Request-Type: partial  — targeted fragment swap (return HTML fragment only)
//	HX-Request-Type: full     — full-page / boosted navigation (return full layout)
//
// Absent header means a normal browser navigation; treat that as a full page.
const headerHXRequestType = "HX-Request-Type"

// IsPartialRequest reports whether r is an HTMX 4 partial (fragment) request.
func IsPartialRequest(r *http.Request) bool {
	return r.Header.Get(headerHXRequestType) == "partial"
}
