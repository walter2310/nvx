# nvx

nvx is a fast and lightweight Node.js version manager written in Go. It allows you to install, switch, and manage multiple Node.js versions easily on Windows (macOS and Linux coming soon).

## ✨ Features

- 📦 Install and manage multiple Node.js versions
- 🔄 Quickly switch between versions
- 🪟 Native support for Windows
- 🛠️ No external dependencies (just a single binary)

## 📥 Installation

### Windows (PowerShell)

Run this command in PowerShell to install nvx:

```powershell
iwr -useb https://raw.githubusercontent.com/walter2310/nvx/main/installer.ps1 | iex
```
This will:

- Download nvx.exe to C:\Users\<user>\.nvx\bin
- Add ~\.nvx\bin to your user PATH
- Make nvx available in your terminal
- ⚠️ **Restart** your terminal after installation.

## 🚀 Usage


### Install a specific Node.js version:

```bash
nvx install 22.20.
```

### Use a specific Node.js version:

```bash
nvx use 22.20.
```

### List all Node.js versions:

```bash
nvx list
```

### Uninstall a specific Node.js version:

```bash
nvx use 22.20.
```

## 🤝 Contributing

Contributions are welcome!
Open an issue or submit a pull request.

### How to Contribute

1. **Fork the repository**
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Commit your changes** (`git commit -m 'Add some amazing feature'`)
4. **Push to the branch** (`git push origin feature/amazing-feature`)
5. **Open a Pull Request**

### Reporting Issues

- Use the GitHub issue tracker
- Provide detailed description and steps to reproduce
- Include your operating system and nvx version
