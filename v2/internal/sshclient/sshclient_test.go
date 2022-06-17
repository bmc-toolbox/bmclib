package sshclient

import "testing"

func Test_checkAndBuildAddr(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "OK: only IPv4 address",
			args: args{
				"127.0.0.1",
			},
			want:    "127.0.0.1:22",
			wantErr: false,
		},
		{
			name: "OK: only IPv6 address",
			args: args{
				"fe80::1",
			},
			want:    "[fe80::1]:22",
			wantErr: false,
		},
		{
			name: "OK: only host",
			args: args{
				"localhost",
			},
			want:    "localhost:22",
			wantErr: false,
		},
		{
			name: "OK: IPv4 address with port",
			args: args{
				"127.0.0.1:2222",
			},
			want:    "127.0.0.1:2222",
			wantErr: false,
		},
		{
			name: "OK: IPv6 address with port",
			args: args{
				"[fe80::1]:2222",
			},
			want:    "[fe80::1]:2222",
			wantErr: false,
		},
		{
			name: "OK: host with port",
			args: args{
				"localhost:2222",
			},
			want:    "localhost:2222",
			wantErr: false,
		},
		{
			name: "Not OK: empty addr",
			args: args{
				"",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkAndBuildAddr(tt.args.addr)
			t.Log(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkAndBuildAddr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("checkAndBuildAddr() got = %v, want %v", got, tt.want)
			}
		})
	}
}
