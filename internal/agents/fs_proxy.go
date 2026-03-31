package agents

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"Lumaestro/internal/config"
)

// FSProxy gerencia as "mãos" do Lumaestro (acesso ao disco e terminal).
type FSProxy struct {
	// Config removido para ser dinâmico
}

// NewFSProxy cria um novo proxy dinâmico.
func NewFSProxy() *FSProxy {
	return &FSProxy{}
}

// getSecurityConfig recupera as permissões mais recentes do disco.
func (p *FSProxy) getSecurityConfig() config.SecurityConfig {
	cfg, err := config.Load()
	if err != nil {
		return config.SecurityConfig{} // Bloqueio total por segurança se falhar
	}
	return cfg.Security
}

// isPathAuthorized verifica se o caminho está dentro dos workspaces permitidos.
func (p *FSProxy) isPathAuthorized(path string) bool {
	sc := p.getSecurityConfig()
	if sc.FullMachineAccess {
		return true
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	for _, ws := range sc.Workspaces {
		absWS, err := filepath.Abs(ws)
		if err != nil {
			continue
		}
		if strings.HasPrefix(absPath, absWS) {
			return true
		}
	}

	return false
}

// ReadFile executa a leitura se autorizado.
func (p *FSProxy) ReadFile(path string) (string, error) {
	sc := p.getSecurityConfig()
	if !sc.AllowRead {
		return "", fmt.Errorf("🛡️ LEITURA BLOQUEADA: Ative 'Permitir Leitura' nas configurações.")
	}
	if !p.isPathAuthorized(path) {
		return "", fmt.Errorf("caminho fora do workspace autorizado: %s", path)
	}

	data, err := os.ReadFile(path)
	return string(data), err
}

// WriteFile executa a escrita se autorizado.
func (p *FSProxy) WriteFile(path string, content string) error {
	sc := p.getSecurityConfig()
	
	// Verifica se o arquivo existe
	_, statErr := os.Stat(path)
	fileExists := !os.IsNotExist(statErr)

	if fileExists {
		if !sc.AllowWrite {
			return fmt.Errorf("🛡️ ESCRITA BLOQUEADA: Ative 'Permitir Escrita' nas configurações.")
		}
	} else {
		if !sc.AllowCreate {
			return fmt.Errorf("🛡️ CRIAÇÃO BLOQUEADA: Ative 'Permitir Criação' nas configurações.")
		}
	}

	if !p.isPathAuthorized(path) {
		return fmt.Errorf("caminho fora do workspace autorizado: %s", path)
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// DeleteFile executa a deleção.
func (p *FSProxy) DeleteFile(path string) error {
	sc := p.getSecurityConfig()
	if !sc.AllowDelete {
		return fmt.Errorf("🛡️ DELEÇÃO BLOQUEADA: Ative permissão nas configurações.")
	}
	if !p.isPathAuthorized(path) {
		return fmt.Errorf("caminho fora do workspace autorizado: %s", path)
	}

	return os.Remove(path)
}

// MoveFile executa o renomeio/movimentação.
func (p *FSProxy) MoveFile(oldPath, newPath string) error {
	sc := p.getSecurityConfig()
	if !sc.AllowMove {
		return fmt.Errorf("🛡️ MOVIMENTAÇÃO BLOQUEADA.")
	}
	if !p.isPathAuthorized(oldPath) || !p.isPathAuthorized(newPath) {
		return fmt.Errorf("origem ou destino fora do workspace autorizado")
	}

	return os.Rename(oldPath, newPath)
}

// RunCommand executa comandos de terminal.
func (p *FSProxy) RunCommand(command string, args []string) (string, error) {
	sc := p.getSecurityConfig()
	if !sc.AllowRunCommands {
		return "", fmt.Errorf("🛡️ EXECUÇÃO BLOQUEADA: Ative 'Executar Comandos' nas configurações.")
	}

	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
