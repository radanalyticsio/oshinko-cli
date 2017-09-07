package version

var appName string
var gitTag string
var sparkImage string

func GetAppName() string {
	return appName
}

func GetVersion() string {
	return gitTag
}

func GetSparkImage() string {
	return sparkImage
}
