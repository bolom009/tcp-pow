package protocol

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParseMessage(t *testing.T) {
	tests := []struct {
		name    string
		msg     string
		want    *Message
		wantErr error
	}{
		{
			name:    "#1: Should return message with header",
			msg:     fmt.Sprintf("%d|", Quit),
			want:    &Message{Header: Quit},
			wantErr: nil,
		},
		{
			name:    "#2: Should return message with header and payload",
			msg:     fmt.Sprintf("%d|%s", RequestChallenge, "payload"),
			want:    &Message{Header: RequestChallenge, Payload: "payload"},
			wantErr: nil,
		},
		{
			name:    "#3: Should return invalid header",
			msg:     "",
			want:    nil,
			wantErr: ErrInvalidHeader,
		},
		{
			name:    "#4: Should return invalid message",
			msg:     "||",
			want:    nil,
			wantErr: ErrInvalidMessage,
		},
		{
			name:    "#5: Should return invalid header with header str",
			msg:     "some|",
			want:    nil,
			wantErr: ErrInvalidHeader,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMessage(tt.msg)
			if err != nil && err != tt.wantErr {
				t.Errorf("ParseMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseMessage() got = %v, want %v", got, tt.want)
			}
		})
	}
}
