package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store"
)

// Sessao é o resultado de um login: o token (entregue só ao cliente) e o usuário.
type Sessao struct {
	Token   string
	Usuario Usuario
}

// Login confere a senha e abre uma sessão de 12 horas. ip identifica o cliente no rate limit, que
// conta só as credenciais inválidas: por IP e login e, com folga maior, por IP.
func (s *Service) Login(ctx context.Context, login, senha, ip string) (Sessao, error) {
	agora := s.relogio()
	conta := ip + "\x00" + login
	if !s.tentativasPorConta.permitir(conta, agora) {
		return Sessao{}, ErrMuitasTentativas
	}
	if !s.tentativasPorIP.permitir(ip, agora) {
		s.tentativasPorConta.devolver(conta)
		return Sessao{}, ErrMuitasTentativas
	}

	// A tentativa é contada antes da verificação, para pedidos simultâneos não passarem do limite,
	// e devolvida quando o resultado não é de credenciais inválidas.
	u, err := s.autenticar(ctx, login, senha)
	if !errors.Is(err, ErrCredenciaisInvalidas) {
		s.tentativasPorConta.devolver(conta)
		s.tentativasPorIP.devolver(ip)
	}
	if err != nil {
		return Sessao{}, err
	}

	if err := s.repo.Sessoes.ApagarExpiradas(ctx, agora); err != nil {
		return Sessao{}, err
	}
	token := novoToken()
	if err := s.repo.Sessoes.Inserir(ctx, hashToken(token), u.ID, agora.Add(validadeSessao)); err != nil {
		return Sessao{}, err
	}
	return Sessao{Token: token, Usuario: paraUsuario(u)}, nil
}

// autenticar confere login e senha; login inexistente gasta o mesmo custo de hash de uma senha errada.
func (s *Service) autenticar(ctx context.Context, login, senha string) (store.Usuario, error) {
	var u store.Usuario
	err := store.ErrNaoEncontrado // nenhum login cadastrado tem NUL, e o Postgres recusaria a consulta
	if !temNUL(login) {
		u, err = s.repo.Usuarios.BuscarPorLogin(ctx, login)
	}
	if errors.Is(err, store.ErrNaoEncontrado) {
		gastarCustoDeHash(senha, s.iteracoes)
		return store.Usuario{}, ErrCredenciaisInvalidas
	}
	if err != nil {
		return store.Usuario{}, err
	}
	if !senhaConfere(u.SenhaHash, senha) {
		return store.Usuario{}, ErrCredenciaisInvalidas
	}
	return u, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	return s.repo.Sessoes.Apagar(ctx, hashToken(token))
}

// UsuarioDaSessao devolve o dono de um token válido ou ErrNaoAutenticado.
func (s *Service) UsuarioDaSessao(ctx context.Context, token string) (Usuario, error) {
	u, err := s.repo.Sessoes.BuscarUsuario(ctx, hashToken(token), s.relogio())
	if errors.Is(err, store.ErrNaoEncontrado) {
		return Usuario{}, ErrNaoAutenticado
	}
	if err != nil {
		return Usuario{}, err
	}
	return paraUsuario(u), nil
}

// CriarUsuario cadastra uma conta com a senha guardada como hash PBKDF2.
func (s *Service) CriarUsuario(ctx context.Context, nome, login, senha, papel string) (int64, error) {
	nome = strings.TrimSpace(nome)
	switch {
	case !tamanhoEntre(nome, 1, 120):
		return 0, invalido("O nome deve ter entre 1 e 120 caracteres.")
	case !tamanhoEntre(login, 1, 60) || strings.ContainsAny(login, " \t\n"):
		return 0, invalido("O login deve ter entre 1 e 60 caracteres, sem espaços.")
	case !tamanhoEntre(senha, 8, 256):
		return 0, invalido("A senha deve ter entre 8 e 256 caracteres.")
	case papel != PapelAdmin && papel != PapelEleitor:
		return 0, invalido("Papel inválido: use admin ou eleitor.")
	}

	hash, err := gerarHashSenha(senha, s.iteracoes)
	if err != nil {
		return 0, err
	}
	id, err := s.repo.Usuarios.Inserir(ctx, store.Usuario{Nome: nome, Login: login, SenhaHash: hash, Papel: papel})
	if errors.Is(err, store.ErrDuplicado) {
		return 0, invalido("Já existe um usuário com este login.")
	}
	return id, err
}

// ListarUsuarios devolve todas as contas em ordem de cadastro.
func (s *Service) ListarUsuarios(ctx context.Context) ([]Usuario, error) {
	usuarios, err := s.repo.Usuarios.Listar(ctx)
	if err != nil {
		return nil, err
	}
	resultado := make([]Usuario, len(usuarios))
	for i, u := range usuarios {
		resultado[i] = paraUsuario(u)
	}
	return resultado, nil
}

func (s *Service) ExistemUsuarios(ctx context.Context) (bool, error) {
	total, err := s.repo.Usuarios.Contar(ctx)
	return total > 0, err
}

func (s *Service) ExistemVotacoes(ctx context.Context) (bool, error) {
	total, err := s.repo.Votacoes.Contar(ctx)
	return total > 0, err
}
