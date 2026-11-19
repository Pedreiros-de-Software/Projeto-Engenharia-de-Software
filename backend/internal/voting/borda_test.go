package voting

import (
	"reflect"
	"testing"
)

func TestCalcularApuracaoBorda(t *testing.T) {
	opcoesPadrao := []string{"Projeto A", "Projeto B", "Projeto C"}

	tests := []struct {
		name          string
		opcoesValidas []string
		votos         []VotoBorda
		want          []OpcaoPontuada
		wantErr       bool
		errEsperado   error
	}{
		{
			name:          "Apuração bem-sucedida com 3 projetos e 3 eleitores",
			opcoesValidas: opcoesPadrao,
			votos: []VotoBorda{
				{
					EleitorID:    "E1",
					Preferencias: []string{"Projeto A", "Projeto B", "Projeto C"}, // A=3, B=2, C=1 (1º A)
				},
				{
					EleitorID:    "E2",
					Preferencias: []string{"Projeto B", "Projeto A", "Projeto C"}, // B=3, A=2, C=1 (1º B)
				},
				{
					EleitorID:    "E3",
					Preferencias: []string{"Projeto A", "Projeto C", "Projeto B"}, // A=3, C=2, B=1 (1º A)
				},
			},
			want: []OpcaoPontuada{
				{Opcao: "Projeto A", Pontos: 8, PrimeirasPreferencias: 2},
				{Opcao: "Projeto B", Pontos: 6, PrimeirasPreferencias: 1},
				{Opcao: "Projeto C", Pontos: 4, PrimeirasPreferencias: 0},
			},
			wantErr: false,
		},
		{
			name:          "Desempate por Primeiras Preferências (Issue #24)",
			opcoesValidas: []string{"Projeto X", "Projeto Y", "Projeto Z"},
			votos: []VotoBorda{
				// Eleitor 1 vota X > Y > Z (X=3, Y=2, Z=1) -> 1º lugar pra X
				{EleitorID: "E1", Preferencias: []string{"Projeto X", "Projeto Y", "Projeto Z"}},
				// Eleitor 2 vota Y > Z > X (Y=3, Z=2, X=1) -> 1º lugar pra Y
				{EleitorID: "E2", Preferencias: []string{"Projeto Y", "Projeto Z", "Projeto X"}},
				// Eleitor 3 vota Z > X > Y (Z=3, X=2, Y=1) -> 1º lugar pra Z
				{EleitorID: "E3", Preferencias: []string{"Projeto Z", "Projeto X", "Projeto Y"}},
				// Eleitor 4 vota X > Z > Y (X=3, Z=2, Y=1) -> 1º lugar pra X
				{EleitorID: "E4", Preferencias: []string{"Projeto X", "Projeto Z", "Projeto Y"}},
			},
			// Totais:
			// X: 3+1+2+3 = 9 pontos (2 primeiras preferências)
			// Z: 1+2+3+2 = 8 pontos
			// Y: 2+3+1+1 = 7 pontos
			want: []OpcaoPontuada{
				{Opcao: "Projeto X", Pontos: 9, PrimeirasPreferencias: 2},
				{Opcao: "Projeto Z", Pontos: 8, PrimeirasPreferencias: 1},
				{Opcao: "Projeto Y", Pontos: 7, PrimeirasPreferencias: 1},
			},
			wantErr: false,
		},
		{
			name:          "Rejeitar voto com lista incompleta de opções (Issue #23)",
			opcoesValidas: opcoesPadrao,
			votos: []VotoBorda{
				{
					EleitorID:    "E1",
					Preferencias: []string{"Projeto A", "Projeto B"}, // Faltou C -> Incompleto
				},
			},
			want:        nil,
			wantErr:     true,
			errEsperado: ErrVotoInvalido,
		},
		{
			name:          "Rejeitar voto com opção duplicada",
			opcoesValidas: opcoesPadrao,
			votos: []VotoBorda{
				{
					EleitorID:    "E1",
					Preferencias: []string{"Projeto A", "Projeto A", "Projeto C"},
				},
			},
			want:        nil,
			wantErr:     true,
			errEsperado: ErrVotoInvalido,
		},
		{
			name:          "Retornar erro se a lista de opções válidas for vazia",
			opcoesValidas: []string{},
			votos: []VotoBorda{
				{EleitorID: "E1", Preferencias: []string{"Projeto A"}},
			},
			want:        nil,
			wantErr:     true,
			errEsperado: ErrOpcoesVazias,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalcularApuracaoBorda(tt.opcoesValidas, tt.votos)

			if (err != nil) != tt.wantErr {
				t.Errorf("CalcularApuracaoBorda() erro = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if tt.errEsperado != nil && err != tt.errEsperado {
				t.Errorf("CalcularApuracaoBorda() erro obtido = %v, esperado = %v", err, tt.errEsperado)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalcularApuracaoBorda() = %v, want %v", got, tt.want)
			}
		})
	}
}
