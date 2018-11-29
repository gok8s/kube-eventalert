package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gok8s/eventalert/pkg/config"
	"github.com/gok8s/eventalert/pkg/store"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/sirupsen/logrus"
)

type EventApi struct {
	idc store.InfluxDBClient
}

func NewEventApi(conf config.Config) EventApi {
	nea := EventApi{
		idc: *store.NewInfluxDBClient(conf.InfluxDBAddr,
			conf.InfluxDBUsernName, conf.InfluxDBPassword, conf.InfluxDBName),
	}
	return nea
}

func (ea *EventApi) GetAllEvent(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]string)
	m["namespace"] = r.FormValue("namespace")
	m["kind"] = r.FormValue("kind")
	m["kind_name"] = r.FormValue("kind_name")
	//limit, _ := strconv.ParseInt(r.FormValue("limit"), 10, 0)
	limit, _ := strconv.Atoi(r.FormValue("limit"))
	if limit == 0 {
		limit = 10
	}
	//offset, _ := strconv.ParseInt(r.FormValue("offset"), 10, 0)
	of, _ := strconv.Atoi(r.FormValue("offset"))
	offset := of * limit
	var sql string
	sql = fmt.Sprintf("select * from k8sevents  where %s order by desc limit %d offset %d", getSqlString(m), limit, offset)
	res := ea.getEvents(sql)
	json.NewEncoder(w).Encode(res)
	//json.NewEncoder(w).Encode(res)
	//fmt.Fprintf(w, "{res:%v}", res)
}

func getSqlString(params map[string]string) string {
	paramstr := ""
	if params != nil {
		for k, v := range params {
			if v != "" {
				if k == "kind_name" {
					paramstr += fmt.Sprintf("%s=~/%s/ and ", k, v)
					continue
				}
				paramstr += fmt.Sprintf("%s='%s' and ", k, v)
			}
		}
	}
	paramstr = strings.TrimSuffix(paramstr, "and ")
	logrus.Debugf("getevent sql paramstr:%s", params)
	return paramstr
}

/*
获取事件
*/
func (ea *EventApi) getEvents(sql string) (retRes []map[string]interface{}) {
	res, err := ea.idc.Get(sql)
	if err != nil {
		logrus.Infof("InfluxDb Get: %s encounter an error: %v", sql, err)
		return retRes
	}
	return extractRes(res)
}

func extractRes(res []client.Result) (retRes []map[string]interface{}) {

	if res[0].Series != nil {
		cs := res[0].Series[0].Columns
		//retRes := make([]map[string]interface)
		for _, row := range res[0].Series[0].Values {
			tm := make(map[string]interface{})
			timestamp, _ := row[0].(json.Number).Int64()
			t := convertTime(timestamp)
			tm["time"] = t
			for n := 1; n < len(cs); n++ {
				if cs[n] == "firstTimestamp" || cs[n] == "lastTimestamp" {
					ts, _ := row[n].(json.Number).Int64()
					tm[cs[n]] = convertTime(ts)
				} else {
					tm[cs[n]] = row[n]
				}
			}
			retRes = append(retRes, tm)
		}
	}
	return retRes
}

func convertTime(tm int64) string {
	t := time.Unix(tm, 0).Format("2006-01-02 15:04:05")
	return t

}
