package controller

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gok8s/eventalert/utils/flog"

	"github.com/gok8s/eventalert/pkg/config"
	"github.com/gok8s/eventalert/pkg/store"

	"github.com/sirupsen/logrus"

	api_v1 "k8s.io/api/core/v1"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const maxRetries = 5

var serverStartTime time.Time

const (
	// CreateEvent event associated with new objects in an informer
	CreateEvent = "CREATE"
	// UpdateEvent event associated with an object update in an informer
	UpdateEvent = "UPDATE"
	// DeleteEvent event associated when an object is removed from an informer
	DeleteEvent = "DELETE"
	// ConfigurationEvent event associated when a configuration object is created or updated
	ConfigurationEvent = "CONFIGURATION"
)

// Controller object
type Controller struct {
	//logger    *logrus.Entry
	clientset      kubernetes.Interface
	queue          workqueue.RateLimitingInterface
	informer       cache.SharedIndexInformer
	config         config.Config
	MQClient       *store.RabbitMQClient
	InfluxdbClient *store.InfluxDBClient
}
type Event struct {
	Type string
	Obj  *api_v1.Event
	Key  string
}

func NewResourceController(client kubernetes.Interface, informer cache.SharedIndexInformer, config config.Config) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	var newEvent Event
	var err error
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			evt := obj.(*api_v1.Event)
			newEvent.Obj = evt
			newEvent.Type = CreateEvent
			newEvent.Key, err = cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				logrus.Debugf("准备写入队列，类型为:CreateEvent,详情为:%+v", evt)
				queue.Add(newEvent.Key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			oldEvt := old.(*api_v1.Event)
			newEvt := new.(*api_v1.Event)
			//ks.ids.RecordEventToInflux(delEvt)
			//generateEvtRecord(ks, CreateEvent, newEvt, delEvt)
			if !reflect.DeepEqual(newEvt.Source, oldEvt.Source) && oldEvt.Reason != newEvt.Reason {
				flog.Log().Infof("====Old evt:   %v  ", oldEvt)
				flog.Log().Infof("====Cur evt:   %v  ", newEvt)
				newEvent.Obj = newEvt
				newEvent.Type = UpdateEvent
				newEvent.Key, err = cache.MetaNamespaceKeyFunc(old)
				if err == nil {
					flog.Log().Debugf("准备写入队列，类型为:UpdateEvent,详情为:%+v", newEvt)
					queue.Add(newEvent.Key)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			evt := obj.(*api_v1.Event)
			flog.Log().Infof("%+v", evt)
			newEvent.Obj = evt
			newEvent.Type = DeleteEvent
			newEvent.Key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				flog.Log().Debugf("不写入队列，类型为:DeleteEvent,详情为:%+v", evt)
				//queue.Add(newEvent.Key)
			}
		},
	})
	return &Controller{
		//logger:       logrus.WithField("pkg", "kubewatch-"+resourceType),
		clientset: client,
		informer:  informer,
		queue:     queue,
		config:    config,
	}
}

// Run starts the kubewatch controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()
	flog.Log().Info("Starting kubewatch controller")
	serverStartTime = time.Now().Local()
	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}
	flog.Log().Info("Kubewatch controller synced and ready")
	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced is required for the cache.Controller interface.
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

// LastSyncResourceVersion is required for the cache.Controller interface.
func (c *Controller) LastSyncResourceVersion() string {
	return c.informer.LastSyncResourceVersion()
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *Controller) processNextItem() bool {
	newEvent, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(newEvent)
	msg, evt, err := c.process(newEvent.(string))

	if err != nil {
		flog.Log().Error(err)
		return true
	}
	err = c.AlertWorker(msg, evt)
	if err == nil {
		c.queue.Forget(newEvent)
	} else if c.queue.NumRequeues(newEvent) < maxRetries {
		obj, _, _ := c.informer.GetIndexer().GetByKey(newEvent.(string))
		evt := obj.(*api_v1.Event)
		message := evt.Message
		flog.Log().Errorf("Error processing %s (will retry): %v", message, err)
		c.queue.AddRateLimited(newEvent)
	} else {
		// err != nil and too many retries
		flog.Log().Errorf("Error processing %s (giving up): %v", newEvent, err)
		c.queue.Forget(newEvent)
		utilruntime.HandleError(err)
	}
	return true
}

func (c *Controller) process(key string) (map[string]string, *api_v1.Event, error) {
	flog.Log().Info("process-写入mq+influxdb")
	msg := make(map[string]string)
	obj, exists, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return nil, nil, fmt.Errorf("获取对象%+v失败: %v", key, err)
	}
	flog.Log().Debugf("obj is :%+v:", obj)
	if !exists {
		return nil, nil, fmt.Errorf("对象%v不存在,不做处理", key)
	}
	evt := obj.(*api_v1.Event)
	reason := evt.Reason

	msg["firstTimestamp"] = time.Unix(evt.FirstTimestamp.Unix(), 0).Format("2006-01-02 15:04:05")
	msg["lastTimestamp"] = time.Unix(evt.LastTimestamp.Unix(), 10).Format("2006-01-02 15:04:05")
	msg["createdTimestamp"] = time.Unix(evt.CreationTimestamp.Unix(), 0).Format("2006-01-02 15:04:05")
	msg["lastTimestampOri"] = string(evt.LastTimestamp.Unix())

	msg["evtName"] = evt.Name
	//msg["clusterName"] = evt.ObjectMeta.ClusterName
	msg["reason"] = reason
	msg["namespace"] = evt.Namespace
	msg["kind"] = evt.Kind
	msg["name"] = evt.Name

	//新增字段给到graylog
	msg["count"] = fmt.Sprintf("%d", evt.Count)
	msg["source"] = evt.Source.Host
	msg["component"] = evt.Source.Component
	msg["short_message"] = evt.Message

	if strings.ToLower(evt.Kind) == "node" {
		msg["namespace"] = "default"
	} else {
		msg["namespace"] = evt.Namespace
	}
	//TODO A类；写入mq和inflxudb也应做重试和异常捕捉
	msgbytes, _ := json.Marshal(msg)
	c.MQClient.Publish(msgbytes)
	c.InfluxdbClient.RecordEventToInflux(evt)
	return msg, evt, nil
}
