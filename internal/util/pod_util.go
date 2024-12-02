package util

import(
	"os"
	"strconv"
	"net"
	"context"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/go-credit/internal/core"
)

var childLogger = log.With().Str("util", "util").Logger()

func GetInfoPod() (	core.InfoPod,
					core.Server, 
					core.RestEndpoint,
					core.AwsServiceConfig) {
	childLogger.Debug().Msg("GetInfoPod")

	err := godotenv.Load(".env")
	if err != nil {
		childLogger.Info().Err(err).Msg("env file not found !!!")
	}

	var infoPod 	core.InfoPod
	var server		core.Server
	var restEndpoint core.RestEndpoint
	var awsServiceConfig core.AwsServiceConfig

	server.ReadTimeout = 60
	server.WriteTimeout = 60
	server.IdleTimeout = 60
	server.CtxTimeout = 60

	if os.Getenv("API_VERSION") !=  "" {
		infoPod.ApiVersion = os.Getenv("API_VERSION")
	}
	if os.Getenv("POD_NAME") !=  "" {
		infoPod.PodName = os.Getenv("POD_NAME")
	}
	if os.Getenv("SETPOD_AZ") == "false" {	
		infoPod.IsAZ = false
	} else {
		infoPod.IsAZ = true
	}
	if os.Getenv("ENV") !=  "" {	
		infoPod.Env = os.Getenv("ENV")
	}
	
	// Get IP
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Error().Err(err).Msg("error to get the POD IP address !!!")
		os.Exit(3)
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				infoPod.IPAddress = ipnet.IP.String()
			}
		}
	}
	infoPod.OSPID = strconv.Itoa(os.Getpid())

	// Get AZ only if localtest is true
	if (infoPod.IsAZ) {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			childLogger.Error().Err(err).Msg("fatal error get Context !!!")
			os.Exit(3)
		}
		client := imds.NewFromConfig(cfg)
		response, err := client.GetInstanceIdentityDocument(context.TODO(), &imds.GetInstanceIdentityDocumentInput{})
		if err != nil {
			childLogger.Error().Err(err).Msg("unable to retrieve the region from the EC2 instance !!!")
			os.Exit(3)
		}
		infoPod.AvailabilityZone = response.AvailabilityZone	
	} else {
		infoPod.AvailabilityZone = "-"
	}

	if os.Getenv("PORT") !=  "" {
		intVar, _ := strconv.Atoi(os.Getenv("PORT"))
		server.Port = intVar
	}

	if os.Getenv("SERVICE_URL_DOMAIN") !=  "" {	
		restEndpoint.ServiceUrlDomain = os.Getenv("SERVICE_URL_DOMAIN")
	}
	if os.Getenv("X_APIGW_API_ID") !=  "" {	
		restEndpoint.XApigwId = os.Getenv("X_APIGW_API_ID")
	}
	if os.Getenv("SERVICE_URL_DOMAIN_CB") !=  "" {	
		restEndpoint.ServiceUrlDomainCB = os.Getenv("SERVICE_URL_DOMAIN_CB")
	}
	if os.Getenv("X_APIGW_API_ID") !=  "" {	
		restEndpoint.XApigwIdCB = os.Getenv("X_APIGW_API_ID_CB")
	}

	if os.Getenv("SERVICE_URL_JWT_SA") !=  "" {	
		awsServiceConfig.ServiceUrlJwtSA = os.Getenv("SERVICE_URL_JWT_SA")
	}
	if os.Getenv("SECRET_JWT_SA_CREDENTIAL") !=  "" {	
		awsServiceConfig.SecretJwtSACredential = os.Getenv("SECRET_JWT_SA_CREDENTIAL")
	}
	if os.Getenv("AWS_REGION") !=  "" {
		awsServiceConfig.AwsRegion = os.Getenv("AWS_REGION")
	}

	return infoPod, server, restEndpoint, awsServiceConfig
}

func ConvertToDate(date_str string) (*time.Time, error){
	
	layout := "2006-01-02"

	date, err := time.Parse(layout, date_str)
	if err != nil {
		log.Error().Err(err).Msg("error parsing date !!!")
		return nil, err
	}

	return &date, nil
}