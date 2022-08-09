package core

import (
	"github.com/JobberRT/pxier_fetcher/public"
	"github.com/sirupsen/logrus"
)

type Executor interface {
	Fetch() []*proxy
	Type() string
}

func NewExecutor(typ string) Executor {
	switch typ {
	case public.ExecutorTypeSTR:
		return newSTRExecutor()
	case public.ExecutorTypeCPL:
		return newCPLExecutor()
	case public.ExecutorTypeTSX:
		return newTSXExecutor()
	case public.ExecutorTypeIHuan:
		return newIHuanExecutor()
	default:
		logrus.WithField("type", typ).Error("unknown executor type")
		return nil
	}
}
