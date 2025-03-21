// SPDX-License-Identifier: AGPL-3.0-only
// Provenance-includes-location: https://github.com/cortexproject/cortex/blob/master/pkg/ruler/rulestore/local/local_test.go
// Provenance-includes-license: Apache-2.0
// Provenance-includes-copyright: The Cortex Authors.

package local

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/rulefmt"
	promRules "github.com/prometheus/prometheus/rules"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/grafana/mimir/pkg/ruler/rulespb"
	"github.com/grafana/mimir/pkg/ruler/rulestore"
)

func TestClient_LoadRuleGroups(t *testing.T) {
	user1 := "user"
	user2 := "second-user"

	namespace1 := "ns"
	namespace2 := "z-another" // This test relies on the fact that os.ReadDir() returns files sorted by name.

	dir := t.TempDir()

	ruleGroups := rulefmt.RuleGroups{
		Groups: []rulefmt.RuleGroup{
			{
				Name:     "rule",
				Interval: model.Duration(100 * time.Second),
				Rules: []rulefmt.Rule{
					{
						Record: "test_rule",
						Expr:   "up",
					},
				},
			},
		},
	}

	b, err := yaml.Marshal(ruleGroups)
	require.NoError(t, err)

	err = os.MkdirAll(path.Join(dir, user1), 0777)
	require.NoError(t, err)

	// Link second user to first.
	err = os.Symlink(user1, path.Join(dir, user2))
	require.NoError(t, err)

	err = os.WriteFile(path.Join(dir, user1, namespace1), b, 0777)
	require.NoError(t, err)

	const ignoredDir = "ignored-dir"
	err = os.Mkdir(path.Join(dir, user1, ignoredDir), os.ModeDir|0644)
	require.NoError(t, err)

	err = os.Symlink(ignoredDir, path.Join(dir, user1, "link-to-dir"))
	require.NoError(t, err)

	// Link second namespace to first.
	err = os.Symlink(namespace1, path.Join(dir, user1, namespace2))
	require.NoError(t, err)

	client, err := NewLocalRulesClient(rulestore.LocalStoreConfig{
		Directory: dir,
	}, promRules.FileLoader{})
	require.NoError(t, err)

	t.Run("all rule groups", func(t *testing.T) {
		ctx := context.Background()

		for _, u := range []string{user1, user2} {
			rgs, err := client.ListRuleGroupsForUserAndNamespace(ctx, u, "") // Client loads rules in its List method.
			require.NoError(t, err)

			require.Equal(t, 2, len(rgs))
			// We rely on the fact that files are parsed in alphabetical order, and our namespace1 < namespace2.
			require.Equal(t, rulespb.ToProto(u, namespace1, ruleGroups.Groups[0]), rgs[0])
			require.Equal(t, rulespb.ToProto(u, namespace2, ruleGroups.Groups[0]), rgs[1])
		}
	})

	t.Run("all rule groups in namespace", func(t *testing.T) {
		ctx := context.Background()

		for _, u := range []string{user1, user2} {
			rgs, err := client.ListRuleGroupsForUserAndNamespace(ctx, u, namespace2) // Client loads rules in its List method.
			require.NoError(t, err)

			require.Equal(t, 1, len(rgs))
			require.Equal(t, rulespb.ToProto(u, namespace2, ruleGroups.Groups[0]), rgs[0])
		}
	})

	t.Run("single rule group in namespace", func(t *testing.T) {
		rg, err := client.GetRuleGroup(context.Background(), user1, namespace1, "rule")
		require.NoError(t, err)
		require.Equal(t, rulespb.ToProto(user1, namespace1, ruleGroups.Groups[0]), rg)
	})

	t.Run("single rule group that not exists", func(t *testing.T) {
		_, err := client.GetRuleGroup(context.Background(), user1, namespace1, "unknown-rule")
		require.ErrorIs(t, err, rulestore.ErrGroupNotFound)
	})
}
