package huehttp

import (
	"reflect"
	"testing"

	"hueshelly/hue"
)

func TestParseRoomName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{name: "valid", path: "/toggle/lights/group/Living Room", want: "Living Room"},
		{name: "invalid prefix", path: "/groups/Living Room", wantErr: true},
		{name: "empty", path: "/toggle/lights/group/", wantErr: true},
		{name: "contains slash", path: "/toggle/lights/group/living/room", wantErr: true},
		{name: "too long", path: "/toggle/lights/group/123456789012345678901234567890123", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseRoomName(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseRoomName() error = nil, want non-nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseRoomName() error = %v, want nil", err)
			}
			if got != tt.want {
				t.Fatalf("parseRoomName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseLightID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		want    int
		wantErr bool
	}{
		{name: "valid", path: "/toggle/light/3", want: 3},
		{name: "invalid prefix", path: "/toggle/lights/3", wantErr: true},
		{name: "empty", path: "/toggle/light/", wantErr: true},
		{name: "not integer", path: "/toggle/light/a", wantErr: true},
		{name: "zero", path: "/toggle/light/0", wantErr: true},
		{name: "contains slash", path: "/toggle/light/3/extra", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseLightID(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseLightID() error = nil, want non-nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseLightID() error = %v, want nil", err)
			}
			if got != tt.want {
				t.Fatalf("parseLightID() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestIsToggleMethod(t *testing.T) {
	t.Parallel()

	if !isToggleMethod("GET") {
		t.Fatalf("isToggleMethod(GET) = false, want true")
	}
	if !isToggleMethod("POST") {
		t.Fatalf("isToggleMethod(POST) = false, want true")
	}
	if isToggleMethod("DELETE") {
		t.Fatalf("isToggleMethod(DELETE) = true, want false")
	}
}

func TestCollectRooms(t *testing.T) {
	t.Parallel()

	groups := []hue.Group{
		{Name: "Kitchen"},
		{Name: "Bedroom"},
		{Name: "Living Room"},
	}

	got := collectRooms(groups)
	want := []roomResponse{
		{Name: "Bedroom"},
		{Name: "Kitchen"},
		{Name: "Living Room"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("collectRooms() = %#v, want %#v", got, want)
	}
}

func TestCollectLights(t *testing.T) {
	t.Parallel()

	groups := []hue.Group{
		{
			Name: "Kitchen",
			Lights: []hue.Light{
				{Name: "Counter", ID: 3},
				{Name: "Ceiling", ID: 2},
			},
		},
		{
			Name: "Bedroom",
			Lights: []hue.Light{
				{Name: "Bedside", ID: 1},
			},
		},
	}

	got := collectLights(groups)
	want := []lightResponse{
		{ID: 1, Name: "Bedside", Room: "Bedroom"},
		{ID: 2, Name: "Ceiling", Room: "Kitchen"},
		{ID: 3, Name: "Counter", Room: "Kitchen"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("collectLights() = %#v, want %#v", got, want)
	}
}
