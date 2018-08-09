package baking

type InstanceGroupsService struct {
	logger logger
	reader directoryReader
}

func NewInstanceGroupsService(logger logger, reader directoryReader) InstanceGroupsService {
	return InstanceGroupsService{
		logger: logger,
		reader: reader,
	}
}

func (igs InstanceGroupsService) FromDirectories(directories []string) (map[string]interface{}, error) {
	if len(directories) == 0 {
		return nil, nil
	}

	igs.logger.Println("Reading instance group files...")

	instanceGroups := map[string]interface{}{}
	for _, directory := range directories {
		directoryInstanceGroups, err := igs.reader.Read(directory)
		if err != nil {
			return nil, err
		}

		for _, instanceGroup := range directoryInstanceGroups {
			instanceGroups[instanceGroup.Name] = instanceGroup.Metadata
		}
	}

	return instanceGroups, nil
}
