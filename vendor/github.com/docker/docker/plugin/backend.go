// +build experimental

package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/stringid"
	"github.com/docker/docker/plugin/distribution"
	"github.com/docker/docker/reference"
	"github.com/docker/engine-api/types"
)

// Disable deactivates a plugin, which implies that they cannot be used by containers.
func (pm *Manager) Disable(name string) error {
	p, err := pm.get(name)
	if err != nil {
		return err
	}
	if err := pm.disable(p); err != nil {
		return err
	}
	pm.pluginEventLogger(p.PluginObj.ID, name, "disable")
	return nil
}

// Enable activates a plugin, which implies that they are ready to be used by containers.
func (pm *Manager) Enable(name string) error {
	p, err := pm.get(name)
	if err != nil {
		return err
	}
	if err := pm.enable(p, false); err != nil {
		return err
	}
	pm.pluginEventLogger(p.PluginObj.ID, name, "enable")
	return nil
}

// Inspect examines a plugin manifest
func (pm *Manager) Inspect(name string) (tp types.Plugin, err error) {
	p, err := pm.get(name)
	if err != nil {
		return tp, err
	}
	return p.PluginObj, nil
}

// Pull pulls a plugin and computes the privileges required to install it.
func (pm *Manager) Pull(name string, metaHeader http.Header, authConfig *types.AuthConfig) (types.PluginPrivileges, error) {
	ref, err := reference.ParseNamed(name)
	if err != nil {
		logrus.Debugf("error in reference.ParseNamed: %v", err)
		return nil, err
	}
	name = ref.String()

	if p, _ := pm.get(name); p != nil {
		logrus.Debugf("plugin already exists")
		return nil, fmt.Errorf("%s exists", name)
	}

	pluginID := stringid.GenerateNonCryptoID()

	if err := os.MkdirAll(filepath.Join(pm.libRoot, pluginID), 0755); err != nil {
		logrus.Debugf("error in MkdirAll: %v", err)
		return nil, err
	}

	pd, err := distribution.Pull(name, pm.registryService, metaHeader, authConfig)
	if err != nil {
		logrus.Debugf("error in distribution.Pull(): %v", err)
		return nil, err
	}

	if err := distribution.WritePullData(pd, filepath.Join(pm.libRoot, pluginID), true); err != nil {
		logrus.Debugf("error in distribution.WritePullData(): %v", err)
		return nil, err
	}

	p := pm.newPlugin(ref, pluginID)
	if err := pm.initPlugin(p); err != nil {
		return nil, err
	}

	pm.Lock()
	pm.plugins[pluginID] = p
	pm.nameToID[name] = pluginID
	pm.save()
	pm.Unlock()

	pm.pluginEventLogger(pluginID, name, "pull")
	return computePrivileges(&p.PluginObj.Manifest), nil
}

// List displays the list of plugins and associated metadata.
func (pm *Manager) List() ([]types.Plugin, error) {
	out := make([]types.Plugin, 0, len(pm.plugins))
	for _, p := range pm.plugins {
		out = append(out, p.PluginObj)
	}
	return out, nil
}

// Push pushes a plugin to the store.
func (pm *Manager) Push(name string, metaHeader http.Header, authConfig *types.AuthConfig) error {
	p, err := pm.get(name)
	if err != nil {
		return err
	}
	dest := filepath.Join(pm.libRoot, p.PluginObj.ID)
	config, err := ioutil.ReadFile(filepath.Join(dest, "manifest.json"))
	if err != nil {
		return err
	}

	var dummy types.Plugin
	err = json.Unmarshal(config, &dummy)
	if err != nil {
		return err
	}

	rootfs, err := archive.Tar(filepath.Join(dest, "rootfs"), archive.Gzip)
	if err != nil {
		return err
	}
	defer rootfs.Close()

	_, err = distribution.Push(name, pm.registryService, metaHeader, authConfig, ioutil.NopCloser(bytes.NewReader(config)), rootfs)
	// XXX: Ignore returning digest for now.
	// Since digest needs to be written to the ProgressWriter.
	return err
}

// Remove deletes plugin's root directory.
func (pm *Manager) Remove(name string, config *types.PluginRmConfig) error {
	p, err := pm.get(name)
	if err != nil {
		return err
	}
	if err := pm.remove(p, config.ForceRemove); err != nil {
		return err
	}
	pm.pluginEventLogger(p.PluginObj.ID, name, "remove")
	return nil
}

// Set sets plugin args
func (pm *Manager) Set(name string, args []string) error {
	p, err := pm.get(name)
	if err != nil {
		return err
	}
	return pm.set(p, args)
}
