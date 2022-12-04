//go:build integration

package pollednotifiertest

import (
	"os"
	"plugin"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/leophys/userz/pkg/notifier"
)

var (
	pluginPath string
)

func init() {
	pluginPath = os.Getenv("PLUGIN_PATH")

	if pluginPath == "" {
		panic("PLUGIN_PATH is mandatory")
	}
}

func TestPolledNotifier(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	plug, err := plugin.Open(pluginPath)
	require.NoError(err)

	sym, err := plug.Lookup("Provider")
	require.NoError(err)

	_, ok := sym.(*notifier.Notifier)
	if !assert.True(ok) {
		t.Logf("Type: %T", sym)
	}
}
