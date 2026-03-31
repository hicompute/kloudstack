package dns

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/gocql/gocql"
	"github.com/hicompute/kloudstack/pkg/daemons/dns/pb"
	"github.com/scylladb/gocqlx/v2"
	"github.com/spf13/viper"
	grpc "google.golang.org/grpc"
)

type dnsServer struct {
	db gocqlx.Session
	pb.UnimplementedDnsServiceServer
}

func Start(socketPath string) error {
	os.RemoveAll(socketPath)
	hosts := viper.GetStringSlice("SCYLLADB_HOSTS")
	scylladbCluster := gocql.NewCluster(
		hosts...,
	)
	scylladbCluster.Keyspace = viper.GetString("SCYLLADB_KEYSPACE")
	// scylladbCluster.Authenticator = gocql.PasswordAuthenticator{Username: "cassandra", Password: "cassandra"}
	scylladbCluster.PoolConfig.HostSelectionPolicy = gocql.DCAwareRoundRobinPolicy(
		viper.GetString("SCYLLADB_DATACENTER"),
	)
	if len(hosts) == 1 {
		scylladbCluster.Consistency = gocql.One
	}
	scylladbCluster.Keyspace = viper.GetString("SCYLLADB_KEYSPACE")
	scylladbCluster.Timeout = 5 * time.Second
	scylladbCluster.ConnectTimeout = 5 * time.Second

	session, err := gocqlx.WrapSession(scylladbCluster.CreateSession())
	if err != nil {
		return err
	}

	// listener, err := net.Listen("unix", socketPath)
	listener, err := net.Listen("tcp", ":5000")
	if err != nil {
		return fmt.Errorf("failed to create unix socket: %v", err)
	}
	defer listener.Close()

	grpcServer := grpc.NewServer()
	pb.RegisterDnsServiceServer(grpcServer, &dnsServer{db: session})
	return grpcServer.Serve(listener)
}
