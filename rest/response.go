package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"
)

func writeJSONSuccessResponse(w http.ResponseWriter, r *http.Request, statusCode int, body interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if body != nil {
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(true)

		if err := enc.Encode(body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func writeJSONErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err != nil {
		preparedError := prepareErrorDetails(err, statusCode)
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(true)

		if err := enc.Encode(preparedError); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func prepareErrorDetails(err error, errorCode int) app.ErrorBody {
	var result app.ErrorBody
	result.Error.ErrorCode = strconv.Itoa(errorCode)
	result.Error.ErrorMessage = err.Error()
	if newErr, ok := err.(*exterr.Error); ok {
		result.Error.ErrorMessage = newErr.Error()
	}
	return result
}

func writeBinaryDataResponse(w http.ResponseWriter, r *http.Request, data []byte, filename string) {
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("Content-Type", "application/octet-stream;")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Write(data)
}
