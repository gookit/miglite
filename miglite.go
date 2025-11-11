package miglite

// Version represents the version of the application
var Version = "0.1.0"
var BuildTime = "2025-11-05T09:00:00Z"
var GitCommit = "ab3cd4ef"
var GoVersion = "1.21"

// InitInfo initializes the version, build time, and git commit
func InitInfo(version, goVer, buildTime, gitCommit string) {
	Version = version
	GoVersion = goVer
	BuildTime = buildTime
	GitCommit = gitCommit
}
