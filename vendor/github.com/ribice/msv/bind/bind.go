package bind

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Binder binds json
type Binder interface {
	Bind() error
}

// JSON binds request into interface v, and validates the request
//  If an error occurs during json unmarshalling, http 500 is responded to client
//  If an error occurs during binding json, http 400 is responded to client
func JSON(w http.ResponseWriter, r *http.Request, v Binder) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		http.Error(w, fmt.Sprintf("error decoding json: %v", err), 500)
		return err
	}
	if err := v.Bind(); err != nil {
		http.Error(w, fmt.Sprintf("error binding request: %v", err), 400)
		return err
	}
	return nil
}
