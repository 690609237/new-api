package common

import (
	"net/http"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestEmbedFileSystemLeavesHTMLEntrypointsForRouterRendering(t *testing.T) {
	fileSystem := &embedFileSystem{FileSystem: http.FS(fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("index")},
	})}

	require.False(t, fileSystem.Exists("/", "/"))
	require.False(t, fileSystem.Exists("/", "/index.html"))
}
