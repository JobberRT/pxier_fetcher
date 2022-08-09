package main

import (
	"fmt"
	"github.com/JobberRT/pxier_fetcher/core"
	nFormatter "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"runtime"
	"strings"
	"time"
)

func init() {
	viper.AutomaticEnv()
	viper.AllowEmptyEnv(true)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		logrus.Panic(err)
		os.Exit(-1)
	}

	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetOutput(os.Stdout)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&nFormatter.Formatter{
		NoColors:        false,
		HideKeys:        false,
		TimestampFormat: time.Stamp,
		CallerFirst:     true,
		CustomCallerFormatter: func(frame *runtime.Frame) string {
			filename := ""
			slash := strings.LastIndex(frame.File, "/")
			if slash >= 0 {
				filename = frame.File[slash+1:]
			}
			return fmt.Sprintf("「%s:%d」", filename, frame.Line)
		},
	})
}

func main() {
	f := core.NewFactory()
	selected := viper.GetStringSlice("factory.selected_executor")
	if len(selected) == 0 {
		logrus.Panic("missing factory's selected_executor param")
	}
	for _, s := range selected {
		e := core.NewExecutor(strings.ToUpper(s))
		f.RegisterExecutor(e)
	}
	f.Start()
}
