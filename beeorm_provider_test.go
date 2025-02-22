package zephyrix

import (
	"testing"
)

func TestBeeormProvider(t *testing.T) {
	bee := beeormProvider()

	_ = bee
	t.Logf("BeeORM engine created\n")
}
