package information

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Rookout/GoSDK/pkg/logger"
)

const (
	_GIT_FOLDER = ".git"
	_GIT_HEAD   = "HEAD"
	_GIT_CONFIG = "config"
)

var r = regexp.MustCompile("\\[remote \"origin\"]\\s*url\\s*=\\s(?P<url>\\S*)")

func isGit(path string) bool {
	return isDir(filepath.Join(path, _GIT_FOLDER))
}

func isDir(path string) bool {
	if pathAbs, err := filepath.Abs(path); err == nil {
		if fileInfo, err := os.Stat(pathAbs); !os.IsNotExist(err) && fileInfo.IsDir() {
			return true
		}
	}
	return false
}

func FindRoot(strPath string) string {
	if isGit(strPath) {
		return strPath
	} else {
		parentPath := filepath.Dir(strPath)
		if parentPath == strPath {
			return ""
		}
		return FindRoot(parentPath)
	}
}

func followSymLinks(root string, link string) string {
	content := ""
	fileContent, err := ioutil.ReadFile(filepath.Join(root, link))
	content = string(fileContent)
	if err != nil {
		logger.Logger().WithError(err).Debugln("Error reading git information from file system")
	}
	if strings.HasPrefix(content, "ref:") {
		splitContent := strings.Split(content, " ")
		if len(splitContent) > 1 {
			nextLink := strings.TrimSpace(splitContent[1])
			return followSymLinks(root, nextLink)
		}
	}
	return strings.TrimSpace(content)
}

func GetRevision(path string) string {
	return followSymLinks(filepath.Join(path, _GIT_FOLDER), _GIT_HEAD)
}

func GetRemoteOrigin(path string) string {
	content := ""
	fileContent, err := ioutil.ReadFile(filepath.Join(path, _GIT_FOLDER, _GIT_CONFIG))
	content = string(fileContent)
	if err != nil {
		logger.Logger().Debugf("Error reading git config from file system: %s", err)
		return ""
	}

	matches := r.FindStringSubmatch(content)
	urlIndex := r.SubexpIndex("url")
	if urlIndex < len(matches) {
		return matches[urlIndex]
	} else {
		return ""
	}
}
