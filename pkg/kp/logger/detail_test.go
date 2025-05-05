package logger

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	HTTPProto              = "HTTP/1.1"
	Expected_RawData_to_be = "Expected RawData to be"
)

func TestNewDetailLog(t *testing.T) {
	// Set up a temporary configuration for testing
	configLog = LogConfig{
		ProjectName: "test_project",
		Detail: DetailLogConfig{
			RawData:    true,
			LogFile:    false,
			LogConsole: false,
		},
	}

	tests := []struct {
		name       string
		Session    string
		initInvoke string
		scenario   string
		autoEnd    bool
	}{
		{
			name:       "All parameters provided",
			Session:    "test_session",
			initInvoke: "test_invoke",
			scenario:   "test_scenario",
			autoEnd:    true,
		},
		{
			name:       "Empty Session",
			Session:    "",
			initInvoke: "test_invoke",
			scenario:   "test_scenario",
			autoEnd:    true,
		},
		{
			name:       "Empty initInvoke",
			Session:    "test_session",
			initInvoke: "",
			scenario:   "test_scenario",
			autoEnd:    true,
		},
		{
			name:       "Empty Session and initInvoke",
			Session:    "",
			initInvoke: "",
			scenario:   "test_scenario",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dl := NewDetailLog(tc.Session, tc.initInvoke, tc.scenario, tc.autoEnd).(*detailLog)

			if tc.Session == "" {
				expectedSession := "default_" + time.Now().Format("20060102150405")
				if dl.Session[:8] != "default_" {
					t.Errorf("Expected Session to start with %s, but got %s", expectedSession, dl.Session)
				}
			} else {
				if dl.Session != tc.Session {
					t.Errorf("Expected Session to be %s, but got %s", tc.Session, dl.Session)
				}
			}

			if tc.initInvoke == "" {
				expectedInitInvoke := configLog.ProjectName + "_" + time.Now().Format("20060102150405")
				if dl.InitInvoke[:len(configLog.ProjectName)+1] != configLog.ProjectName+"_" {
					t.Errorf("Expected InitInvoke to start with %s, but got %s", expectedInitInvoke, dl.InitInvoke)
				}
			} else {
				if dl.InitInvoke != tc.initInvoke {
					t.Errorf("Expected InitInvoke to be %s, but got %s", tc.initInvoke, dl.InitInvoke)
				}
			}

			if dl.Scenario != tc.scenario {
				t.Errorf("Expected Scenario to be %s, but got %s", tc.scenario, dl.Scenario)
			}

			if dl.AppName != configLog.ProjectName {
				t.Errorf("Expected AppName to be %s, but got %s", configLog.ProjectName, dl.AppName)
			}

			host, _ := os.Hostname()
			if dl.Host != host {
				t.Errorf("Expected Host to be %s, but got %s", host, dl.Host)
			}

			if dl.conf != configLog.Detail {
				t.Errorf("Expected conf to be %+v, but got %+v", configLog.Detail, dl.conf)
			}
		})
	}
}
func TestIsRawDataEnabled(t *testing.T) {
	tests := []struct {
		name     string
		rawData  bool
		expected bool
	}{
		{
			name:     "RawData enabled",
			rawData:  true,
			expected: true,
		},
		{
			name:     "RawData disabled",
			rawData:  false,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configLog = LogConfig{
				ProjectName: "test_project",
				Detail: DetailLogConfig{
					RawData: tc.rawData,
				},
			}

			dl := NewDetailLog("test_session", "test_invoke", "test_scenario", false).(*detailLog)
			result := dl.IsRawDataEnabled()

			if result != tc.expected {
				t.Errorf("Expected IsRawDataEnabled to be %v, but got %v", tc.expected, result)
			}
		})
	}
}
func TestAddInputHttpRequest(t *testing.T) {

	tests := []struct {
		name      string
		node      string
		cmd       string
		invoke    string
		rawData   bool
		req       *http.Request
		expected  InComing
		expectRaw bool
	}{
		{
			name:    "Valid HTTP request with raw data",
			node:    "test_node",
			cmd:     "test_cmd",
			invoke:  "test_invoke",
			rawData: true,
			req: &http.Request{
				Method: "POST",
				Proto:  HTTPProto,
				Header: http.Header{
					"Content-Type": []string{ContentTypeJSON},
				},
				URL: &url.URL{
					RawQuery: "param1=value1&param2=value2",
				},
				Body: io.NopCloser(bytes.NewBufferString(`{"key":"value"}`)),
			},
			expected: InComing{
				// Header: http.Header{
				// 	"Content-Type": []string{ContentTypeJSON},
				// },
				QueryString: url.Values{
					"param1": []string{"value1"},
					"param2": []string{"value2"},
				},
				Body: map[string]interface{}{
					"key": "value",
				},
				Url: "?param1=value1&param2=value2",
			},
			expectRaw: true,
		},
		{
			name:    "Valid HTTP request without raw data",
			node:    "test_node",
			cmd:     "test_cmd",
			invoke:  "test_invoke",
			rawData: false,
			req: &http.Request{
				Method: "GET",
				Proto:  HTTPProto,
				Header: http.Header{
					"Accept": []string{ContentTypeJSON},
				},
				URL: &url.URL{
					RawQuery: "param1=value1&param2=value2",
				},
				Body: io.NopCloser(bytes.NewBufferString(`{"key":"value"}`)),
			},
			expected: InComing{
				Header: map[string]any{},
				QueryString: url.Values{
					"param1": []string{"value1"},
					"param2": []string{"value2"},
				},
				Body: map[string]interface{}{
					"key": "value",
				},
				Url: "?param1=value1\u0026param2=value2",
			},
			expectRaw: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configLog = LogConfig{
				ProjectName: "test_project",
				Detail: DetailLogConfig{
					RawData: tc.rawData,
				},
			}

			dl := NewDetailLog("test_session", "test_invoke", "test_scenario", false).(*detailLog)
			dl.AddInputHttpRequest(tc.node, tc.cmd, tc.invoke, InComing{}, tc.rawData, tc.req.Proto, tc.req.Method)

			if len(dl.Input) != 1 {
				t.Fatalf("expected 1 input log, but got %d", len(dl.Input))
			}

			inputLog := dl.Input[0]

			tcEvent := fmt.Sprintf("%s.%s", tc.node, tc.cmd)
			if inputLog.Event != tcEvent {
				t.Errorf(Expected_RawData_to_be+" %s, But got %s", tcEvent, inputLog.Event)
			}

			if inputLog.Invoke != tc.invoke {
				t.Errorf("expected invoke to be %s, but got %s", tc.invoke, inputLog.Invoke)
			}

			// if inputLog.Protocol == "" || strings.ToLower("http."+tc.req.Method) != fmt.Sprintf("%s.%s", tc.req.Proto, tc.req.Method) {
			// 	t.Errorf("Expected Protocol to be %s, but got %v", tc.req.Proto, inputLog.Protocol)
			// }

			// inputLog.Data convert to InComing
			// data := inputLog.Data.(InComing)
			// fmt.Println(data)

			// if inputLog.Data != tc.expected.Body {
			// 	t.Errorf("Expected Data to be %+v, but got %+v", tc.expected, inputLog.Data)
			// }

			// if tc.expectRaw {
			// 	expectedRaw := ToJson(tc.expected)
			// 	if inputLog.RawData != expectedRaw {
			// 		t.Errorf(Expected_RawData_to_be+" %s, but got %v", expectedRaw, inputLog.RawData)
			// 	}
			// } else {
			// 	if inputLog.RawData != nil {
			// 		t.Errorf(Expected_RawData_to_be+" nil, but got %v", inputLog.RawData)
			// 	}
			// }
		})
	}
}
func TestAddInputRequest(t *testing.T) {
	tests := []struct {
		name      string
		node      string
		cmd       string
		invoke    string
		rawData   interface{}
		data      interface{}
		expectRaw bool
	}{
		{
			name:      "Valid input request with raw data",
			node:      "test_node",
			cmd:       "test_cmd",
			invoke:    "test_invoke",
			rawData:   map[string]interface{}{"key": "value"},
			data:      map[string]interface{}{"key": "value"},
			expectRaw: true,
		},
		{
			name:      "Valid input request without raw data",
			node:      "test_node",
			cmd:       "test_cmd",
			invoke:    "test_invoke",
			rawData:   nil,
			data:      map[string]interface{}{"key": "value"},
			expectRaw: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configLog = LogConfig{
				ProjectName: "test_project",
				Detail: DetailLogConfig{
					RawData: tc.expectRaw,
				},
			}

			dl := NewDetailLog("test_session", "test_invoke", "test_scenario", false).(*detailLog)
			dl.AddInputRequest(tc.node, tc.cmd, tc.invoke, tc.rawData, tc.data, "", "")

			if len(dl.Input) != 1 {
				t.Fatalf("Expected 1 input log, but got %d", len(dl.Input))
			}

			inputLog := dl.Input[0]

			if inputLog.Event != fmt.Sprintf("%s.%s", tc.node, tc.cmd) {
				t.Errorf(Expected_RawData_to_be+" %s, but got %s", fmt.Sprintf("%s.%s", tc.node, tc.cmd), inputLog.Event)
			}

			if inputLog.Invoke != tc.invoke {
				t.Errorf("expected Invoke to be %s, but got %s", tc.invoke, inputLog.Invoke)
			}

			if inputLog.Type != "req" {
				t.Errorf("Expected Type to be req, but got %s", inputLog.Type)
			}

			if tc.expectRaw {
				expectedRaw := ToJson(tc.rawData)

				if inputLog.RawData != expectedRaw {
					t.Errorf("expected RawData to be %s, but got %v", expectedRaw, inputLog.RawData)
				}
			} else {
				if inputLog.RawData != nil {
					t.Errorf("expected RawData to be nil, but got %v", inputLog.RawData)
				}
			}

			if !reflect.DeepEqual(inputLog.Data, ToStruct(tc.data)) {
				t.Errorf("maps are not equal. Expected: %+v, Got: %+v", ToStruct(tc.data), inputLog.Data)
			}
		})
	}
}
func TestAddInputResponse(t *testing.T) {
	tests := []struct {
		name           string
		node           string
		cmd            string
		invoke         string
		rawData        interface{}
		data           interface{}
		protocol       string
		protocolMethod string
		expectRaw      bool
	}{
		{
			name:           "Valid input response with raw data",
			node:           "test_node",
			cmd:            "test_cmd",
			invoke:         "test_invoke",
			rawData:        map[string]interface{}{"key": "value"},
			data:           map[string]interface{}{"key": "value"},
			protocol:       HTTPProto,
			protocolMethod: "POST",
			expectRaw:      true,
		},
		{
			name:           "Valid input response without raw data",
			node:           "test_node",
			cmd:            "test_cmd",
			invoke:         "test_invoke",
			rawData:        nil,
			data:           map[string]interface{}{"key": "value"},
			protocol:       HTTPProto,
			protocolMethod: "GET",
			expectRaw:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configLog = LogConfig{
				ProjectName: "test_project",
				Detail: DetailLogConfig{
					RawData: tc.expectRaw,
				},
			}

			dl := NewDetailLog("test_session", "test_invoke", "test_scenario", false).(*detailLog)
			dl.AddInputResponse(tc.node, tc.cmd, tc.invoke, tc.rawData, tc.data)

			if len(dl.Input) != 1 {
				t.Fatalf("Expected 1 input log, but got %d", len(dl.Input))
			}

			inputLog := dl.Input[0]

			if inputLog.Event != fmt.Sprintf("%s.%s", tc.node, tc.cmd) {
				t.Errorf(Expected_RawData_to_be+" %s, but got %s", fmt.Sprintf("%s.%s", tc.node, tc.cmd), inputLog.Event)
			}

			if inputLog.Invoke != tc.invoke {
				t.Errorf("expected Invoke to be %s, but got %s", tc.invoke, inputLog.Invoke)
			}

			if inputLog.Type != "res" {
				t.Errorf("Expected Type to be res, but got %s", inputLog.Type)
			}

			// expectedProtocol := fmt.Sprintf("%s.%s", tc.protocol, tc.protocolMethod)
			// if inputLog.Protocol == "" || inputLog.Protocol != expectedProtocol {
			// 	t.Errorf("Expected Protocol to be %s, but got %v", expectedProtocol, inputLog.Protocol)
			// }

			if tc.expectRaw {
				expectedRaw := ToJson(tc.rawData)
				if inputLog.RawData != expectedRaw {
					t.Errorf("expected RawData to be %s, but got %v", expectedRaw, inputLog.RawData)
				}
			} else {
				if inputLog.RawData != nil {
					t.Errorf("expected RawData to be nil, but got %v", inputLog.RawData)
				}
			}

			if !reflect.DeepEqual(inputLog.Data, ToStruct(tc.data)) {
				t.Errorf("maps are not equal. Expected: %+v, Got: %+v", ToStruct(tc.data), inputLog.Data)
			}
		})
	}
}
func TestAddOutputResponse(t *testing.T) {
	tests := []struct {
		name      string
		node      string
		cmd       string
		invoke    string
		rawData   interface{}
		data      interface{}
		expectRaw bool
	}{
		{
			name:      "Valid output response with raw data",
			node:      "test_node",
			cmd:       "test_cmd",
			invoke:    "test_invoke",
			rawData:   map[string]interface{}{"key": "value"},
			data:      map[string]interface{}{"key": "value"},
			expectRaw: true,
		},
		{
			name:      "Valid output response without raw data",
			node:      "test_node",
			cmd:       "test_cmd",
			invoke:    "test_invoke",
			rawData:   nil,
			data:      map[string]interface{}{"key": "value"},
			expectRaw: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configLog = LogConfig{
				ProjectName: "test_project",
				Detail: DetailLogConfig{
					RawData: tc.expectRaw,
				},
			}

			dl := NewDetailLog("test_session", "test_invoke", "test_scenario", false).(*detailLog)
			dl.AddOutputResponse(tc.node, tc.cmd, tc.invoke, tc.rawData, tc.data)

			if len(dl.Output) != 1 {
				t.Fatalf("Expected 1 output log, but got %d", len(dl.Output))
			}

			outputLog := dl.Output[0]

			if outputLog.Event != fmt.Sprintf("%s.%s", tc.node, tc.cmd) {
				t.Errorf(Expected_RawData_to_be+" %s.%s, but got %s", tc.node, tc.cmd, outputLog.Event)
			}

			if outputLog.Invoke != tc.invoke {
				t.Errorf("Expected Invoke to be %s, but got %s", tc.invoke, outputLog.Invoke)
			}

			if outputLog.Type != "res" {
				t.Errorf("Expected Type to be res, but got %s", outputLog.Type)
			}

			if tc.expectRaw {
				expectedRaw := ToJson(tc.rawData)
				if outputLog.RawData != expectedRaw {
					t.Errorf("Expected RawData to be %s, but got %v", expectedRaw, outputLog.RawData)
				}
			} else {
				if outputLog.RawData != nil {
					t.Errorf("Expected RawData to be nil, but got %v", outputLog.RawData)
				}
			}

			if !reflect.DeepEqual(outputLog.Data, ToStruct(tc.data)) {
				t.Errorf("Maps are not equal. Expected: %+v, Got: %+v", ToStruct(tc.data), outputLog.Data)
			}
		})
	}
}
func TestAddOutputRequest(t *testing.T) {
	tests := []struct {
		name      string
		node      string
		cmd       string
		invoke    string
		rawData   interface{}
		data      interface{}
		expectRaw bool
	}{
		{
			name:      "Valid output request with raw data",
			node:      "test_node",
			cmd:       "test_cmd",
			invoke:    "test_invoke",
			rawData:   map[string]interface{}{"key": "value"},
			data:      map[string]interface{}{"key": "value"},
			expectRaw: true,
		},
		{
			name:      "Valid output request without raw data",
			node:      "test_node",
			cmd:       "test_cmd",
			invoke:    "test_invoke",
			rawData:   nil,
			data:      map[string]interface{}{"key": "value"},
			expectRaw: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configLog = LogConfig{
				ProjectName: "test_project",
				Detail: DetailLogConfig{
					RawData: tc.expectRaw,
				},
			}

			dl := NewDetailLog("test_session", "test_invoke", "test_scenario", false).(*detailLog)
			dl.AddOutputRequest(tc.node, tc.cmd, tc.invoke, tc.rawData, tc.data, "", "")

			if len(dl.Output) != 1 {
				t.Fatalf("Expected 1 output log, but got %d", len(dl.Output))
			}

			outputLog := dl.Output[0]

			if outputLog.Event != fmt.Sprintf("%s.%s", tc.node, tc.cmd) {
				t.Errorf(Expected_RawData_to_be+" %s.%s, but got %s", tc.node, tc.cmd, outputLog.Event)
			}

			if outputLog.Invoke != tc.invoke {
				t.Errorf("Expected Invoke to be %s, but got %s", tc.invoke, outputLog.Invoke)
			}

			if outputLog.Type != "rep" {
				t.Errorf("Expected Type to be rep, but got %s", outputLog.Type)
			}

			if tc.expectRaw {
				expectedRaw := ToJson(tc.rawData)
				if outputLog.RawData != expectedRaw {
					t.Errorf("Expected RawData to be %s, but got %v", expectedRaw, outputLog.RawData)
				}
			} else {
				if outputLog.RawData != nil {
					t.Errorf("Expected RawData to be nil, but got %v", outputLog.RawData)
				}
			}

			if !reflect.DeepEqual(outputLog.Data, ToStruct(tc.data)) {
				t.Errorf("Maps are not equal. Expected: %+v, Got: %+v", ToStruct(tc.data), outputLog.Data)
			}
		})
	}
}
func TestEndDetail(t *testing.T) {
	tests := []struct {
		name             string
		inputLogs        []InputOutputLog
		outputLogs       []InputOutputLog
		expectLogFile    bool
		expectLogConsole bool
	}{
		{
			name: "End with input and output logs",
			inputLogs: []InputOutputLog{
				{
					Invoke: "test_invoke",
					Event:  "test_event",
					Type:   "req",
					Data:   map[string]interface{}{"key": "value"}},
			},
			outputLogs: []InputOutputLog{
				{
					Invoke: "test_invoke",
					Event:  "test_event",
					Type:   "res",
					Data:   map[string]interface{}{"key": "value"}},
			},

			expectLogFile:    true,
			expectLogConsole: true,
		},
		{
			name:             "End without input and output logs",
			inputLogs:        []InputOutputLog{},
			outputLogs:       []InputOutputLog{},
			expectLogFile:    false,
			expectLogConsole: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configLog = LogConfig{
				ProjectName: "test_project",
				Detail: DetailLogConfig{
					LogFile:    tc.expectLogFile,
					LogConsole: tc.expectLogConsole,
				},
			}

			dl := NewDetailLog("test_session", "test_invoke", "test_scenario", false)
			dl.AddInputRequest("test_node", "test_cmd", "test_invoke", "", map[string]interface{}{"key": "value"}, "", "")
			dl.AddOutputRequest("test_node", "test_cmd", "test_invoke", "", map[string]interface{}{"key": "value"}, "", "")

			// Ensure LogDetail is properly initialized
			// if LogDetail == nil {
			// 	LogDetail = log.New(os.Stdout, "", log.LstdFlags)
			// }
			dl.End()
		})
	}
}
func TestAutoEnd(t *testing.T) {
	tests := []struct {
		name            string
		inputLogs       []InputOutputLog
		outputLogs      []InputOutputLog
		expectedAutoEnd bool
	}{
		{
			name: "AutoEnd with input and output logs",
			inputLogs: []InputOutputLog{
				{
					Invoke: "test_invoke",
					Event:  "test_event",
					Type:   "req",
					Data:   map[string]interface{}{"key": "value"},
				},
			},
			outputLogs: []InputOutputLog{
				{
					Invoke: "test_invoke",
					Event:  "test_event",
					Type:   "res",
					Data:   map[string]interface{}{"key": "value"},
				},
			},
			expectedAutoEnd: true,
		},
		{
			name:            "AutoEnd without input and output logs",
			inputLogs:       []InputOutputLog{},
			outputLogs:      []InputOutputLog{},
			expectedAutoEnd: false,
		},
		{
			name: "AutoEnd with only input logs",
			inputLogs: []InputOutputLog{
				{
					Invoke: "test_invoke",
					Event:  "test_event",
					Type:   "req",
					Data:   map[string]interface{}{"key": "value"},
				},
			},
			outputLogs:      []InputOutputLog{},
			expectedAutoEnd: true,
		},
		{
			name:      "AutoEnd with only output logs",
			inputLogs: []InputOutputLog{},
			outputLogs: []InputOutputLog{
				{
					Invoke: "test_invoke",
					Event:  "test_event",
					Type:   "res",
					Data:   map[string]interface{}{"key": "value"},
				},
			},
			expectedAutoEnd: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configLog = LogConfig{
				ProjectName: "test_project",
				Detail: DetailLogConfig{
					LogFile:    false,
					LogConsole: false,
				},
			}

			dl := NewDetailLog("test_session", "test_invoke", "test_scenario", false).(*detailLog)
			dl.Input = tc.inputLogs
			dl.Output = tc.outputLogs

			result := dl.AutoEnd()

			if result != tc.expectedAutoEnd {
				t.Errorf("Expected AutoEnd to be %v, but got %v", tc.expectedAutoEnd, result)
			}
		})
	}
}

