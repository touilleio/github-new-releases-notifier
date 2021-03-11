package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/sqooba/go-common/logging"
	"github.com/sqooba/go-common/version"
	"github.com/touilleio/github-new-releases-notifier/model"
	"github.com/touilleio/github-new-releases-notifier/notification"
	"github.com/touilleio/github-new-releases-notifier/releases"
	"github.com/touilleio/github-new-releases-notifier/storage"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	setLogLevel = flag.String("set-log-level", "", "Change log level. Possible values are trace,debug,info,warn,error,fatal,panic")
	log         = logging.NewLogger()
)

type envConfig struct {
	Port             string `envconfig:"PORT" default:"8080"`
	LogLevel         string `envconfig:"LOG_LEVEL" default:"info"`
	MetricsNamespace string `envconfig:"METRICS_NAMESPACE" default:"releasenotifier"`
	MetricsSubsystem string `envconfig:"METRICS_SUBSYSTEM" default:""`
	MetricsPath      string `envconfig:"METRICS_PATH" default:"/metrics"`

	ConfigFilePath string `envconfig:"CONFIG_FILE_PATH"`
	DBStoragePath  string `envconfig:"DB_STORAGE_PATH"`
	DBBucketName   string `envconfig:"DB_BUCKET_NAME" default:"bucket"`
	NotifyAllTags  bool   `envconfig:"NOTIFY_ALL_TAGS" default:"false"` // notify all tags when a new project is added?
}

func main() {
	log.Println("github-new-releases-notifier")
	log.Printf("Version    : %s", version.Version)
	log.Printf("Commit     : %s", version.GitCommit)
	log.Printf("Build date : %s", version.BuildDate)
	log.Printf("OSarch     : %s", version.OsArch)

	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s\n", err)
		return
	}

	flag.Parse()
	err := logging.SetLogLevel(log, env.LogLevel)
	if err != nil {
		log.Fatalf("Logging level %s do not seem to be right. Err = %v", env.LogLevel, err)
	}

	if *setLogLevel != "" {
		logging.SetRemoteLogLevelAndExit(log, env.Port, *setLogLevel)
	}

	// Special endpoint to change the verbosity at runtime, i.e. curl -X PUT --data debug ...
	logging.InitVerbosityHandler(log, http.DefaultServeMux)
	initPrometheus(env, http.DefaultServeMux)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	var wg sync.WaitGroup

	config, err := parseConfigFile(env.ConfigFilePath)
	if err != nil {
		log.Fatalf("Error while parsing the config file %s, err = %v", env.ConfigFilePath, err)
	}

	notificationChan := make(chan model.TagToNotify)

	storageHandler, err := storage.NewStorageHandler(env.DBStoragePath, env.DBBucketName, log)
	if err != nil {
		log.Fatalf("Error while instantiating storage handler, err = %v", err)
	}
	http.DefaultServeMux.HandleFunc("/delete", storage.DeleteTagOnDemand(storageHandler))
	http.DefaultServeMux.HandleFunc("/list", storage.ListAllNotifiedTags(storageHandler))

	notificationHandler, err := notification.NewNotificationHandler(config.Notification.Uri, notificationChan, log)
	if err != nil {
		log.Fatalf("Error while instantiating storage handler, err = %v", err)
	}

	releasesHandler, err := releases.NewReleasesHandler(config.Projects, config.PollFrequency, env.NotifyAllTags,
		storageHandler, counterNewTag, notificationChan, log)
	if err != nil {
		log.Fatalf("Error while instantiating storage handler, err = %v", err)
	}
	// TODO: http.DefaultServeMux.HandleFunc("/trigger", triggerOnDemand(releasesHandler))

	go func() {
		perr := notificationHandler.Handle()
		if perr != nil {
			log.Errorf("Got an error in return of notificationHandler.Handle(), err = %v", perr)
			signalChan <- syscall.Signal(2) // SIGINT
		}
	}()

	go func() {
		perr := releasesHandler.Handle()
		if perr != nil {
			log.Errorf("Got an error in return of releasesHandler.Handle(), err = %v", perr)
			signalChan <- syscall.Signal(2) // SIGINT
		}
	}()

	s := http.Server{Addr: fmt.Sprint(":", env.Port)}
	go func() {
		log.Fatal(s.ListenAndServe())
	}()

	<-signalChan
	log.Printf("Shutdown signal received, exiting...")

	close(notificationChan)

	err = s.Shutdown(context.Background())
	if err != nil {
		log.Fatalf("Got an error while shutting down: %v\n", err)
	}

	// Wait for processing to complete properly
	wg.Wait()
}

func parseConfigFile(configFilePath string) (model.ReleaseNotifierConfig, error) {

	var config model.ReleaseNotifierConfig

	yamlFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Errorf("Got an error while reading the config file %s, err = %v", configFilePath, err)
		return config, err
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Got an error while unmarshalling config file %s, err = %v", configFilePath, err)
		return config, err
	}
	return config, nil
}
