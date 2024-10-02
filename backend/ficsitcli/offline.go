package ficsitcli

import (
	"fmt"
	"slices"
	"strings"

	ficsitcache "github.com/satisfactorymodding/ficsit-cli/cli/cache"
	"github.com/satisfactorymodding/ficsit-cli/cli/provider"

	"github.com/satisfactorymodding/SatisfactoryModManager/backend/settings"
)

func (f *ficsitCLI) GetOffline() bool {
	return f.ficsitCli.Provider.IsOffline()
}

func (f *ficsitCLI) SetOffline(offline bool) {
	f.ficsitCli.Provider.(*provider.MixedProvider).Offline = offline
	settings.Settings.Offline = offline
	_ = settings.SaveSettings()
}

type Mod struct {
	ModReference string       `json:"mod_reference"`
	Name         string       `json:"name"`
	Logo         *string      `json:"logo"` // Base64 encoded
	Authors      []string     `json:"authors"`
	Versions     []ModVersion `json:"versions"`
}

type ModVersion struct {
	Version      string                 `json:"version"`
	GameVersion  string                 `json:"game_version"`
	Size         int64                  `json:"size"`
	Dependencies []ModVersionDependency `json:"dependencies"`
}

type ModVersionDependency struct {
	ModReference string `json:"mod_id"`
	Condition    string `json:"condition"`
	Optional     bool   `json:"optional"`
}

func (f *ficsitCLI) OfflineGetMods() ([]Mod, error) {
	cache, err := ficsitcache.GetCache()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	mods := make([]Mod, 0)
	cache.Range(func(modReference string, modFiles []ficsitcache.File) bool {
		mods = append(mods, convertCacheFileToMod(modFiles))
		return true
	})
	return mods, nil
}

func (f *ficsitCLI) OfflineGetModsByReferences(modReferences []string) ([]Mod, error) {
	cache, err := ficsitcache.GetCache()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	mods := make([]Mod, 0)
	cache.Range(func(modReference string, modFiles []ficsitcache.File) bool {
		if !slices.Contains(modReferences, modReference) {
			return true
		}
		mods = append(mods, convertCacheFileToMod(modFiles))
		return true
	})
	return mods, nil
}

func (f *ficsitCLI) OfflineGetMod(modReference string) (Mod, error) {
	modFiles, err := ficsitcache.GetCacheMod(modReference)
	if err != nil {
		return Mod{}, fmt.Errorf("failed to get cache: %w", err)
	}
	if modFiles == nil {
		return Mod{}, fmt.Errorf("mod not found")
	}
	return convertCacheFileToMod(modFiles), nil
}

func convertCacheFileToMod(files []ficsitcache.File) Mod {
	authors := make([]string, 0)

	for _, author := range strings.Split(files[0].Plugin.CreatedBy, ",") {
		authors = append(authors, strings.TrimSpace(author))
	}

	versions := make([]ModVersion, 0)
	for _, file := range files {
		dependencies := make([]ModVersionDependency, 0)

		for _, dependency := range file.Plugin.Plugins {
			if dependency.BasePlugin {
				continue
			}
			dependencies = append(dependencies, ModVersionDependency{
				ModReference: dependency.Name,
				Condition:    dependency.SemVersion,
				Optional:     dependency.Optional,
			})
		}

		versions = append(versions, ModVersion{
			Version:      file.Plugin.SemVersion,
			GameVersion:  file.Plugin.GameVersion,
			Size:         file.Size,
			Dependencies: dependencies,
		})
	}

	return Mod{
		Name:         files[0].Plugin.FriendlyName,
		ModReference: files[0].ModReference,
		Authors:      authors,
		Logo:         files[0].Icon,
		Versions:     versions,
	}
}
