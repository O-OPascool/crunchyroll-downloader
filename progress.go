package main

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

type ProgressBar struct {
	total      int64
	current    atomic.Int64
	bytesTotal atomic.Int64
	label      string
	start      time.Time
}

func NewProgressBar(total int64, label string) *ProgressBar {
	return &ProgressBar{
		total: total,
		label: label,
		start: time.Now(),
	}
}

func (p *ProgressBar) Add(bytes int64) {
	p.current.Add(1)
	p.bytesTotal.Add(bytes)
	p.render()
}

func (p *ProgressBar) SetDone(count int64) {
	p.current.Store(count)
}

func (p *ProgressBar) render() {
	current := p.current.Load()
	bytes := p.bytesTotal.Load()

	width := 28
	pct := float64(current) / float64(p.total)
	filled := int(pct * float64(width))
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	speed := ""
	elapsed := time.Since(p.start).Seconds()
	if elapsed > 0 && bytes > 0 {
		mbps := float64(bytes) / elapsed / 1024 / 1024
		speed = fmt.Sprintf(" • %.1f MB/s", mbps)
	}

	fmt.Printf("\r  %-7s [%s] %d/%d (%.0f%%)%s   ",
		p.label, bar, current, p.total, pct*100, speed)
}

func (p *ProgressBar) Done() {
	p.render()
	fmt.Println()
}
