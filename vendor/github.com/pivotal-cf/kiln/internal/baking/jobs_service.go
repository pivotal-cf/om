package baking

type JobsService struct {
	logger logger
	reader directoryReader
}

func NewJobsService(logger logger, reader directoryReader) JobsService {
	return JobsService{
		logger: logger,
		reader: reader,
	}
}

func (js JobsService) FromDirectories(directories []string) (map[string]interface{}, error) {
	if len(directories) == 0 {
		return nil, nil
	}

	js.logger.Println("Reading jobs files...")

	jobs := map[string]interface{}{}
	for _, directory := range directories {
		directoryJobs, err := js.reader.Read(directory)
		if err != nil {
			return nil, err
		}

		for _, job := range directoryJobs {
			jobs[job.Name] = job.Metadata
		}
	}

	return jobs, nil
}
