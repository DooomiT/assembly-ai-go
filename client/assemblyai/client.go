package assemblyai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type AssemblyAI interface {
	// UploadLocalFile uploads binary data to AssemblyAI
	// It returs the upload_url
	UploadLocalFile(content []byte) (string, error)
	// Transcript creates a transcription job at AssemblyAI
	// It returns the id of the job
	Transcript(audioUrl string) (string, error)
	// Transcript polls a transcription job at AssemblyAI
	// It returns the result of the job
	PollTranscript(id string, pollSettings *PollSettings) (string, error)
}

type AssemblyAImpl struct {
	http.Client
	baseUrl string
	token   string
}

// Creates a new AssemblyAI client.
// baseUrl is the base api url of AssemblyAI e.g. "https://api.AssemblyAI.com/v2".
// token is your AssemblyAI api token.
// client lets you configure your own http client to use, by default it uses the basic go http.Client with a 15 seconds timeout.
func New(baseUrl, token string, client *http.Client) AssemblyAI {
	if client == nil {
		client = &http.Client{
			Timeout: time.Second * 15,
		}
	}
	return &AssemblyAImpl{*client, baseUrl, token}
}

func isValidStatus(statusCode int) bool {
	okStatusRegex := regexp.MustCompile(`^2..`)
	s := strconv.Itoa(statusCode)
	return okStatusRegex.MatchString(s)
}

func getBody(response *http.Response) ([]byte, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, err
}

func getData[T any](response *http.Response) (*T, error) {
	body, err := getBody(response)
	if err != nil {
		return nil, err
	}

	if !isValidStatus(response.StatusCode) {
		return nil, fmt.Errorf(string(body))
	}

	var data T
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

type UploadLocalFileResponse struct {
	UploadUrl string `json:"upload_url"`
}

// Uploads the content to AssemblyAI following the AssemblyAI documentation https://www.AssemblyAI.com/docs/walkthroughs#uploading-local-files-for-transcription.
// Returns the upload_url
func (client *AssemblyAImpl) UploadLocalFile(content []byte) (string, error) {
	req, err := http.NewRequest("POST", client.baseUrl+"/upload", bytes.NewBuffer(content))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", client.token)
	req.Header.Set("transfer-encoding", "chunked")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := getData[UploadLocalFileResponse](resp)
	if err != nil {
		return "", err
	}
	return data.UploadUrl, nil
}

type TranscriptResponse struct {
	Id     string `json:"id"`
	Status string `json:"status"`
	Text   string `json:"text"`
	Error  string `json:"error"`
}

type PollSettings struct {
	frequency time.Duration
	timeout   time.Duration
}

type TranscriptionStatus string

const (
	Err       TranscriptionStatus = "error"
	Queued                        = "queued"
	Completed                     = "completed"
)

// Polls the transcription job based on a id.
// Optionally you can provide pollSettings to define the poll frequency and timeout
// pollSettings.frequency defines the poll frequency and defaults to 5 seconds
// pollSettings.timeout defines the maximum polling time and defaults to 1 minute
// returns the transcribed text if the status is completed
func (client *AssemblyAImpl) PollTranscript(id string, pollSettings *PollSettings) (string, error) {
	if pollSettings == nil {
		pollSettings = &PollSettings{frequency: time.Second * 5, timeout: time.Minute}
	}
	url := fmt.Sprintf("%s/transcript/%s", client.baseUrl, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("authorization", client.token)
	timeoutTime := time.Now().Add(pollSettings.timeout)
	for time.Now().Before(timeoutTime) {
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		data, err := getData[TranscriptResponse](resp)
		if err != nil {
			return "", err
		}
		switch TranscriptionStatus(data.Status) {
		case Err:
			return "", errors.New(data.Error)
		case Completed:
			return data.Text, nil
		case Queued:
			time.Sleep(pollSettings.frequency)

		}
	}
	return "", fmt.Errorf("timeout, transcription not finished in %s", pollSettings.timeout)
}

type TranscriptDto struct {
	AudioUrl string `json:"audio_url"`
}

// Submits a audio file for transcription follwing the AssemblyAI documentation https://www.AssemblyAI.com/docs/walkthroughs#submitting-files-for-transcription.
// Returns the id of the transcription job
func (client *AssemblyAImpl) Transcript(audioUrl string) (string, error) {
	dto := TranscriptDto{AudioUrl: audioUrl}
	body, err := json.Marshal(dto)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", client.baseUrl+"/transcript", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", client.token)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := getData[TranscriptResponse](resp)
	if err != nil {
		return "", err
	}
	if data.Id == "" {
		return "", errors.New("response did not include an id")
	}
	if data.Status == "error" {
		return "", errors.New(data.Error)
	}
	return data.Id, nil
}
