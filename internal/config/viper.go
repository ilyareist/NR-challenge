package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	vp := viper.GetViper()
	vp.AutomaticEnv()

	pflag.String("listen.addr", "localhost:4000", "")
	pflag.String("log.name", "numbers.log", "")
	pflag.Int("report.int", 10, "")
	pflag.Int("max.conn", 5, "")
	_ = vp.BindPFlags(pflag.CommandLine)
	pflag.Parse()
}

func ListenAddr() string {
	return viper.GetViper().GetString("listen.addr")
}

func ClientsLimit() int {
	return viper.GetViper().GetInt("max.conn")
}

func ReportInt() int {
	return viper.GetViper().GetInt("report.int")
}

func LogName() string {
	return viper.GetViper().GetString("log.name")
}
