package restapi

import(
	"errors"
	"net/http"
	"time"
	"encoding/json"
	"bytes"
	"context"
	"crypto/x509"
	"crypto/tls"
	"encoding/base64"

	"github.com/go-credit/internal/lib"
	"github.com/rs/zerolog/log"
	"github.com/go-credit/internal/erro"
	"github.com/go-credit/internal/core"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var childLogger = log.With().Str("adapter/restapi", "restApiService").Logger()
//----------------------------------------------
type RestApiService struct {
}

func NewRestApiService(	) *RestApiService{
	childLogger.Debug().Msg("*** NewRestApiService")

	return &RestApiService {
	}
}
//----------------------------------------------
func loadClientCertsTLS(cert *core.Cert) (*tls.Config, error){
	childLogger.Debug().Msg("loadClientCertsTLS")

	caPEM_Raw, err := base64.StdEncoding.DecodeString(string(cert.CaPEM))
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro caPEM_Raw !!!")
		return nil, err
	}

	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(caPEM_Raw)

	clientTLSConf := &tls.Config{
		RootCAs: certpool,
	}

	return clientTLSConf ,nil
}

func (r *RestApiService) GetData(ctx context.Context, 
								urlDomain string, 
								xApigwId string,
								data interface{} ) (interface{}, error) {
	childLogger.Debug().Msg("GetData")
	
	span := lib.Span(ctx, "adapter.GetData")	
    defer span.End()

	data_interface, err := r.makeGet(ctx, urlDomain, http.MethodGet, xApigwId, data)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return nil, errors.New(err.Error())
	}
    
	return data_interface, nil
}

func (r *RestApiService) PostData(	ctx context.Context, 
									urlDomain string, 
									xApigwId string, 
									data interface{}) (interface{}, error) {
	childLogger.Debug().Msg("PostData")

	span := lib.Span(ctx, "adapter.PostData")	
    defer span.End()

	data_interface, err := r.makePost(ctx, urlDomain, http.MethodPost, xApigwId, data)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return nil, errors.New(err.Error())
	}
    
	return data_interface, nil
}

func (r *RestApiService) makeGet(	ctx context.Context, 
									url string, 
									method string,
									xApigwId string,
									data interface{}) (interface{}, error) {
	childLogger.Debug().Msg("makeGet")
	childLogger.Debug().Str("url : ", url).Msg("")
	childLogger.Debug().Str("xApigwId : ", xApigwId).Msg("")

	/*transport := &http.Transport{
		TLSClientConfig: r.ClientTLSConf,
	}
	client := &http.Client{Timeout: time.Second * 5 , Transport: transport}*/

	span, ctxSpan := lib.SpanCtx(ctx, "adapter.GetData.makeGet: " + url)	
    defer span.End()

	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequestWithContext(ctxSpan, method, url, nil)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return false, errors.New(err.Error())
	}
	req.Header.Add("Content-Type", "application/json;charset=UTF-8");
	req.Header.Add("x-apigw-api-id", xApigwId);

	resp, err := client.Do(req.WithContext(ctxSpan))
	if err != nil {
		childLogger.Error().Err(err).Msg("error Do Request")
		return false, errors.New(err.Error())
	}
	defer resp.Body.Close()

	childLogger.Debug().Int("StatusCode :", resp.StatusCode).Msg("")
	switch (resp.StatusCode) {
		case 401:
			return false, erro.ErrHTTPForbiden
		case 403:
			return false, erro.ErrHTTPForbiden
		case 200:
		case 400:
			return false, erro.ErrNotFound
		case 404:
			return false, erro.ErrNotFound
		default:
			return false, erro.ErrServer
	}

	result := data
	err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
		childLogger.Error().Err(err).Msg("error no ErrUnmarshal")
		return false, errors.New(err.Error())
    }

	return result, nil
}

func (r *RestApiService) makePost(	ctx context.Context, 
									url string, 
									method string,
									xApigwId string,
									data interface{}) (interface{}, error) {
	childLogger.Debug().Msg("makePost")
	childLogger.Debug().Str("url : ", url).Msg("")
	childLogger.Debug().Str("xApigwId : ", xApigwId).Msg("")

	/*transport := &http.Transport{
		TLSClientConfig: r.ClientTLSConf,
	})
	client := &http.Client{Timeout: time.Second * 5, Transport: transport}*/
	
	span := lib.Span(ctx, url)	
    defer span.End()

	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout: time.Second * 10,
	}

	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(data)

	req, err := http.NewRequestWithContext(ctx, method, url, payload)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return false, errors.New(err.Error())
	}

	req.Header.Add("Content-Type", "application/json;charset=UTF-8");
	req.Header.Add("x-apigw-api-id", xApigwId);
	
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		childLogger.Error().Err(err).Msg("error Do Request")
		return false, errors.New(err.Error())
	}
	defer resp.Body.Close()

	childLogger.Debug().Int("StatusCode :", resp.StatusCode).Msg("")
	switch (resp.StatusCode) {
		case 401:
			return false, erro.ErrHTTPForbiden
		case 403:
			return false, erro.ErrHTTPForbiden
		case 200:
		case 400:
			return false, erro.ErrNotFound
		case 404:
			return false, erro.ErrNotFound
		default:
			return false, erro.ErrHTTPForbiden
	}

	result := data
	err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
		childLogger.Error().Err(err).Msg("error no ErrUnmarshal")
		return false, errors.New(err.Error())
    }

	return result, nil
}