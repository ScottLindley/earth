package nasa

import "testing"

func Test_getPreviousDate(t *testing.T) {
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
			want: "2020-11-14",
		},
		{
			name: "When the date on the first of a month",
			args: args{date: "2020-11-01"},
			want: "2020-10-31",
		},
		{
			name: "When the date is on the first of a year",
			args: args{date: "2020-01-01"},
			want: "2019-12-31",
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
			got, err := getPreviousDate(tt.args.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPreviousDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getPreviousDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
