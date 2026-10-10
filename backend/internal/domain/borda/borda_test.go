package borda

import (
	"errors"
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
				{EleitorID: "E1", Preferencias: []string{"Projeto A", "Projeto B", "Projeto C"}}, // A=3, B=2, C=1
				{EleitorID: "E2", Preferencias: []string{"Projeto B", "Projeto A", "Projeto C"}}, // B=3, A=2, C=1
				{EleitorID: "E3", Preferencias: []string{"Projeto A", "Projeto C", "Projeto B"}}, // A=3, C=2, B=1
			},
			// Totais: Projeto A = 3+2+3 = 8 | Projeto B = 2+3+1 = 6 | Projeto C = 1+1+2 = 4
			want: []OpcaoPontuada{
				{Opcao: "Projeto A", Pontos: 8},
				{Opcao: "Projeto B", Pontos: 6},
				{Opcao: "Projeto C", Pontos: 4},
			},
			wantErr: false,
		},
		{
			name:          "Empate de pontuação - Desempate por ordem alfabética",
			opcoesValidas: []string{"Projeto X", "Projeto Y"},
			votos: []VotoBorda{
				{EleitorID: "E1", Preferencias: []string{"Projeto X", "Projeto Y"}}, // X=2, Y=1
				{EleitorID: "E2", Preferencias: []string{"Projeto Y", "Projeto X"}}, // Y=2, X=1
			},
			// Totais: X=3, Y=3. Desempate alfabético coloca Projeto X em 1º
			want: []OpcaoPontuada{
				{Opcao: "Projeto X", Pontos: 3},
				{Opcao: "Projeto Y", Pontos: 3},
			},
			wantErr: false,
		},
		{
			name:          "Voto contendo opção duplicada deve ser ignorado na apuração",
			opcoesValidas: opcoesPadrao,
			votos: []VotoBorda{
				{EleitorID: "E1", Preferencias: []string{"Projeto A", "Projeto A", "Projeto C"}}, // Duplicado -> Inválido
				{EleitorID: "E2", Preferencias: []string{"Projeto B", "Projeto A", "Projeto C"}}, // B=3, A=2, C=1
			},
			// Apenas E2 é considerado
			want: []OpcaoPontuada{
				{Opcao: "Projeto B", Pontos: 3},
				{Opcao: "Projeto A", Pontos: 2},
				{Opcao: "Projeto C", Pontos: 1},
			},
			wantErr: false,
		},
		{
			name:          "Voto contendo projeto inexistente deve ser ignorado",
			opcoesValidas: opcoesPadrao,
			votos: []VotoBorda{
				{EleitorID: "E1", Preferencias: []string{"Projeto Desconhecido", "Projeto A", "Projeto B"}}, // Inválido
				{EleitorID: "E2", Preferencias: []string{"Projeto A", "Projeto B", "Projeto C"}},            // A=3, B=2, C=1
			},
			want: []OpcaoPontuada{
				{Opcao: "Projeto A", Pontos: 3},
				{Opcao: "Projeto B", Pontos: 2},
				{Opcao: "Projeto C", Pontos: 1},
			},
			wantErr: false,
		},
		{
			name:          "Retorna erro se a lista de opções válidas for vazia",
			opcoesValidas: []string{},
			votos: []VotoBorda{
				{EleitorID: "E1", Preferencias: []string{"Projeto A"}},
			},
			want:        nil,
			wantErr:     true,
			errEsperado: ErrOpcoesVazias,
		},
		{
			name:          "Retorna erro se todos os votos submetidos forem inválidos",
			opcoesValidas: opcoesPadrao,
			votos: []VotoBorda{
				{EleitorID: "E1", Preferencias: []string{"Projeto Inexistente"}},
			},
			want:        nil,
			wantErr:     true,
			errEsperado: ErrSemVotosValidos,
		},
	}

	for _, tt := range tests {
		// A linha abaixo foi corrigida. O LLM original havia gerado "[link suspeito removido]"
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalcularApuracaoBorda(tt.opcoesValidas, tt.votos)

			// Verificação de erro
			if (err != nil) != tt.wantErr {
				t.Fatalf("CalcularApuracaoBorda() erro = %v, wantErr = %v", err, tt.wantErr)
			}

			if tt.errEsperado != nil && !errors.Is(err, tt.errEsperado) {
				t.Fatalf("CalcularApuracaoBorda() erro obtido = %v, esperado = %v", err, tt.errEsperado)
			}

			// Verificação do resultado
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalcularApuracaoBorda() = %v, want %v", got, tt.want)
			}
		})
	}
}
