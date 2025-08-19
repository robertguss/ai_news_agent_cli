package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOK(t *testing.T) {
	assert.True(t, OK())
}
