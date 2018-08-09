package fakes

type MD5SumCalculator struct {
	ChecksumCall struct {
		CallCount int
		Receives  struct {
			Path string
		}
		Returns struct {
			Sum   string
			Error error
		}
	}
}

func (m *MD5SumCalculator) Checksum(path string) (string, error) {
	m.ChecksumCall.CallCount++
	m.ChecksumCall.Receives.Path = path
	return m.ChecksumCall.Returns.Sum, m.ChecksumCall.Returns.Error
}
