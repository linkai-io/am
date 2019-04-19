package am

import (
	"context"
)

const (
	CustomWebFlowServiceKey = "customwebflowservice"
)

// Match Types
const (
	CustomMatchStatusCode int32 = 1
	CustomMatchString     int32 = 2
	CustomMatchRegex      int32 = 3
)

type CustomRequestConfig struct {
	Method     string            `json:"method"`
	URI        string            `json:"uri"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Match      map[int32]string  `json:"match"` // match type -> value
	OnlyPort   int32             `json:"only_port"`
	OnlyScheme string            `json:"only_scheme"`
}

type CustomWebFlowConfig struct {
	OrgID         int                  `json:"org_id"`
	GroupID       int                  `json:"group_id"`
	WebFlowID     int32                `json:"web_flow_id"`
	WebFlowName   string               `json:"web_flow_name"`
	CreationTime  int64                `json:"creation_time"`
	ModifiedTime  int64                `json:"modified_time"`
	Deleted       bool                 `json:"deleted"`
	Configuration *CustomRequestConfig `json:"configuration"`
}

type CustomRequestResult struct {
	Response     string   `json:"response"`
	Matched      bool     `json:"matched"`
	MatchType    int32    `json:"match_type"`
	MatchResults []string `json:"match_results"`
}

type CustomWebFlowResults struct {
	ResultID          int64                  `json:"result_id"`
	OrgID             int                    `json:"org_id"`
	GroupID           int                    `json:"group_id"`
	WebFlowID         int32                  `json:"web_flow_id"`
	URL               string                 `json:"url"`
	LoadURL           string                 `json:"load_url"`
	LoadHostAddress   string                 `json:"load_host_address"`
	LoadIPAddress     string                 `json:"load_ip_address"`
	RequestedPort     int32                  `json:"requested_port"`
	ResponsePort      int32                  `json:"response_port"`
	ResponseTimestamp int64                  `json:"response_timestamp"`
	Result            []*CustomRequestResult `json:"result"`
	ResponseBodyHash  string                 `json:"response_body_hash"`
	ResponseBodyLink  string                 `json:"response_body_link"`
	Error             string                 `json:"error"`
}

const (
	WebFlowStatusStopped int32 = 1
	WebFlowStatusRunning int32 = 2
)

type CustomWebStatus struct {
	StatusID             int64 `json:"status_id"`
	OrgID                int   `json:"org_id"`
	GroupID              int   `json:"group_id"`
	WebFlowID            int32 `json:"web_flow_id"`
	LastUpdatedTimestamp int64 `json:"last_updated_timestamp"`
	StartedTimestamp     int64 `json:"started_timestamp"`
	FinishedTimestamp    int64 `json:"finished_timestamp"`
	WebFlowStatus        int32 `json:"web_flow_status"`
	Total                int32 `json:"total"`
	InProgress           int32 `json:"in_progress"`
	Completed            int32 `json:"completed"`
}

type CustomWebFilter struct {
}

type CustomWebFlowService interface {
	Init(config []byte) error
	Create(ctx context.Context, userContext UserContext, config *CustomWebFlowConfig) (int, error)
	Update(ctx context.Context, userContext UserContext, config *CustomWebFlowConfig) (int, error)
	Delete(ctx context.Context, userContext UserContext, webFlowID int32) (int, error)
	Start(ctx context.Context, userContext UserContext, webFlowID int32) (int, error)
	Stop(ctx context.Context, userContext UserContext, webFlowID int32) (int, error)
	GetStatus(ctx context.Context, userContext UserContext, webFlowID int32) (int, *CustomWebStatus, error)
	GetResults(ctx context.Context, userContext UserContext, filter *CustomWebFilter) (int, []*CustomWebFlowResults, error)
}
