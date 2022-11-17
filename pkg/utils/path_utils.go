package utils

import (
	"os"
	"path/filepath"
	"strings"
)











func GetPathMatchingScore(path1, path2 string) int {
	path1List := splitPath(path1)
	path2List := splitPath(path2)
	L1 := len(path1List)
	L2 := len(path2List)
	shorterLen := L1
	if L2 < shorterLen {
		shorterLen = L2
	}
	matchingScore := 0
	for i := 1; i <= shorterLen; i++ {
		if path1List[L1-i] == path2List[L2-i] {
			matchingScore++
		} else {
			break
		}
	}
	return matchingScore
}

func splitPath(p string) []string {
	sep := string(os.PathSeparator)
	p = filepath.Clean(p)
	
	
	splitted := strings.Split(p, sep) 
	
	if len(splitted) > 0 && len(splitted[0]) == 0 {
		splitted = splitted[1:]
	}
	
	if len(splitted) == 1 && len(splitted[0]) == 0 {
		splitted[0] = "/"
	}
	return splitted
}

type FileMatcherUpdateRes int

const (
	NewBestMatch FileMatcherUpdateRes = iota
	SameBestMatch
	NotBestMatch
)

func isInternalModuleFile(filename string) bool {
	const externalGoModulePathIndicator = "pkg/mod/"
	return !strings.Contains(filename, externalGoModulePathIndicator)
}

type FileMatcher struct {
	uniqueFile       bool
	bestFile         string
	bestScore        int
	bestFileInternal bool
}

func NewFileMatcher() *FileMatcher {
	return &FileMatcher{uniqueFile: false, bestFile: "", bestScore: -1, bestFileInternal: false}
}

func (f *FileMatcher) UpdateMatch(matchScore int, filename string) FileMatcherUpdateRes {
	if matchScore > f.bestScore {
		f.resetBest(matchScore, filename)
		return NewBestMatch
	}

	if matchScore < f.bestScore {
		
		return NotBestMatch
	}

	

	if filename == f.GetBestFile() && f.IsUnique() {
		
		return SameBestMatch
	}

	if !f.bestFileInternal && isInternalModuleFile(filename) {
		
		
		f.resetBest(matchScore, filename)
		return NewBestMatch
	}

	if f.bestFileInternal && !isInternalModuleFile(filename) {
		
		return NotBestMatch
	}
	
	f.uniqueFile = false
	return NotBestMatch
}

func (f *FileMatcher) IsUnique() bool {
	return f.uniqueFile
}

func (f *FileMatcher) GetBestFile() string {
	return f.bestFile
}

func (f *FileMatcher) AnyMatch() bool {
	return f.bestScore > 0
}

func (f *FileMatcher) resetBest(score int, filename string) {
	f.bestScore = score
	f.bestFile = filename
	f.uniqueFile = true
	f.bestFileInternal = isInternalModuleFile(filename)
}
