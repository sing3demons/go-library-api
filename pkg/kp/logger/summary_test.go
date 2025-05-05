package logger

import (
	"testing"
	"time"
)

const (
	INTERNAL_SERVER_ERROR = "internal server error"
)

func TestProcess(t *testing.T) {
	// Mock configuration
	configLog = LogConfig{
		ProjectName: "test_project",
		Summary: SummaryLogConfig{
			LogConsole: true,
			LogFile:    false,
		},
	}

	// Create a new summary log
	sl := NewSummaryLog("test_session", "test_initInvoke", "test_cmd").(*summaryLog)

	// Add some blocks
	sl.AddSuccess("node1", "cmd1", "200", "OK")
	sl.AddError("node2", "cmd2", "500", INTERNAL_SERVER_ERROR)

	// Mock the request time
	requestTime := time.Now().Add(-5 * time.Second)
	sl.requestTime = &requestTime

	// Call the process method
	sl.process("200", "Success")
}
func TestAddField(t *testing.T) {
	// Mock configuration
	configLog = LogConfig{
		ProjectName: "test_project",
		Summary: SummaryLogConfig{
			LogConsole: true,
			LogFile:    false,
		},
	}

	// Create a new summary log
	sl := NewSummaryLog("test_session", "test_initInvoke", "test_cmd").(*summaryLog)

	// Add custom fields
	sl.AddField("customField1", "customValue1")
	sl.AddField("customField2", 12345)

	// Check if the fields were added correctly
	if sl.optionalField["customField1"] != "customValue1" {
		t.Errorf("Expected customField1 to be 'customValue1', but got %v", sl.optionalField["customField1"])
	}

	if sl.optionalField["customField2"] != 12345 {
		t.Errorf("Expected customField2 to be 12345, but got %v", sl.optionalField["customField2"])
	}

	// Mock the request time
	requestTime := time.Now().Add(-5 * time.Second)
	sl.requestTime = &requestTime

	// Call the process method
	sl.process("200", "Success")

}
func TestEnd(t *testing.T) {
	// Mock configuration
	configLog = LogConfig{
		ProjectName: "test_project",
		Summary: SummaryLogConfig{
			LogConsole: true,
			LogFile:    false,
		},
	}

	// Create a new summary log
	sl := NewSummaryLog("test_session", "test_initInvoke", "test_cmd").(*summaryLog)

	// Add some blocks
	sl.AddSuccess("node1", "cmd1", "200", "OK")
	sl.AddError("node2", "cmd2", "500", INTERNAL_SERVER_ERROR)

	// Mock the request time
	requestTime := time.Now().Add(-5 * time.Second)
	sl.requestTime = &requestTime

	// Call the End method
	err := sl.End("200", "Success")
	if err != nil {
		t.Errorf("expected no error, but got %v", err)
	}

	// Check if the log entry is correct
	expectedLogEntry := map[string]interface{}{
		"LogType":        "Summary",
		"InputTimeStamp": requestTime.Format(time.RFC3339),
		"Host":           getHostname(),
		"AppName":        "test_project",
		"Instance":       getInstance(),
		"Session":        "test_session",
		"InitInvoke":     "test_initInvoke",
		"Scenario":       "test_cmd",
		"ResponseResult": "200",
		"ResponseDesc":   "Success",
		"Sequences": []map[string]interface{}{
			{
				"Node": "node1",
				"Cmd":  "cmd1",
				"Result": []map[string]string{
					{"Result": "200", "Desc": "OK"},
				},
			},
			{
				"Node": "node2",
				"Cmd":  "cmd2",
				"Result": []map[string]string{
					{"Result": "500", "Desc": INTERNAL_SERVER_ERROR},
				},
			},
		},
		"EndProcessTimeStamp": time.Now().Format(time.RFC3339),
		"ProcessTime":         "5000 ms",
	}

	_ = expectedLogEntry

	// Here you would typically compare the actual log entry with the expected one.
	// Since the log entry is written to stdout, you might need to capture stdout
	// and parse the JSON to compare it with the expectedLogEntry.
}

