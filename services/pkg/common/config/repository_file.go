package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	config_api "github.com/safedep/gateway/services/gen"
	"github.com/safedep/gateway/services/pkg/common/logger"
	"github.com/safedep/gateway/services/pkg/common/utils"
)

type configFileRepository struct {
	path                 string
	gatewayConfiguration *config_api.GatewayConfiguration
	m                    sync.RWMutex
}

func NewConfigFileRepository(path string, lazy bool, monitorForChange bool) (ConfigRepository, error) {
	r := &configFileRepository{path: path}
	var err error

	if !lazy {
		err = r.load()
	}

	if err == nil && monitorForChange {
		err = r.monitorForChange()
	}

	return r, err
}

func (c *configFileRepository) LoadGatewayConfiguration() (*config_api.GatewayConfiguration, error) {
	var err error = nil
	if c.gatewayConfiguration == nil {
		_ = c.load()
	}

	if c.gatewayConfiguration == nil {
		err = fmt.Errorf("gateway configuration is not loaded")
	}

	c.m.RLock()
	defer c.m.RUnlock()

	return c.gatewayConfiguration, err
}

func (c *configFileRepository) SaveGatewayConfiguration(config *config_api.GatewayConfiguration) error {
	return fmt.Errorf("persisting gateway configuration is not supported")
}

func (c *configFileRepository) load() error {
	file, err := os.Open(c.path)
	if err != nil {
		return err
	}

	defer file.Close()

	var gatewayConfiguration config_api.GatewayConfiguration
	err = utils.FromPbJson(file, &gatewayConfiguration)
	if err != nil {
		return err
	}

	err = gatewayConfiguration.Validate()
	if err != nil {
		return err
	}

	c.m.Lock()
	defer c.m.Unlock()

	c.gatewayConfiguration = &gatewayConfiguration
	return nil
}

func (c *configFileRepository) monitorForChange() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// We are OK to leak the goroutine as the watcher will
	// never terminate
	wcb := func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					logger.Errorf("Failed to read from events channel")
					return
				}

				if event.Has(fsnotify.Write) {
					logger.Debugf("Detected changes in configuration file")
					err := c.load()

					if err != nil {
						logger.Errorf("Failed to reload config: %v", err)
					} else {
						logger.Debugf("Successfully reloaded gateway config")
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					logger.Errorf("Failed to read from errors channel")
					return
				}

				logger.Errorf("Watcher returned error: %v", err)
			}
		}
	}

	err = watcher.Add(c.path)
	if err == nil {
		go wcb()
	}

	logger.Debugf("Watcher initialized with error: %v", err)
	return err
}
