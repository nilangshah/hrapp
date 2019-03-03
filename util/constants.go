package util

import "os"

const (
	LACONFIGKEY = "listen-address"
)

var (
	SERVICENAME    = getServiceName()
	SERVICEVERSION = getServiceVersion()
)

func getServiceName() string {
	name := os.Getenv("SERVICE-NAME")
	if name != "" {
		return name
	} else {
		return "hrapp"
	}
}

func getServiceVersion() string {
	ver := os.Getenv("SERVICE-VERSION")
	if ver != "" {
		return ver
	} else {
		return "v1-0"
	}
}
