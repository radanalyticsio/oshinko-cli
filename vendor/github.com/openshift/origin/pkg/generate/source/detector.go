package source

import (
	"os"
	"path/filepath"
)

// Info is detected platform information from a source directory
type Info struct {
	Platform string
	Version  string
}

// DetectorFunc is a function that returns source Info from a given directory.
// It returns true if it was able to detect the code in the given directory.
type DetectorFunc func(dir string) (*Info, bool)

// Detectors is a set of DetectorFunc that is used to detect the
// language/platform for a given source directory
type Detectors []DetectorFunc

// DefafultDetectors is a default set of Detector functions
var DefaultDetectors = Detectors{
	DetectRuby,
	DetectJava,
	DetectNodeJS,
	DetectPHP,
	DetectPython,
	DetectPerl,
	DetectScala,
	DetectDotNet,
	DetectLiteralDotNet,
}

// DetectRuby detects Ruby source
func DetectRuby(dir string) (*Info, bool) {
	return detect("ruby", dir, "Gemfile", "Rakefile", "config.ru")
}

// DetectJava detects Java source
func DetectJava(dir string) (*Info, bool) {
	return detect("jee", dir, "pom.xml")
}

// DetectNodeJS detects NodeJS source
func DetectNodeJS(dir string) (*Info, bool) {
	return detect("nodejs", dir, "app.json", "package.json")
}

// DetectPHP detects PHP source
func DetectPHP(dir string) (*Info, bool) {
	return detect("php", dir, "index.php", "composer.json")
}

// DetectPython detects Python source
func DetectPython(dir string) (*Info, bool) {
	return detect("python", dir, "requirements.txt", "setup.py")
}

// DetectPerl detects Perl source
func DetectPerl(dir string) (*Info, bool) {
	return detect("perl", dir, "index.pl", "cpanfile")
}

// DetectScala detects Scala source
func DetectScala(dir string) (*Info, bool) {
	return detect("scala", dir, "build.sbt")
}

// DetectDotNet detects .NET source and matches it to a dotnet supported annotatin or dotnet imagestream name
func DetectDotNet(dir string) (*Info, bool) {
	return detect("dotnet", dir, "project.json")
}

// DetectLiteralDotNet detects .NET source and matches it to a .net supported annotation
func DetectLiteralDotNet(dir string) (*Info, bool) {
	return detect(".net", dir, "project.json")
}

// detect returns an Info object with the given platform if the source at dir contains any of the argument files
func detect(platform string, dir string, files ...string) (*Info, bool) {
	if filesPresent(dir, files) {
		return &Info{
			Platform: platform,
		}, true
	}
	return nil, false
}

func filesPresent(dir string, files []string) bool {
	for _, f := range files {
		_, err := os.Stat(filepath.Join(dir, f))
		if err == nil {
			return true
		}
	}
	return false
}
