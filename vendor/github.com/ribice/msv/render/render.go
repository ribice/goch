package render

import (
	"encoding/json"
	"net/http"
)

// JSON marshals 'v' to JSON, automatically escaping HTML and setting the
// Content-Type as application/json.
func JSON(w http.ResponseWriter, r *http.Request, v interface{}) {
	bts, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write(bts)
}
