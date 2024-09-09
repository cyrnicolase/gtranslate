package gtranslate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/joyparty/gokit"
	"github.com/robertkrimen/otto"
	"golang.org/x/text/language"
)

var (
	defaultOption = Option{
		GoogleHost: "google.com",
		Delay:      0,
		TryTimes:   2,
		Timeout:    time.Second * 10,
	}

	defaultHttpClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns: 2,
		},
	}
)

var ttk otto.Value

func init() {
	ttk = gokit.MustReturn(otto.ToValue("0"))
}

// NewTranslate is a function to create a new translate
func NewTranslate(options ...Option) *Translate {
	option := defaultOption
	if len(options) > 0 {
		option = options[0]
	}

	client := defaultHttpClient
	if option.Timeout > 0 {
		client.Timeout = option.getTimeout()
	}

	return &Translate{
		client:   client,
		host:     option.getGoogleHost(),
		delay:    option.getDelay(),
		tryTimes: option.getTryTimes(),
	}
}

// Translate is a struct to hold the translation
type Translate struct {
	host     string
	delay    time.Duration
	tryTimes uint
	client   *http.Client
}

// Run is a method to run the translation
func (t Translate) Run(ctx context.Context, text string, from, to language.Tag) (*TransResult, error) {
	return t.doTranslate(ctx, text, from, to)
}

func (t Translate) doTranslate(ctx context.Context, text string, from, to language.Tag) (*TransResult, error) {
	// build request data
	params, err := t.buildRequestData(text, from, to)
	if err != nil {
		return nil, fmt.Errorf("build request data error: %w", err)
	}

	// build request url
	url2 := t.requestURL()
	u, err := url.Parse(url2)
	if err != nil {
		return nil, fmt.Errorf("parse url error: %w", err)
	}
	u.RawQuery = params.Encode()

	// do http request
	body, err := t.doRequest(ctx, u.String())
	if err != nil {
		return nil, fmt.Errorf("do request error: %w", err)
	}

	// format result
	return t.formatResult(body)
}

func (t Translate) requestURL() string {
	return "https://translate." + t.host + "/translate_a/single"
}

func (t Translate) buildRequestData(text string, from, to language.Tag) (url.Values, error) {
	vt, err := otto.ToValue(text)
	if err != nil {
		return url.Values{}, fmt.Errorf("convert text to otto value error: %w", err)
	}
	token := get(vt, ttk)
	data := map[string]string{
		"client": "gtx",
		"sl":     from.String(),
		"tl":     to.String(),
		"hl":     to.String(),
		"ie":     "UTF-8",
		"oe":     "UTF-8",
		"otf":    "1",
		"ssel":   "0",
		"tsel":   "0",
		"kc":     "7",
		"q":      text,
	}

	params := url.Values{}
	for k, v := range data {
		params.Add(k, v)
	}
	for _, v := range []string{"at", "bd", "ex", "ld", "md", "qca", "rw", "rm", "ss", "t"} {
		params.Add("dt", v)
	}
	params.Add("tk", token)

	return params, nil
}

func (t Translate) doRequest(ctx context.Context, url string) (io.Reader, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	tryTimes := t.tryTimes
	for tryTimes > 0 {
		resp, err := t.client.Do(r)
		if err != nil {
			return nil, fmt.Errorf("http request error: %w", err)
		}
		if resp.StatusCode == http.StatusOK {
			return resp.Body, nil
		}

		tryTimes--
		time.Sleep(t.delay)
	}

	return nil, errBadRequest
}

func (t Translate) formatResult(body io.Reader) (*TransResult, error) {
	var data []any
	if err := json.NewDecoder(body).Decode(&data); err != nil {
		return nil, fmt.Errorf("json decode error: %w", err)
	}

	objs := data[0].([]any)
	objCount := len(objs)

	requestText := ""
	requestTongue := ""
	responseText := ""
	responseTongue := ""
	for i, obj := range objs {
		if len(obj.([]any)) == 0 {
			break
		}
		t, ok := obj.([]any)[0].(string)
		if ok {
			responseText += t
		}
		tr, ok := obj.([]any)[1].(string)
		if ok {
			requestText += tr
		}

		if i == objCount-1 {
			items := obj.([]any)
			requestTongue += obj.([]any)[len(items)-1].(string)
			responseTongue += obj.([]any)[len(items)-2].(string)
		}
	}

	return &TransResult{
		RequestText:    requestText,
		ResponseText:   responseText,
		RequestTongue:  requestTongue,
		ResponseTongue: responseTongue,
	}, nil
}

// Option is a struct to hold the options for the translation
type Option struct {
	GoogleHost string        // GoogleHost is the host of the google translate
	Delay      time.Duration // Delay is the time to wait before retrying the translation
	TryTimes   uint          // Tries is the number of times to retry the translation
	Timeout    time.Duration // Timeout is the time to wait for the translation to complete
}

func (o Option) getGoogleHost() string {
	if o.GoogleHost == "" {
		return defaultOption.GoogleHost
	}
	return o.GoogleHost
}

func (o Option) getTryTimes() uint {
	if o.TryTimes <= 0 {
		return defaultOption.TryTimes
	}
	return o.TryTimes
}

func (o Option) getDelay() time.Duration {
	if o.Delay <= 0 {
		return defaultOption.Delay
	}
	return o.Delay
}

func (o Option) getTimeout() time.Duration {
	if o.Timeout <= 0 {
		return defaultHttpClient.Timeout
	}
	return o.Timeout
}

// TransResult 翻译结果
type TransResult struct {
	RequestText    string `json:"rawText"`
	ResponseText   string `json:"text"`
	RequestTongue  string `json:"rawTongue"`
	ResponseTongue string `json:"tongue"`
}
