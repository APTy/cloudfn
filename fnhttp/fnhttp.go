package fnhttp

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/APTy/cloudfn/fnerrors"
)

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
	b, err := json.Marshal(res)
	if err != nil {
		WriteErr(w, err)
		return
	}
	fmt.Fprintf(w, "%s\n", b)
}

// WriteErr writes an fnerror with the correct status code.
func WriteErr(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if err, ok := err.(*fnerrors.Error); ok {
		status = err.HTTPStatus()
	}
	w.WriteHeader(status)
	fmt.Fprintf(w, "%v\n", err)
}

// GetPostData gets the POST data from the body.
func GetPostData(r *http.Request, ifcPtr interface{}) error {
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return nil
	}
	if err := json.Unmarshal(b, ifcPtr); err != nil {
		return err
	}
	return nil
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