// SPDX-License-Identifier: Apache-2.0
package v1

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/kikudesuyo/error-agent/api/entity"
)

const maxErrorEventBodyBytes = 64 * 1024

type ErrorEventAcceptor interface {
	AcceptErrorEvent(context.Context, entity.PostErrorEventRequest) (entity.PostErrorEventResponse, error)
}

type ErrorEventHandler struct {
	service ErrorEventAcceptor
}

func NewErrorEventHandler(errorEventService ErrorEventAcceptor) *ErrorEventHandler {
	return &ErrorEventHandler{service: errorEventService}
}

// HandlePostErrorEvent validates and durably accepts a production error event.
func (handler *ErrorEventHandler) HandlePostErrorEvent(writer http.ResponseWriter, request *http.Request) {
	mediaType, _, mediaTypeErr := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if mediaTypeErr != nil || mediaType != "application/json" {
		http.Error(writer, "content type must be application/json", http.StatusUnsupportedMediaType)
		return
	}
	request.Body = http.MaxBytesReader(writer, request.Body, maxErrorEventBodyBytes)
	var requestData entity.PostErrorEventRequest
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestData); err != nil {
		http.Error(writer, "invalid event", http.StatusBadRequest)
		return
	}
	if err := ensureSingleJSONValue(decoder); err != nil {
		http.Error(writer, "invalid event", http.StatusBadRequest)
		return
	}
	if err := requestData.ValidateAndParse(); err != nil {
		http.Error(writer, "invalid event", http.StatusBadRequest)
		return
	}

	response, err := handler.service.AcceptErrorEvent(request.Context(), requestData)
	if err != nil {
		http.Error(writer, "event could not be accepted", http.StatusServiceUnavailable)
		return
	}
	writer.Header().Set("content-type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(writer).Encode(response)
}

func ensureSingleJSONValue(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON value")
	}
	return nil
}
