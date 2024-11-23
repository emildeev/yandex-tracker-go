package tracker

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

var (
	_ Client = (*TrackerClient)(nil)
)

const (
	baseUrl        = "https://api.tracker.yandex.net"
	ticketUrl      = "https://api.tracker.yandex.net/v2/issues/"
	ticketComments = "/comments"
)

type Client interface {
	// GetTicket - get Yandex.Tracker ticket by ticket keys
	GetTicket(ticketKey string) (ticket Ticket, err error)
	// PatchTicket - patch Yandex.Tracker ticket by ticket key
	PatchTicket(ticketKey string, body map[string]string) (ticket Ticket, err error)
	// GetTicketComments - get Yandex.Tracker ticket comments by ticket key
	GetTicketComments(ticketKey string) (comments TicketComments, err error)
	// Myself - get information about the current Yandex.Tracker user
	Myself() (user *User, err error)
	// CreateIssue - create Yandex.Tracker issue
	CreateIssue(opts *CreateIssueOptions) (issue *Issue, response *resty.Response, err error)
	// FindIssues - search Yandex.Tracker issues
	FindIssues(opts *FindIssuesOptions, listOpts *ListOptions) (issues []*Issue, response *resty.Response, err error)
	// GetIssue - get Yandex.Tracker issue by key
	GetIssue(issueKey string) (*Issue, *resty.Response, error)

	WithLogger(l resty.Logger)
	WithDebug(d bool)
}

func New(token, xOrgID, xCloudOrgID string) *TrackerClient {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": token,
	}

	switch {
	case xCloudOrgID != "":
		headers["X-Cloud-Org-ID"] = xCloudOrgID
	default:
		headers["X-Org-Id"] = xOrgID
	}

	return &TrackerClient{
		client:  resty.New(),
		headers: headers,
	}
}

type TrackerClient struct {
	headers map[string]string
	client  *resty.Client
}

func (t *TrackerClient) WithLogger(l resty.Logger) {
	t.client.SetLogger(l)
}

func (t *TrackerClient) WithDebug(d bool) {
	t.client.SetDebug(d)
}

func (t *TrackerClient) NewRequest(method, path string, opt interface{}) *resty.Request {
	req := t.client.R()
	req.Method = method
	req.URL = baseUrl + path
	if opt != nil {
		req.SetBody(opt)
	}
	return req.SetHeaders(t.headers)
}

func (t *TrackerClient) Do(req *resty.Request, v interface{}) (*resty.Response, error) {
	resp, err := req.Send()
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf(
			"wrong status code: %d, message=%s, headers=%s", resp.StatusCode(), string(resp.Body()), t.headers,
		)
	}
	if err := json.Unmarshal(resp.Body(), v); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}
	return resp, nil
}

func (t *TrackerClient) GetTicket(ticketKey string) (Ticket, error) {
	request := t.client.R().SetHeaders(t.headers)
	resp, err := request.Get(ticketUrl + ticketKey)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("wrong status code: %d, message=%s", resp.StatusCode(), string(resp.Body()))
	}

	var result Ticket
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return result, nil
}

func (t *TrackerClient) PatchTicket(ticketKey string, body map[string]string) (Ticket, error) {
	request := t.client.R().SetHeaders(t.headers)
	resp, err := request.
		SetBody(body).
		Patch(ticketUrl + ticketKey)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("wrong status code: %d, message=%s", resp.StatusCode(), string(resp.Body()))
	}

	var result Ticket
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return result, nil
}

func (t *TrackerClient) GetTicketComments(ticketKey string) (TicketComments, error) {
	request := t.client.R().SetHeaders(t.headers)
	resp, err := request.Get(ticketUrl + ticketKey + ticketComments)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("wrong status code: %d, message=%s", resp.StatusCode(), string(resp.Body()))
	}

	var result TicketComments
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return result, nil
}
