package elements

import (
	"github.com/boz/kubetop/ui/theme"
	"github.com/boz/kubetop/util"
	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/views"
)

type Panes interface {
	views.Widget
	theme.Themeable

	Widgets() []views.Widget

	Append(views.Widget)
	Prepend(views.Widget)
	Remove(views.Widget)
	InsertBefore(views.Widget, views.Widget)
	InsertAfter(views.Widget, views.Widget)
}

type panesChild struct {
	widget views.Widget
	view   *views.ViewPort
}

type panes struct {
	children []*panesChild

	orientation views.Orientation

	expand bool

	width  int
	height int

	view  views.View
	theme theme.Theme
	env   util.Env
	views.WidgetWatchers
}

func NewVPanes(env util.Env, expand bool) Panes {
	return NewPanes(env, views.Vertical, expand)
}

func NewHPanes(env util.Env, expand bool) Panes {
	return NewPanes(env, views.Horizontal, expand)
}

func NewPanes(env util.Env, o views.Orientation, expand bool) Panes {
	return &panes{orientation: o, expand: expand, env: env}
}

func (p *panes) SetTheme(th theme.Theme) {
	p.theme = th
	for _, c := range p.children {
		if w, ok := c.widget.(theme.Themeable); ok {
			w.SetTheme(th)
		}
	}
}

func (p *panes) Widgets() []views.Widget {
	children := make([]views.Widget, 0, len(p.children))
	for _, c := range p.children {
		children = append(children, c.widget)
	}
	return children
}

func (p *panes) Draw() {
	for _, c := range p.children {
		c.widget.Draw()
	}
}

func (p *panes) Resize() {
	p.layout()

	p.PostEventWidgetResize(p)
}

func (p *panes) HandleEvent(ev tcell.Event) bool {
	switch ev.(type) {
	case *views.EventWidgetContent:
		p.layout()
		p.PostEventWidgetContent(p)
		return true
	}
	for _, c := range p.children {
		if c.widget.HandleEvent(ev) {
			return true
		}
	}
	return false
}

func (p *panes) SetView(view views.View) {
	p.view = view
	for _, c := range p.children {
		c.view.SetView(view)
	}
	p.Resize()
}

func (p *panes) Size() (int, int) {
	return p.width, p.height
}

func (p *panes) Append(w views.Widget) {
	cnew := p.newChild(w)
	p.children = append(p.children, cnew)
	p.afterModify()
}

func (p *panes) Prepend(w views.Widget) {
	cnew := p.newChild(w)
	p.children = append([]*panesChild{cnew}, p.children...)
	p.afterModify()
}

func (p *panes) Remove(w views.Widget) {
	changed := false

	for i, c := range p.children {
		if c.widget == w {
			changed = true
			p.children = append(p.children[:i], p.children[i+1:]...)
		}
	}

	if changed {
		p.afterModify()
	}
}

func (p *panes) InsertBefore(mark views.Widget, w views.Widget) {
	for i, c := range p.children {
		if c.widget == mark {
			cnew := p.newChild(w)

			p.children = append(p.children, nil)
			copy(p.children[i+1:], p.children[i:])
			p.children[i] = cnew

			p.afterModify()
			return
		}
	}
}

func (p *panes) InsertAfter(mark views.Widget, w views.Widget) {
	for i, c := range p.children {
		if c.widget == mark {
			cnew := p.newChild(w)

			p.children = append(p.children, nil)
			copy(p.children[i+2:], p.children[i+1:])
			p.children[i+1] = cnew

			p.afterModify()
			return
		}
	}
}

func (p *panes) newChild(w views.Widget) *panesChild {
	c := &panesChild{
		view:   views.NewViewPort(p.view, 0, 0, 0, 0),
		widget: w,
	}
	w.SetView(c.view)
	w.Watch(p)
	return c
}

func (p *panes) afterModify() {
	p.layout()
	p.PostEventWidgetContent(p)
}

func (p *panes) layout() {
	switch p.orientation {
	case views.Horizontal:
		p.hlayout()
	default:
		p.vlayout()
	}
}

func (p *panes) vlayout() {
	if p.view == nil {
		return
	}

	vx, vy := p.view.Size()

	px, py := 0, 0

	for i, c := range p.children {
		wx, wy := c.widget.Size()

		if wx > px {
			px = wx
		}

		// if wy+py > vy {
		// 	logrus.StandardLogger().Debugf("vlayout: %v+%v>%v", wy, py, vy)
		// 	wy = vy - py
		// }

		if p.expand && i == len(p.children)-1 && vy-py > wy {
			wy = vy - py
		}

		c.view.Resize(0, py, vx, wy)
		c.widget.Resize()

		py += wy
	}

	p.width = px
	p.height = py
}

func (p *panes) hlayout() {
	if p.view == nil {
		return
	}

	vx, vy := p.view.Size()

	px, py := 0, 0

	for i, c := range p.children {
		wx, wy := c.widget.Size()

		if wy > py {
			py = wy
		}

		// if wx+px > vx {
		// 	wx = vx - px
		// }

		if p.expand && i == len(p.children)-1 && vx-px > wx {
			wx = vx - px
		}

		c.view.Resize(px, 0, wx, vy)
		c.widget.Resize()

		px += wx
	}

	p.width = px
	p.height = py
}
