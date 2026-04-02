package agents

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"Lumaestro/internal/config"
)

// Indexer define a interface para o motor de RAG/Grafo de Contexto.
type Indexer interface {
	IndexVault(ctx context.Context) error
}

// Tool representa uma função executável por um agente.
type Tool struct {
	Name        string
	Description string
	Function    func(args map[string]interface{}) (string, error)
}

// ToolRegistry mantém as ferramentas disponíveis.
type ToolRegistry struct {
	Tools   map[string]Tool
	Indexer Indexer
	Ctx     context.Context
}

// ═══════════════════════════════════════════════════════════════
// HELPERS DE SEGURANÇA
// ═══════════════════════════════════════════════════════════════

// checkPermission verifica se a ação é permitida pelas políticas de segurança.
func checkPermission(action string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("falha ao carregar config de segurança: %w", err)
	}

	switch action {
	case "read":
		if !cfg.Security.AllowRead {
			return fmt.Errorf("🔒 BLOQUEADO: Leitura de arquivos não autorizada. Ative em Configurações > Segurança")
		}
	case "write":
		if !cfg.Security.AllowWrite {
			return fmt.Errorf("🔒 BLOQUEADO: Escrita em arquivos não autorizada. Ative em Configurações > Segurança")
		}
	case "create":
		if !cfg.Security.AllowCreate {
			return fmt.Errorf("🔒 BLOQUEADO: Criação de arquivos não autorizada. Ative em Configurações > Segurança")
		}
	case "delete":
		if !cfg.Security.AllowDelete {
			return fmt.Errorf("🔒 BLOQUEADO: Exclusão de arquivos não autorizada. Ative em Configurações > Segurança")
		}
	case "command":
		if !cfg.Security.AllowRunCommands {
			return fmt.Errorf("🔒 BLOQUEADO: Execução de comandos não autorizada. Ative em Configurações > Segurança")
		}
	}
	return nil
}

// getFullNotePath garante que o arquivo tenha extensão .md e esteja dentro do vault.
func getFullNotePath(vaultPath, fileName string) string {
	if !strings.HasSuffix(strings.ToLower(fileName), ".md") {
		fileName += ".md"
	}
	return filepath.Join(vaultPath, fileName)
}

// ═══════════════════════════════════════════════════════════════
// REGISTRO DE FERRAMENTAS
// ═══════════════════════════════════════════════════════════════

