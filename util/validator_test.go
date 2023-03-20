package util

import "testing"

func TestValidateDatasetName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"valid", true},
		{"valid-dash", true},
		{"123-this-is-valid", true},
		{"1-2-3-4", true},
		{"spaces are invalid", false},
		{"--", false},
		{"-", false},
		{"", false},
		{"Invalid", false},
		{"@nvalid", false},
		{"-invalid-", false},
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
