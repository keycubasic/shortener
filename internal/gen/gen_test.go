package gen

import "testing"

func TestGenKey(t *testing.T) {
	tests := []struct {
		name string
		link string
		want string
	}{
		{
			name: "TestGenKeyOK",
			link: "https://practicum.yandex.ru",
			want: "a9tbDian",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenKey(tt.link); got != tt.want {
				t.Errorf("GenKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
