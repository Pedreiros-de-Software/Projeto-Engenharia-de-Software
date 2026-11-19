package store

import "context"

type Unidade struct {
	ID              int64
	Identificador   string
	FracaoIdeal     float64
	Adimplente      bool
	ResponsavelID   *int64 // nil quando a unidade não tem responsável
	ResponsavelNome string
}

// Unidades é o repositório das unidades do condomínio.
type Unidades struct{ db executor }

const selectUnidade = `SELECT un.id, un.identificador, un.fracao_ideal, un.adimplente,
	un.responsavel_id, COALESCE(us.nome, '')
	FROM unidades un LEFT JOIN usuarios us ON us.id = un.responsavel_id`

func lerUnidade(l linha) (Unidade, error) {
	var u Unidade
	err := l.Scan(&u.ID, &u.Identificador, &u.FracaoIdeal, &u.Adimplente, &u.ResponsavelID, &u.ResponsavelNome)
	return u, err
}

func (r Unidades) Inserir(ctx context.Context, u Unidade) (int64, error) {
	id, err := lerID(r.db.QueryRowContext(ctx,
		`INSERT INTO unidades (identificador, fracao_ideal, adimplente, responsavel_id)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		u.Identificador, u.FracaoIdeal, u.Adimplente, u.ResponsavelID))
	return id, traduzirErro(err)
}

// Listar devolve todas as unidades em ordem de cadastro.
func (r Unidades) Listar(ctx context.Context) ([]Unidade, error) {
	rows, err := r.db.QueryContext(ctx, selectUnidade+" ORDER BY un.id")
	if err != nil {
		return nil, err
	}
	return coletar(rows, lerUnidade)
}

func (r Unidades) Buscar(ctx context.Context, id int64) (Unidade, error) {
	u, err := lerUnidade(r.db.QueryRowContext(ctx, selectUnidade+" WHERE un.id = $1", id))
	return u, traduzirErro(err)
}

// BuscarPorResponsavel devolve a unidade sob responsabilidade do usuário.
func (r Unidades) BuscarPorResponsavel(ctx context.Context, usuarioID int64) (Unidade, error) {
	u, err := lerUnidade(r.db.QueryRowContext(ctx, selectUnidade+" WHERE un.responsavel_id = $1", usuarioID))
	return u, traduzirErro(err)
}

// PesoTotal soma as frações ideais das unidades com responsável (as que votam na pauta).
func (r Unidades) PesoTotal(ctx context.Context) (float64, error) {
	var total float64
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(fracao_ideal), 0) FROM unidades WHERE responsavel_id IS NOT NULL").Scan(&total)
	return total, err
}

func (r Unidades) AlterarAdimplencia(ctx context.Context, id int64, adimplente bool) error {
	resultado, err := r.db.ExecContext(ctx, "UPDATE unidades SET adimplente = $2 WHERE id = $1", id, adimplente)
	if err != nil {
		return err
	}
	return exigirUmaLinha(resultado)
}
