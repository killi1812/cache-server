package app

import (
	"go.uber.org/dig"
	"go.uber.org/zap"
)

var digContainer *dig.Container = nil

// Test configures app so that it can be used in unit testing
func Test() {
	digContainer = dig.New()
}

// Provide is a wrapper around digs Provide on a global container
func Provide(service any, opts ...dig.ProvideOption) {
	if err := digContainer.Provide(service, opts...); err != nil {
		zap.S().Panicf("Faild to provide service %T, err = %+v", service, err)
	}
}

// Invoke is a wrapper around digs Invoke on a global container
func Invoke(service any, opts ...dig.InvokeOption) {
	if err := digContainer.Invoke(service, opts...); err != nil {
		zap.S().Panicf("Faild to provide service %T, err = %+v", service, err)
	}
}
