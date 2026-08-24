package instman

import (
	"context"
	"testing"
)

func TestUnknownInstanceIsRejected(t *testing.T) {
	mgr, err := NewInstanceMgr(context.Background(), &Config{})
	if err != nil {
		t.Fatalf("failed to create instance manager: %v", err)
	}

	if err := mgr.StartInstance(context.Background(), "missing"); err == nil {
		t.Fatal("StartInstance accepted an unknown instance name")
	}
	if err := mgr.StopInstance(context.Background(), "missing"); err == nil {
		t.Fatal("StopInstance accepted an unknown instance name")
	}
	if err := mgr.UntrackInstance(context.Background(), "missing"); err != nil {
		t.Fatalf("UntrackInstance failed for an unknown instance name: %v", err)
	}
}
