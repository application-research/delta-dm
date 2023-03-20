package util

import "testing"

func TestValidateDatasetName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"valid", true},
		{"valid-dash", true},
		{"spaces are invalid", false},
		{"--", false},
		{"-", false},
		{"", false},
		{"Invalid", false},
		{"@nvalid", false},
		{"-invalid-", false},
		{"123-this-is-invalid", false},
		{"doubledash--is-invalid", false},
		{"fxbaohisiclyxslhlnblyytlmnuvjzchhptbouaccokwzcitoessnddsikdguqocctmahdjftcsuunaaaxrucloyarsseykmkixyveacahiecsfsseeivcwxiyfmfebtdaqxbbdduiutqttebyviankzkhqxksmueqlqzacfllnrwvanautlpdaoucumgzxmburdxfhdhbykhjetqarrqpehiypkehxjefdlfmgoerotnnyqbzmxrzcvvestxrq", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateDatasetName(tt.name); got != tt.want {
				t.Errorf("ValidateDatasetName(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
