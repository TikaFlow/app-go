package app

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

var defaultChromeArgs = []string{
	"--disable-background-networking",
	"--disable-background-timer-throttling",
	"--disable-backgrounding-occluded-windows",
	"--disable-breakpad",
	"--disable-client-side-phishing-detection",
	"--disable-default-apps",
	"--disable-dev-shm-usage",
	"--disable-infobars",
	"--disable-extensions",
	"--disable-features=site-per-process",
	"--disable-hang-monitor",
	"--disable-ipc-flooding-protection",
	"--disable-popup-blocking",
	"--disable-prompt-on-repost",
	"--disable-renderer-backgrounding",
	"--disable-sync",
	"--disable-windows10-custom-titlebar",
	"--metrics-recording-only",
	"--no-first-run",
	"--no-default-browser-check",
	"--safebrowsing-disable-auto-update",
	"--silent-debugger-extension-api",
	"--password-store=basic",
	"--use-mock-keychain",
	"--disable-translate",
	"--disable-features=Translate",
	// "--kiosk", // fullscreen mode
}

type UI struct {
	chrome *exec.Cmd
	tmpDir string
}

func (this *UI) IsOpen() bool {
	return this.chrome != nil
}

func (this *UI) Open(url, dir string, width, height int, customArgs ...string) error {
	if this.chrome != nil {
		return nil
	}
	if dir == "" {
		name, err := os.MkdirTemp("", "App-Go")
		if err != nil {
			return err
		}
		dir = name
	}
	args := append(defaultChromeArgs, fmt.Sprintf("--app=%s", url))
	args = append(args, fmt.Sprintf("--user-data-dir=%s", dir))
	args = append(args, fmt.Sprintf("--window-size=%d,%d", width, height))
	args = append(args, customArgs...)

	w, err := newChromeWithArgs(locateChrome(), args...)
	if err != nil {
		return err
	}

	go func() {
		_ = w.Wait()
		_ = this.Close()
	}()

	this.chrome = w
	this.tmpDir = dir
	return nil
}

func (this *UI) Close() error {
	if this.chrome == nil {
		return nil
	}
	if state := this.chrome.ProcessState; state == nil || !state.Exited() {
		_ = this.chrome.Process.Kill()
	}
	this.chrome = nil
	if this.tmpDir != "" {
		if err := os.RemoveAll(this.tmpDir); err != nil {
			return err
		}
	}
	return nil
}

func newChromeWithArgs(chromeBinary string, args ...string) (*exec.Cmd, error) {
	if chromeBinary == "" {
		Notify("Chrome not found, please install Google Chrome or Chromium")
		os.Exit(1)
	}

	// start chrome
	c := exec.Command(chromeBinary, args...)
	if err := c.Start(); err != nil {
		return nil, err
	}

	return c, nil
}

// returns a path to Chrome, or empty string if not found.
func locateChrome() string {
	var paths []string
	switch runtime.GOOS {
	case "darwin":
		paths = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/google-chrome",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
		}
	case "windows":
		paths = []string{
			os.Getenv("LocalAppData") + "/Google/Chrome/Application/chrome.exe",
			os.Getenv("ProgramFiles") + "/Google/Chrome/Application/chrome.exe",
			os.Getenv("ProgramFiles(x86)") + "/Google/Chrome/Application/chrome.exe",
			os.Getenv("LocalAppData") + "/Chromium/Application/chrome.exe",
			os.Getenv("ProgramFiles") + "/Chromium/Application/chrome.exe",
			os.Getenv("ProgramFiles(x86)") + "/Chromium/Application/chrome.exe",
			os.Getenv("ProgramFiles(x86)") + "/Microsoft/Edge/Application/msedge.exe",
		}
	default:
		paths = []string{
			"/usr/bin/google-chrome-stable",
			"/usr/bin/google-chrome",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/snap/bin/chromium",
		}
	}

	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		return path
	}
	return ""
}
