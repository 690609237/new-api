package common

import "sync"

var publicHomePageSnapshot struct {
	sync.RWMutex
	content string
	ready   bool
}

// SetHomePageContentSnapshot stores the public home-page content captured at
// startup. The public page intentionally keeps this value stable until the
// next process restart.
func SetHomePageContentSnapshot(content string) {
	publicHomePageSnapshot.Lock()
	publicHomePageSnapshot.content = content
	publicHomePageSnapshot.ready = true
	publicHomePageSnapshot.Unlock()
}

// GetHomePageContentSnapshot returns the startup snapshot when one has been
// initialized. The fallback keeps tests and tools that call the controller
// without the normal startup sequence compatible with the live option map.
func GetHomePageContentSnapshot() string {
	publicHomePageSnapshot.RLock()
	if publicHomePageSnapshot.ready {
		content := publicHomePageSnapshot.content
		publicHomePageSnapshot.RUnlock()
		return content
	}
	publicHomePageSnapshot.RUnlock()

	OptionMapRWMutex.RLock()
	content := OptionMap["HomePageContent"]
	OptionMapRWMutex.RUnlock()
	return content
}
