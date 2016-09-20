package fakes

type Writer struct {
	WriteCall struct {
		Receives struct {
			Bytes []byte
		}
		Returns struct {
			BytesWritten int
			Error        error
		}
	}
}

func (w *Writer) Write(b []byte) (int, error) {
	w.WriteCall.Receives.Bytes = b

	return w.WriteCall.Returns.BytesWritten, w.WriteCall.Returns.Error
}
