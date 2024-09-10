package qyweixin

import (
	"context"
	"os"
	"testing"
)

func TestQYWeiXinReport_Report(t *testing.T) {
	type args struct {
		worker  string
		content string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"", args{"test", "test error"}, false},
	}
	ctx := context.Background()
	r := NewReport(os.Getenv("QY_WECHAT_TOKEN"))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := r.Report(ctx, tt.args.worker, tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("Report() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
