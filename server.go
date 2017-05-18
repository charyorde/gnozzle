package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"golang.org/x/net/context"

	"github.com/golang/protobuf/proto"

	"github.com/cloudfoundry/noaa/consumer"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

	pb "github.com/charyorde/gnozzle/gnozzle"
)

var token = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImtleS0xIiwidHlwIjoiSldUIn0.eyJqdGkiOiIxMTM5YjFjNDJkNjA0Yzk1YWUwMDcxMzE0NGI3OGUyZCIsInN1YiI6IjQ1MTZmOTBhLTE4YTQtNGI0Ni1hOWU2LWVjYTcwNzhkN2RkNiIsInNjb3BlIjpbInJvdXRpbmcucm91dGVyX2dyb3Vwcy5yZWFkIiwiY2xvdWRfY29udHJvbGxlci5yZWFkIiwicGFzc3dvcmQud3JpdGUiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwib3BlbmlkIiwicm91dGluZy5yb3V0ZXJfZ3JvdXBzLndyaXRlIiwiZG9wcGxlci5maXJlaG9zZSIsInNjaW0ud3JpdGUiLCJzY2ltLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLmFkbWluIiwidWFhLnVzZXIiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImF6cCI6ImNmIiwiZ3JhbnRfdHlwZSI6InBhc3N3b3JkIiwidXNlcl9pZCI6IjQ1MTZmOTBhLTE4YTQtNGI0Ni1hOWU2LWVjYTcwNzhkN2RkNiIsIm9yaWdpbiI6InVhYSIsInVzZXJfbmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbiIsInJldl9zaWciOiJjNjJiZTU1MyIsImlhdCI6MTQ5NTAyNjQ3MCwiZXhwIjoxNDk1MDMzNjcwLCJpc3MiOiJodHRwczovL3VhYS5zeXN0ZW0ueW9va29yZS5uZXQvb2F1dGgvdG9rZW4iLCJ6aWQiOiJ1YWEiLCJhdWQiOlsiY2xvdWRfY29udHJvbGxlciIsInNjaW0iLCJwYXNzd29yZCIsImNmIiwidWFhIiwib3BlbmlkIiwiZG9wcGxlciIsInJvdXRpbmcucm91dGVyX2dyb3VwcyJdfQ.KgFZGS4oxjcFFaeywpqMbiSR6UqQKWK6kuEnO-xq02y5rE2gTkIZp6YST9ErF9eotQU919EYz3CHFYn9tyzCO6zLXRZ3RCBDRrniwgeALFkHDkl_MUIuIJGHuBvUyRz62brpdBBjz82_c2MQC-sk5yeyFlGCgOaP81CoZExUYXKTe9QhId3htEMehO2kc0DL9ODSNy_0NEXyVYIdRSDVEyVpAVAtyu9HtIXjDQmrgaT-aeOXKP6bINn-RM6bfnuX3vzB9hJFc5yOkgI7R9ZSH0cA_TD1lL-HSu9YHkGcglNOYcY_cBPpcPpc2ucpGZHAZQJGnLC9xlZxcLl7gK4xwQ"

var (
	ssl        = flag.Bool("ssl", false, "Connection uses TLS if true, else plain TCP")
	certFile   = flag.String("cert_file", "testdata/server1.pem", "The TLS cert file")
	keyFile    = flag.String("key_file", "testdata/server1.key", "The TLS key file")
	jsonDBFile = flag.String("json_db_file", "testdata/route_guide_db.json", "A json file containing a list of features")
	port       = flag.Int("port", 50051, "The server port")
)

type GNozzle struct {
	reqEntity pb.LogRequest
	resp      pb.LogResponse
}

type Config struct {
	ApiAddress             string `json:"api_url"`
	Username               string `json:"user"`
	Password               string `json:"password"`
	ClientID               string `json:"client_id"`
	ClientSecret           string `json:"client_secret"`
	SkipSslValidation      bool   `json:"skip_ssl_validation"`
	HttpClient             *http.Client
	Token                  string `json:"auth_token"`
	FirehoseSubscriptionID string
	UserAgent              string `json:"user_agent"`
	DopplerEndpoint        string
	LoggingEndpoint        string
	TokenEndpoint          string
	AuthEndpoint           string
}

func DefaultConfig() *Config {
	return &Config{
		ApiAddress:             "https://api.system.yookore.net",
		Username:               "admin",
		Password:               "admin",
		Token:                  "",
		SkipSslValidation:      true,
		HttpClient:             http.DefaultClient,
		UserAgent:              "Go-CF-client/1.1",
		FirehoseSubscriptionID: "firehose",
		DopplerEndpoint:        "wss://doppler.system.yookore.net:443",
		LoggingEndpoint:        "wss://loggregator.system.yookore.net:443",
		TokenEndpoint:          "https://uaa.system.yookore.net",
		AuthEndpoint:           "https://login.system.yookore.net",
	}
}

func (s *GNozzle) AppLogs(ctx context.Context, req *pb.LogRequest) {
	logger.Printf("Fetching app logs: %v\n", req)
	conf := DefaultConfig()
	token := s.reqEntity.token
	if token == "" {
		token := conf.Token
	}
	noaaConsumer := consumer.New(config.DopplerEndpoint, &tls.Config{
		InsecureSkipVerify: conf.SkipSslValidation,
	}, nil)
	events, errs := noaaConsumer.Firehose(conf.FirehoseSubscriptionID, token)
	return s.reqEntity.token
}

func main() {
	logger := log.New(os.Stdout, ">>> ", 0)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	if *ssl {
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			grpclog.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	grpcServer := grpc.NewServer(opts...)
	grpcServer.Serve(lis)
}
