package rand

import "testing"

func TestString(t *testing.T) {
	type args struct {
		nBytes int
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"success", args{nBytes: 5}, "", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, err := String(tt.args.nBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
