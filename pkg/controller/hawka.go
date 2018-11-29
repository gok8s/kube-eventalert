package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/gok8s/eventalert/utils/flog"

	api_v1 "k8s.io/api/core/v1"

	"net/http"
	"strings"
)

/*
	//判断事件性质，区分出管理员类别admin和用户类别user
	//1，基于alerttype进行初步区分，匹配adminType的设置为admin
	//2，对于usertype的，排除正常的kill
	//3，属于管理员的ns则设置为admin
	//4，其他不发送报警
*/
func (c *Controller) AlertWorker(msg map[string]string, evt *api_v1.Event) error {
	flog.Log().Info("AlertWorker start")
	reason := evt.Reason
	if time.Now().Unix()-evt.LastTimestamp.Unix() > 1800 {
		flog.Log().Debugf("事件时间在30分钟以前不做报警，详情为:%v", msg)
		return nil
	}
	var describe string
	var ok bool
	flog.Log().Debugf("%+v\n", msg)

	if describe, ok = UserAlertReasonType[msg["reason"]]; ok {
		if msg["reason"] == "Killing" {
			if strings.Contains(msg["short_message"], "is unhealthy, it will be killed and re-created") {
				describe = "容器因unhealthy被删除"
			} else {
				flog.Log().Debugf("正常killing，不报警:%s", msg["short_message"])
				return nil
			}
		}
		if _, ok = AdminAlertNS[evt.Namespace]; ok {
			msg["alertType"] = "admin"
		} else {
			msg["alertType"] = "user"
		}
	} else if describe, ok = AdminAlertReasonType[reason]; ok {
		msg["alertType"] = "admin"
		flog.Log().Infof("Admin类型报警触发，Reason: %v, Message: %s", reason, msg["content"])
	} else {
		flog.Log().Debugf("其他事件，不做报警,evt.Reason为:%v", reason)
		return nil
	}

	msg["subject"] = fmt.Sprintf("%s-%s/%s", describe, msg["namespace"], msg["name"])
	status, respBytes, err := callAlertSpeaker(msg, c.config.AlertSpeaker)
	if status != http.StatusOK || err != nil {
		flog.Log().Errorf("调用alert-speaker接口失败:%v,%v", status, respBytes)
		errmessage := fmt.Sprintf("调用alert-speaker接口失败:%v", err)
		return errors.New(errmessage)
	} else {
		flog.Log().Info("调用alert-speaker接口成功")
	}
	return nil
}

func callAlertSpeaker(msg map[string]string, url string) (int, []byte, error) {
	defer func() { //catch or finally
		if r := recover(); r != nil { //catch
			fmt.Println("Recover Triggered: ", r)
		}
	}()

	client := &http.Client{}
	jsonValue, _ := json.Marshal(msg)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		flog.Log().Error(err)
	}
	resp, err := client.Do(req)

	if resp != nil {
		defer resp.Body.Close()
	} else {
		return 503, nil, err
	}

	if err != nil {
		flog.Log().Error(err)
	}

	if resp.StatusCode != http.StatusOK {
		flog.Log().Errorf("wrong status %s code:%d", url, resp.StatusCode)
	}
	bodyBytes, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		flog.Log().Error(err)
	}
	return resp.StatusCode, bodyBytes, err
}
