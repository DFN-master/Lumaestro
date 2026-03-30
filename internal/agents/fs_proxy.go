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
	Config config.SecurityConfig
}

// NewFSProxy cria um novo proxy com as configurações carregadas.
func NewFSProxy() *FSProxy {
	cfg, _ := config.Load()
	return &FSProxy{
		Config: cfg.Security,
	}
}

// isPathAuthorized verifica se o caminho está dentro dos workspaces permitidos.
func (p *FSProxy) isPathAuthorized(path string) bool {
	if p.Config.FullMachineAccess {
		return true
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	for _, ws := range p.Config.Workspaces {
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
	if !p.Config.AllowRead {
		return "", fmt.Errorf("permissão de leitura desativada")
	}
	if !p.isPathAuthorized(path) {
		return "", fmt.Errorf("caminho fora do workspace autorizado: %s", path)
	}

	data, err := os.ReadFile(path)
	return string(data), err
}

// WriteFile executa a escrita se autorizado.
func (p *FSProxy) WriteFile(path string, content string) error {
	if !p.Config.AllowWrite {
		return fmt.Errorf("permissão de escrita desativada")
	}
	if !p.isPathAuthorized(path) {
		return fmt.Errorf("caminho fora do workspace autorizado: %s", path)
	}

	// Se o arquivo não existir, verifica permissão de criação
	if _, err := os.Stat(path); os.IsNotExist(err) && !p.Config.AllowCreate {
		return fmt.Errorf("permissão para criar arquivos desativada")
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// DeleteFile executa a deleção. IMPORTANTE: Sempre requer confirmação via UI (será tratado no orquestrador).
func (p *FSProxy) DeleteFile(path string) error {
	if !p.Config.AllowDelete {
		return fmt.Errorf("permissão de deleção desativada")
	}
	if !p.isPathAuthorized(path) {
		return fmt.Errorf("caminho fora do workspace autorizado: %s", path)
	}

	return os.Remove(path)
}

// MoveFile executa o renomeio/movimentação.
func (p *FSProxy) MoveFile(oldPath, newPath string) error {
	if !p.Config.AllowMove {
		return fmt.Errorf("permissão de movimentação desativada")
	}
	if !p.isPathAuthorized(oldPath) || !p.isPathAuthorized(newPath) {
		return fmt.Errorf("origem ou destino fora do workspace autorizado")
	}

	return os.Rename(oldPath, newPath)
}

// RunCommand executa comandos de terminal.
func (p *FSProxy) RunCommand(command string, args []string) (string, error) {
	if !p.Config.AllowRunCommands {
		return "", fmt.Errorf("permissão para rodar comandos desativada")
	}

	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
