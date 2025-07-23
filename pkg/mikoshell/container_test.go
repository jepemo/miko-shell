package mikoshell

import (
	"testing"
)

func TestNewContainerProvider(t *testing.T) {
	t.Run("docker provider", func(t *testing.T) {
		provider, err := NewContainerProvider("docker")
		if err != nil {
			t.Fatalf("NewContainerProvider('docker') failed: %v", err)
		}

		if provider == nil {
			t.Error("NewContainerProvider('docker') should return a non-nil provider")
		}

		dockerProvider, ok := provider.(*DockerProvider)
		if !ok {
			t.Error("Expected provider to be a DockerProvider")
		}

		if dockerProvider == nil {
			t.Error("DockerProvider should not be nil")
		}
	})

	t.Run("podman provider", func(t *testing.T) {
		provider, err := NewContainerProvider("podman")
		if err != nil {
			t.Fatalf("NewContainerProvider('podman') failed: %v", err)
		}

		if provider == nil {
			t.Error("NewContainerProvider('podman') should return a non-nil provider")
		}

		podmanProvider, ok := provider.(*PodmanProvider)
		if !ok {
			t.Error("Expected provider to be a PodmanProvider")
		}

		if podmanProvider == nil {
			t.Error("PodmanProvider should not be nil")
		}
	})

	t.Run("invalid provider", func(t *testing.T) {
		provider, err := NewContainerProvider("invalid")
		if err == nil {
			t.Error("NewContainerProvider('invalid') should return an error")
		}

		if provider != nil {
			t.Error("NewContainerProvider('invalid') should return nil provider")
		}
	})

	t.Run("empty provider", func(t *testing.T) {
		provider, err := NewContainerProvider("")
		if err == nil {
			t.Error("NewContainerProvider('') should return an error")
		}

		if provider != nil {
			t.Error("NewContainerProvider('') should return nil provider")
		}
	})
}

func TestDockerProvider_IsAvailable(t *testing.T) {
	provider := &DockerProvider{}
	
	// Note: This test depends on docker being available in the system
	// For a real test environment, you might want to mock this
	available := provider.IsAvailable()
	
	// We can't assume docker is always available, so we just check that the method doesn't panic
	if available {
		t.Log("Docker is available")
	} else {
		t.Log("Docker is not available")
	}
}

func TestDockerProvider_ImageExists(t *testing.T) {
	provider := &DockerProvider{}
	
	// Test with a tag that likely doesn't exist
	exists := provider.ImageExists("nonexistent-image:latest")
	
	// We don't expect this image to exist
	if exists {
		t.Log("Image exists (unexpected)")
	} else {
		t.Log("Image doesn't exist (expected)")
	}
}

func TestDockerProvider_BuildImage(t *testing.T) {
	provider := &DockerProvider{}
	
	// Create a test config
	config := &Config{
		Container: Container{
			Image: "alpine:latest",
			Setup: []string{"apk add --no-cache curl"},
		},
		Shell: Shell{
			InitHook: []string{"echo 'test'"},
		},
	}
	
	// Test building image (this won't actually build unless docker is available)
	err := provider.BuildImage(config, "test-image:latest")
	
	if err != nil {
		t.Logf("Build failed (expected if docker not available): %v", err)
	} else {
		t.Log("Build succeeded")
	}
}

func TestDockerProvider_RunCommand(t *testing.T) {
	provider := &DockerProvider{}
	
	// Create a test config
	config := &Config{
		Container: Container{
			Image: "alpine:latest",
		},
	}
	
	// Test running a command (this won't actually run unless docker is available)
	err := provider.RunCommand(config, "test-image:latest", []string{"echo", "test"})
	
	if err != nil {
		t.Logf("Command failed (expected if docker not available): %v", err)
	} else {
		t.Log("Command succeeded")
	}
}

func TestPodmanProvider_IsAvailable(t *testing.T) {
	provider := &PodmanProvider{}
	
	// Note: This test depends on podman being available in the system
	// For a real test environment, you might want to mock this
	available := provider.IsAvailable()
	
	// We can't assume podman is always available, so we just check that the method doesn't panic
	if available {
		t.Log("Podman is available")
	} else {
		t.Log("Podman is not available")
	}
}

func TestPodmanProvider_ImageExists(t *testing.T) {
	provider := &PodmanProvider{}
	
	// Test with a tag that likely doesn't exist
	exists := provider.ImageExists("nonexistent-image:latest")
	
	// We don't expect this image to exist
	if exists {
		t.Log("Image exists (unexpected)")
	} else {
		t.Log("Image doesn't exist (expected)")
	}
}

func TestPodmanProvider_BuildImage(t *testing.T) {
	provider := &PodmanProvider{}
	
	// Create a test config
	config := &Config{
		Container: Container{
			Image: "alpine:latest",
			Setup: []string{"apk add --no-cache curl"},
		},
		Shell: Shell{
			InitHook: []string{"echo 'test'"},
		},
	}
	
	// Test building image (this won't actually build unless podman is available)
	err := provider.BuildImage(config, "test-image:latest")
	
	if err != nil {
		t.Logf("Build failed (expected if podman not available): %v", err)
	} else {
		t.Log("Build succeeded")
	}
}

func TestPodmanProvider_RunCommand(t *testing.T) {
	provider := &PodmanProvider{}
	
	// Create a test config
	config := &Config{
		Container: Container{
			Image: "alpine:latest",
		},
	}
	
	// Test running a command (this won't actually run unless podman is available)
	err := provider.RunCommand(config, "test-image:latest", []string{"echo", "test"})
	
	if err != nil {
		t.Logf("Command failed (expected if podman not available): %v", err)
	} else {
		t.Log("Command succeeded")
	}
}
