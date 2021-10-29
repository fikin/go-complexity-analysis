package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fikin/go-complexity-analysis"
	"gopkg.in/yaml.v3"
)

var skipFiles []*regexp.Regexp
var skipDirs []*regexp.Regexp
var theConfig *ConfigFile

// ConfigFile is representing gocomplexity.yml file
// format is similar to golangci-lint configuration file.
type ConfigFile struct {
	LintersSettings struct {
		Complexity struct {
			CycloOver  *int `yaml:"cyclo-over,omitempty" json:"cyclo-over,omitempty"`
			MaintUnder *int `yaml:"maint-under,omitempty" json:"maint-under,omitempty"`
		} `yaml:"complexity" json:"complexity"`
	} `yaml:"linters-settings" json:"linters-settings"`
	Run struct {
		SkipDirs  []string `yaml:"skip-dirs" json:"skip-dirs"`
		SkipFiles []string `yaml:"skip-files" json:"skip-files"`
		BuildTags []string `yaml:"build-tags" json:"build-tags"`
		Tests     bool     `yaml:"tests" json:"tests"`
	} `yaml:"run" json:"run"`
	Issues struct {
		ExcludeRules []struct {
			Path string `yaml:"path" json:"path"`
		} `yaml:"exclude-rules" json:"exclude-rules"`
	} `yaml:"issues" json:"issues"`
}

func parseConfig(filename string) (*ConfigFile, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var unmarshal func(in []byte, out interface{}) (err error)
	if strings.HasSuffix(filename, "yaml") || strings.HasSuffix(filename, "yml") || strings.HasSuffix(filename, "toml") {
		unmarshal = yaml.Unmarshal
	} else if strings.HasSuffix(filename, ".json") {
		unmarshal = json.Unmarshal
	} else {
		return nil, fmt.Errorf("unsupported file type for configuration file (yaml,yml,toml,json) : %s", filename)
	}

	c := &ConfigFile{}
	err = unmarshal(buf, c)
	if err != nil {
		return nil, fmt.Errorf("in file %q: %v", filename, err)
	}

	return c, nil
}

func configureConfigIfGiven() (err error) {
	if configfile != "" {
		theConfig, err = parseConfig(configfile)
		if err != nil {
			return err
		}
		if theConfig.LintersSettings.Complexity.CycloOver != nil {
			complexity.CycloOver = *theConfig.LintersSettings.Complexity.CycloOver
		}
		if theConfig.LintersSettings.Complexity.MaintUnder != nil {
			complexity.MaintUnder = *theConfig.LintersSettings.Complexity.MaintUnder
		}
		skipFiles, err = stringArrToRegex(theConfig.Run.SkipFiles)
		if err != nil {
			return err
		}
		skipDirs, err = stringArrToRegex(theConfig.Run.SkipDirs)
		if err != nil {
			return err
		}
		complexity.SkipFileFnc = filterFile
	}
	return nil
}

func stringArrToRegex(patterns []string) ([]*regexp.Regexp, error) {
	var patternsRe []*regexp.Regexp
	for _, p := range patterns {
		p = normalizePathInRegex(p)
		patternRe, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("can't compile regexp %q: %s", p, err)
		}
		patternsRe = append(patternsRe, patternRe)
	}

	return patternsRe, nil
}

var separatorToReplace = regexp.QuoteMeta(string(filepath.Separator))

func normalizePathInRegex(path string) string {
	if filepath.Separator == '/' {
		return path
	}

	// This replacing should be safe because "/" are disallowed in Windows
	// https://docs.microsoft.com/ru-ru/windows/win32/fileio/naming-a-file
	return strings.ReplaceAll(path, "/", separatorToReplace)
}

func filterFile(filename string) bool {
	fn := getRelativeFileName(filename, currDir)
	return isAnyMatching(skipFiles, fn) || isAnyMatching(skipDirs, filepath.Dir(fn))
}

func isAnyMatching(arr []*regexp.Regexp, str string) bool {
	for _, re := range arr {
		if re.MatchString(str) {
			return true
		}
	}
	return false
}
