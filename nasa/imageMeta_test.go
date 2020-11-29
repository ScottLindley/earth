package nasa

import "testing"

func Test_getNextDate(t *testing.T) {
	type args struct {
		date string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "When the date is in the middle of a month",
			args: args{date: "2020-11-15"},
			want: "2020-11-16",
		},
		{
			name: "When the date on the last day of a month",
			args: args{date: "2020-10-31"},
			want: "2020-11-01",
		},
		{
			name: "When the date is on the lat day of a year",
			args: args{date: "2019-12-31"},
			want: "2020-01-01",
		},
		{
			name:    "When the date malformed",
			args:    args{date: "2020-01-"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getNextDate(tt.args.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNextDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getNextDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
