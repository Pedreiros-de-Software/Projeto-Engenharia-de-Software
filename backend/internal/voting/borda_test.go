package voting

import (
	"errors"
	"reflect"
	"testing"
)

func TestCalcularApuracaoBorda(t *testing.T) {
	opcoesPadrao := []string{"Projeto A", "Projeto B", "Projeto C"}

	testes := []struct {
		nome          string
		opcoesValidas []string
		votos         []VotoBorda
		esperado      []OpcaoPontuada
		erroEsperado  error
	}{
		{
			nome:          "Apuração bem-sucedida com 3 projetos e 3 eleitores",
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
			esperado: []OpcaoPontuada{
				{Opcao: "Projeto A", Pontos: 8, PrimeirasPreferencias: 2},
				{Opcao: "Projeto B", Pontos: 6, PrimeirasPreferencias: 1},
				{Opcao: "Projeto C", Pontos: 4, PrimeirasPreferencias: 0},
			},
		},
		{
			nome:          "Ranking por pontos com contagem de primeiras preferências",
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
			esperado: []OpcaoPontuada{
				{Opcao: "Projeto X", Pontos: 9, PrimeirasPreferencias: 2},
				{Opcao: "Projeto Z", Pontos: 8, PrimeirasPreferencias: 1},
				{Opcao: "Projeto Y", Pontos: 7, PrimeirasPreferencias: 1},
			},
		},
		{
			nome:          "Empate de pontos desfeito por primeiras preferências",
			opcoesValidas: opcoesPadrao,
			votos: []VotoBorda{
				{EleitorID: "E1", Preferencias: []string{"Projeto B", "Projeto C", "Projeto A"}},
				{EleitorID: "E2", Preferencias: []string{"Projeto C", "Projeto A", "Projeto B"}},
				{EleitorID: "E3", Preferencias: []string{"Projeto C", "Projeto A", "Projeto B"}},
			},
			// A: 1+2+2 = 5 (0 primeiras); B: 3+1+1 = 5 (1 primeira); C: 2+3+3 = 8.
			// B fica à frente de A, embora A tenha sido cadastrado antes.
			esperado: []OpcaoPontuada{
				{Opcao: "Projeto C", Pontos: 8, PrimeirasPreferencias: 2},
				{Opcao: "Projeto B", Pontos: 5, PrimeirasPreferencias: 1},
				{Opcao: "Projeto A", Pontos: 5, PrimeirasPreferencias: 0},
			},
		},
		{
			nome:          "Empate persistente mantém a ordem de cadastro e marca Empatado",
			opcoesValidas: []string{"Projeto Z", "Projeto Y"},
			votos: []VotoBorda{
				{EleitorID: "E1", Preferencias: []string{"Projeto Z", "Projeto Y"}},
				{EleitorID: "E2", Preferencias: []string{"Projeto Y", "Projeto Z"}},
			},
			esperado: []OpcaoPontuada{
				{Opcao: "Projeto Z", Pontos: 3, PrimeirasPreferencias: 1, Empatado: true},
				{Opcao: "Projeto Y", Pontos: 3, PrimeirasPreferencias: 1, Empatado: true},
			},
		},
		{
			nome:          "Mesmos pontos com primeiras preferências diferentes não é empate",
			opcoesValidas: opcoesPadrao,
			votos: []VotoBorda{
				{EleitorID: "E1", Preferencias: []string{"Projeto A", "Projeto B", "Projeto C"}},
				{EleitorID: "E2", Preferencias: []string{"Projeto C", "Projeto B", "Projeto A"}},
			},
			// Todos com 4 pontos; A e C com 1 primeira preferência, B com nenhuma.
			esperado: []OpcaoPontuada{
				{Opcao: "Projeto A", Pontos: 4, PrimeirasPreferencias: 1, Empatado: true},
				{Opcao: "Projeto C", Pontos: 4, PrimeirasPreferencias: 1, Empatado: true},
				{Opcao: "Projeto B", Pontos: 4, PrimeirasPreferencias: 0},
			},
		},
		{
			nome:          "Sem votos todas empatam na ordem de cadastro",
			opcoesValidas: []string{"Projeto C", "Projeto A", "Projeto B"},
			esperado: []OpcaoPontuada{
				{Opcao: "Projeto C", Empatado: true},
				{Opcao: "Projeto A", Empatado: true},
				{Opcao: "Projeto B", Empatado: true},
			},
		},
		{
			nome:          "Opção única recebe 1 ponto por cédula",
			opcoesValidas: []string{"Projeto A"},
			votos: []VotoBorda{
				{EleitorID: "E1", Preferencias: []string{"Projeto A"}},
				{EleitorID: "E2", Preferencias: []string{"Projeto A"}},
			},
			esperado: []OpcaoPontuada{
				{Opcao: "Projeto A", Pontos: 2, PrimeirasPreferencias: 2},
			},
		},
		{
			nome:          "Rejeitar voto com lista incompleta de opções",
			opcoesValidas: opcoesPadrao,
			votos: []VotoBorda{
				{
					EleitorID:    "E1",
					Preferencias: []string{"Projeto A", "Projeto B"}, // Faltou C -> Incompleto
				},
			},
			erroEsperado: ErrVotoInvalido,
		},
		{
			nome:          "Rejeitar voto com opção duplicada",
			opcoesValidas: opcoesPadrao,
			votos: []VotoBorda{
				{
					EleitorID:    "E1",
					Preferencias: []string{"Projeto A", "Projeto A", "Projeto C"},
				},
			},
			erroEsperado: ErrVotoInvalido,
		},
		{
			nome:          "Rejeitar voto com opção desconhecida",
			opcoesValidas: opcoesPadrao,
			votos: []VotoBorda{
				{EleitorID: "E1", Preferencias: []string{"Projeto A", "Projeto B", "Projeto D"}},
			},
			erroEsperado: ErrVotoInvalido,
		},
		{
			nome:          "Rejeitar a apuração inteira se uma cédula depois de outras válidas for inválida",
			opcoesValidas: opcoesPadrao,
			votos: []VotoBorda{
				{EleitorID: "E1", Preferencias: []string{"Projeto A", "Projeto B", "Projeto C"}},
				{EleitorID: "E2", Preferencias: []string{"Projeto C"}},
			},
			erroEsperado: ErrVotoInvalido,
		},
		{
			nome:          "Retornar erro se a lista de opções válidas for vazia",
			opcoesValidas: []string{},
			votos: []VotoBorda{
				{EleitorID: "E1", Preferencias: []string{"Projeto A"}},
			},
			erroEsperado: ErrOpcoesVazias,
		},
		{
			nome:          "Retornar erro se a lista de opções válidas tiver itens repetidos",
			opcoesValidas: []string{"Projeto A", "Projeto B", "Projeto A"},
			erroEsperado:  ErrOpcoesDuplicadas,
		},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			obtido, err := CalcularApuracaoBorda(tt.opcoesValidas, tt.votos)
			if !errors.Is(err, tt.erroEsperado) {
				t.Fatalf("erro: esperado %v, obtido %v", tt.erroEsperado, err)
			}
			if !reflect.DeepEqual(obtido, tt.esperado) {
				t.Errorf("resultado: esperado %+v, obtido %+v", tt.esperado, obtido)
			}
		})
	}
}

