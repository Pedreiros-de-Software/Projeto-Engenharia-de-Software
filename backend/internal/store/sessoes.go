package store

import (
	"context"
	"time"
)

// Sessoes é o repositório de sessões de login, identificadas pelo hash do token.
type Sessoes struct{ db executor }

func (r Sessoes) Inserir(ctx context.Context, tokenHash string, usuarioID int64, expiraEm time.Time) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO sessoes (token_hash, usuario_id, expira_em) VALUES ($1, $2, $3)",
		tokenHash, usuarioID, expiraEm)
	return err
}

// BuscarUsuario devolve o dono de uma sessão ainda válida no instante agora.
func (r Sessoes) BuscarUsuario(ctx context.Context, tokenHash string, agora time.Time) (Usuario, error) {
	u, err := lerUsuario(r.db.QueryRowContext(ctx,
		"SELECT "+colunasUsuario+` FROM sessoes s JOIN usuarios u ON u.id = s.usuario_id
		WHERE s.token_hash = $1 AND s.expira_em > $2`, tokenHash, agora))
	return u, traduzirErro(err)
}

func (r Sessoes) Apagar(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sessoes WHERE token_hash = $1", tokenHash)
	return err
}

func (r Sessoes) ApagarExpiradas(ctx context.Context, agora time.Time) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sessoes WHERE expira_em <= $1", agora)
	return err
}
