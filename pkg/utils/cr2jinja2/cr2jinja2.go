package cr2jinja2

import (
	brokerv2alpha2 "github.com/artemiscloud/activemq-artemis-operator/pkg/apis/broker/v2alpha2"
)

/* return a yaml string */
func MakeBrokerCfgOverrides(customeResource *brokerv2alpha2.ActiveMQArtemis) string {
	return "broker_home: /opt/artemis-2.14.0"
}