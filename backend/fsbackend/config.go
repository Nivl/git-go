package fsbackend

import (
	"path/filepath"

	"github.com/Nivl/git-go/backend"
	"github.com/Nivl/git-go/internal/gitpath"
	"golang.org/x/xerrors"
	"gopkg.in/ini.v1"
)

// setDefaultCfg set and persists the default git configuration for
// the repository
func (b *Backend) setDefaultCfg() error {
	cfg := ini.Empty()

	// Core
	core, err := cfg.NewSection(backend.CfgCore)
	if err != nil {
		return xerrors.Errorf("could not create core section: %w", err)
	}
	coreCfg := map[string]string{
		backend.CfgCoreFormatVersion:     "0",
		backend.CfgCoreFileMode:          "true",
		backend.CfgCoreBare:              "false",
		backend.CfgCoreLogAllRefUpdate:   "true",
		backend.CfgCoreIgnoreCase:        "true",
		backend.CfgCorePrecomposeUnicode: "true",
	}
	for k, v := range coreCfg {
		if _, err := core.NewKey(k, v); err != nil {
			return xerrors.Errorf("could not set %s: %w", k, err)
		}
	}
	return cfg.SaveTo(filepath.Join(b.root, gitpath.ConfigPath))
}