func TestMockDetailLogIsRawDataEnabled(t *testing.T) {
	mockLog := new(MockDetailLog)

	// Set up expectations
	mockLog.On("IsRawDataEnabled").Return(true)

	// Call the mocked method
	isEnabled := mockLog.IsRawDataEnabled()

	// Assert expectations
	mockLog.AssertExpectations(t)
	assert.True(t, isEnabled)
}
func TestMockDetailLogAddInputRequest(t *testing.T) {
	mockLog := new(MockDetailLog)

	// Set up expectations
	mockLog.On("AddInputRequest", "node1", "cmd1", "invoke1", "rawData", "data").Return()

	// Call the mocked method
	mockLog.AddInputRequest("node1", "cmd1", "invoke1", "rawData", "data")

	// Assert expectations
	mockLog.AssertExpectations(t)
}

func TestMockDetailLogEnd(t *testing.T) {
	mockLog := new(MockDetailLog)

	// Set up expectations
	mockLog.On("End").Return()

	// Call the mocked method
	mockLog.End()

	// Assert expectations
	mockLog.AssertExpectations(t)
}

func TestMockDetailLogAutoEnd(t *testing.T) {
	mockLog := new(MockDetailLog)

	// Set up expectations
	mockLog.On("AutoEnd").Return(false)

	// Call the mocked method
	autoEnd := mockLog.AutoEnd()

	// Assert expectations
	mockLog.AssertExpectations(t)
	assert.False(t, autoEnd)
}

func TestMockDetailLogAddInputHttpRequest(t *testing.T) {
	mockLog := new(MockDetailLog)
	req, _ := http.NewRequest("POST", "http://example.com", nil)

	// Set up expectations
	mockLog.On("AddInputHttpRequest", "node1", "cmd1", "invoke1", req, true).Return()

	// Call the mocked method
	mockLog.AddInputHttpRequest("node1", "cmd1", "invoke1", req, true)

	// Assert expectations
	mockLog.AssertExpectations(t)
}
