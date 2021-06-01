package ginternals

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/Nivl/git-go/ginternals/config"
)

// .git/ Files and directories
// We keep the refs paths in unix format since they must be stored
// this way. The backend is in charge to convert this to the current
// system when needed
const (
	refsDirName      = "refs"
	refsTagsRelPath  = refsDirName + "/tags"
	refsHeadsRelPath = refsDirName + "/heads"
)

// LocalTagFullName returns the full name of a tag
// ex. for `my-tag` returns `refs/tags/my-tag`
func LocalTagFullName(shortName string) string {
	return path.Join(refsTagsRelPath, shortName)
}

// LocalTagShortName returns the short name of a tag
// ex. for refs/tags/my-tag returns my-tag
func LocalTagShortName(fullName string) string {
	return strings.TrimPrefix(fullName, refsTagsRelPath)
}

// LocalBranchFullName returns the full name of branch
// ex. for `main` returns `refs/heads/main`
func LocalBranchFullName(shortName string) string {
	return path.Join(refsHeadsRelPath, shortName)
}

// LocalBranchShortName returns the short name of a branch
// ex. for `refs/heads/main` returns `main`
func LocalBranchShortName(fullName string) string {
	return strings.TrimPrefix(fullName, refsHeadsRelPath)
}

// RefFullName returns the UNIX path of a ref
func RefFullName(shortName string) string {
	return path.Join("refs", shortName)
}

// RefsPath return the path to the directory that contains all the refs
func RefsPath(cfg *config.Config) string {
	return filepath.Join(cfg.GitCommonDirPath, "refs")
}

// RefPath return the path of a reference
func RefPath(cfg *config.Config, name string) string {
	return filepath.Join(cfg.GitCommonDirPath, "refs")
}

// PackedRefsPath return the local path of a the packed-refs file
func PackedRefsPath(cfg *config.Config) string {
	return filepath.Join(cfg.GitCommonDirPath, "packed-refs")
}

// TagsPath returns the path to the directory that contains the tags
func TagsPath(cfg *config.Config) string {
	return filepath.Join(RefsPath(cfg), "tags")
}

// DotGitPath returns the path to the dotgit directory
func DotGitPath(cfg *config.Config) string {
	return cfg.GitDirPath
}

// LocalBranchesPath returns the path to the directory containing the
// local branches
func LocalBranchesPath(cfg *config.Config) string {
	return filepath.Join(RefsPath(cfg), "heads")
}

// ObjectsPath returns the path to the directory that contains
// the object
func ObjectsPath(cfg *config.Config) string {
	return cfg.ObjectDirPath
}

// ObjectsInfoPath returns the path to the directory that contains
// the info about the objects
func ObjectsInfoPath(cfg *config.Config) string {
	return filepath.Join(cfg.ObjectDirPath, "info")
}

// ObjectsPacksPath returns the path to the directory that contains
// the packfiles
func ObjectsPacksPath(cfg *config.Config) string {
	return filepath.Join(cfg.ObjectDirPath, "pack")
}

// PackfilePath returns the path of a packfiles
func PackfilePath(cfg *config.Config, name string) string {
	return filepath.Join(ObjectsPacksPath(cfg), name)
}

// ConfigPath returns the path to the local config file
func ConfigPath(cfg *config.Config, name string) string {
	return cfg.LocalConfig
}

// DescriptionFilePath returns the path to the description file
func DescriptionFilePath(cfg *config.Config) string {
	return filepath.Join(DotGitPath(cfg), "description")
}

// LooseObjectPath returns the path of a loose object.
// Path is .git/objects/first_2_chars_of_sha/remaining_chars_of_sha
//
// Ex. path of fcfe68a0e44e04bd7fd564fc0b75f1ae457e18b3 is:
// .git/objects/fc/fe68a0e44e04bd7fd564fc0b75f1ae457e18b3
func LooseObjectPath(cfg *config.Config, sha string) string {
	return filepath.Join(ObjectsPath(cfg), sha[:2], sha[2:])
}
