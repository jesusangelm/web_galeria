package main

import (
	"flag"
	"html/template"
	"os"
	"sync"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"jesusmarin.dev/galeria/internal/data"
	filestorage "jesusmarin.dev/galeria/internal/file_storage"
	"jesusmarin.dev/galeria/internal/jsonlog"
)

type config struct {
	port int
	db   struct {
		dsn      string
		maxConns int
	}
	env string
	s3  struct {
		bucket            string
		region            string
		endpoint          string
		access_key_id     string
		secret_access_key string
	}
	cdn_host string
}

type application struct {
	config         config
	logger         *jsonlog.Logger
	templateCache  map[string]*template.Template
	sessionManager *scs.SessionManager
	models         data.Models
	s3Manager      filestorage.S3
	wg             sync.WaitGroup
}

func main() {
	// declare the cfg var with configs for the app
	var cfg config

	// Flag parameters
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.IntVar(&cfg.port, "port", 3000, "Network Port to listen")
	// postgres://dbuser:dbpass@dbserver/galeria?sslmode=disable
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxConns, "db-max-conns", 25, "PostgreSQL max connections in the pool")
	flag.StringVar(&cfg.s3.bucket, "s3_bucket", "", "S3 Bucket Name")
	flag.StringVar(&cfg.s3.region, "s3_region", "", "S3 Region")
	flag.StringVar(&cfg.s3.endpoint, "s3_endpoint", "", "S3 Endpoint")
	flag.StringVar(&cfg.s3.access_key_id, "s3_akid", "", "S3 Access Key ID")
	flag.StringVar(&cfg.s3.secret_access_key, "s3_sak", "", "S3 Secret Access Key")
	flag.StringVar(&cfg.cdn_host, "cdn_host", "", "CDN Host")
	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Create the DB Connection Pool
	dbConn, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	// Close the connection pool before exit the main() function
	defer dbConn.Close()
	logger.PrintInfo("database connection pool established", nil)

	// use a SessionManager for session manage and storage
	// in the PostgreSQL DB
	sessionManager := scs.New()
	sessionManager.Store = pgxstore.New(dbConn)
	sessionManager.Lifetime = 12 * time.Hour

	// Initialize a new template cache...
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	s3Session, err := createS3Session(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	s3Manager := filestorage.NewS3Manager(s3Session, cfg.s3.bucket, cfg.env, cfg.cdn_host)

	// Initialize the application struct
	// for application config
	app := application{
		config:         cfg,
		logger:         logger,
		templateCache:  templateCache,
		sessionManager: sessionManager,
		models:         data.NewModels(dbConn, s3Manager),
		s3Manager:      s3Manager,
	}

	// call app.serve() to start the server
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}
