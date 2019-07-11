package render

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Binder binds json
type Binder interface {
	Bind() error
}

// Bind binds JSON request into interface v, and validates the request
//  If an error occurs during json unmarshalling, http 500 is responded to client
//  If an error occurs during binding json, http 400 is responded to client
func Bind(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		http.Error(w, fmt.Sprintf("error decoding json: %v", err), 500)
		return err
	}

	if binder, ok := interface{}(v).(Binder); ok {
		if err := binder.Bind(); err != nil {
			http.Error(w, fmt.Sprintf("error binding request: %v", err), 400)
			return err
		}
	}
	return nil
}
