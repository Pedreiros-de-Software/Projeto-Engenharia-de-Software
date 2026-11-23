package store

import "context"

type Usuario struct {
	ID        int64
	Nome      string
	Login     string
	SenhaHash string
	Papel     string
}

// Usuarios é o repositório de contas de acesso.
type Usuarios struct{ db executor }

const colunasUsuario = "u.id, u.nome, u.login, u.senha_hash, u.papel"

func lerUsuario(l linha) (Usuario, error) {
	var u Usuario
	err := l.Scan(&u.ID, &u.Nome, &u.Login, &u.SenhaHash, &u.Papel)
	return u, err
}

// Inserir devolve ErrDuplicado quando o login já existe.
func (r Usuarios) Inserir(ctx context.Context, u Usuario) (int64, error) {
	id, err := lerID(r.db.QueryRowContext(ctx,
		"INSERT INTO usuarios (nome, login, senha_hash, papel) VALUES ($1, $2, $3, $4) RETURNING id",
		u.Nome, u.Login, u.SenhaHash, u.Papel))
	return id, traduzirErro(err)
}

func (r Usuarios) BuscarPorLogin(ctx context.Context, login string) (Usuario, error) {
	u, err := lerUsuario(r.db.QueryRowContext(ctx,
		"SELECT "+colunasUsuario+" FROM usuarios u WHERE u.login = $1", login))
	return u, traduzirErro(err)
}

// Listar devolve todos os usuários em ordem de cadastro.
func (r Usuarios) Listar(ctx context.Context) ([]Usuario, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT "+colunasUsuario+" FROM usuarios u ORDER BY u.id")
	if err != nil {
		return nil, err
	}
	return coletar(rows, lerUsuario)
}

func (r Usuarios) Contar(ctx context.Context) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT count(*) FROM usuarios").Scan(&total)
	return total, err
}
