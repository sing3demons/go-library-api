package logger

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetModuleNameFromGoMod(t *testing.T) {
	originalDir, _ := os.Getwd() // Save the original working directory

	tests := []struct {
		name         string
		goModContent string
		expected     string
	}{
		{
			name:         "Valid go.mod with module name",
			goModContent: "module github.com/example/myproject\n",
			expected:     "myproject",
		},
		{
			name:         "Empty go.mod file",
			goModContent: "",
			expected:     "",
		},
		{
			name:         "go.mod without module line",
			goModContent: "// Comment only\nrequire something v1.0.0\n",
			expected:     "",
		},
		{
			name:         "Module name with spaces",
			goModContent: "module github.com/example/myproject    \n",
			expected:     "myproject",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary directory for the test
			tmpDir, err := os.MkdirTemp("", "testgomod")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir) // Clean up the directory after the test

			// Write go.mod file in the temp directory
			goModPath := filepath.Join(tmpDir, "go.mod")
			if err := os.WriteFile(goModPath, []byte(tc.goModContent), 0644); err != nil {
				t.Fatalf("Failed to write go.mod file: %v", err)
			}

			// Change working directory to the temp directory
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("Failed to change directory: %v", err)
			}
			defer os.Chdir(originalDir) // Restore the original directory after the test

			// Ensure the go.mod file exists in the temp directory
			if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
				t.Fatalf("go.mod file does not exist in temp directory: %v", err)
			}

			// Run the function to get the module name
			result := getModuleNameFromGoMod()
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestLoadLogConfig(t *testing.T) {
	// Test case 1: Load configuration with all fields set
	cfg := LogConfig{
		ProjectName: "test_project",
		Namespace:   "test_namespace",
		AppLog: AppLog{
			Name:       "test_app_log",
			LogFile:    true,
			LogConsole: true,
		},
		Detail: DetailLogConfig{
			Name:       "test_detail_log",
			RawData:    true,
			LogFile:    true,
			LogConsole: true,
		},
		Summary: SummaryLogConfig{
			Name:       "test_summary_log",
			RawData:    true,
			LogFile:    true,
			LogConsole: true,
		},
	}

	loadedConfig := LoadLogConfig(cfg)

	if loadedConfig.ProjectName != cfg.ProjectName {
		t.Errorf("Expected ProjectName to be %s, but got %s", cfg.ProjectName, loadedConfig.ProjectName)
	}
	if loadedConfig.Namespace != cfg.Namespace {
		t.Errorf("Expected Namespace to be %s, but got %s", cfg.Namespace, loadedConfig.Namespace)
	}
	if loadedConfig.AppLog.Name != cfg.AppLog.Name {
		t.Errorf("Expected AppLog.Name to be %s, but got %s", cfg.AppLog.Name, loadedConfig.AppLog.Name)
	}
	if loadedConfig.AppLog.LogFile != cfg.AppLog.LogFile {
		t.Errorf("Expected AppLog.LogFile to be %v, but got %v", cfg.AppLog.LogFile, loadedConfig.AppLog.LogFile)
	}
	if loadedConfig.AppLog.LogConsole != cfg.AppLog.LogConsole {
		t.Errorf("Expected AppLog.LogConsole to be %v, but got %v", cfg.AppLog.LogConsole, loadedConfig.AppLog.LogConsole)
	}
	if loadedConfig.Detail.Name != cfg.Detail.Name {
		t.Errorf("Expected Detail.Name to be %s, but got %s", cfg.Detail.Name, loadedConfig.Detail.Name)
	}
	if loadedConfig.Detail.RawData != cfg.Detail.RawData {
		t.Errorf("Expected Detail.RawData to be %v, but got %v", cfg.Detail.RawData, loadedConfig.Detail.RawData)
	}
	if loadedConfig.Detail.LogFile != cfg.Detail.LogFile {
		t.Errorf("Expected Detail.LogFile to be %v, but got %v", cfg.Detail.LogFile, loadedConfig.Detail.LogFile)
	}
	if loadedConfig.Detail.LogConsole != cfg.Detail.LogConsole {
		t.Errorf("Expected Detail.LogConsole to be %v, but got %v", cfg.Detail.LogConsole, loadedConfig.Detail.LogConsole)
	}
	if loadedConfig.Summary.Name != cfg.Summary.Name {
		t.Errorf("Expected Summary.Name to be %s, but got %s", cfg.Summary.Name, loadedConfig.Summary.Name)
	}
	if loadedConfig.Summary.RawData != cfg.Summary.RawData {
		t.Errorf("Expected Summary.RawData to be %v, but got %v", cfg.Summary.RawData, loadedConfig.Summary.RawData)
	}
	if loadedConfig.Summary.LogFile != cfg.Summary.LogFile {
		t.Errorf("Expected Summary.LogFile to be %v, but got %v", cfg.Summary.LogFile, loadedConfig.Summary.LogFile)
	}
	if loadedConfig.Summary.LogConsole != cfg.Summary.LogConsole {
		t.Errorf("Expected Summary.LogConsole to be %v, but got %v", cfg.Summary.LogConsole, loadedConfig.Summary.LogConsole)
	}

	// clean up
	os.RemoveAll(cfg.AppLog.Name)
	os.RemoveAll(cfg.Detail.Name)
	os.RemoveAll(cfg.Summary.Name)
	os.RemoveAll("logs")

	// Test case 2: Load configuration with only some fields set
	cfg = LogConfig{
		ProjectName: "partial_project",
		AppLog: AppLog{
			Name:    "partial_app_log",
			LogFile: true,
		},
	}

	loadedConfig = LoadLogConfig(cfg)

	if loadedConfig.ProjectName != cfg.ProjectName {
		t.Errorf("Expected ProjectName to be %s, but got %s", cfg.ProjectName, loadedConfig.ProjectName)
	}
	if loadedConfig.AppLog.Name != cfg.AppLog.Name {
		t.Errorf("Expected AppLog.Name to be %s, but got %s", cfg.AppLog.Name, loadedConfig.AppLog.Name)
	}
	if loadedConfig.AppLog.LogFile != cfg.AppLog.LogFile {
		t.Errorf("Expected AppLog.LogFile to be %v, but got %v", cfg.AppLog.LogFile, loadedConfig.AppLog.LogFile)
	}

	// clean up

	os.RemoveAll(cfg.AppLog.Name)
	os.RemoveAll(cfg.Detail.Name)
	os.RemoveAll(cfg.Summary.Name)
	os.RemoveAll("logs")

}
func TestEnsureLogDirExists(t *testing.T) {
	// Test case 1: Directory does not exist and should be created
	dirPath := "./test_logs"
	err := ensureLogDirExists(dirPath)
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	// Check if directory was created
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Errorf("Expected directory %s to be created, but it does not exist", dirPath)
	}

	// Clean up
	os.RemoveAll(dirPath)

	// Test case 2: Directory already exists
	err = os.Mkdir(dirPath, os.ModePerm)
	if err != nil {
		t.Fatalf("Failed to create directory for test case 2: %v", err)
	}

	err = ensureLogDirExists(dirPath)
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	// Clean up -> directory should still exist
	os.RemoveAll(dirPath)
}
