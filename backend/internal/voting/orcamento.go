package voting

import (
	"errors"
	"fmt"
	"math"
)

// ProjetoOrcado é uma linha do ranking Borda com o custo e a situação em relação ao orçamento.
type ProjetoOrcado struct {
	OpcaoPontuada
	Custo, CustoAcumulado int64 // centavos
	DentroDoOrcamento     bool
}

// ResultadoOrcamento é o ranking com a linha de corte orçamentária aplicada (valores em centavos).
type ResultadoOrcamento struct {
	Projetos                           []ProjetoOrcado
	Orcamento, TotalSelecionado, Saldo int64
	EmpateNaLinhaDeCorte               bool
}

var (
	ErrOrcamentoInvalido = errors.New("orcamento deve ser maior ou igual a zero")
	ErrCustoInvalido     = errors.New("custo do projeto ausente ou negativo")
)

// AplicarCorteOrcamentario percorre o ranking acumulando custos com uma linha de corte única:
// o primeiro projeto que não cabe no orçamento e todos abaixo dele ficam fora.
func AplicarCorteOrcamentario(ranking []OpcaoPontuada, custos map[string]int64, orcamento int64) (ResultadoOrcamento, error) {
	if orcamento < 0 {
		return ResultadoOrcamento{}, ErrOrcamentoInvalido
	}

	resultado := ResultadoOrcamento{
		Projetos:  make([]ProjetoOrcado, 0, len(ranking)),
		Orcamento: orcamento,
	}
	var acumulado int64
	dentro := true
	for _, opcao := range ranking {
		custo, err := custoDoProjeto(custos, opcao.Opcao, acumulado)
		if err != nil {
			return ResultadoOrcamento{}, err
		}

		acumulado += custo
		dentro = dentro && acumulado <= orcamento
		if dentro {
			resultado.TotalSelecionado = acumulado
		}
		resultado.Projetos = append(resultado.Projetos, ProjetoOrcado{
			OpcaoPontuada:     opcao,
			Custo:             custo,
			CustoAcumulado:    acumulado,
			DentroDoOrcamento: dentro,
		})
	}

	resultado.Saldo = orcamento - resultado.TotalSelecionado
	resultado.EmpateNaLinhaDeCorte = empateNaLinhaDeCorte(resultado.Projetos)
	return resultado, nil
}

// custoDoProjeto rejeita custo ausente, negativo ou que estouraria a soma acumulada.
func custoDoProjeto(custos map[string]int64, projeto string, acumulado int64) (int64, error) {
	custo, informado := custos[projeto]
	if !informado || custo < 0 || custo > math.MaxInt64-acumulado {
		return 0, fmt.Errorf("projeto %q: %w", projeto, ErrCustoInvalido)
	}
	return custo, nil
}

// empateNaLinhaDeCorte indica se o último projeto dentro e o primeiro fora estão empatados.
func empateNaLinhaDeCorte(projetos []ProjetoOrcado) bool {
	for i := 1; i < len(projetos); i++ {
		ultimoDentro, primeiroFora := projetos[i-1], projetos[i]
		if ultimoDentro.DentroDoOrcamento && !primeiroFora.DentroDoOrcamento {
			return empatadas(ultimoDentro.OpcaoPontuada, primeiroFora.OpcaoPontuada)
		}
	}
	return false
}
