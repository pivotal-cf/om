package baking

type PropertiesService struct {
	logger logger
	reader directoryReader
}

func NewPropertiesService(logger logger, reader directoryReader) PropertiesService {
	return PropertiesService{
		logger: logger,
		reader: reader,
	}
}

func (ps PropertiesService) FromDirectories(directories []string) (map[string]interface{}, error) {
	if len(directories) == 0 {
		return nil, nil
	}

	ps.logger.Println("Reading property blueprint files...")

	properties := map[string]interface{}{}
	for _, directory := range directories {
		directoryProperties, err := ps.reader.Read(directory)
		if err != nil {
			return nil, err
		}

		for _, property := range directoryProperties {
			properties[property.Name] = property.Metadata
		}
	}

	return properties, nil
}
