package skills

import (
    "fmt"
    "strings"
    "sync"
)

// Skill representa a estrutura de uma habilidade em Go
type Skill struct {
    Name        string
    Category    string
    Content     string
    Description string
}

var (
    registry = make(map[string]Skill)
    mu       sync.RWMutex
)

// Register adiciona uma skill ao arsenal global
func Register(s Skill) {
    mu.Lock()
    defer mu.Unlock()
    registry[strings.ToLower(s.Name)] = s
}

// GetSkill busca o conteúdo de uma habilidade pelo nome (ex: "brainstorming")
func GetSkill(name string) (string, error) {
    mu.RLock()
    defer mu.RUnlock()
    name = strings.TrimPrefix(strings.ToLower(name), "@")
    s, ok := registry[name]
    if !ok {
        return "", fmt.Errorf("skill '%s' não encontrada no arsenal nativo", name)
    }
    return s.Content, nil
}

// ListSkills retorna a lista de todos os nomes de skills disponíveis
func ListSkills() []string {
    mu.RLock()
    defer mu.RUnlock()
    keys := make([]string, 0, len(registry))
    for k := range registry {
        keys = append(keys, k)
    }
    return keys
}

// GetDetailedSkill retorna a struct completa da skill
func GetDetailedSkill(name string) (Skill, error) {
    mu.RLock()
    defer mu.RUnlock()
    name = strings.TrimPrefix(strings.ToLower(name), "@")
    s, ok := registry[name]
    if !ok {
        return Skill{}, fmt.Errorf("skill '%s' não encontrada", name)
    }
    return s, nil
}
