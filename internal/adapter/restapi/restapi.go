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
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/go-credit/internal/erro"
	"github.com/go-credit/internal/core"
	"github.com/aws/aws-xray-sdk-go/xray"
)

var childLogger = log.With().Str("adapter/restapi", "restapi").Logger()

type RestApiSConfig struct {
	ServerUrlDomain			string
	XApigwId				string
	ClientTLSConf 			*tls.Config
}

func NewRestApi(serverUrlDomain string, 
				xApigwId string,
				cert	*core.Cert) (*RestApiSConfig){
	childLogger.Debug().Msg("*** NewRestApi")

	fmt.Println(string(cert.CaPEM))

	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(cert.CaPEM)
	clientTLSConf := &tls.Config{
		RootCAs: certpool,
	}

	return &RestApiSConfig {
		ServerUrlDomain: 	serverUrlDomain,
		XApigwId: 			xApigwId,
		ClientTLSConf:		clientTLSConf,
	}
}

func (r *RestApiSConfig) GetData(ctx context.Context, serverUrlDomain string, xApigwId string ,path string , id string ) (interface{}, error) {
	childLogger.Debug().Msg("GetData")

	domain := serverUrlDomain + path +"/" + id

	data_interface, err := r.makeGet(ctx, domain, xApigwId ,id)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return nil, errors.New(err.Error())
	}
    
	return data_interface, nil
}

func (r *RestApiSConfig) PostData(ctx context.Context, serverUrlDomain string, xApigwId string, path string, data interface{}) (interface{}, error) {
	childLogger.Debug().Msg("PostData")

	domain := serverUrlDomain + path 

	data_interface, err := r.makePost(ctx, domain, xApigwId ,data)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return nil, errors.New(err.Error())
	}
    
	return data_interface, nil
}

func (r *RestApiSConfig) makeGet(ctx context.Context, url string, xApigwId string ,id interface{}) (interface{}, error) {
	childLogger.Debug().Msg("makeGet")
	childLogger.Debug().Str("url : ", url).Msg("")
	childLogger.Debug().Str("xApigwId : ", xApigwId).Msg("")

	/*transport := &http.Transport{
		TLSClientConfig: r.ClientTLSConf,
	}
	client := xray.Client(&http.Client{Timeout: time.Second * 5 , Transport: transport})*/
	client := xray.Client(&http.Client{Timeout: time.Second * 5})

	req, err := http.NewRequest("GET", url, nil)
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

	result := id
	err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
		childLogger.Error().Err(err).Msg("error no ErrUnmarshal")
		return false, errors.New(err.Error())
    }

	return result, nil
}

func (r *RestApiSConfig) makePost(ctx context.Context, url string, xApigwId string ,data interface{}) (interface{}, error) {
	childLogger.Debug().Msg("makePost")
	childLogger.Debug().Str("url : ", url).Msg("")
	childLogger.Debug().Str("xApigwId : ", xApigwId).Msg("")

	/*transport := &http.Transport{
		TLSClientConfig: r.ClientTLSConf,
	})
	client := xray.Client(&http.Client{Timeout: time.Second * 5, Transport: transport})*/

	client := xray.Client(&http.Client{Timeout: time.Second * 5})
	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(data)

	req, err := http.NewRequest("POST", url, payload)
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
