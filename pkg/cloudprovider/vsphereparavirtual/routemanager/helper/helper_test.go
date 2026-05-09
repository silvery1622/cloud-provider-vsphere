package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testClustername = "test-cluster"
	testCIDR        = "100.96.0.0/24"
	testNodeName    = "fakeNode1"
)

func TestGetRouteName(t *testing.T) {
	name := GetRouteName(testNodeName, testCIDR, testClustername)
	expectedName := testNodeName + "-100.96.0.0-24-" + testClustername
	assert.Equal(t, name, expectedName)
}

func TestStripFamilySuffix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"node-1", "node-1"},
		{"node-1" + SuffixIPv6, "node-1"},
		{"my-node" + SuffixIPv6, "my-node"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := StripFamilySuffix(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
