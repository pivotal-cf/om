package builder

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	pathpkg "path"

	yaml "gopkg.in/yaml.v2"
)

type MetadataPartsDirectoryReader struct {
	topLevelKey string
	orderKey    string
}

type Part struct {
	File     string
	Name     string
	Metadata interface{}
}

func NewMetadataPartsDirectoryReader() MetadataPartsDirectoryReader {
	return MetadataPartsDirectoryReader{}
}

func NewMetadataPartsDirectoryReaderWithTopLevelKey(topLevelKey string) MetadataPartsDirectoryReader {
	return MetadataPartsDirectoryReader{topLevelKey: topLevelKey}
}

func NewMetadataPartsDirectoryReaderWithOrder(topLevelKey, orderKey string) MetadataPartsDirectoryReader {
	return MetadataPartsDirectoryReader{topLevelKey: topLevelKey, orderKey: orderKey}
}

func (r MetadataPartsDirectoryReader) Read(path string) ([]Part, error) {
	parts, err := r.readMetadataRecursivelyFromDir(path)
	if err != nil {
		return []Part{}, err
	}

	if r.orderKey != "" {
		return r.orderWithOrderFromFile(path, parts)
	}

	return r.orderAlphabeticallyByName(path, parts)
}

func (r MetadataPartsDirectoryReader) readMetadataRecursivelyFromDir(path string) ([]Part, error) {
	parts := []Part{}

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || filepath.Ext(filePath) != ".yml" || pathpkg.Base(filePath) == "_order.yml" {
			return nil
		}

		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}

		var vars interface{}
		if r.topLevelKey != "" {
			var fileVars map[string]interface{}
			err = yaml.Unmarshal([]byte(data), &fileVars)
			if err != nil {
				return fmt.Errorf("cannot unmarshal: %s", err)
			}

			var ok bool
			vars, ok = fileVars[r.topLevelKey]
			if !ok {
				return fmt.Errorf("not a %s file: %q", r.topLevelKey, filePath)
			}
		} else {
			err = yaml.Unmarshal([]byte(data), &vars)
			if err != nil {
				return fmt.Errorf("cannot unmarshal: %s", err)
			}
		}

		parts, err = r.readMetadataIntoParts(pathpkg.Base(filePath), vars, parts)
		if err != nil {
			return fmt.Errorf("file '%s' with top-level key '%s' has an invalid format: %s", filePath, r.topLevelKey, err)
		}

		return nil
	})

	return parts, err
}

func (r MetadataPartsDirectoryReader) readMetadataIntoParts(fileName string, vars interface{}, parts []Part) ([]Part, error) {
	switch v := vars.(type) {
	case []interface{}:
		for _, item := range v {
			i, ok := item.(map[interface{}]interface{})
			if !ok {
				return []Part{}, fmt.Errorf("metadata item '%v' must be a map", item)
			}

			part, err := r.buildPartFromMetadata(i, fileName)
			if err != nil {
				return []Part{}, err
			}

			parts = append(parts, part)
		}
	case map[interface{}]interface{}:
		part, err := r.buildPartFromMetadata(v, fileName)
		if err != nil {
			return []Part{}, err
		}
		parts = append(parts, part)
	default:
		return []Part{}, fmt.Errorf("expected either slice or map value")
	}

	return parts, nil
}

func (r MetadataPartsDirectoryReader) buildPartFromMetadata(metadata map[interface{}]interface{}, legacyFilename string) (Part, error) {
	name, ok := metadata["alias"].(string)
	if !ok {
		name, ok = metadata["name"].(string)
		if !ok {
			return Part{}, fmt.Errorf("metadata item '%v' does not have a `name` field", metadata)
		}
	}
	delete(metadata, "alias")

	return Part{File: legacyFilename, Name: name, Metadata: metadata}, nil
}

func (r MetadataPartsDirectoryReader) orderWithOrderFromFile(path string, parts []Part) ([]Part, error) {
	orderPath := filepath.Join(path, "_order.yml")
	f, err := os.Open(orderPath)
	if err != nil {
		return []Part{}, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return []Part{}, err
	}

	var files map[string][]interface{}
	err = yaml.Unmarshal([]byte(data), &files)
	if err != nil {
		return []Part{}, fmt.Errorf("Invalid format for '%s': %s", orderPath, err)
	}

	orderedNames, ok := files[r.orderKey]
	if !ok {
		return []Part{}, fmt.Errorf("Could not find top-level order key '%s' in '%s'", r.orderKey, orderPath)
	}

	var outputs []Part
	for _, name := range orderedNames {
		found := false
		for _, part := range parts {
			if part.Name == name {
				found = true
				outputs = append(outputs, part)
			}
		}
		if !found {
			return []Part{}, fmt.Errorf("file specified in _order.yml %q does not exist in %q", name, path)
		}
	}

	return outputs, err
}

func (r MetadataPartsDirectoryReader) orderAlphabeticallyByName(path string, parts []Part) ([]Part, error) {
	var orderedKeys []string
	for _, part := range parts {
		orderedKeys = append(orderedKeys, part.Name)
	}
	sort.Strings(orderedKeys)

	var outputs []Part
	for _, name := range orderedKeys {
		for _, part := range parts {
			if part.Name == name {
				outputs = append(outputs, part)
			}
		}
	}

	return outputs, nil
}
