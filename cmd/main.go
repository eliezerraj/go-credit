package main

import(
	"time"
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-credit/internal/circuitbreaker"
	"github.com/go-credit/internal/handler"
	"github.com/go-credit/internal/util"
	"github.com/go-credit/internal/core"
	"github.com/go-credit/internal/service"
	"github.com/go-credit/internal/repository/postgre"
	"github.com/go-credit/internal/adapter/restapi"
	
)

var(
	logLevel 	= 	zerolog.DebugLevel
	appServer	core.AppServer
)

func init(){
	log.Debug().Msg("init")
	zerolog.SetGlobalLevel(logLevel)

	infoPod , server, restEndpoint := util.GetInfoPod()
	database := util.GetDatabaseEnv()
	configOTEL := util.GetOtelEnv()
	cert := util.GetCertEnv()

	appServer.Cert = &cert
	appServer.InfoPod = &infoPod
	appServer.Database = &database
	appServer.Server = &server
	appServer.RestEndpoint = &restEndpoint
	appServer.Server.Cert = &cert
	appServer.ConfigOTEL = &configOTEL
}

func main() {
	log.Debug().Msg("----------------------------------------------------")
	log.Debug().Msg("main")
	log.Debug().Msg("----------------------------------------------------")
	log.Debug().Interface("appServer :",appServer).Msg("")
	log.Debug().Msg("----------------------------------------------------")

	ctx, cancel := context.WithTimeout(context.Background(), 
									time.Duration( appServer.Server.ReadTimeout ) * time.Second)
	defer cancel()

	// Open Database
	count := 1
	var databaseHelper	postgre.DatabaseHelper
	var err error
	for {
		databaseHelper, err = postgre.NewDatabaseHelper(ctx, appServer.Database)
		if err != nil {
			if count < 3 {
				log.Error().Err(err).Msg("Erro open Database... trying again !!")
			} else {
				log.Error().Err(err).Msg("Fatal erro open Database aborting")
				panic(err)
			}
			time.Sleep(3 * time.Second)
			count = count + 1
			continue
		}
		break
	}
	
	repoDB := postgre.NewWorkerRepository(databaseHelper)

	// Setup workload
	circuitBreaker := circuitbreaker.CircuitBreakerConfig()
	restApiService	:= restapi.NewRestApiService()
	workerService := service.NewWorkerService(	&repoDB, 
												appServer.RestEndpoint,
												restApiService, 
												circuitBreaker)

	httpWorkerAdapter := handler.NewHttpWorkerAdapter(workerService)
	httpServer := handler.NewHttpAppServer(appServer.Server)

	httpServer.StartHttpAppServer(	ctx, 
									&httpWorkerAdapter,
									&appServer)
}