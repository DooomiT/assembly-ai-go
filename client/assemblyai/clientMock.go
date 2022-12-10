package assemblyai

type AssemblyAIMock struct {
	UploadLocalFileMock func() (string, error)
	TranscriptMock      func() (string, error)
	PollTranscriptMock  func() (string, error)
}

func (client *AssemblyAIMock) UploadLocalFile(content []byte) (string, error) {
	return client.UploadLocalFileMock()
}

func (client *AssemblyAIMock) Transcript(audioUrl string) (string, error) {
	return client.TranscriptMock()
}

func (client *AssemblyAIMock) PollTranscript(id string, pollSettings *PollSettings) (string, error) {
	return client.PollTranscriptMock()
}
func mockFunction(data string, err error) func() (string, error) {
	return func() (string, error) {
		return data, err
	}
}

func NewMock(uploadFileUrl string, uploadFileError error, transcribedText string, transcribedTextError error, pollText string, pollError error) AssemblyAI {
	return &AssemblyAIMock{
		UploadLocalFileMock: mockFunction(uploadFileUrl, uploadFileError),
		TranscriptMock:      mockFunction(transcribedText, transcribedTextError),
		PollTranscriptMock:  mockFunction(pollText, pollError),
	}
}
