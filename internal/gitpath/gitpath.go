// Package gitpath contains consts and methods to work with path inside
// the .git directory
package gitpath

import "os"

// .git/ Files and directories
const (
	DotGitPath      = ".git"
	ConfigPath      = "config"
	DescriptionPath = "description"
	PackedRefsPath  = "packed-refs"
	HEADPath        = "HEAD"
	ObjectsPath     = "objects"
	ObjectsInfoPath = ObjectsPath + string(os.PathSeparator) + "info"
	ObjectsPackPath = ObjectsPath + string(os.PathSeparator) + "pack"
	RefsPath        = "refs"
	RefsTagsPath    = RefsPath + "/tags"
	RefsHeadsPath   = RefsPath + "/heads"
	RefsRemotesPath = RefsPath + "/heads"
)
