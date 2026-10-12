package common

import (
	"net/http"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestEmbedFileSystemLeavesHTMLEntrypointsForDynamicRendering(t *testing.T) {
	fileSystem := &embedFileSystem{FileSystem: http.FS(fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("index")},
	})}

	require.False(t, fileSystem.Exists("/", "/"))
	require.False(t, fileSystem.Exists("/", "/index.html"))
}

func TestHomePageContentSnapshotStaysStableUntilReset(t *testing.T) {
	publicHomePageSnapshot.Lock()
	previousContent := publicHomePageSnapshot.content
	previousReady := publicHomePageSnapshot.ready
	publicHomePageSnapshot.Unlock()
	OptionMapRWMutex.Lock()
	previousOptions := OptionMap
	OptionMap = map[string]string{"HomePageContent": "live content"}
	OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		publicHomePageSnapshot.Lock()
		publicHomePageSnapshot.content = previousContent
		publicHomePageSnapshot.ready = previousReady
		publicHomePageSnapshot.Unlock()
		OptionMapRWMutex.Lock()
		OptionMap = previousOptions
		OptionMapRWMutex.Unlock()
	})

	SetHomePageContentSnapshot("startup content")
	OptionMapRWMutex.Lock()
	OptionMap["HomePageContent"] = "administrator update"
	OptionMapRWMutex.Unlock()

	require.Equal(t, "startup content", GetHomePageContentSnapshot())
	SetHomePageContentSnapshot("next startup content")
	require.Equal(t, "next startup content", GetHomePageContentSnapshot())
}
