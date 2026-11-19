package store

import (
	"context"
	"time"
)

type Votacao struct {
	ID                int64
	Tipo              string
	Titulo            string
	Descricao         string
	Inicio            time.Time
	Fim               time.Time
	Turno             int
	VotacaoAnteriorID *int64
	OrcamentoCentavos *int64
	SegundoTurnoID    *int64 // votação de 2º turno criada a partir desta, se houver
}

type Opcao struct {
	ID            int64
	VotacaoID     int64
	Nome          string
	Descricao     string
	CustoCentavos *int64
	Posicao       int
}

// Votacoes é o repositório das votações, com suas opções e eleitores aptos.
type Votacoes struct{ db executor }

const selectVotacao = `SELECT v.id, v.tipo, v.titulo, v.descricao, v.inicio, v.fim, v.turno,
	v.votacao_anterior_id, v.orcamento_centavos,
	(SELECT s.id FROM votacoes s WHERE s.votacao_anterior_id = v.id)
	FROM votacoes v`

func lerVotacao(l linha) (Votacao, error) {
	var v Votacao
	err := l.Scan(&v.ID, &v.Tipo, &v.Titulo, &v.Descricao, &v.Inicio, &v.Fim, &v.Turno,
		&v.VotacaoAnteriorID, &v.OrcamentoCentavos, &v.SegundoTurnoID)
	return v, err
}

// Inserir devolve ErrDuplicado quando já existe um 2º turno para a mesma votação anterior.
func (r Votacoes) Inserir(ctx context.Context, v Votacao) (int64, error) {
	id, err := lerID(r.db.QueryRowContext(ctx,
		`INSERT INTO votacoes (tipo, titulo, descricao, inicio, fim, turno, votacao_anterior_id, orcamento_centavos)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		v.Tipo, v.Titulo, v.Descricao, v.Inicio, v.Fim, v.Turno, v.VotacaoAnteriorID, v.OrcamentoCentavos))
	return id, traduzirErro(err)
}

func (r Votacoes) Buscar(ctx context.Context, id int64) (Votacao, error) {
	v, err := lerVotacao(r.db.QueryRowContext(ctx, selectVotacao+" WHERE v.id = $1", id))
	return v, traduzirErro(err)
}

// Listar devolve todas as votações, das mais recentes para as mais antigas.
func (r Votacoes) Listar(ctx context.Context) ([]Votacao, error) {
	rows, err := r.db.QueryContext(ctx, selectVotacao+" ORDER BY v.id DESC")
	if err != nil {
		return nil, err
	}
	return coletar(rows, lerVotacao)
}

// BuscarComTrava lê a votação com FOR SHARE: até o fim da transação, EncerrarSeAberta espera
// por ela, enquanto outras transações de voto seguem em paralelo.
func (r Votacoes) BuscarComTrava(ctx context.Context, id int64) (Votacao, error) {
	v, err := lerVotacao(r.db.QueryRowContext(ctx, selectVotacao+" WHERE v.id = $1 FOR SHARE OF v", id))
	return v, traduzirErro(err)
}

// EncerrarSeAberta move o fim para agora só se inicio < agora < fim, e informa se moveu.
// O UPDATE espera as transações que usaram BuscarComTrava confirmarem.
func (r Votacoes) EncerrarSeAberta(ctx context.Context, id int64, agora time.Time) (bool, error) {
	resultado, err := r.db.ExecContext(ctx,
		"UPDATE votacoes SET fim = $2 WHERE id = $1 AND inicio < $2 AND fim > $2", id, agora)
	if err != nil {
		return false, err
	}
	alteradas, err := resultado.RowsAffected()
	return alteradas > 0, err
}

// Contar devolve o total de votações cadastradas.
func (r Votacoes) Contar(ctx context.Context) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT count(*) FROM votacoes").Scan(&total)
	return total, err
}

func (r Votacoes) InserirOpcao(ctx context.Context, o Opcao) (int64, error) {
	id, err := lerID(r.db.QueryRowContext(ctx,
		`INSERT INTO opcoes (votacao_id, nome, descricao, custo_centavos, posicao)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		o.VotacaoID, o.Nome, o.Descricao, o.CustoCentavos, o.Posicao))
	return id, traduzirErro(err)
}

// ListarOpcoes devolve as opções na ordem de cadastro (posição).
func (r Votacoes) ListarOpcoes(ctx context.Context, votacaoID int64) ([]Opcao, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, votacao_id, nome, descricao, custo_centavos, posicao
		FROM opcoes WHERE votacao_id = $1 ORDER BY posicao`, votacaoID)
	if err != nil {
		return nil, err
	}
	return coletar(rows, func(l linha) (Opcao, error) {
		var o Opcao
		err := l.Scan(&o.ID, &o.VotacaoID, &o.Nome, &o.Descricao, &o.CustoCentavos, &o.Posicao)
		return o, err
	})
}

func (r Votacoes) InserirApto(ctx context.Context, votacaoID, usuarioID int64) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO aptos (votacao_id, usuario_id) VALUES ($1, $2)", votacaoID, usuarioID)
	return traduzirErro(err)
}

// CopiarAptos habilita na votação destino os mesmos eleitores da votação origem.
func (r Votacoes) CopiarAptos(ctx context.Context, origemID, destinoID int64) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO aptos (votacao_id, usuario_id) SELECT $2, usuario_id FROM aptos WHERE votacao_id = $1",
		origemID, destinoID)
	return err
}

func (r Votacoes) EstaApto(ctx context.Context, votacaoID, usuarioID int64) (bool, error) {
	var apto bool
	err := r.db.QueryRowContext(ctx,
		"SELECT EXISTS (SELECT 1 FROM aptos WHERE votacao_id = $1 AND usuario_id = $2)",
		votacaoID, usuarioID).Scan(&apto)
	return apto, err
}

func (r Votacoes) ContarAptos(ctx context.Context, votacaoID int64) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT count(*) FROM aptos WHERE votacao_id = $1", votacaoID).Scan(&total)
	return total, err
}

// IDsAptas devolve as votações em que o usuário está habilitado a votar.
func (r Votacoes) IDsAptas(ctx context.Context, usuarioID int64) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT votacao_id FROM aptos WHERE usuario_id = $1", usuarioID)
	if err != nil {
		return nil, err
	}
	return coletar(rows, lerID)
}
