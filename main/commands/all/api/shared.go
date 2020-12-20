package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/v2fly/v2ray-core/v4/main/commands/base"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

var (
	apiServerAddrPtr string
	apiTimeout       int
)

func setSharedFlags(cmd *base.Command) {
	cmd.Flag.StringVar(&apiServerAddrPtr, "s", "127.0.0.1:8080", "")
	cmd.Flag.StringVar(&apiServerAddrPtr, "server", "127.0.0.1:8080", "")
	cmd.Flag.IntVar(&apiTimeout, "t", 3, "")
	cmd.Flag.IntVar(&apiTimeout, "timeout", 3, "")
}

func dialAPIServer() (conn *grpc.ClientConn, ctx context.Context, close func()) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(apiTimeout)*time.Second)
	conn, err := grpc.DialContext(ctx, apiServerAddrPtr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		base.Fatalf("failed to dial %s", apiServerAddrPtr)
	}
	close = func() {
		cancel()
		conn.Close()
	}
	return
}

func showResponese(m proto.Message) {
	if isEmpty(m) {
		// avoid outputs like `{}`, `{"key":{}}`
		return
	}
	b := new(strings.Builder)
	e := json.NewEncoder(b)
	e.SetIndent("", "  ")
	e.SetEscapeHTML(false)
	err := e.Encode(m)
	if err != nil {
		fmt.Fprintf(os.Stdout, "%v\n", m)
		base.Fatalf("error encode json: %s", err)
		return
	}
	fmt.Println(strings.TrimSpace(b.String()))
}

// isEmpty checks if the response is empty (all zero values).
// proto.Message types always "omitempty" on fields,
// there's no chance for a response to show zero-value messages,
// so we can perform isZero test here
func isEmpty(response interface{}) bool {
	s := reflect.Indirect(reflect.ValueOf(response))
	if s.Kind() == reflect.Invalid {
		return true
	}
	switch s.Kind() {
	case reflect.Struct:
		for i := 0; i < s.NumField(); i++ {
			f := s.Type().Field(i)
			if f.Name[0] < 65 || f.Name[0] > 90 {
				// continue if not exported.
				continue
			}
			field := s.Field(i)
			if !isEmpty(field.Interface()) {
				return false
			}
		}
	case reflect.Array, reflect.Slice:
		for i := 0; i < s.Len(); i++ {
			if !isEmpty(s.Index(i).Interface()) {
				return false
			}
		}
	default:
		if !s.IsZero() {
			return false
		}
	}
	return true
}
