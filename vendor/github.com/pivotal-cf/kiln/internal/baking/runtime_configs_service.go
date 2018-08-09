package baking

type RuntimeConfigsService struct {
	logger logger
	reader directoryReader
}

func NewRuntimeConfigsService(logger logger, reader directoryReader) RuntimeConfigsService {
	return RuntimeConfigsService{
		logger: logger,
		reader: reader,
	}
}

func (rcs RuntimeConfigsService) FromDirectories(directories []string) (map[string]interface{}, error) {
	if len(directories) == 0 {
		return nil, nil
	}

	rcs.logger.Println("Reading runtime config files...")

	runtimeConfigs := map[string]interface{}{}
	for _, directory := range directories {
		directoryRuntimeConfigs, err := rcs.reader.Read(directory)
		if err != nil {
			return nil, err
		}

		for _, runtimeConfig := range directoryRuntimeConfigs {
			runtimeConfigs[runtimeConfig.Name] = runtimeConfig.Metadata
		}
	}

	return runtimeConfigs, nil
}
