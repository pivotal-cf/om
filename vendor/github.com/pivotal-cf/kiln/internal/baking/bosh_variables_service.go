package baking

type BOSHVariablesService struct {
	logger logger
	reader directoryReader
}

func NewBOSHVariablesService(logger logger, reader directoryReader) BOSHVariablesService {
	return BOSHVariablesService{
		logger: logger,
		reader: reader,
	}
}

func (s BOSHVariablesService) FromDirectories(directories []string) (map[string]interface{}, error) {
	if len(directories) == 0 {
		return nil, nil
	}
	boshVariables := map[string]interface{}{}

	for _, directory := range directories {
		directoryVariables, err := s.reader.Read(directory)
		if err != nil {
			return nil, err
		}

		for _, boshVariable := range directoryVariables {
			boshVariables[boshVariable.Name] = boshVariable.Metadata
		}
	}

	return boshVariables, nil
}
