package conq

type HelpFn = func(h func(Helper))

type Helper interface {
	Help(HelpCtx) string
}

// localization settings from LOCALE, etc
type HelpCtx struct {
}
