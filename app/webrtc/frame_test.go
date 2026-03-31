package webrtc

import "testing"

func TestWireFrameRoundTrip(t *testing.T) {
	t.Run("open", func(t *testing.T) {
		encoded := encodeWireFrame(wireFrame{
			FrameType: frameTypeOpen,
			StreamID:  7,
			Tag:       "dns",
		})

		decoded, err := decodeWireFrame(encoded)
		if err != nil {
			t.Fatal(err)
		}
		if decoded.FrameType != frameTypeOpen || decoded.StreamID != 7 || decoded.Tag != "dns" {
			t.Fatalf("unexpected open frame: %#v", decoded)
		}
	})

	t.Run("data", func(t *testing.T) {
		payload := []byte("payload")
		encoded := encodeWireFrame(wireFrame{
			FrameType: frameTypeData,
			StreamID:  9,
			Payload:   payload,
		})

		decoded, err := decodeWireFrame(encoded)
		if err != nil {
			t.Fatal(err)
		}
		if decoded.FrameType != frameTypeData || decoded.StreamID != 9 || string(decoded.Payload) != string(payload) {
			t.Fatalf("unexpected data frame: %#v", decoded)
		}
	})

	t.Run("close", func(t *testing.T) {
		encoded := encodeWireFrame(wireFrame{
			FrameType: frameTypeClose,
			StreamID:  11,
		})

		decoded, err := decodeWireFrame(encoded)
		if err != nil {
			t.Fatal(err)
		}
		if decoded.FrameType != frameTypeClose || decoded.StreamID != 11 {
			t.Fatalf("unexpected close frame: %#v", decoded)
		}
	})
}

func TestDecodeWireFrameRejectsInvalidLength(t *testing.T) {
	if _, err := decodeWireFrame([]byte{frameTypeData, 0, 0, 0, 1, 0, 0, 0, 4, 1, 2}); err == nil {
		t.Fatal("expected invalid data frame length error")
	}
}

func TestDrainCandidateSetHandlesClosedChannel(t *testing.T) {
	candidates := make(chan []byte)
	close(candidates)

	got := drainCandidateSet(candidates, map[string]struct{}{}, nil)
	if len(got) != 0 {
		t.Fatalf("expected no candidates, got %d", len(got))
	}
}
