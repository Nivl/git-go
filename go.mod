module github.com/Nivl/git-go

go 1.14

require (
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e
	github.com/spf13/afero v1.3.2
	// We use a specific commit because 1.0.0 contains a vulnerability
	// that has been fixed in the targeted commit
	github.com/spf13/cobra v1.0.1-0.20200713175500-884edc58ad08
	github.com/stretchr/testify v1.6.1
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/ini.v1 v1.57.0
)
