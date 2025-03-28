package api

import (
	"encoding/json"
	"net/http"
	"github.com/rs/zerolog/log"
	"github.com/go-credit/internal/core/service"
	"github.com/go-credit/internal/core/model"
	"github.com/go-credit/internal/core/erro"
	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_tools "github.com/eliezerraj/go-core/tools"
	"github.com/eliezerraj/go-core/coreJson"
	"github.com/gorilla/mux"
)

var childLogger = log.With().Str("component", "go-credit").Str("package", "internal.adapter.api").Logger()

var core_json coreJson.CoreJson
var core_apiError coreJson.APIError
var core_tools go_core_tools.ToolsCore
var tracerProvider go_core_observ.TracerProvider

type HttpRouters struct {
	workerService 	*service.WorkerService
}

func NewHttpRouters(workerService *service.WorkerService) HttpRouters {
	childLogger.Info().Str("func","NewHttpRouters").Send()

	return HttpRouters{
		workerService: workerService,
	}
}

// About return a health
func (h *HttpRouters) Health(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Health").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	json.NewEncoder(rw).Encode(model.MessageRouter{Message: "true"})
}

// About return a live
func (h *HttpRouters) Live(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Live").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	json.NewEncoder(rw).Encode(model.MessageRouter{Message: "true"})
}

// About show all header received
func (h *HttpRouters) Header(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Header").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()	
	
	json.NewEncoder(rw).Encode(req.Header)
}

// About add credit 
func (h *HttpRouters) AddCredit(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","AddCredit").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	//trace
	span := tracerProvider.Span(req.Context(), "adapter.api.AddCredit")
	defer span.End()

	// prepare body
	credit := model.AccountStatement{}
	err := json.NewDecoder(req.Body).Decode(&credit)
    if err != nil {
		core_apiError = core_apiError.NewAPIError(err, http.StatusBadRequest)
		return &core_apiError
    }
	defer req.Body.Close()

	//call service
	res, err := h.workerService.AddCredit(req.Context(), &credit)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, http.StatusNotFound)
		case erro.ErrTransInvalid:
			core_apiError = core_apiError.NewAPIError(err, http.StatusConflict)
		case erro.ErrInvalidAmount:
			core_apiError = core_apiError.NewAPIError(err, http.StatusConflict)
		case erro.ErrInvalidAmount:
			core_apiError = core_apiError.NewAPIError(err, http.StatusConflict)	
		default:
			core_apiError = core_apiError.NewAPIError(err, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About list all credit 
func (h *HttpRouters) ListCredit(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","ListCredit").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	// trace
	span := tracerProvider.Span(req.Context(), "adapter.api.ListCredit")
	defer span.End()

	//parameters
	vars := mux.Vars(req)
	varID := vars["id"]

	credit := model.AccountStatement{}
	credit.AccountID = varID

	// call service
	res, err := h.workerService.ListCredit(req.Context(), &credit)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About list all credits per date
func (h *HttpRouters) ListCreditPerDate(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","ListCreditPerDate").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	//Trace
	span := tracerProvider.Span(req.Context(), "adapter.api.ListCreditPerDate")
	defer span.End()

	// parameter
	params := req.URL.Query()
	varAcc := params.Get("account")
	varDate := params.Get("date_start")

	credit := model.AccountStatement{}
	credit.AccountID = varAcc

	convertDate, err := core_tools.ConvertToDate(varDate)
	if err != nil {
		core_apiError = core_apiError.NewAPIError(erro.ErrUnmarshal, http.StatusBadRequest)
		return &core_apiError
	}
	credit.ChargeAt = *convertDate

	//service
	res, err := h.workerService.ListCreditPerDate(req.Context(), &credit)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}