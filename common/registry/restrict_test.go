package registry

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestEnforceRestrictionWithOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		config        proto.Message
		allow         bool
		allowIfSet    string
		wantErrSubstr string
	}{
		{
			name:       "unconditional option allows load",
			config:     &descriptorpb.FileOptions{},
			allow:      true,
			allowIfSet: "field_that_does_not_exist",
		},
		{
			name:       "conditional bool true allows load",
			config:     &descriptorpb.FileOptions{JavaMultipleFiles: proto.Bool(true)},
			allowIfSet: "java_multiple_files",
		},
		{
			name:          "conditional bool false denies load",
			config:        &descriptorpb.FileOptions{JavaMultipleFiles: proto.Bool(false)},
			allowIfSet:    "java_multiple_files",
			wantErrSubstr: "component has not opted in for load in restricted mode",
		},
		{
			name:          "unset conditional bool denies load",
			config:        &descriptorpb.FileOptions{},
			allowIfSet:    "java_multiple_files",
			wantErrSubstr: "component has not opted in for load in restricted mode",
		},
		{
			name:          "unset conditional bool with true schema default denies load",
			config:        &descriptorpb.FileOptions{},
			allowIfSet:    "cc_enable_arenas",
			wantErrSubstr: "component has not opted in for load in restricted mode",
		},
		{
			name:       "explicit conditional bool matching true schema default allows load",
			config:     &descriptorpb.FileOptions{CcEnableArenas: proto.Bool(true)},
			allowIfSet: "cc_enable_arenas",
		},
		{
			name:          "no opt-in denies load",
			config:        &descriptorpb.FileOptions{},
			wantErrSubstr: "component has not opted in for load in restricted mode",
		},
		{
			name:          "unknown conditional field fails closed",
			config:        &descriptorpb.FileOptions{},
			allowIfSet:    "field_that_does_not_exist",
			wantErrSubstr: "allow_restricted_mode_load_if_set references unknown field: field_that_does_not_exist",
		},
		{
			name:          "JSON field name is not accepted",
			config:        &descriptorpb.FileOptions{JavaMultipleFiles: proto.Bool(true)},
			allowIfSet:    "javaMultipleFiles",
			wantErrSubstr: "allow_restricted_mode_load_if_set references unknown field: javaMultipleFiles",
		},
		{
			name:          "non-bool conditional field fails closed",
			config:        &descriptorpb.FileOptions{JavaPackage: proto.String("example")},
			allowIfSet:    "java_package",
			wantErrSubstr: "allow_restricted_mode_load_if_set must reference a singular bool field: java_package",
		},
		{
			name:          "repeated conditional field fails closed",
			config:        &descriptorpb.FileDescriptorProto{},
			allowIfSet:    "dependency",
			wantErrSubstr: "allow_restricted_mode_load_if_set must reference a singular bool field: dependency",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := enforceRestrictionWithOptions(test.config, test.allow, test.allowIfSet)
			if test.wantErrSubstr == "" {
				if err != nil {
					t.Fatalf("enforceRestrictionWithOptions() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("enforceRestrictionWithOptions() error = nil, want error containing %q", test.wantErrSubstr)
			}
			if !strings.Contains(err.Error(), test.wantErrSubstr) {
				t.Fatalf("enforceRestrictionWithOptions() error = %q, want error containing %q", err, test.wantErrSubstr)
			}
		})
	}
}
