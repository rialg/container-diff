/*
Copyright 2018 Google, Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package differs

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	pkgutil "github.com/GoogleContainerTools/container-diff/pkg/util"
	"github.com/GoogleContainerTools/container-diff/util"
)

type PhpAnalyzer struct {
}

func (a PhpAnalyzer) Name() string {
	return "PhpAnalyzer"
}

// PhpDiff compares PHP extensions between layers of two different images.
func (a PhpAnalyzer) Diff(image1, image2 pkgutil.Image) (util.Result, error) {
	diff, err := multiVersionDiff(image1, image2, a)
	return diff, err
}

func (a PhpAnalyzer) Analyze(image pkgutil.Image) (util.Result, error) {
	analysis, err := multiVersionAnalysis(image, a)
	return analysis, err
}

func (a PhpAnalyzer) getPackages(image pkgutil.Image) (map[string]map[string]util.PackageInfo, error) {
	packages := make(map[string]map[string]util.PackageInfo)
	var extensionsPath string
	config, err := image.Image.ConfigFile()

	if err != nil {
		return packages, err
	}
	if config.Config.Env != nil {
		confPath := getPhpConfPaths(config.Config.Env)
		// File walk to the dir where the extensions are installed
		extensionsPath = filepath.Join(confPath, "../../lib/php/extensions")
	}
	phpVersion := getPhpVersion(config.Config.Env)

	files, err := os.ReadDir(extensionsPath)
	if err != nil {
		return packages, err
	}

	for _, file := range files {

		if file.IsDir() {

			nextFile, err := os.ReadDir(filepath.Join(extensionsPath, file.Name()))
			if err != nil {
				return packages, err
			}
			files = append(files, nextFile...)
			continue
		}

		info, err := file.Info()
		if err != nil {
			return packages, err
		}

		currPackage := util.PackageInfo{Version: phpVersion, Size: info.Size()}
		mapPath := filepath.Join(extensionsPath, file.Name())
		packageName := strings.Split(file.Name(), ".")[0]
		addToMap(packages, packageName, mapPath, currPackage)
	}

	return packages, nil
}

func getPhpVersion(vars []string) string {
	var version string
	for _, envVar := range vars {
		phpVersionPattern := regexp.MustCompile("^PHP_VERSION=(.*)")
		match := phpVersionPattern.FindStringSubmatch(envVar)
		if len(match) != 0 {
			version = match[1]
			break
		}
	}
	return version
}

func getPhpConfPaths(vars []string) string {
	var path string
	for _, envVar := range vars {
		phpIniPathPattern := regexp.MustCompile("^PHP_INI_DIR=(.*)")
		match := phpIniPathPattern.FindStringSubmatch(envVar)
		if len(match) != 0 {
			path = match[1]
			break
		}
	}
	return path
}
