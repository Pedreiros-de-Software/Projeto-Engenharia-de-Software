package voting

import (
	"errors"
	"slices"
	"sort"
)

const (
	OpcaoBranco = "branco"
	OpcaoNulo   = "nulo"
)

// VotosOpcao é a quantidade de votos recebida por uma opção.
type VotosOpcao struct {
	Opcao string
	Votos int
}

// ResultadoEleicao é a apuração de um turno de eleição por maioria absoluta dos votos válidos.
type ResultadoEleicao struct {
	Chapas           []VotosOpcao // ordem decrescente de votos; empate mantém a ordem de cadastro
	Brancos, Nulos   int
	Validos          int // soma das chapas (exclui branco e nulo)
	Maioria          ResultadoSegundoTurno
	Vencedora        string   // vazio quando não há
	SegundoTurno     []string // as duas chapas, quando necessário e definido
	EmpateIndefinido bool     // 2º turno necessário, mas empate impede escolher as duas chapas
}

// Boletim é o resultado público de um turno, com comparecimento e percentuais sobre os válidos.
type Boletim struct {
	Aptos, Participantes, Abstencoes int
	PercentualComparecimento         float64            // participantes / aptos * 100
	PercentuaisValidos               map[string]float64 // por chapa, sobre Validos
	Resultado                        ResultadoEleicao
}

var ErrTotaisInconsistentes = errors.New("totais do boletim inconsistentes")

// ApurarEleicao conta os votos de chapas, brancos e nulos e aplica a maioria absoluta dos válidos.
func ApurarEleicao(chapas []string, votos []Voto) (ResultadoEleicao, error) {
	if err := validarChapas(chapas); err != nil {
		return ResultadoEleicao{}, err
	}

	opcoes := slices.Concat(chapas, []string{OpcaoBranco, OpcaoNulo})
	contagem, err := ContarVotos(opcoes, votos, PesoUnitario{})
	if err != nil {
		return ResultadoEleicao{}, err
	}

	resultado := ResultadoEleicao{
		Chapas:  ordenarPorVotos(chapas, contagem),
		Brancos: int(contagem[OpcaoBranco]),
		Nulos:   int(contagem[OpcaoNulo]),
	}
	for _, chapa := range resultado.Chapas {
		resultado.Validos += chapa.Votos
	}

	lider := resultado.Chapas[0]
	resultado.Maioria = VerificarMaioriaAbsoluta(lider.Votos, resultado.Validos-lider.Votos)
	if resultado.Maioria.VencedorDefinido {
		resultado.Vencedora = lider.Opcao
	}
	if resultado.Maioria.PrecisaSegundoTurno {
		resultado.SegundoTurno, resultado.EmpateIndefinido = escolherSegundoTurno(resultado.Chapas)
	}
	return resultado, nil
}

// MontarBoletim calcula abstenções e percentuais de um turno já apurado.
func MontarBoletim(aptos, participantes int, r ResultadoEleicao) (Boletim, error) {
	if !totaisConsistentes(aptos, participantes, r) {
		return Boletim{}, ErrTotaisInconsistentes
	}

	percentuais := make(map[string]float64, len(r.Chapas))
	for _, chapa := range r.Chapas {
		percentuais[chapa.Opcao] = percentual(chapa.Votos, r.Validos)
	}

	return Boletim{
		Aptos:                    aptos,
		Participantes:            participantes,
		Abstencoes:               aptos - participantes,
		PercentualComparecimento: percentual(participantes, aptos),
		PercentuaisValidos:       percentuais,
		Resultado:                r,
	}, nil
}

// validarChapas exige ao menos uma chapa, sem nomes repetidos nem reservados a branco e nulo.
func validarChapas(chapas []string) error {
	if len(chapas) == 0 {
		return ErrOpcoesVazias
	}

	usadas := map[string]bool{OpcaoBranco: true, OpcaoNulo: true}
	for _, chapa := range chapas {
		if usadas[chapa] {
			return ErrOpcoesDuplicadas
		}
		usadas[chapa] = true
	}
	return nil
}

func ordenarPorVotos(chapas []string, contagem map[string]float64) []VotosOpcao {
	resultado := make([]VotosOpcao, len(chapas))
	for posicao, chapa := range chapas {
		resultado[posicao] = VotosOpcao{Opcao: chapa, Votos: int(contagem[chapa])}
	}

	sort.SliceStable(resultado, func(i, j int) bool {
		return resultado[i].Votos > resultado[j].Votos
	})
	return resultado
}

// escolherSegundoTurno devolve as duas chapas mais votadas ou sinaliza empate pelo segundo lugar.
func escolherSegundoTurno(chapas []VotosOpcao) ([]string, bool) {
	if len(chapas) > 2 && chapas[1].Votos == chapas[2].Votos {
		return nil, true
	}
	return []string{chapas[0].Opcao, chapas[1].Opcao}, false
}

func totaisConsistentes(aptos, participantes int, r ResultadoEleicao) bool {
	if aptos < 0 || participantes < 0 || participantes > aptos {
		return false
	}
	if r.Brancos < 0 || r.Nulos < 0 || participantes != r.Brancos+r.Nulos+r.Validos {
		return false
	}

	somaChapas := 0
	for _, chapa := range r.Chapas {
		if chapa.Votos < 0 {
			return false
		}
		somaChapas += chapa.Votos
	}
	return somaChapas == r.Validos
}

// percentual devolve parte / total * 100, ou 0 quando o total é zero.
func percentual(parte, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(parte) / float64(total) * 100
}
