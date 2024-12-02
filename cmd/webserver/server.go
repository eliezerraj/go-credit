package webserver

import(
	"time"
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-credit/internal/handler/controller"
	"github.com/go-credit/internal/circuitbreaker"
	"github.com/go-credit/internal/handler"
	"github.com/go-credit/internal/util"
	"github.com/go-credit/internal/core"
	"github.com/go-credit/internal/service"
	"github.com/go-credit/internal/repository/pg"
	"github.com/go-credit/internal/repository/storage"	
	"github.com/go-credit/internal/adapter/restapi"
)

var(
	logLevel 	= 	zerolog.DebugLevel
	appServer	core.AppServer
	restApiCallData core.RestApiCallData
)

func init(){
	log.Debug().Msg("init")
	zerolog.SetGlobalLevel(logLevel)

	infoPod , server, restEndpoint, awsServiceConfig := util.GetInfoPod()
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
	appServer.AwsServiceConfig = &awsServiceConfig
}

func Server() {
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
	var databasePG	pg.DatabasePG
	var err error
	for {
		databasePG, err = pg.NewDatabasePGServer(ctx, appServer.Database)
		if err != nil {
			if count < 3 {
				log.Error().Err(err).Msg("error open database... trying again !!")
			} else {
				log.Error().Err(err).Msg("fatal error open Database aborting")
				panic(err)
			}
			time.Sleep(3 * time.Second)
			count = count + 1
			continue
		}
		break
	}
	
	repoDatabase := storage.NewWorkerRepository(databasePG)
	// Setup workload
	circuitBreaker := circuitbreaker.CircuitBreakerConfig()
	restApiService	:= restapi.NewRestApiService(&appServer)
	workerService := service.NewWorkerService(&repoDatabase, &appServer, restApiService,circuitBreaker)

	httpWorkerAdapter := controller.NewHttpWorkerAdapter(workerService)
	httpServer := handler.NewHttpAppServer(appServer.Server)

	httpServer.StartHttpAppServer(ctx, &httpWorkerAdapter, &appServer)
}