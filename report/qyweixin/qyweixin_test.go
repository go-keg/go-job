package qyweixin

import (
	"context"
	"errors"
	"os"
	"testing"
)

func TestQYWeiXinReport_Report(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"", args{errors.New("test error")}, false},
	}
	ctx := context.Background()
	r := NewReport(os.Getenv("QY_WECHAT_TOKEN"))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := r.Report(ctx, tt.args.err); (err != nil) != tt.wantErr {
				t.Errorf("Report() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
