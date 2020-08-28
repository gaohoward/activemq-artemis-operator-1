package controller

import (
	v3alpha1 "github.com/artemiscloud/activemq-artemis-operator/pkg/controller/broker/v3alpha1/activemqartemis"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, v3alpha1.Add)
}
