#!/bin/bash

# Example script showing different ways to use the miko-shell CLI tool

echo "=== Miko-Shell CLI Tool Demo ==="
echo

# Initialize a new project
echo "1. Initializing new project..."
./miko-shell init

echo
echo "2. Building container image..."
./miko-shell image build

echo
echo "3. Running simple commands..."
./miko-shell run echo "Hello from container!"
./miko-shell run whoami
./miko-shell run pwd

echo
echo "4. Running predefined script..."
./miko-shell run lint

echo
echo "5. Testing with different container providers..."
# Create a podman config
cat > miko-shell-podman.yaml << 'EOF'
provider: podman
image: alpine:latest
bootstrap:
  - apk add curl
shell:
  shell-init:
    - echo "Using Podman!"
  scripts:
    - name: test-curl
      cmds:
        - curl --version
EOF

echo "   - Created podman configuration"

# You can test with podman if available:
# cp miko-shell-podman.yaml miko-shell.yaml
# ./miko-shell run test-curl

echo
echo "6. Interactive shell example (uncomment to try):"
echo "   ./miko-shell shell"

echo
echo "Demo completed!"
echo "Check the README.md for more examples and documentation."
