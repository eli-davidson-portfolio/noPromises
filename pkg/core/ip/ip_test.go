package ip

import "testing"

func TestNewIP(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		wantType Type
	}{
		{
			name:     "normal IP",
			data:     "test data",
			wantType: Normal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.data)
			if got.Type != tt.wantType {
				t.Errorf("New() type = %v, want %v", got.Type, tt.wantType)
			}
			if got.Data != tt.data {
				t.Errorf("New() data = %v, want %v", got.Data, tt.data)
			}
			if got.Metadata == nil {
				t.Error("New() metadata should not be nil")
			}
		})
	}
}

func TestNewInitialIP(t *testing.T) {
	data := "config data"
	ip := NewInitial(data)

	if ip.Type != Initial {
		t.Errorf("NewInitial() type = %v, want %v", ip.Type, Initial)
	}
	if ip.Data != data {
		t.Errorf("NewInitial() data = %v, want %v", ip.Data, data)
	}
	if ip.Metadata == nil {
		t.Error("NewInitial() metadata should not be nil")
	}
}
