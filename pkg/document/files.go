package document

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type files struct {
	baseDir    string
	foundFiles map[string]*fileEntry
}

type fileEntry struct {
	Path string
	data []byte
}

func getFiles(dir string) (files, error) {
	result := files{
		baseDir:    dir,
		foundFiles: make(map[string]*fileEntry),
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		result.foundFiles[path] = &fileEntry{Path: path}
		return nil
	})

	if err != nil {
		return files{}, err
	}

	return result, nil
}

func (f *fileEntry) GetData() []byte {
	if f.data == nil {
		data, err := os.ReadFile(f.Path)
		if err != nil {
			log.Warnf("Error reading file contents for %s: %s", f.Path, err.Error())
			return []byte{}
		}
		f.data = data
	}

	return f.data
}

func normalizePath(p string) string {
	return strings.ReplaceAll(p, "\\", "/") // Convert all backslashes to forward slashes
}

func (f files) GetBytes(name string) []byte {
	normalizedKey := normalizePath(filepath.Join(f.baseDir, name))
	if v, ok := f.foundFiles[normalizedKey]; ok {
		return v.GetData()
	}
	return []byte{}
}

func (f files) Get(name string) string {
	return string(f.GetBytes(name))
}

func (f files) Glob(pattern string) files {
	result := files{
		baseDir:    f.baseDir,
		foundFiles: make(map[string]*fileEntry),
	}

	normalizedPattern := normalizePath(filepath.Join(f.baseDir, pattern))
	g, err := glob.Compile(normalizedPattern, '/')
	if err != nil {
		log.Warnf("Error compiling Glob pattern %s: %s", normalizedPattern, err.Error())
		return result
	}

	for filePath, entry := range f.foundFiles {
		normalizedFilePath := normalizePath(filePath)
		if g.Match(normalizedFilePath) {
			result.foundFiles[normalizedFilePath] = entry
		}
	}

	return result
}

func (f files) AsConfig() string {
	if len(f.foundFiles) == 0 {
		return ""
	}

	m := make(map[string]string)

	// Explicitly convert to strings, and file names
	for k, v := range f.foundFiles {
		m[filepath.Base(k)] = string(v.GetData())
	}

	return toYAML(m)
}

func (f files) AsSecrets() string {
	if len(f.foundFiles) == 0 {
		return ""
	}

	m := make(map[string]string)

	for k, v := range f.foundFiles {
		m[filepath.Base(k)] = base64.StdEncoding.EncodeToString(v.GetData())
	}

	return toYAML(m)
}

func (f files) AsMap() map[string]*fileEntry {
	return f.foundFiles
}

func (f files) Lines(path string) []string {
	if len(f.foundFiles) == 0 {
		return []string{}
	}
	entry, exists := f.foundFiles[path]
	if !exists {
		return []string{}
	}

	return strings.Split(string(entry.GetData()), "\n")
}

func toYAML(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside a template.
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}
