package version

import (
)


//var appName string
//var tag string
//
//
//var (
//	// These variables are initialized via the linker -X flag in the
//	// top-level Makefile when compiling release binaries.
//	tag         = "unknown" // Tag of this build (git describe)
//	time        string      // Build time in UTC (year/month/day hour:min:sec)
//	platform    = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
//	appName		string
//)
//
//type Info struct {
//	GoVersion    string
//	Tag          string
//	Time         string
//	Platform     string
//	AppName      string
//}
//
//
//func (b Info) Short() string {
//	return fmt.Sprintf("oshinko-cli %s (%s, built %s, %s)", b.Tag, b.Platform, b.Time, b.GoVersion)
//}
//
//
//// GetInfo returns an Info struct populated with the build information.
//func GetInfo() Info {
//	return Info{
//		GoVersion:    runtime.Version(),
//		Tag:          tag,
//		Time:         time,
//		Platform:     platform,
//		AppName:      appName,
//	}
//}



var appName string
var gitTag string

func GetAppName() string {
return appName
}

func GetVersion() string {
return gitTag
}