func TestValidarCedulaBorda(t *testing.T) {
	opcoes := []string{"Projeto A", "Projeto B", "Projeto C"}

	testes := []struct {
		nome         string
		opcoes       []string
		preferencias []string
		erroEsperado error
	}{
		{
			nome:         "Cédula completa e válida",
			opcoes:       opcoes,
			preferencias: []string{"Projeto C", "Projeto A", "Projeto B"},
		},
		{
			nome:         "Opção única ordenada",
			opcoes:       []string{"Projeto A"},
			preferencias: []string{"Projeto A"},
		},
		{
			nome:         "Votação sem opções",
			opcoes:       nil,
			preferencias: []string{"Projeto A"},
			erroEsperado: ErrOpcoesVazias,
		},
		{
			nome:         "Opções da votação repetidas",
			opcoes:       []string{"Projeto A", "Projeto A"},
			preferencias: []string{"Projeto A", "Projeto A"},
			erroEsperado: ErrOpcoesDuplicadas,
		},
		{
			nome:         "Cédula vazia",
			opcoes:       opcoes,
			preferencias: nil,
			erroEsperado: ErrVotoInvalido,
		},
		{
			nome:         "Cédula incompleta",
			opcoes:       opcoes,
			preferencias: []string{"Projeto A", "Projeto B"},
			erroEsperado: ErrVotoInvalido,
		},
		{
			nome:         "Cédula com opção a mais",
			opcoes:       opcoes,
			preferencias: []string{"Projeto A", "Projeto B", "Projeto C", "Projeto D"},
			erroEsperado: ErrVotoInvalido,
		},
		{
			nome:         "Cédula com opção repetida",
			opcoes:       opcoes,
			preferencias: []string{"Projeto A", "Projeto B", "Projeto B"},
			erroEsperado: ErrVotoInvalido,
		},
		{
			nome:         "Cédula com opção desconhecida",
			opcoes:       opcoes,
			preferencias: []string{"Projeto A", "Projeto B", "Projeto D"},
			erroEsperado: ErrVotoInvalido,
		},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			if err := ValidarCedulaBorda(tt.opcoes, tt.preferencias); !errors.Is(err, tt.erroEsperado) {
				t.Errorf("erro: esperado %v, obtido %v", tt.erroEsperado, err)
			}
		})
	}
}
