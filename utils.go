package main

import (
	"image/color"
	"math/rand"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func rainbowBg(base lipgloss.Style, s string, colors []color.Color) string {
	var str string

	for i, ss := range s {
		color, _ := colorful.MakeColor(colors[i%len(colors)])
		str = str + base.Background(lipgloss.Color(color.Hex())).Render(string(ss))
	}

	return str
}
