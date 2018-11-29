package store

import (
	"fmt"

	"github.com/gok8s/eventalert/utils/flog"

	"github.com/gok8s/eventalert/pkg/config"

	"github.com/streadway/amqp"
	"k8s.io/apimachinery/pkg/util/rand"
)

type RabbitMQClient struct {
	conf config.Config
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewRabbitMQClient(conf config.Config) (*RabbitMQClient, error) {
	rmc := &RabbitMQClient{
		conf: conf,
	}
	ch, conn, err := rmc.getChannel()
	//defer ch.Close()
	if err != nil {
		flog.Log().Errorf("初始化Channel失败，无法发送消息,error is: %v", err)
		return nil, err
	}
	rmc.conn = conn
	rmc.ch = ch
	return rmc, nil
}

func (rmc *RabbitMQClient) Publish(msgBody []byte) {
	if rmc.ch == nil {
		flog.Log().Error("channel已关闭，将重连...")
		ch, conn, err := rmc.getChannel()
		if err != nil {
			flog.Log().Errorf("重连Channel失败，无法发送消息,error is: %v", err)
		}
		rmc.conn = conn
		rmc.ch = ch
	}
	queue, qerr := rmc.ch.QueueDeclare(rmc.conf.RabbitRouteKey, rmc.conf.RabbitDurable, false, false, false, nil)

	if qerr != nil {
		flog.Log().Errorf("创建Queue %s 错误，错误： %v", rmc.conf.RabbitMQTopicName, qerr)
	}

	qberr := rmc.ch.QueueBind(queue.Name, rmc.conf.RabbitRouteKey, rmc.conf.RabbitMQTopicName, false, nil)
	if qberr != nil {
		flog.Log().Errorf("QueueBind %s 错误，错误： %v", rmc.conf.RabbitMQTopicName, qberr)
	}

	err := rmc.ch.Publish(
		rmc.conf.RabbitMQTopicName, // exchange
		rmc.conf.RabbitRouteKey,    // routing key
		false,                      // mandatory
		false,                      // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         msgBody,
			DeliveryMode: 2,
		})

	if err != nil {
		flog.Log().Errorf("发送消息失败,消息为 ,错误为： %v", err)
	}
}

func (rmc *RabbitMQClient) getChannel() (*amqp.Channel, *amqp.Connection, error) {
	for i := 0; i < 3; i++ {
		var mqhost string
		if len(rmc.conf.RabbitMQHosts) > 1 {
			mqhost = rmc.conf.RabbitMQHosts[rand.IntnRange(0, len(rmc.conf.RabbitMQHosts))]
		} else {
			mqhost = rmc.conf.RabbitMQHosts[0]
		}

		connstr := fmt.Sprintf("amqp://%s:%s@%s/%s", rmc.conf.RabbitUser, rmc.conf.RabbitPassword, mqhost, rmc.conf.RabbitVhost)
		//connstr := fmt.Sprintf("amqp://%s:%s@%s/", rmc.conf.RabbitUser, rmc.conf.RabbitPassword, mqhost)
		conn, err := amqp.Dial(connstr)
		if err != nil {
			flog.Log().Errorf("不能连接到MQ: %s,会重试3遍，现在是重试第%d遍!! 错误:  %v", connstr, i, err)
			if i == 2 {
				return nil, nil, err
			}
			continue
		}

		ch, err := conn.Channel()
		if err != nil {
			flog.Log().Errorf("初始化到%s MQ Connection成功，但初始化channel失败，会重试3遍，现在是重试第%d遍!! 错误:  %v", mqhost, i, err)
			if i == 2 {
				return nil, nil, err
			}
			continue
		}
		_, err = ch.QueueDeclare(
			rmc.conf.RabbitRouteKey,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			flog.Log().Errorf("声明%s queue失败，会重试3遍，现在是重试第%d遍,错误:%v", rmc.conf.RabbitRouteKey, i, err)
			return ch, conn, err
		} else {
			flog.Log().Debugf("声明%s queue成功", rmc.conf.RabbitRouteKey)
		}
		err = ch.ExchangeDeclare(
			rmc.conf.RabbitMQTopicName,  // name
			rmc.conf.RabbitExchangeType, // type
			rmc.conf.RabbitDurable,      // durable
			false,                       // auto-deleted
			false,                       // internal
			false,                       // no-wait
			nil,                         // arguments
		)
		if err != nil {
			flog.Log().Errorf("声明%s Exchange失败:,会重试3遍，现在是重试第%d遍!! 错误:  %v", mqhost, i, err)
			return ch, conn, err
		}
		return ch, conn, nil
	}
	return nil, nil, nil
}

func (rmc *RabbitMQClient) Close() {
	rmc.ch.Close()
	rmc.conn.Close()
}
