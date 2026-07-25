package settings

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"
)

type fakeCeph struct{ payload json.RawMessage }

func (f fakeCeph) Raw(context.Context, string, string, url.Values, any) (json.RawMessage, error) {
	return f.payload, nil
}

func TestListRedactsSensitiveValues(t *testing.T) {
	service := New(fakeCeph{payload: json.RawMessage(`[{"name":"GRAFANA_API_PASSWORD","value":"secret"}]`)}, nil)
	result, err := service.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	items := result.([]Setting)
	if len(items) != 1 || items[0].Value != "********" || !items[0].Sensitive {
		t.Fatalf("List() = %#v", items)
	}
}

func TestGetGroupRejectsUnknownGroup(t *testing.T) {
	service := New(fakeCeph{}, nil)
	if _, err := service.GetGroup(context.Background(), "missing"); err != ErrGroupNotFound {
		t.Fatalf("GetGroup() error = %v", err)
	}
}
