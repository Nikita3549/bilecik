package bot

func RegisterHandlers(r *Router, h *Handlers) {
	r.Handle("start", h.Start)
	r.Handle("help", h.Help)
	r.Handle("subscribe", h.Subscribe)
	r.Handle("list", h.List)
	r.Handle("unsubscribe", h.Unsubscribe)
	r.Handle("cancel", h.Cancel)
	r.Fallback(h.Fallback)
	r.Intercept(h.Interceptor())
}
