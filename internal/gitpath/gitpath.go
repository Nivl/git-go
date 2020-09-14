// Package gitpath contains consts and methods to work with path inside
// the .git directory
package gitpath

import (
	"os"
	"path"
)

// .git/ Files and directories
// We keep the refs paths in unix format since they must be stored
// this way. The backend is in charge to convert this to the current
// system when needed
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

// LocalTag returns the UNIX path of a tag
func LocalTag(name string) string {
	return path.Join(RefsTagsPath, name)
}

// LocalBranch returns the UNIX path of branch
func LocalBranch(name string) string {
	return path.Join(RefsHeadsPath, name)
}

// Ref returns the UNIX path of a ref
func Ref(name string) string {
	return path.Join(RefsPath, name)
}
