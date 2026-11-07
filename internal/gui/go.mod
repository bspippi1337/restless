module github.com/bspippi1337/restless/internal/gui

go 1.24.0

// Optional Fyne desktop UI module. This module is intentionally separate
// so the core CLI does not require graphical/CGo dependencies.
require fyne.io/fyne/v2 v2.8.2
