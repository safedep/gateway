package main

import (
	"context"
	"log"
	"os"
	"strconv"

	api "github.com/safedep/gateway/services/gen"

	common_adapters "github.com/safedep/gateway/services/pkg/common/adapters"
	"github.com/safedep/gateway/services/pkg/common/config"
	"github.com/safedep/gateway/services/pkg/common/logger"
	"github.com/safedep/gateway/services/pkg/common/obs"
	"google.golang.org/grpc"

	"github.com/safedep/gateway/services/pkg/common/db"
	"github.com/safedep/gateway/services/pkg/common/db/adapters"
	"github.com/safedep/gateway/services/pkg/pds"
)

func main() {
	logger.Init("dcs")
	config.Bootstrap("", true)

	tracerShutDown := obs.InitTracing()
	defer tracerShutDown(context.Background())

	mysqlPort, err := strconv.ParseInt(os.Getenv("MYSQL_SERVER_PORT"), 0, 16)
	if err != nil {
		log.Fatalf("Failed to parse mysql server port: %v", err)
	}

	mysqlAdapter, err := adapters.NewMySqlAdapter(adapters.MySqlAdapterConfig{
		Host:     os.Getenv("MYSQL_SERVER_HOST"),
		Port:     int16(mysqlPort),
		Username: os.Getenv("MYSQL_USER"),
		Password: os.Getenv("MYSQL_PASSWORD"),
		Database: os.Getenv("MYSQL_DATABASE"),
	})

	if err != nil {
		log.Fatalf("Failed to initialize MySQL adapter: %v", err)
	}

	repository, err := db.NewVulnerabilityRepository(mysqlAdapter)
	if err != nil {
		log.Fatalf("Failed to create vulnerability repository")
	}

	pdService, err := pds.NewPolicyDataService(repository)
	if err != nil {
		log.Fatalf("Failed to create policy data service")
	}

	common_adapters.StartGrpcMtlsServer("PDS", os.Getenv("PDS_SERVER_NAME"), "0.0.0.0", "9002",
		[]grpc.ServerOption{grpc.MaxConcurrentStreams(5000)}, func(s *grpc.Server) {
			api.RegisterPolicyDataServiceServer(s, pdService)
		})
}
