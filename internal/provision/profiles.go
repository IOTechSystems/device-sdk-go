package provision

import (
	"fmt"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func LoadProfiles(path string) {
	if path == "" {
		path = "./res"
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("profiles: couldn't create absolute path for: %s; %v\n", path, err))
		return
	}

	profiles, err := common.DevPrfCli.DeviceProfiles()
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("profiles: couldn't read device profiles from Core Metadata: %v\n", err))
		return
	}

	fileInfo, err := ioutil.ReadDir(absPath)
	if err != nil {
		common.LogCli.Error(fmt.Sprintf("profiles: couldn't read directory: %s; %v\n", path, err))
		return
	}

	for _, file := range fileInfo {
		var profile models.DeviceProfile

		name := file.Name()
		if strings.HasSuffix(name, yamlExt) || strings.HasSuffix(name, yamlExtUpper) {
			path := absPath + "/" + name
			yamlFile, err := ioutil.ReadFile(path)

			if err != nil {
				common.LogCli.Error(fmt.Sprintf("profiles: couldn't read file: %s; %v\n", name, err))
			}

			err = yaml.Unmarshal(yamlFile, &profile)
			if err != nil {
				common.LogCli.Error(fmt.Sprintf("profiles: invalid deviceprofile: %s; %v\n", name, err))
			}

			// if profile already exists in metadata, skip it
			// TODO: optimize by making profiles a map
			if findProfile(profile.Name, profiles) {
				continue
			}

			// add profile to metadata
			id, err := common.DevPrfCli.Add(&profile)
			if err != nil {
				common.LogCli.Error(fmt.Sprintf("profiles: Add device profile: %s to Core Metadata failed: %v\n", name, err))
				continue
			}

			if len(id) != 24 || !bson.IsObjectIdHex(id) {
				common.LogCli.Error("Add deviceprofile returned invalid Id: " + id)
				return
			}

			profile.Id = bson.ObjectIdHex(id)
			pc.profiles[profile.Name] = profile
		}
	}
}
