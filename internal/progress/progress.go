package progress

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

// Indicator displays generation progress on stderr for TTY output.
type Indicator struct {
	w     io.Writer
	name  string
	total int
	large bool
	start time.Time
	last  time.Time
}

// New creates a progress indicator. Returns nil if stdout is not a TTY
// or count < 1000 or quiet mode is active.
func New(name string, total int, quiet bool) *Indicator {
	if quiet || total < 1000 {
		return nil
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return nil
	}
	return &Indicator{
		w:     os.Stderr,
		name:  name,
		total: total,
		large: total >= 1000000,
		start: time.Now(),
	}
}

// Update reports progress. Only redraws at most every 100ms.
func (p *Indicator) Update(current int) {
	if p == nil {
		return
	}
	now := time.Now()
	if now.Sub(p.last) < 100*time.Millisecond && current < p.total {
		return
	}
	p.last = now
	p.render(current, now)
}

// Done clears the progress line.
func (p *Indicator) Done() {
	if p == nil {
		return
	}
	fmt.Fprintf(p.w, "\r\033[K")
}

func (p *Indicator) render(current int, now time.Time) {
	pct := current * 100 / p.total
	elapsed := now.Sub(p.start)

	var throughput float64
	if elapsed > 0 {
		throughput = float64(current) / elapsed.Seconds()
	}

	var eta time.Duration
	if throughput > 0 {
		remaining := p.total - current
		eta = time.Duration(float64(remaining)/throughput) * time.Second
	}

	line := fmt.Sprintf("\r[%s] %s / %s | %d%% | %s/sec | ETA %s",
		p.name,
		formatNum(current),
		formatNum(p.total),
		pct,
		formatNum(int(throughput)),
		formatDuration(eta),
	)

	if p.large {
		line += " " + asciiBar(pct, 20)
	}

	fmt.Fprint(p.w, line+"\033[K")
}

func asciiBar(pct, width int) string {
	filled := pct * width / 100
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("=", filled) + strings.Repeat(" ", width-filled) + "]"
}

func formatNum(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	return strings.Join(parts, ",")
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return "<1ms"
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
