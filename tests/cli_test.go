package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildCLI(t *testing.T) string {
	binPath := filepath.Join(t.TempDir(), "gonest.exe")
	cmd := exec.Command("go", "build", "-o", binPath, "github.com/linuxerlv/gonest/cmd/gonest")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build CLI: %v\n%s", err, output)
	}
	return binPath
}

func TestCLI_PrintUsage(t *testing.T) {
	binPath := buildCLI(t)

	cmd := exec.Command(binPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\n%s", err, output)
	}

	if !strings.Contains(string(output), "gonest new") {
		t.Error("Expected usage to contain 'gonest new'")
	}

	if !strings.Contains(string(output), "gonest controller") {
		t.Error("Expected usage to contain 'gonest controller'")
	}
}

func TestCLI_NewProject(t *testing.T) {
	binPath := buildCLI(t)

	tmpDir := t.TempDir()
	projectName := "mytestproject"

	cmd := exec.Command(binPath, "new", projectName)
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\n%s", err, output)
	}

	projectPath := filepath.Join(tmpDir, projectName)

	expectedDirs := []string{
		projectPath,
		filepath.Join(projectPath, "cmd"),
		filepath.Join(projectPath, "internal/controllers"),
		filepath.Join(projectPath, "internal/services"),
		filepath.Join(projectPath, "internal/middleware"),
		filepath.Join(projectPath, "internal/models"),
		filepath.Join(projectPath, "wire"),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected directory %s to exist", dir)
		}
	}

	expectedFiles := []string{
		filepath.Join(projectPath, "go.mod"),
		filepath.Join(projectPath, "cmd/main.go"),
		filepath.Join(projectPath, "wire/wire.go"),
		filepath.Join(projectPath, "internal/controllers/app.controller.go"),
		filepath.Join(projectPath, "internal/middleware/logger.middleware.go"),
		filepath.Join(projectPath, "internal/services/user.service.go"),
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", file)
		}
	}

	goModContent, err := os.ReadFile(filepath.Join(projectPath, "go.mod"))
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	if !strings.Contains(string(goModContent), "module "+projectName) {
		t.Error("Expected go.mod to contain module name")
	}
}

func TestCLI_NewProject_NoName(t *testing.T) {
	binPath := buildCLI(t)

	cmd := exec.Command(binPath, "new")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\n%s", err, output)
	}

	if !strings.Contains(string(output), "Usage: gonest new") {
		t.Error("Expected usage message for 'new' command")
	}
}

func TestCLI_GenerateController(t *testing.T) {
	binPath := buildCLI(t)

	tmpDir := t.TempDir()
	controllersDir := filepath.Join(tmpDir, "internal", "controllers")
	os.MkdirAll(controllersDir, 0755)

	cmd := exec.Command(binPath, "controller", "user")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\n%s", err, output)
	}

	expectedFile := filepath.Join(controllersDir, "user.controller.go")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected controller file %s to exist", expectedFile)
	}

	content, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read controller file: %v", err)
	}

	if !strings.Contains(string(content), "type UserController struct") {
		t.Error("Expected controller file to contain UserController struct")
	}

	if !strings.Contains(string(content), "func NewUserController") {
		t.Error("Expected controller file to contain NewUserController function")
	}
}

func TestCLI_GenerateController_NoName(t *testing.T) {
	binPath := buildCLI(t)

	cmd := exec.Command(binPath, "controller")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\n%s", err, output)
	}

	if !strings.Contains(string(output), "Usage: gonest controller") {
		t.Error("Expected usage message for 'controller' command")
	}
}

func TestCLI_GenerateService(t *testing.T) {
	binPath := buildCLI(t)

	tmpDir := t.TempDir()
	servicesDir := filepath.Join(tmpDir, "internal", "services")
	os.MkdirAll(servicesDir, 0755)

	cmd := exec.Command(binPath, "service", "auth")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\n%s", err, output)
	}

	expectedFile := filepath.Join(servicesDir, "auth.service.go")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected service file %s to exist", expectedFile)
	}

	content, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read service file: %v", err)
	}

	if !strings.Contains(string(content), "type AuthService interface") {
		t.Error("Expected service file to contain AuthService interface")
	}

	if !strings.Contains(string(content), "func NewAuthService") {
		t.Error("Expected service file to contain NewAuthService function")
	}
}

func TestCLI_GenerateService_NoName(t *testing.T) {
	binPath := buildCLI(t)

	cmd := exec.Command(binPath, "service")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\n%s", err, output)
	}

	if !strings.Contains(string(output), "Usage: gonest service") {
		t.Error("Expected usage message for 'service' command")
	}
}

func TestCLI_GenerateMiddleware(t *testing.T) {
	binPath := buildCLI(t)

	tmpDir := t.TempDir()
	middlewareDir := filepath.Join(tmpDir, "internal", "middleware")
	os.MkdirAll(middlewareDir, 0755)

	cmd := exec.Command(binPath, "middleware", "auth")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\n%s", err, output)
	}

	expectedFile := filepath.Join(middlewareDir, "auth.middleware.go")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected middleware file %s to exist", expectedFile)
	}

	content, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read middleware file: %v", err)
	}

	if !strings.Contains(string(content), "func NewAuthMiddleware") {
		t.Error("Expected middleware file to contain NewAuthMiddleware function")
	}
}

func TestCLI_GenerateMiddleware_NoName(t *testing.T) {
	binPath := buildCLI(t)

	cmd := exec.Command(binPath, "middleware")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\n%s", err, output)
	}

	if !strings.Contains(string(output), "Usage: gonest middleware") {
		t.Error("Expected usage message for 'middleware' command")
	}
}

func TestCLI_GenerateGuard(t *testing.T) {
	binPath := buildCLI(t)

	tmpDir := t.TempDir()
	guardsDir := filepath.Join(tmpDir, "internal", "guards")
	os.MkdirAll(guardsDir, 0755)

	cmd := exec.Command(binPath, "guard", "admin")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\n%s", err, output)
	}

	expectedFile := filepath.Join(guardsDir, "admin.guard.go")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected guard file %s to exist", expectedFile)
	}

	content, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read guard file: %v", err)
	}

	if !strings.Contains(string(content), "func NewAdminGuard") {
		t.Error("Expected guard file to contain NewAdminGuard function")
	}
}

func TestCLI_GenerateGuard_NoName(t *testing.T) {
	binPath := buildCLI(t)

	cmd := exec.Command(binPath, "guard")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\n%s", err, output)
	}

	if !strings.Contains(string(output), "Usage: gonest guard") {
		t.Error("Expected usage message for 'guard' command")
	}
}

func TestCLI_UnknownCommand(t *testing.T) {
	binPath := buildCLI(t)

	cmd := exec.Command(binPath, "unknown")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\n%s", err, output)
	}

	if !strings.Contains(string(output), "Usage:") {
		t.Error("Expected usage message for unknown command")
	}
}

func TestCLI_ControllerHyphenName(t *testing.T) {
	binPath := buildCLI(t)

	tmpDir := t.TempDir()
	controllersDir := filepath.Join(tmpDir, "internal", "controllers")
	os.MkdirAll(controllersDir, 0755)

	cmd := exec.Command(binPath, "controller", "user-profile")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\n%s", err, output)
	}

	expectedFile := filepath.Join(controllersDir, "user-profile.controller.go")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected controller file %s to exist", expectedFile)
	}

	content, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read controller file: %v", err)
	}

	if !strings.Contains(string(content), "type UserProfileController struct") {
		t.Error("Expected controller file to contain UserProfileController struct (PascalCase)")
	}
}
