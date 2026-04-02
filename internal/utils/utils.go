package utils

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

var (
	stdout = colorable.NewColorableStdout()
	isTTY  = isatty.IsTerminal(os.Stdout.Fd()) || os.Getenv("FORCE_COLOR") == "1"
)

// Cores ANSI
const (
	Reset   = "\033[0m"
	Bold    = "\033[1m"
	Dim     = "\033[2m"
	Italic  = "\033[3m"
	Cyan    = "\033[36m"
	Yellow  = "\033[33m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Magenta = "\033[35m"
	Blue    = "\033[34m"
	White   = "\033[37m"
)

func color(c string, text string) string {
	if !isTTY {
		return text
	}
	return c + text + Reset
}

func getTimestamp() string {
	return color(Dim, time.Now().Format("15:04:05"))
}

// LogInfo exibe uma mensagem informativa
func LogInfo(message string, icon ...string) {
	emoji := "ℹ️"
	if len(icon) > 0 {
		emoji = icon[0]
	}
	fmt.Fprintf(stdout, "%s %s %s %s\n", getTimestamp(), emoji, color(Cyan, "[INFO]"), message)
}

// LogSuccess exibe uma mensagem de sucesso
func LogSuccess(message string, icon ...string) {
	emoji := "✅"
	if len(icon) > 0 {
		emoji = icon[0]
	}
	fmt.Fprintf(stdout, "%s %s %s %s\n", getTimestamp(), emoji, color(Green, "[OK]"), color(Bold+Green, message))
}

// LogWarning exibe uma mensagem de alerta
func LogWarning(message string, icon ...string) {
	emoji := "⚠️"
	if len(icon) > 0 {
		emoji = icon[0]
	}
	fmt.Fprintf(stdout, "%s %s %s %s\n", getTimestamp(), emoji, color(Yellow, "[AVISO]"), message)
}

// LogError exibe uma mensagem de erro
func LogError(message string, icon ...string) {
	emoji := "❌"
	if len(icon) > 0 {
		emoji = icon[0]
	}
	fmt.Fprintf(stdout, "%s %s %s %s\n", getTimestamp(), emoji, color(Red, "[ERRO]"), color(Bold+Red, message))
}

// LogStep exibe o progresso de um round ou etapa
func LogStep(current, total int, message string) {
	percentage := float64(current) / float64(total) * 100
	barLen := 10
	filled := int(float64(barLen) * (float64(current) / float64(total)))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barLen-filled)

	fmt.Fprintf(stdout, "%s 🔄 %s %s %s %s\n",
		getTimestamp(),
		color(Blue, fmt.Sprintf("Progresso %s %.0f%%", bar, percentage)),
		color(Dim, fmt.Sprintf("(%d/%d)", current, total)),
		"-",
		color(Italic+White, message),
	)
}

// LogSection abre uma nova seção visual com um painel simulado
func LogSection(title string) {
	line := strings.Repeat("━", len(title)+4)
	fmt.Fprintf(stdout, "\n%s\n", color(Magenta, "┏"+line+"┓"))
	fmt.Fprintf(stdout, "%s %s %s\n", color(Magenta, "┃"), color(Bold+Magenta, "  "+strings.ToUpper(title)+"  "), color(Magenta, "┃"))
	fmt.Fprintf(stdout, "%s\n", color(Magenta, "┗"+line+"┛"))
}

// LogKeyValue exibe um par chave-valor formatado
func LogKeyValue(key string, value interface{}) {
	fmt.Fprintf(stdout, "    %s %s: %s\n", color(Dim, "•"), color(Bold+Cyan, key), fmt.Sprint(value))
}

// 📡 NetworkLogger agrega logs de rede para evitar flooding
type NetworkLogger struct {
	Interval     time.Duration
	RequestCount int
	LastLogTime  time.Time
}

func NewNetworkLogger(interval time.Duration) *NetworkLogger {
	return &NetworkLogger{
		Interval:    interval,
		LastLogTime: time.Now(),
	}
}

func (n *NetworkLogger) LogRequest() {
	n.RequestCount++
	now := time.Now()
	duration := now.Sub(n.LastLogTime)
	
	if duration >= n.Interval {
		reqPerSec := float64(n.RequestCount) / duration.Seconds()
		fmt.Fprintf(stdout, "%s 📡 %s %s requisições capturadas em %s (%.1f req/s)\n",
			getTimestamp(),
			color(Cyan, "[NETWORK]"),
			color(Bold, fmt.Sprintf("%d", n.RequestCount)),
			color(Bold, fmt.Sprintf("%.1fs", duration.Seconds())),
			reqPerSec,
		)
		n.RequestCount = 0
		n.LastLogTime = now
	}
}

// 🧠 AIPerformance rastreia a velocidade de resposta do modelo
func LogAIPerformance(model string, duration time.Duration, points int) {
	status := Green
	speedText := "Rápido"
	speedIcon := "⚡"

	if duration.Seconds() > 30 {
		status = Red
		speedText = "EXTREMAMENTE LENTO"
		speedIcon = "🐌"
	} else if duration.Seconds() > 15 {
		status = Yellow
		speedText = "Lento"
		speedIcon = "🐢"
	}

	fmt.Fprintf(stdout, "  %s %s %s (%s): %s | Resultado: %s pontos encontrados\n",
		speedIcon,
		color(status, speedText),
		color(Bold+White, "IA"),
		model,
		color(Bold, fmt.Sprintf("%.2fs", duration.Seconds())),
		color(Cyan, fmt.Sprintf("%d", points)),
	)
	
	// Diagnóstico de lentidão
	if duration.Seconds() > 15 {
		LogWarning("IA lenta detectada. Possíveis causas:", "🐌")
		fmt.Fprintf(stdout, "    • %s\n", color(Dim, "Rate limit da API ou quota excedida"))
		fmt.Fprintf(stdout, "    • %s\n", color(Dim, "Modelo sob alta carga (Busy)"))
		LogInfo("Tentando otimizar próxima chamada...", "🔄")
	}
}

var (
	modelRegex = regexp.MustCompile(`(?i)model:\s*([\w\.\-]+)`)
	retryRegex = regexp.MustCompile(`(?i)retry\s+in\s+([\d\.]+\w+)`)
)

// IsQuotaError verifica se o erro é relacionado a limites de API ou exaustão de cota.
func IsQuotaError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "429") ||
		strings.Contains(msg, "quota") ||
		strings.Contains(msg, "resource_exhausted") ||
		strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "too many requests")
}

// FormatGenAIError transforma os erros verbosos do Google GenAI em logs concisos e amigáveis.
func FormatGenAIError(err error) string {
	if err == nil {
		return ""
	}

	if !IsQuotaError(err) {
		msg := err.Error()
		if len(msg) > 100 {
			return msg[:97] + "..."
		}
		return msg
	}

	msg := err.Error()
	// Extração de metadados via Regex para erro 429
	model := "modelo desconhecido"
	if m := modelRegex.FindStringSubmatch(msg); len(m) > 1 {
		model = m[1]
	}

	retry := ""
	if r := retryRegex.FindStringSubmatch(msg); len(r) > 1 {
		retry = " Tente em " + r[1] + "."
	}

	// Limpeza de URLs e metadados brutos (Details: [map...])
	// Focamos no que importa para o usuário
	return fmt.Sprintf("[Gemini 429] Cota excedida para '%s'.%s", model, retry)
}
