package version

var appName string
var gitTag string

func GetAppName() string {
	return appName
}

func GetVersion() string {
	return gitTag
}
