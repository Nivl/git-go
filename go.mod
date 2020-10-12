module github.com/Nivl/git-go

go 1.15

require (
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e
	github.com/spf13/afero v1.4.1
	// We use a specific commit because 1.0.0 contains a vulnerability
	// that has been fixed in the targeted commit
	github.com/spf13/cobra v1.0.1-0.20200713175500-884edc58ad08
	github.com/stretchr/testify v1.6.1
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/ini.v1 v1.61.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)
