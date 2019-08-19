package git

import "os"

// .git/ Files and directories
const (
	DotGitPath      = ".git"
	ConfigPath      = "config"
	DescriptionPath = "description"
	HEADPath        = "HEAD"
	BranchesPath    = "branches"
	ObjectsPath     = "objects"
	ObjectsInfoPath = ObjectsPath + string(os.PathSeparator) + "info"
	ObjectsPackPath = ObjectsPath + string(os.PathSeparator) + "pack"
	RefsPath        = "refs"
	RefsTagsPath    = RefsPath + string(os.PathSeparator) + "tags"
	RefsHeadsPath   = RefsPath + string(os.PathSeparator) + "heads"
)

// .git/Config config keys
const (
	cfgCore                  = "core"
	cfgCoreFormatVersion     = "repositoryformatversion"
	cfgCoreFileMode          = "filemode"
	cfgCoreBare              = "bare"
	cfgCoreLogAllRefUpdate   = "logallrefupdates"
	cfgCoreIgnoreCase        = "ignorecase"
	cfgCorePrecomposeUnicode = "precomposeunicode"
)

// list of file extensions
const (
	ExtPackfile = ".pack"
	ExtIndex    = ".idx"
)
