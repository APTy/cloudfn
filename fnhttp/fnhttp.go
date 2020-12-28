package fnhttp

import (
	"bytes"
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/APTy/cloudfn/fnerrors"
)

// FnHttper performs cloud function http duties.
type FnHttper struct {
	// CORSOrigins is a list of allowed origin domains for CORS headers.
	CORSOrigins []string
	corsOrigins map[string]struct{}
}

// CORSMiddleware is middleware that handles CORS requests.
func (fn *FnHttper) CORSMiddleware(w http.ResponseWriter, r *http.Request) bool {
	if len(fn.CORSOrigins) == 0 {
		panic("fnhttp: missing allowed cors origins")
	}
	// default to first origin in the list
	origin := fn.CORSOrigins[0]

	// construct map for faster lookups
	if fn.corsOrigins == nil {
		fn.corsOrigins = make(map[string]struct{})
	}

	// use the origin if it's in our list of allowed origins
	if _, ok := fn.corsOrigins[r.Header.Get("Origin")]; ok {
		origin = r.Header.Get("Origin")
	}

	// look through the cors origins list, and add it to the map if we find it
	for _, allowedOrigin := range fn.CORSOrigins {
		if strings.Contains(r.Header.Get("Origin"), allowedOrigin) {
			origin = r.Header.Get("Origin")
			fn.corsOrigins[origin] = struct{}{}
		}
	}

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Max-Age", "3600")
	w.Header().Set("Vary", "Origin")

	// Return early for the preflight request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return true
	}

	return false
}

// NewCtx returns the context of the request.
func NewCtx(r *http.Request) context.Context {
	// TODO(tyler): add support for extracting auth headers.
	return r.Context()
}

// WriteRes writes the response unless err is non-nil.
func WriteRes(w http.ResponseWriter, res interface{}, err error) {
	if err != nil {
		WriteErr(w, err)
		return
	}

	var b []byte
	if marshaler, ok := res.(encoding.BinaryMarshaler); ok {
		b, err = marshaler.MarshalBinary()
	} else {
		b, err = json.Marshal(res)
	}
	if err != nil {
		WriteErr(w, err)
		return
	}
	fmt.Fprintf(w, "%s\n", b)
}

// WriteErr writes an fnerror with the correct status code.
func WriteErr(w http.ResponseWriter, err error) {
	msg := err.Error()
	status := http.StatusInternalServerError
	if err, ok := err.(*fnerrors.Error); ok {
		status = err.HTTPStatus()
		msg = err.JSONResponse()
	}
	w.WriteHeader(status)
	fmt.Fprintf(w, "%v\n", msg)
}

// GetPostData gets the POST data from the body.
func GetPostData(r *http.Request, ifcPtr interface{}) error {
	body := r.Body
	defer body.Close()
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return fnerrors.NewBadRequest("read body", err)
	}
	if len(b) == 0 {
		return nil
	}

	// copy bytes back to the request so it can be read again later
	r.Body = ioutil.NopCloser(bytes.NewBuffer(b))

	if err := json.Unmarshal(b, ifcPtr); err != nil {
		return fnerrors.NewBadRequest("json decode", err)
	}
	return nil
}

// GetPathID returns an id from the path of form "/foo/<id>"
func GetPathID(r *http.Request) string {
	return strings.TrimPrefix(r.URL.Path, "/")
}

// HandleOptionsRequestAndCORS responds with CORS headers to OPTIONS requests, and sets the appropriate headers otherwise.
func HandleOptionsRequestAndCORS(w http.ResponseWriter, r *http.Request) bool {
	// Set CORS headers for the preflight request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	return false
}