// NewToolRegistry inicializa a biblioteca de ferramentas do Lumaestro.
func NewToolRegistry() *ToolRegistry {
	r := &ToolRegistry{Tools: make(map[string]Tool)}

	// ─────────────────────────────────────────────────
	// 📂 LEITURA (Requer: allow_read)
	// ─────────────────────────────────────────────────

	r.Tools["ListVaultFiles"] = Tool{
		Name:        "ListVaultFiles",
		Description: "Lista todos os arquivos do Obsidian Vault.",
		Function: func(args map[string]interface{}) (string, error) {
			vaultPath, _ := args["path"].(string)
			var files []string
			filepath.Walk(vaultPath, func(p string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				files = append(files, info.Name())
				return nil
			})
			return fmt.Sprintf("Arquivos encontrados (%d): %v", len(files), files), nil
		},
	}

	r.Tools["ReadNote"] = Tool{
		Name:        "ReadNote",
		Description: "Lê o conteúdo bruto de uma nota do cofre. Args: file (nome da nota).",
		Function: func(args map[string]interface{}) (string, error) {
			if err := checkPermission("read"); err != nil {
				return "", err
			}
			file, _ := args["file"].(string)
			vaultPath, _ := args["path"].(string)
			if file == "" {
				return "", fmt.Errorf("argumento 'file' é obrigatório")
			}
			
			fullPath := getFullNotePath(vaultPath, file)
			content, err := os.ReadFile(fullPath)
			if err != nil {
				return "", fmt.Errorf("falha ao ler nota '%s': %w", file, err)
			}
			return string(content), nil
		},
	}

	r.Tools["ReadDaily"] = Tool{
		Name:        "ReadDaily",
		Description: "Lê a nota diária de hoje no cofre (formato YYYY-MM-DD.md).",
		Function: func(args map[string]interface{}) (string, error) {
			if err := checkPermission("read"); err != nil {
				return "", err
			}
			vaultPath, _ := args["path"].(string)
			today := time.Now().Format("2006-01-02")
			fullPath := getFullNotePath(vaultPath, today)
			
			content, err := os.ReadFile(fullPath)
			if err != nil {
				return "", fmt.Errorf("nota diária de hoje (%s) não encontrada ou inacessível", today)
			}
			return string(content), nil
		},
	}

	r.Tools["GetBacklinks"] = Tool{
		Name:        "GetBacklinks",
		Description: "Retorna as notas que linkam para a nota especificada. Args: file (nome da nota).",
		Function: func(args map[string]interface{}) (string, error) {
			if err := checkPermission("read"); err != nil {
				return "", err
			}
			file, _ := args["file"].(string)
			if file == "" {
				return "", fmt.Errorf("argumento 'file' é obrigatório")
			}
			return runObsidianCLI("backlinks", "file="+file)
		},
	}

	r.Tools["GetTags"] = Tool{
		Name:        "GetTags",
		Description: "Lista tags encontradas no vault (via busca de texto simples).",
		Function: func(args map[string]interface{}) (string, error) {
			if err := checkPermission("read"); err != nil {
				return "", err
			}
			vaultPath, _ := args["path"].(string)
			tags := make(map[string]int)
			
			filepath.WalkDir(vaultPath, func(path string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
					return nil
				}
				content, _ := os.ReadFile(path)
				words := strings.Fields(string(content))
				for _, word := range words {
					if strings.HasPrefix(word, "#") && len(word) > 1 {
						tags[word]++
					}
				}
				return nil
			})
			return fmt.Sprintf("Tags encontradas: %v", tags), nil
		},
	}

	r.Tools["GetTasks"] = Tool{
		Name:        "GetTasks",
		Description: "Lista tarefas pendentes (- [ ]) encontradas no vault.",
		Function: func(args map[string]interface{}) (string, error) {
			if err := checkPermission("read"); err != nil {
				return "", err
			}
			vaultPath, _ := args["path"].(string)
			var tasks []string
			
			filepath.WalkDir(vaultPath, func(path string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
					return nil
				}
				content, _ := os.ReadFile(path)
				lines := strings.Split(string(content), "\n")
				for _, line := range lines {
					if strings.Contains(line, "- [ ]") {
						tasks = append(tasks, fmt.Sprintf("[%s]: %s", d.Name(), strings.TrimSpace(line)))
					}
				}
				return nil
			})
			return fmt.Sprintf("Tarefas pendentes:\n%s", strings.Join(tasks, "\n")), nil
		},
	}

	// ─────────────────────────────────────────────────
	// 🔍 BUSCA
	// ─────────────────────────────────────────────────

	r.Tools["SearchVault"] = Tool{
		Name:        "SearchVault",
		Description: "Busca texto bruto em todos os arquivos markdown do cofre. Args: query.",
		Function: func(args map[string]interface{}) (string, error) {
			if err := checkPermission("read"); err != nil {
				return "", err
			}
			query, _ := args["query"].(string)
			vaultPath, _ := args["path"].(string)
			if query == "" {
				return "", fmt.Errorf("argumento 'query' é obrigatório")
			}
			
			var matches []string
			filepath.WalkDir(vaultPath, func(path string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
					return nil
				}
				content, _ := os.ReadFile(path)
				if strings.Contains(strings.ToLower(string(content)), strings.ToLower(query)) {
					matches = append(matches, d.Name())
				}
				if len(matches) >= 10 { // Limite simples para não sobrecarregar
					return filepath.SkipDir
				}
				return nil
			})
			return fmt.Sprintf("Resultados para '%s' (%d): %v", query, len(matches), matches), nil
		},
	}

	// Alias de compatibilidade
	r.Tools["ObsidianSearch"] = r.Tools["SearchVault"]

	// ─────────────────────────────────────────────────
	// ✍️ ESCRITA (Requer: allow_write ou allow_create)
	// ─────────────────────────────────────────────────

	r.Tools["CreateNote"] = Tool{
		Name:        "CreateNote",
		Description: "Cria uma nova nota .md no cofre. Args: name, content.",
		Function: func(args map[string]interface{}) (string, error) {
			if err := checkPermission("create"); err != nil {
				return "", err
			}
			name, _ := args["name"].(string)
			content, _ := args["content"].(string)
			vaultPath, _ := args["path"].(string)
			if name == "" {
				return "", fmt.Errorf("argumento 'name' é obrigatório")
			}

			fullPath := getFullNotePath(vaultPath, name)
			
			// Criar subpastas se necessário
			os.MkdirAll(filepath.Dir(fullPath), 0755)

			err := os.WriteFile(fullPath, []byte(content), 0644)
			if err != nil {
				return "", fmt.Errorf("erro ao criar arquivo: %w", err)
			}
			return fmt.Sprintf("✅ Nota '%s' criada no cofre nativo.", name), nil
		},
	}

	r.Tools["AppendToNote"] = Tool{
		Name:        "AppendToNote",
		Description: "Adiciona texto ao final de uma nota. Args: file, content.",
		Function: func(args map[string]interface{}) (string, error) {
			if err := checkPermission("write"); err != nil {
				return "", err
			}
			file, _ := args["file"].(string)
			content, _ := args["content"].(string)
			vaultPath, _ := args["path"].(string)
			if file == "" || content == "" {
				return "", fmt.Errorf("argumentos 'file' e 'content' são obrigatórios")
			}
			
			fullPath := getFullNotePath(vaultPath, file)
			f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return "", fmt.Errorf("erro ao abrir nota para append: %w", err)
			}
			defer f.Close()

			if _, err := f.WriteString("\n" + content); err != nil {
				return "", fmt.Errorf("erro ao escrever na nota: %w", err)
			}
			
			return fmt.Sprintf("✅ Conteúdo adicionado à nota '%s'.", file), nil
		},
	}

	r.Tools["AppendDaily"] = Tool{
		Name:        "AppendDaily",
		Description: "Adiciona conteúdo à nota diária de hoje no cofre.",
		Function: func(args map[string]interface{}) (string, error) {
			if err := checkPermission("write"); err != nil {
				return "", err
			}
			content, _ := args["content"].(string)
			vaultPath, _ := args["path"].(string)
			if content == "" {
				return "", fmt.Errorf("argumento 'content' é obrigatório")
			}
			
			today := time.Now().Format("2006-01-02")
			fullPath := getFullNotePath(vaultPath, today)
			
			f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return "", fmt.Errorf("erro ao abrir nota diária: %w", err)
			}
			defer f.Close()

			if _, err := f.WriteString("\n" + content); err != nil {
				return "", fmt.Errorf("erro ao escrever na nota diária: %w", err)
			}
			
			return fmt.Sprintf("✅ Conteúdo adicionado à nota diária %s.", today), nil
		},
	}

	r.Tools["SetProperty"] = Tool{
		Name:        "SetProperty",
		Description: "Adiciona ou atualiza metadados no frontmatter (YAML) da nota. Args: file, name, value.",
		Function: func(args map[string]interface{}) (string, error) {
			if err := checkPermission("write"); err != nil {
				return "", err
			}
			// Placeholder: Em modo nativo, precisaríamos de um parser YAML para não destruir o arquivo.
			// Por enquanto, vamos retornar que está em desenvolvimento para o modo nativo ou fazer um append simples.
			return "⚠️ SetProperty em modo nativo requer parser YAML (em desenvolvimento). Use AppendToNote para adicionar informações.", nil
		},
	}

	// ─────────────────────────────────────────────────
	// 🌐 WEB SCRAPING (Defuddle)
	// ─────────────────────────────────────────────────

	r.Tools["WebScrape"] = Tool{
		Name:        "WebScrape",
		Description: "Extrai conteúdo limpo (markdown) de uma URL usando Defuddle. Args: url.",
		Function: func(args map[string]interface{}) (string, error) {
			if err := checkPermission("command"); err != nil {
				return "", err
			}
			url, _ := args["url"].(string)
			if url == "" {
				return "", fmt.Errorf("argumento 'url' é obrigatório")
			}
			cmd := exec.Command("defuddle", "parse", url, "--md")
			output, err := cmd.CombinedOutput()
			if err != nil {
				return "", fmt.Errorf("defuddle erro: %w — saída: %s", err, string(output))
			}
			return strings.TrimSpace(string(output)), nil
		},
	}

	// ─────────────────────────────────────────────────
	// 🧠 INDEXAÇÃO / RAG (Motor Interno)
	// ─────────────────────────────────────────────────

	indexTool := Tool{
		Name:        "IndexVault",
		Description: "Analisa o Vault inteiro, extrai fatos complexos (triplas) e constrói o Context Graph.",
		Function: func(args map[string]interface{}) (string, error) {
			if r.Indexer == nil {
				return "", fmt.Errorf("motor de RAG não inicializado")
			}
			vaultPath, _ := args["path"].(string)
			if vaultPath == "" {
				return "", fmt.Errorf("caminho do vault não fornecido")
			}

			// Dispara em background para não travar o RPC
			go r.Indexer.IndexVault(r.Ctx)

			return "🚀 Indexação de Grafo de Contexto INICIADA em background. Verifique os logs de progresso.", nil
		},
	}
	r.Tools["IndexVault"] = indexTool
	r.Tools["ScanVault"] = indexTool // Alias para compatibilidade com o Maestro

	return r
}
