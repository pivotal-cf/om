package baking

type FormsService struct {
	logger logger
	reader directoryReader
}

func NewFormsService(logger logger, reader directoryReader) FormsService {
	return FormsService{
		logger: logger,
		reader: reader,
	}
}

func (fs FormsService) FromDirectories(directories []string) (map[string]interface{}, error) {
	if len(directories) == 0 {
		return nil, nil
	}

	fs.logger.Println("Reading form files...")

	forms := map[string]interface{}{}
	for _, directory := range directories {
		directoryForms, err := fs.reader.Read(directory)
		if err != nil {
			return nil, err
		}

		for _, directoryForm := range directoryForms {
			forms[directoryForm.Name] = directoryForm.Metadata
		}
	}

	return forms, nil
}
