// SPDX-License-Identifier: AGPL-3.0-only
// Provenance-includes-location: https://github.com/cortexproject/cortex/blob/master/pkg/util/shard_test.go
// Provenance-includes-license: Apache-2.0
// Provenance-includes-copyright: The Cortex Authors.

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShuffleShardExpectedInstancesPerZone(t *testing.T) {
	tests := []struct {
		shardSize int
		numZones  int
		expected  int
	}{
		{
			shardSize: 1,
			numZones:  1,
			expected:  1,
		},
		{
			shardSize: 1,
			numZones:  3,
			expected:  1,
		},
		{
			shardSize: 3,
			numZones:  3,
			expected:  1,
		},
		{
			shardSize: 4,
			numZones:  3,
			expected:  2,
		},
		{
			shardSize: 6,
			numZones:  3,
			expected:  2,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, ShuffleShardExpectedInstancesPerZone(test.shardSize, test.numZones))
	}
}

func TestShuffleShardExpectedInstances(t *testing.T) {
	tests := []struct {
		shardSize int
		numZones  int
		expected  int
	}{
		{
			shardSize: 1,
			numZones:  1,
			expected:  1,
		},
		{
			shardSize: 1,
			numZones:  3,
			expected:  3,
		},
		{
			shardSize: 3,
			numZones:  3,
			expected:  3,
		},
		{
			shardSize: 4,
			numZones:  3,
			expected:  6,
		},
		{
			shardSize: 6,
			numZones:  3,
			expected:  6,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, ShuffleShardExpectedInstances(test.shardSize, test.numZones))
	}
}

func TestYoloBuf(t *testing.T) {
	s := yoloBuf("hello world")

	require.Equal(t, []byte("hello world"), s)
}
