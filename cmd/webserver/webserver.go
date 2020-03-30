package main

import (
	_ "database/sql"
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"runtime/pprof"
	"shoppinglist/pkg/api"
	"shoppinglist/pkg/endpoint"
	"shoppinglist/pkg/service"
	"shoppinglist/pkg/transport"
	"text/tabwriter"
	"time"
)

var (
	debugPort   string
	port        string
	serviceName = "Shopping-List"
)

func init() {
	flag.StringVar(&port, "port", "8000", "specify port to run this server on")
	flag.StringVar(&debugPort, "debug_port", "8080", "specify port to run debug server on")
}

func initCache() {
	// Initialize the redis connection to a redis instance running on local machine
	conn, err := redis.DialURL("redis://localhost")
	if err != nil {
		panic(err)
	}
	// Assign the connection to the package level `cache` variable
	api.Cache = conn
}

func usageFor(short string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  %s\n", short)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")
		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		w.Flush()
		fmt.Fprintf(os.Stderr, "\n")
	}
}

func buildConfigFromEnv() (*service.Config, error) {
	dbconn := viper.GetString("DB_CONNECTION_URL")
	dbport := viper.GetString("DB_CONNECTION_PORT")
	dbuser := viper.GetString("DB_USER")
	dbpass := viper.GetString("DB_PASSWORD")
	c := &service.Config{
		DBConn:     dbconn,
		DBPort:     dbport,
		DBUser:     dbuser,
		DBPassword: dbpass,
	}
	return c, nil
}

func newWebServer(addr string, debugAddr string) {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewSyncLogger(logger)
		logger = log.With(logger,
			"service", serviceName,
			"time:", log.DefaultTimestampUTC(),
			"caller", log.DefaultCaller())
	}
	level.Info(logger).Log("msg", "service started")
	defer level.Info(logger).Log("msg", "service ended")

	c, err := buildConfigFromEnv()
	if err != nil {
		level.Info(logger).Log("failed to read the environment variable with error:", err)
		os.Exit(1)
	}

	serviceStartTime := time.Now().UTC()

	serviceInfo := &service.Info{
		ServiceName: serviceName,
		Version:     "0.0.0",
		BuildInfo:   "",
		BuildTime:   "",
		StartTime:   serviceStartTime.Format("2006-01-02T15:04:05"),
	}

	//open a database connection
	db, err := sqlx.Open("mysql", "root:root@/shoppinglist_test?parseTime=true")
	if err != nil {
		logger.Log("failed to open database connection with err: ", err)
	}
	defer db.Close()

	var (
		service     = service.New(db, logger, c, serviceInfo)
		endpoints   = endpoint.New(service, logger)
		httpHandler = transport.NewHTTPHandler(endpoints, logger)
	)
	go func() {
		logger.Log("transport", "debug/HTTP", "addr", debugAddr)
		err := http.ListenAndServe(debugAddr, http.DefaultServeMux)
		if err != nil {
			logger.Log("transport", "debug/HTTP", "during", "Listen", "err", err)
			os.Exit(1)
		}
	}()
	logger.Log("transport", "debug/HTTP", "addr", addr)
	err = http.ListenAndServe(addr, httpHandler)
	if err != nil {
		logger.Log("transport", "debug/HTTP", "during", "Listen", "err", err)
		os.Exit(1)
	}
}

func main() {

	flag.Usage = usageFor(os.Args[0] + " [flags]")
	flag.Parse()

	initCache()
	// The debug listener mounts the http.DefaultServeMux, and serves up
	// stuff like the Prometheus metrics route, the Go debug and profiling
	// routes, and so on.
	http.DefaultServeMux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello")
	})
	http.DefaultServeMux.HandleFunc("/goroutines", func(w http.ResponseWriter, r *http.Request) {
		pprof.Lookup("goroutine").WriteTo(w, 1)
	})
	http.DefaultServeMux.HandleFunc("/heap", func(w http.ResponseWriter, r *http.Request) {
		pprof.Lookup("heap").WriteTo(w, 1)
	})
	http.DefaultServeMux.HandleFunc("/threads", func(w http.ResponseWriter, r *http.Request) {
		pprof.Lookup("threadcreate").WriteTo(w, 1)
	})

	addr := ":" + port
	debugAddr := "localhost:" + debugPort

	newWebServer(addr, debugAddr)

}
