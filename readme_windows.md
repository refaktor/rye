# Getting Started with Rye on Windows

This guide provides instructions for installing and running Rye on Windows systems.

## Option 1: Download Pre-built Binary

The simplest way to get started with Rye on Windows is to download a pre-built binary:

1. Visit the [Rye Releases page](https://github.com/refaktor/rye/releases)
2. Download the latest Windows binary (look for `rye_windows_amd64.zip` or similar)
3. Extract the ZIP file to a location of your choice
4. Add the extracted folder to your PATH environment variable (optional but recommended)
   - Right-click on "This PC" or "My Computer" and select "Properties"
   - Click on "Advanced system settings"
   - Click the "Environment Variables" button
   - Under "System variables", find the "Path" variable, select it and click "Edit"
   - Click "New" and add the path to the folder containing the Rye executable
   - Click "OK" on all dialogs to save the changes
5. Open Command Prompt or PowerShell and type `rye` to start the Rye console

## Option 2: Build from Source

If you prefer to build Rye from source:

1. Install Git for Windows
   - Download from [git-scm.com](https://git-scm.com/download/win)
   - Follow the installation wizard with default options

2. Install Go (Golang)
   - Download from [go.dev/dl](https://go.dev/dl/)
   - Follow the installation wizard
   - Verify installation by opening Command Prompt and typing `go version`

3. Clone the Rye repository
   ```
   git clone https://github.com/refaktor/rye.git
   cd rye
   ```

4. Optionally switch to a specific branch
   ```
   git checkout v0.0.90prep
   ```
   This is useful if you want to use a specific version or if the main branch is not what you need.

5. Build Rye
   ```
   go build -o bin/rye.exe
   ```
   
   For a minimal version with fewer dependencies:
   ```
   go build -tags "b_tiny" -o bin/rye.exe
   ```

5. Run Rye
   - To start the Rye console: `bin/rye.exe`
   - To run a Rye script: `bin/rye.exe script.rye`

## Editor Support

For the best development experience, we recommend using Visual Studio Code with the Rye extension:

1. Install Visual Studio Code from [code.visualstudio.com](https://code.visualstudio.com/)
2. Open VS Code and go to the Extensions view (Ctrl+Shift+X)
3. Search for "ryelang" in the marketplace
4. Install the Rye language extension for syntax highlighting and other features

## Next Steps

- Check out the [main README](README.md) for more information about Rye
- Visit [ryelang.org](https://ryelang.org/) for documentation and examples
- Join the [Reddit community](https://reddit.com/r/ryelang/) for discussions and updates
