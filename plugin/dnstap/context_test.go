package dnstap

import (
	"context"
	"testing"
)

func TestDnstapContext(t *testing.T) {
	ctx := contextWithTapper(context.TODO(), Dnstap{})
	if tapper := TapperFromContext(ctx); tapper == nil {
		t.Fatal("Can't get tapper")
	}
}
