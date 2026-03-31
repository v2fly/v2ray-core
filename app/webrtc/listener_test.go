package webrtc

import (
	"testing"

	pionwebrtc "github.com/pion/webrtc/v4"
)

func TestActiveListenerICEConfiguration(t *testing.T) {
	t.Run("stun only", func(t *testing.T) {
		cfg, err := activeListenerICEConfiguration(&LocalWebRTCListener{
			Tag:         "active-a",
			StunServers: []string{"stun:stun.example.com:3478"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := cfg.ICETransportPolicy, pionwebrtc.ICETransportPolicyNoHost; got != want {
			t.Fatalf("ICETransportPolicy = %v, want %v", got, want)
		}
		if got, want := len(cfg.ICEServers), 1; got != want {
			t.Fatalf("len(ICEServers) = %d, want %d", got, want)
		}
		if got, want := cfg.ICEServers[0].URLs[0], "stun:stun.example.com:3478"; got != want {
			t.Fatalf("ICEServers[0].URLs[0] = %q, want %q", got, want)
		}
	})

	t.Run("udp turn only", func(t *testing.T) {
		cfg, err := activeListenerICEConfiguration(&LocalWebRTCListener{
			Tag: "active-a",
			TurnServers: []*WebRTCTURNServer{
				{
					Url:      "turn:turn.example.com:3478?transport=udp",
					Username: "user-a",
					Password: "pass-a",
				},
				{
					Url:      "turn:turn2.example.com",
					Username: "user-b",
					Password: "pass-b",
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(cfg.ICEServers), 2; got != want {
			t.Fatalf("len(ICEServers) = %d, want %d", got, want)
		}
		if got, want := cfg.ICEServers[0].URLs[0], "turn:turn.example.com:3478?transport=udp"; got != want {
			t.Fatalf("ICEServers[0].URLs[0] = %q, want %q", got, want)
		}
		if got, want := cfg.ICEServers[0].Username, "user-a"; got != want {
			t.Fatalf("ICEServers[0].Username = %q, want %q", got, want)
		}
		if got, want := cfg.ICEServers[0].Credential, any("pass-a"); got != want {
			t.Fatalf("ICEServers[0].Credential = %#v, want %#v", got, want)
		}
		if got, want := cfg.ICEServers[0].CredentialType, pionwebrtc.ICECredentialTypePassword; got != want {
			t.Fatalf("ICEServers[0].CredentialType = %v, want %v", got, want)
		}
		if got, want := cfg.ICEServers[1].URLs[0], "turn:turn2.example.com"; got != want {
			t.Fatalf("ICEServers[1].URLs[0] = %q, want %q", got, want)
		}
		if got, want := cfg.ICEServers[1].Username, "user-b"; got != want {
			t.Fatalf("ICEServers[1].Username = %q, want %q", got, want)
		}
		if got, want := cfg.ICEServers[1].Credential, any("pass-b"); got != want {
			t.Fatalf("ICEServers[1].Credential = %#v, want %#v", got, want)
		}
	})

	t.Run("missing ice servers", func(t *testing.T) {
		_, err := activeListenerICEConfiguration(&LocalWebRTCListener{Tag: "active-a"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("missing turn credentials", func(t *testing.T) {
		_, err := activeListenerICEConfiguration(&LocalWebRTCListener{
			Tag: "active-a",
			TurnServers: []*WebRTCTURNServer{
				{Url: "turn:turn.example.com:3478?transport=udp"},
			},
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("reject tcp turn", func(t *testing.T) {
		_, err := activeListenerICEConfiguration(&LocalWebRTCListener{
			Tag: "active-a",
			TurnServers: []*WebRTCTURNServer{
				{
					Url:      "turn:turn.example.com:3478?transport=tcp",
					Username: "user",
					Password: "pass",
				},
			},
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("reject turns", func(t *testing.T) {
		_, err := activeListenerICEConfiguration(&LocalWebRTCListener{
			Tag: "active-a",
			TurnServers: []*WebRTCTURNServer{
				{
					Url:      "turns:turn.example.com:5349",
					Username: "user",
					Password: "pass",
				},
			},
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
