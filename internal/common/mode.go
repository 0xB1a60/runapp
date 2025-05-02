package common

type RunMode string

const (
	RunModeOnce   RunMode = "once"
	RunModeOnBoot RunMode = "on-boot"
)

var PrettyRunMode = map[RunMode]string{
	RunModeOnce:   "1️ Once",
	RunModeOnBoot: "🔌 On-boot",
}
