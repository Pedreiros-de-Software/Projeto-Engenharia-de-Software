package service

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

const (
	prefixoHashSenha = "pbkdf2-sha256"
	tamanhoSal       = 16
	tamanhoChave     = 32
	tamanhoToken     = 32
)

// saltFicticio é usado só para gastar o custo do hash quando o login não existe.
var saltFicticio = make([]byte, tamanhoSal)

// gerarHashSenha devolve "pbkdf2-sha256$<iteracoes>$<sal base64>$<hash base64>" com sal aleatório.
func gerarHashSenha(senha string, iteracoes int) (string, error) {
	sal := make([]byte, tamanhoSal)
	_, _ = rand.Read(sal) // crypto/rand.Read nunca devolve erro (Go 1.24+)

	chave, err := pbkdf2.Key(sha256.New, senha, sal, iteracoes, tamanhoChave)
	if err != nil {
		return "", fmt.Errorf("gerando hash da senha: %w", err)
	}
	return fmt.Sprintf("%s$%d$%s$%s", prefixoHashSenha, iteracoes,
		base64.RawStdEncoding.EncodeToString(sal), base64.RawStdEncoding.EncodeToString(chave)), nil
}

// senhaConfere recalcula o hash com o sal e as iterações gravados e compara em tempo constante.
func senhaConfere(hashGravado, senha string) bool {
	partes := strings.Split(hashGravado, "$")
	if len(partes) != 4 || partes[0] != prefixoHashSenha {
		return false
	}
	iteracoes, err := strconv.Atoi(partes[1])
	if err != nil || iteracoes <= 0 {
		return false
	}
	sal, errSal := base64.RawStdEncoding.DecodeString(partes[2])
	esperada, errChave := base64.RawStdEncoding.DecodeString(partes[3])
	if errSal != nil || errChave != nil {
		return false
	}

	chave, err := pbkdf2.Key(sha256.New, senha, sal, iteracoes, len(esperada))
	return err == nil && subtle.ConstantTimeCompare(chave, esperada) == 1
}

// gastarCustoDeHash faz o mesmo trabalho de uma verificação de senha, para que um login
// inexistente não responda mais rápido que um login existente com senha errada.
func gastarCustoDeHash(senha string, iteracoes int) {
	_, _ = pbkdf2.Key(sha256.New, senha, saltFicticio, iteracoes, tamanhoChave)
}

// novoToken gera um token de sessão aleatório de 32 bytes em base64url.
func novoToken() string {
	b := make([]byte, tamanhoToken)
	_, _ = rand.Read(b) // crypto/rand.Read nunca devolve erro (Go 1.24+)
	return base64.RawURLEncoding.EncodeToString(b)
}

// hashToken é o que fica gravado no banco; o token em si nunca é armazenado.
func hashToken(token string) string {
	soma := sha256.Sum256([]byte(token))
	return hex.EncodeToString(soma[:])
}
