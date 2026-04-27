// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBindEnvDefaults(t *testing.T) {
	t.Run("env var overrides default when flag not set", func(t *testing.T) {
		t.Setenv("TEST_MY_FLAG", "from-env")

		cmd := &cobra.Command{Use: "test"}
		var val string
		cmd.Flags().StringVar(&val, "my-flag", "original", "")
		require.NoError(t, cmd.ParseFlags(nil))

		bindEnvDefaults(cmd, "TEST")
		assert.Equal(t, "from-env", val)
	})

	t.Run("explicit flag wins over env var", func(t *testing.T) {
		t.Setenv("TEST_MY_FLAG", "from-env")

		cmd := &cobra.Command{Use: "test"}
		var val string
		cmd.Flags().StringVar(&val, "my-flag", "original", "")
		require.NoError(t, cmd.ParseFlags([]string{"--my-flag", "from-cli"}))

		bindEnvDefaults(cmd, "TEST")
		assert.Equal(t, "from-cli", val)
	})

	t.Run("no env var keeps original default", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		var val string
		cmd.Flags().StringVar(&val, "some-flag", "default-val", "")
		require.NoError(t, cmd.ParseFlags(nil))

		bindEnvDefaults(cmd, "TEST")
		assert.Equal(t, "default-val", val)
	})
}
