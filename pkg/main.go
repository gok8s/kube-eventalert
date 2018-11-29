package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/gok8s/eventalert/pkg/config"
	"github.com/gok8s/eventalert/pkg/controller"
	"github.com/gok8s/eventalert/pkg/store"

	wapi "github.com/gok8s/eventalert/pkg/api"
	"github.com/gok8s/eventalert/utils"
	api_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

func Start(config config.Config) {
	var kubeClient kubernetes.Interface
	_, err := rest.InClusterConfig()
	if err != nil {
		logrus.Error(err)
		kubeClient = utils.GetClientOutOfCluster(config.APIServerHost, config.KubeConfigFile)
	} else {
		kubeClient = utils.GetClient()
	}

	lw := cache.NewListWatchFromClient(
		kubeClient.CoreV1().RESTClient(), // 客户端
		"events",                         // 被监控资源类型
		"",                               // 被监控命名空间
		fields.Everything())              // 选择器，减少匹配的资源数量

	informer := cache.NewSharedIndexInformer(lw, &api_v1.Event{}, 0, cache.Indexers{})

	c := controller.NewResourceController(kubeClient, informer, config)
	c.MQClient, err = store.NewRabbitMQClient(config)
	defer c.MQClient.Close()

	if err != nil {
		return
	}

	c.InfluxdbClient = store.NewInfluxDBClient(config.InfluxDBAddr, config.InfluxDBUsernName, config.InfluxDBPassword, config.InfluxDBName)

	mux := http.NewServeMux()
	eapi := wapi.NewEventApi(config)
	go registerHandlers(eapi, config.EnableProfiling, config.HttpPort, c, mux)

	stopCh := make(chan struct{})
	defer close(stopCh)
	go c.Run(stopCh)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)
	<-sigterm
}

/*
 * 注册相关的api,profiling
 */
func registerHandlers(eapi wapi.EventApi, enableProfiling bool, port int, wc *controller.Controller, mux *http.ServeMux) {
	mux.HandleFunc("/healthz", wapi.Healthz)
	//mux.HandleFunc("/events", eapi.GetAllEvent)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/build", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		b, _ := json.Marshal("{status: ok}")
		w.Write(b)
	})

	mux.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		if err != nil {
			logrus.Errorf("unexpected error: %v", err)
		}
	})

	if enableProfiling {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/heap", pprof.Index)
		mux.HandleFunc("/debug/pprof/mutex", pprof.Index)
		mux.HandleFunc("/debug/pprof/goroutine", pprof.Index)
		mux.HandleFunc("/debug/pprof/threadcreate", pprof.Index)
		mux.HandleFunc("/debug/pprof/block", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%v", port),
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      300 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	logrus.Fatal(server.ListenAndServe())
}
