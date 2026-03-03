package views

import "math/rand"

type Face struct {
	quiet  bool
	i      int
	frames []string
}

func NewFace(quiet bool) Face {
	return Face{
		quiet: quiet,
		frames: []string{
			"(•‿•)",
			"(•ᴗ•)",
			"(•̀ᴗ•́)و",
			"(•‿•)",
		},
	}
}

func (f *Face) Tick() {
	if f.quiet {
		return
	}
	if rand.Intn(3) == 0 {
		f.i = (f.i + 1) % len(f.frames)
	}
}

func (f Face) View() string {
	if f.quiet {
		return ""
	}
	return f.frames[f.i]
}
