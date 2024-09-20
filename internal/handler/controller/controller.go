package controller

import (
	"encoding/json"
	"net/http"
	"github.com/rs/zerolog/log"
	"github.com/gorilla/mux"

	"github.com/go-credit/internal/service"
	"github.com/go-credit/internal/core"
	"github.com/go-credit/internal/lib"
	"github.com/go-credit/internal/erro"
	"github.com/go-credit/internal/util"
)


var childLogger = log.With().Str("handler", "controller").Logger()

type HttpWorkerAdapter struct {
	workerService 	*service.WorkerService
}

func NewHttpWorkerAdapter(workerService *service.WorkerService) HttpWorkerAdapter {
	childLogger.Debug().Msg("NewHttpWorkerAdapter")
	
	return HttpWorkerAdapter{
		workerService: workerService,
	}
}

type APIError struct {
	StatusCode	int  `json:"statusCode"`
	Msg			any `json:"msg"`
}

func NewAPIError(statusCode int, err error) APIError {
	return APIError{
		StatusCode: statusCode,
		Msg:		err.Error(),
	}
}

func (h *HttpWorkerAdapter) Health(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Health")

	health := true
	json.NewEncoder(rw).Encode(health)
	return
}

func (h *HttpWorkerAdapter) Live(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Live")

	live := true
	json.NewEncoder(rw).Encode(live)
	return
}

func (h *HttpWorkerAdapter) Header(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Header")
	
	json.NewEncoder(rw).Encode(req.Header)
	return
}

func (h *HttpWorkerAdapter) Add( rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Add")

	span := lib.Span(req.Context(), "handler.Add")
	defer span.End()

	credit := core.AccountStatement{}
	err := json.NewDecoder(req.Body).Decode(&credit)
    if err != nil {
		apiError := NewAPIError(400, erro.ErrUnmarshal)
		rw.WriteHeader(apiError.StatusCode)
		json.NewEncoder(rw).Encode(apiError)
		return
    }

	res, err := h.workerService.Add(req.Context(), &credit)
	if err != nil {
		var apiError APIError
		switch err {
			case erro.ErrNotFound:
				apiError = NewAPIError(404, err)
			case erro.ErrTransInvalid:
				apiError = NewAPIError(409, err)
			default:
				apiError = NewAPIError(500, err)
		}
		rw.WriteHeader(apiError.StatusCode)
		json.NewEncoder(rw).Encode(apiError)
		return
	}

	json.NewEncoder(rw).Encode(res)
	return
}

func (h *HttpWorkerAdapter) List(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("List")

	span := lib.Span(req.Context(), "handler.List")
	defer span.End()

	vars := mux.Vars(req)
	varID := vars["id"]

	credit := core.AccountStatement{}
	credit.AccountID = varID
	
	res, err := h.workerService.List(req.Context(), &credit)
	if err != nil {
		var apiError APIError
		switch err {
			case erro.ErrNotFound:
				apiError = NewAPIError(404, err)
			default:
				apiError = NewAPIError(500, err)
		}
		rw.WriteHeader(apiError.StatusCode)
		json.NewEncoder(rw).Encode(apiError)
		return
	}

	json.NewEncoder(rw).Encode(res)
	return
}

func (h *HttpWorkerAdapter) ListPerDate(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("ListPerDate")

	span := lib.Span(req.Context(), "handler.ListPerDate")
	defer span.End()

	params := req.URL.Query()
	varAcc := params.Get("account")
	varDate := params.Get("date_start")

	log.Debug().Interface("******* >>>>> params :", params).Msg("")

	credit := core.AccountStatement{}
	credit.AccountID = varAcc

	convertDate, err := util.ConvertToDate(varDate)
	if err != nil {
		apiError := NewAPIError(400, erro.ErrUnmarshal)
		rw.WriteHeader(apiError.StatusCode)
		json.NewEncoder(rw).Encode(apiError)
		return
	}

	credit.ChargeAt = *convertDate

	res, err := h.workerService.ListPerDate(req.Context(), &credit)
	if err != nil {
		var apiError APIError
		switch err {
			case erro.ErrNotFound:
				apiError = NewAPIError(404, err)
			default:
				apiError = NewAPIError(500, err)
		}
		rw.WriteHeader(apiError.StatusCode)
		json.NewEncoder(rw).Encode(apiError)
		return
	}

	json.NewEncoder(rw).Encode(res)
	return
}