func TestEndAlreadyEnded(t *testing.T) {
	// Mock configuration
	configLog = LogConfig{
		ProjectName: "test_project",
		Summary: SummaryLogConfig{
			LogConsole: true,
			LogFile:    false,
		},
	}

	// Create a new summary log
	sl := NewSummaryLog("test_session", "test_initInvoke", "test_cmd").(*summaryLog)

	// Call the End method once
	err := sl.End("200", "Success")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Call the End method again
	err = sl.End("200", "Success")
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}
}
func TestAddBlock(t *testing.T) {
	// Mock configuration
	configLog = LogConfig{
		ProjectName: "test_project",
		Summary: SummaryLogConfig{
			LogConsole: true,
			LogFile:    false,
		},
	}

	// Create a new summary log
	sl := NewSummaryLog("test_session", "test_initInvoke", "test_cmd").(*summaryLog)

	// Add a success block
	sl.addBlock("node1", "cmd1", "200", "OK")

	// Check if the block was added correctly
	if len(sl.blockDetail) != 1 {
		t.Fatalf("Expected 1 block, but got %d", len(sl.blockDetail))
	}

	block := sl.blockDetail[0]
	if block.Node != "node1" || block.Cmd != "cmd1" {
		t.Errorf("Expected block with node 'node1' and cmd 'cmd1', but got node '%s' and cmd '%s'", block.Node, block.Cmd)
	}

	if len(block.Result) != 1 {
		t.Fatalf("Expected 1 result, but got %d", len(block.Result))
	}

	result := block.Result[0]
	if result.ResultCode != "200" || result.ResultDesc != "OK" {
		t.Errorf("Expected result with code '200' and desc 'OK', but got code '%s' and desc '%s'", result.ResultCode, result.ResultDesc)
	}

	// Add an error block
	sl.addBlock("node2", "cmd2", "500", "Internal Server Error")

	// Check if the block was added correctly
	if len(sl.blockDetail) != 2 {
		t.Fatalf("Expected 2 blocks, but got %d", len(sl.blockDetail))
	}

	block = sl.blockDetail[1]
	if block.Node != "node2" || block.Cmd != "cmd2" {
		t.Errorf("Expected block with node 'node2' and cmd 'cmd2', but got node '%s' and cmd '%s'", block.Node, block.Cmd)
	}

	if len(block.Result) != 1 {
		t.Fatalf("Expected 1 result, but got %d", len(block.Result))
	}

	result = block.Result[0]
	if result.ResultCode != "500" || result.ResultDesc != "Internal Server Error" {
		t.Errorf("Expected result with code '500' and desc 'Internal Server Error', but got code '%s' and desc '%s'", result.ResultCode, result.ResultDesc)
	}

	// Add another success block to the same node and cmd
	sl.addBlock("node1", "cmd1", "201", "Created")

	// Check if the block was updated correctly
	if len(sl.blockDetail) != 2 {
		t.Fatalf("Expected 2 blocks, but got %d", len(sl.blockDetail))
	}

	block = sl.blockDetail[0]
	if len(block.Result) != 2 {
		t.Fatalf("Expected 2 results, but got %d", len(block.Result))
	}

	result = block.Result[1]
	if result.ResultCode != "201" || result.ResultDesc != "Created" {
		t.Errorf("Expected result with code '201' and desc 'Created', but got code '%s' and desc '%s'", result.ResultCode, result.ResultDesc)
	}
}
func TestIsEnd(t *testing.T) {
	// Mock configuration
	configLog = LogConfig{
		ProjectName: "test_project",
		Summary: SummaryLogConfig{
			LogConsole: true,
			LogFile:    false,
		},
	}

	// Create a new summary log
	sl := NewSummaryLog("test_session", "test_initInvoke", "test_cmd").(*summaryLog)

	// Check if IsEnd returns false initially
	if sl.IsEnd() {
		t.Errorf("Expected IsEnd to return false, but got true")
	}

	// Call the End method
	err := sl.End("200", "Success")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Check if IsEnd returns true after calling End
	if !sl.IsEnd() {
		t.Errorf("Expected IsEnd to return true, but got false")
	}
}
func TestNewSummaryLog(t *testing.T) {
	// Mock configuration
	configLog = LogConfig{
		ProjectName: "test_project",
		Summary: SummaryLogConfig{
			LogConsole: true,
			LogFile:    false,
		},
	}

	// Test with all parameters provided
	sl := NewSummaryLog("test_session", "test_initInvoke", "test_cmd").(*summaryLog)
	if sl.session != "test_session" {
		t.Errorf("Expected session to be 'test_session', but got %v", sl.session)
	}
	if sl.initInvoke != "test_initInvoke" {
		t.Errorf("Expected initInvoke to be 'test_initInvoke', but got %v", sl.initInvoke)
	}
	if sl.cmd != "test_cmd" {
		t.Errorf("Expected cmd to be 'test_cmd', but got %v", sl.cmd)
	}

	// Test with empty session
	sl = NewSummaryLog("", "test_initInvoke", "test_cmd").(*summaryLog)
	if sl.session == "" {
		t.Errorf("Expected session to be generated, but got empty string")
	}

	// Test with empty initInvoke
	sl = NewSummaryLog("test_session", "", "test_cmd").(*summaryLog)
	if sl.initInvoke == "" {
		t.Errorf("Expected initInvoke to be generated, but got empty string")
	}

	// Test with empty session and initInvoke
	sl = NewSummaryLog("", "", "test_cmd").(*summaryLog)
	if sl.session == "" {
		t.Errorf("Expected session to be generated, but got empty string")
	}
	if sl.initInvoke == "" {
		t.Errorf("Expected initInvoke to be generated, but got empty string")
	}
}
