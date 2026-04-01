package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	name := "Teste_Isolation"
	cwd, _ := os.Getwd()
	accountPath := filepath.Join(cwd, ".gemini_accounts", name)

	// Simula AddGeminiAccount
	os.MkdirAll(accountPath, 0755)
	fmt.Printf("[TEST] Pasta criada: %s\n", accountPath)

	// Simula LoginGeminiAccount
	binaryPath := "gemini"
	// Procura no PATH ou node_modules
	if _, err := exec.LookPath("gemini"); err != nil {
		binaryPath = filepath.Join(cwd, "node_modules", ".bin", "gemini.cmd")
	}

	script := fmt.Sprintf(`$env:GEMINI_CLI_HOME='%s'; $env:NO_BROWSER='true'; & '%s' login`, accountPath, binaryPath)
	
	fmt.Printf("[TEST] Abrindo terminal com isolamento...\n")
	cmd := exec.Command("cmd", "/c", "start", "powershell", "-NoExit", "-Command", script)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("[ERROR] Falha ao abrir terminal: %v\n", err)
	} else {
		fmt.Println("[SUCCESS] Terminal de login aberto com sucesso!")
	}
}
