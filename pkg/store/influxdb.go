package store

import (
	"strconv"
	"time"

	"github.com/gok8s/eventalert/utils/flog"

	client "github.com/influxdata/influxdb/client/v2"
	api_v1 "k8s.io/api/core/v1"

	"github.com/sirupsen/logrus"
)

type InfluxDBClient struct {
	addr     string
	username string
	password string
	dbname   string
}

func NewInfluxDBClient(addr string, username string, password string, db string) *InfluxDBClient {
	idc := &InfluxDBClient{
		addr:     addr,
		username: username,
		password: password,
		dbname:   db,
	}
	return idc
}

func (idc *InfluxDBClient) getClient() (client.Client, client.BatchPoints) {
	insecureSkipVerify := false
	if idc.username == "" || idc.password == "" {
		insecureSkipVerify = true
	}

	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:               idc.addr,
		Username:           idc.username,
		Password:           idc.password,
		UserAgent:          "",
		Timeout:            0,
		InsecureSkipVerify: insecureSkipVerify,
		TLSConfig:          nil,
		Proxy:              nil,
	})
	//defer c.Close()

	if err != nil {
		logrus.Errorf("Create influxdb client error:  %v", err)
		return nil, nil
	}

	bp, err := client.NewBatchPoints(
		client.BatchPointsConfig{Database: idc.dbname,
			Precision: "s"})

	if err != nil {
		logrus.Errorf("Batch Point return err:  %v", err)
		return nil, nil
	}
	return c, bp
}

func (idc *InfluxDBClient) Write(measurement string, tags map[string]string, fields map[string]interface{}, t time.Time) {
	c, bp := idc.getClient()
	defer c.Close()
	if c == nil || bp == nil {
		logrus.Errorf("InfluxDBClientlog:Create DB Client failed due to client or bp is null!!!!")
		//return
	}
	pt, err := client.NewPoint(measurement, tags, fields, t)
	if err != nil {
		logrus.Fatal("InfluxDBClientlog:influxdb write：  %v", err)
		//return
	}
	bp.AddPoint(pt)
	if err := c.Write(bp); err != nil {
		flog.Log().Fatal("InfluxDBClientlog:InfluxDB write db error----------: %v", err)
		//return
	}
}

func (idc *InfluxDBClient) Get(cmd string) (res []client.Result, err error) {
	c, bp := idc.getClient()
	flog.Log().Infof("sql is:   %s", cmd)
	if c == nil || bp == nil {
		flog.Log().Errorf("Write to DB fail due to client or bp is null!!!!")
		return
	}
	q := client.Query{
		Command:   cmd,
		Database:  idc.dbname,
		Precision: "s",
	}

	if response, err := c.Query(q); err == nil {
		if response.Error() != nil {
			flog.Log().Errorf("Response.res  %v", res)
			return res, response.Error()
		}
		res = response.Results
	} else {
		flog.Log().Errorf("get error:  %v", err)
		return res, err
	}
	return res, nil
}

func (idc *InfluxDBClient) RecordEventToInflux(e *api_v1.Event) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})
	measurement := "k8sevents"
	tags["namespace"] = e.ObjectMeta.Namespace
	tags["kind"] = e.InvolvedObject.Kind
	tags["kind_name"] = e.InvolvedObject.Name
	tags["reason"] = e.Reason
	tags["type"] = e.Type
	fields["message"] = e.Message
	fields["count"] = e.Count
	fields["source_component"] = e.Source.Component
	fields["source_host"] = e.Source.Host
	fields["firstTimestamp"] = e.FirstTimestamp.Unix()
	fields["lastTimestamp"] = e.LastTimestamp.Unix()
	fields["createTimestamp"] = e.CreationTimestamp.UnixNano()
	fields["evt_name"] = e.ObjectMeta.Name
	rv, _ := strconv.Atoi(e.InvolvedObject.ResourceVersion)

	fields["resourceVersion"] = rv
	strcreated := e.CreationTimestamp.Unix() + int64(rv)
	t := time.Unix(strcreated, 0)

	flog.Log().Debugf("InfluxDBClientlog:measurement: %v  ns:  %v; kind:   %v ; Name:   %v; Reason:    %v ;Type:   %v ;"+
		"Message:  %v; Count:    %v;  Component:    %v;  Host:   %v ; FirstTimestamp:   %v, Action: %v ",
		measurement, e.ObjectMeta.Namespace, e.InvolvedObject.Kind, e.InvolvedObject.Name, e.Reason, e.Type,
		e.Message, e.Count, e.Source.Component, e.Source.Host, e.FirstTimestamp, e.Action)
	idc.Write(measurement, tags, fields, t)
}
