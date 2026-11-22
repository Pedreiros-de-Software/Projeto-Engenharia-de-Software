package store

import "context"

type VotoPauta struct {
	UnidadeID int64
	Escolha   string
}

// CedulaEleicao é uma cédula anônima de eleição; OpcaoID só existe quando Tipo é "chapa".
type CedulaEleicao struct {
	Tipo    string
	OpcaoID *int64
}

// Votos é o repositório dos votos nominais da pauta, das participações e das cédulas anônimas.
type Votos struct{ db executor }

// InserirVotoPauta devolve ErrDuplicado quando a unidade já votou na pauta.
func (r Votos) InserirVotoPauta(ctx context.Context, votacaoID, unidadeID int64, escolha string) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO votos_pauta (votacao_id, unidade_id, escolha) VALUES ($1, $2, $3)",
		votacaoID, unidadeID, escolha)
	return traduzirErro(err)
}

func (r Votos) ListarVotosPauta(ctx context.Context, votacaoID int64) ([]VotoPauta, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT unidade_id, escolha FROM votos_pauta WHERE votacao_id = $1 ORDER BY unidade_id", votacaoID)
	if err != nil {
		return nil, err
	}
	return coletar(rows, func(l linha) (VotoPauta, error) {
		var v VotoPauta
		err := l.Scan(&v.UnidadeID, &v.Escolha)
		return v, err
	})
}

// IDsVotadasPelaUnidade devolve as pautas em que a unidade já votou.
func (r Votos) IDsVotadasPelaUnidade(ctx context.Context, unidadeID int64) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT votacao_id FROM votos_pauta WHERE unidade_id = $1", unidadeID)
	if err != nil {
		return nil, err
	}
	return coletar(rows, lerID)
}

// InserirParticipacao registra que o usuário votou; devolve ErrDuplicado se ele já tinha votado.
func (r Votos) InserirParticipacao(ctx context.Context, votacaoID, usuarioID int64) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO participacoes (votacao_id, usuario_id) VALUES ($1, $2)", votacaoID, usuarioID)
	return traduzirErro(err)
}

func (r Votos) ContarParticipacoes(ctx context.Context, votacaoID int64) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx,
		"SELECT count(*) FROM participacoes WHERE votacao_id = $1", votacaoID).Scan(&total)
	return total, err
}

// IDsParticipadas devolve as votações (eleição e consulta) em que o usuário já votou.
func (r Votos) IDsParticipadas(ctx context.Context, usuarioID int64) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT votacao_id FROM participacoes WHERE usuario_id = $1", usuarioID)
	if err != nil {
		return nil, err
	}
	return coletar(rows, lerID)
}

// InserirCedula grava uma cédula anônima e devolve seu id (UUID aleatório gerado pelo banco).
func (r Votos) InserirCedula(ctx context.Context, votacaoID int64, tipo string, opcaoID *int64) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO cedulas (votacao_id, tipo, opcao_id) VALUES ($1, $2, $3) RETURNING id",
		votacaoID, tipo, opcaoID).Scan(&id)
	return id, err
}

func (r Votos) InserirPreferencia(ctx context.Context, cedulaID string, posicao int, opcaoID int64) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO cedula_preferencias (cedula_id, posicao, opcao_id) VALUES ($1, $2, $3)",
		cedulaID, posicao, opcaoID)
	return err
}

func (r Votos) ListarCedulasEleicao(ctx context.Context, votacaoID int64) ([]CedulaEleicao, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT tipo, opcao_id FROM cedulas WHERE votacao_id = $1 AND tipo <> 'ranking'", votacaoID)
	if err != nil {
		return nil, err
	}
	return coletar(rows, func(l linha) (CedulaEleicao, error) {
		var c CedulaEleicao
		err := l.Scan(&c.Tipo, &c.OpcaoID)
		return c, err
	})
}

// ListarCedulasRanking devolve, para cada cédula de consulta, os ids das opções na ordem de preferência.
func (r Votos) ListarCedulasRanking(ctx context.Context, votacaoID int64) ([][]int64, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT cp.cedula_id, cp.opcao_id
		FROM cedula_preferencias cp JOIN cedulas c ON c.id = cp.cedula_id
		WHERE c.votacao_id = $1 AND c.tipo = 'ranking'
		ORDER BY cp.cedula_id, cp.posicao`, votacaoID)
	if err != nil {
		return nil, err
	}

	type preferencia struct {
		cedulaID string
		opcaoID  int64
	}
	preferencias, err := coletar(rows, func(l linha) (preferencia, error) {
		var p preferencia
		err := l.Scan(&p.cedulaID, &p.opcaoID)
		return p, err
	})
	if err != nil {
		return nil, err
	}

	cedulas := [][]int64{}
	cedulaAtual := ""
	for _, p := range preferencias {
		if p.cedulaID != cedulaAtual {
			cedulas = append(cedulas, nil)
			cedulaAtual = p.cedulaID
		}
		ultima := len(cedulas) - 1
		cedulas[ultima] = append(cedulas[ultima], p.opcaoID)
	}
	return cedulas, nil
}
