package mikado

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

var ErrNotRunnable = errors.New("not runnable")

type typeKey struct {
	Type reflect.Type
}

func (k typeKey) String() string {
	return k.Type.String()
}

type App struct {
	logger

	providers []*provider
	runnables []typeKey

	values map[typeKey]reflect.Value
}

func New() *App {
	return &App{
		logger: defaultLogger,
		values: make(map[typeKey]reflect.Value),
	}
}

// AddProvider declares a provider for one or more components.
func (a *App) AddProvider(constructor interface{}) {
	a.providers = append(a.providers, newProvider(constructor))
}

// AddRunnable declares a provider for one or more components that implements the
// github.com/pior/runnable.Runnable interface.
func (a *App) AddRunnable(constructor interface{}) {
	p := newProvider(constructor)
	a.providers = append(a.providers, p)

	for _, output := range p.outputTypes {
		if !_isErrorType(output) {
			a.runnables = append(a.runnables, typeKey{output})
		}
	}
}

// Run instantiates all runnable components previously declared, and all their dependencies.
// Then, the runnable components are started (dependencies first).
// The running components are stopped when the context.Context is cancelled, or when a runnable prematurely exits.
func (a *App) Run(ctx context.Context) error {

	// 1. instantiate all components

	for _, runnable := range a.runnables {
		a.logger("building: %s", runnable)

		_, err := a.getValue(runnable)
		if err != nil {
			return fmt.Errorf("building %s: %w", runnable, err)
		}
	}

	// 2. start all runnable
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	var wg sync.WaitGroup

	for _, runnable := range a.runnables {
		a.logger("starting: %s", runnable)
		err := a.start(ctx, runnable, &wg)
		if err != nil {
			a.logger("runnable failed to start: %s", runnable)
			cancelCtx()
			break
		}
	}

	<-ctx.Done()
	a.logger("context cancelled, starting shutdown")

	// 3. TODO: cancel component in reverse order

	wg.Wait()

	return nil
}

func (a *App) start(ctx context.Context, key typeKey, wg *sync.WaitGroup) error {
	component, err := a.getValue(key)
	if err != nil {
		return err
	}

	if component.Kind() == reflect.Interface {
		component = component.Elem() // the interface may not cover the Runnable interface, but the type may.
	}

	componentRunFunc := component.MethodByName("Run")
	if !componentRunFunc.IsValid() {
		return ErrNotRunnable
	}

	wg.Add(1)
	go func() {
		a.logger("starting runnable: %s", key)

		args := []reflect.Value{reflect.ValueOf(ctx)}
		results := componentRunFunc.Call(args)

		a.logger("component %s returned: %s", component, results)
		wg.Done()
	}()

	return nil
}

func (a *App) getValue(key typeKey) (reflect.Value, error) {
	if v, ok := a.values[key]; ok {
		return v, nil
	}

	p := findProvider(a, key)
	if p == nil {
		return reflect.Value{}, errors.New("not found")
	}
	a.logger("calling provider: %s", p.name)

	err := a.callProvider(p)
	if err != nil {
		return reflect.Value{}, err
	}

	if v, ok := a.values[key]; ok {
		return v, nil
	}

	panic("bug")
}

func (a *App) setValue(key typeKey, value reflect.Value) {
	a.values[key] = value
}

func (a *App) callProvider(p *provider) error {
	var inputs []reflect.Value

	for _, inputType := range p.inputTypes {
		arg, err := a.getValue(typeKey{inputType})
		if err != nil {
			return err
		}

		inputs = append(inputs, arg)
	}

	results := p.call(inputs)
	for idx, output := range p.outputTypes {
		value := results[idx]
		if _isErrorType(output) && !value.IsNil() {
			err := value.Interface().(error)
			return fmt.Errorf("provider %s returned an error: %w", p.name, err)
		}
		a.setValue(typeKey{output}, value)
	}

	return nil
}
