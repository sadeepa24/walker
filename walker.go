package walker

import (
	"errors"
	"reflect"
	"strconv"
)

type Walker struct {
	*Editor
	
	CanSetCheck CansetCheck //checks whether value can be set or not all inbuild types can be set and everything true IsFinal also true
	SetValue SetValue
	SliceTagHook SliceTagHook
	ApendCt ApendCt

	cache map[string]*Editor
}

type CansetCheck func(val reflect.Value, nextItemPath string, wlkr *Walker) bool
type SetValue func(curval reflect.Value, path string,  wlkr *Walker, item any) (reflect.Value, bool)
type SliceTagHook func (val reflect.Value, path string, i int) string
type ApendCt func(val reflect.Value, path string)  bool

type CommonFuncs struct {
	CanSetCheck CansetCheck //checks whether value can be set or not all inbuild types can be set and everything true IsFinal also true
	SetValue SetValue
	SliceTagHook SliceTagHook
	ApendCt ApendCt
}

var (
	ErrRootContruct = errors.New("cannot construct root")
	ErrUknownType = errors.New("cannot find type")
	ErrCanotAddress = errors.New("value is not addresable or cannot set")
)

type Editor struct {
	before *Editor
	val    reflect.Value
	place int
	EndValue bool
	Path string
	canset bool //checks whether support setting whole object default types already support all values true from IsFina also support
	
}

// func (e *Walker) Next() (error, bool) {

// 	return nil, false
// }

func NewWalker(root any) (*Walker, error) {
	e := reflect.ValueOf(root)
	if e.Kind() != reflect.Pointer {
		return nil, errors.New("root must be a pointer")
	}
	for e.Kind() == reflect.Interface || e.Kind() == reflect.Pointer{
		e = e.Elem()
	}
	return &Walker{
		Editor: &Editor{
			place: 0,
			val: e,
			canset: false,
		},
		cache: map[string]*Editor{},
	}, nil
}

func (e *Walker) Items() []string {
	
	fields := []string{}
	it := e.val
	if e.val.Kind() == reflect.Pointer {
		it = e.val.Elem()
	}
	switch it.Kind() {
	case reflect.Struct:
		tp := it.Type()
		for i := 0; i < it.NumField(); i++ {
			fields = append(fields, tp.Field(i).Name)
		}
		return fields
	case reflect.Slice:
		if e.SliceTagHook != nil {
			for i := 0; i < it.Len(); i++ {
				fields = append(fields, e.SliceTagHook(it.Index(i), e.Path, i))
			}
		} else {
			for i := 0; i < it.Len(); i++ {
				fields = append(fields, strconv.Itoa(i))
			}
			return fields
		}
		
	}
	return fields
}

func (e *Walker) WalkInto(name string) bool {
	edtr, loaded := e.cache[e.Path+"."+name]
	if loaded {
		e.Editor = edtr
		return true
	}

	var nextitem reflect.Value 
	switch e.val.Kind() {
	case reflect.Slice:
		if e.SliceTagHook != nil {
			ilp:
			for i := 0; i < e.val.Len(); i++ {
				if e.SliceTagHook(e.val.Index(i), e.Path, i) == name {
					nextitem = e.val.Index(i)
					break ilp
				}
			}
			break
		}
		index, err := strconv.Atoi(name)
		if err != nil {
			return false
		}
		if index >= e.val.Len() {
			return false
		}
		nextitem = e.val.Index(index)
	case reflect.Struct:
		nextitem = e.val.FieldByName(name)
	case reflect.Pointer, reflect.Interface:
		return false
	default:
		return false
	}
	if nextitem.Kind() == reflect.Invalid {
		return false
	}
	for nextitem.Kind() == reflect.Interface || nextitem.Kind() == reflect.Pointer {
		if nextitem.IsNil() {
			break
		}
		nextitem = nextitem.Elem()
	}
	
	if !nextitem.CanSet() || !nextitem.CanAddr() {
		return false
	}
	e.Editor = &Editor{
		before: e.Editor,
		Path: e.Path+("."+name),
		val: nextitem,
		place: e.place+1,
	}
	e.canset = e.checkcanset(nextitem, name)
	
	if !e.EndValue {
		e.cache[e.Path] = e.Editor
	}
	
	return true
}

func (e *Walker) WalkBack() error {
	if e.place == 0 {
		return nil
	}
	e.Editor = e.before
	return nil
}


func (e *Walker) CanSet() bool {
	return e.canset
}

func (e *Walker) Construct(opts ...string) error {
	if e.Path == "" {
		return ErrRootContruct
	}
	if e.val.CanSet(){
		var (
			tp reflect.Value
			err error
		)
		tp, err = e.typconstruct(e.val.Type())
		if err != nil && e.SetValue != nil {
			ok := e.change("");
			if !ok {
				return err
			}
		} else {
			e.val.Set(tp)
		}
		for e.val.Kind() == reflect.Interface || e.val.Kind() == reflect.Pointer {
			if e.val.IsNil() {
				break
			}
			e.val = e.val.Elem()
		}
		e.canset = e.checkcanset(e.val, e.Path)
	} else {
		return ErrCanotAddress
	}
	return nil
}

func (e *Walker) typconstruct(typ reflect.Type) (reflect.Value, error) {
	var (
		endtyp reflect.Value
	)
	switch typ.Kind() {
	case reflect.Pointer:
		nv := reflect.New(typ.Elem())
		endtyp = reflect.NewAt(typ.Elem(), nv.UnsafePointer())
	default:
		endtyp = reflect.New(typ).Elem()
	}
	switch endtyp.Kind() {
	case reflect.Pointer, reflect.Interface:
		if endtyp.IsNil() {
			return endtyp, ErrUknownType
		}
	}
	return endtyp, nil

}


func (e *Walker) AppendZero() (err error) {
	if e.val.Kind() == reflect.Slice {	
		edty, err  := e.typconstruct(e.val.Type().Elem())
		if err != nil {
			return err
		}
		e.val.Set(reflect.Append(e.val, edty))
	}
	return
}

func (e *Walker) AppendCustom() (err error) {
	if e.val.Kind() == reflect.Slice {	
		if e.ApendCt != nil {
			if e.ApendCt(e.val, e.Path) {
				return
			}
		}
		if e.CanSetCheck != nil {
			elemval, err := e.typconstruct(e.val.Type().Elem())
			if err != nil {
				return err
			}
			if e.checkcansetnoend(elemval, e.Path) {
				new, ok := e.SetValue(elemval, e.Path, e, nil)
				if !ok {
					return nil
				} 
				if elemval.CanSet() && elemval.Kind() == new.Kind() {
					elemval.Set(new)
				}
				e.val.Set(reflect.Append(e.val, elemval))
			}
		}
	}
	return
}

func (e *Walker) RemoveFromSLice(place string) {
	if e.val.Kind() == reflect.Slice {	
		if e.SliceTagHook != nil {
			for i := 0; i < e.val.Len(); i++ {
				if e.SliceTagHook(e.val.Index(i), e.Path, i) == place {
					if i < 0 || i >= e.val.Len() {
						return
					}
					newSlice := reflect.AppendSlice(e.val.Slice(0, i), e.val.Slice(i+1, e.val.Len()))
					if newSlice.Len() < e.val.Len() {
						e.val.Set(newSlice)
					}
					return
				}
			}
			return
		}
		idx, err := strconv.Atoi(place)
		if err != nil {
			return
		}

		if idx < 0 || idx >= e.val.Len() {
			return
		}
		newSlice := reflect.AppendSlice(e.val.Slice(0, idx), e.val.Slice(idx+1, e.val.Len()))
		if newSlice.Len() < e.val.Len() {
			e.val.Set(newSlice)
		}
	}
}

func (e *Walker) change(item any) bool {
	if e.val.CanSet(){
		if e.SetValue != nil {
			if val, ok := e.SetValue(e.val, e.Path, e, item); ok {
				switch {
				case val.IsZero():
				case val.Type() == e.val.Type():
					e.val.Set(val)
					return true
				case val.Kind() == reflect.Pointer:
					if !val.IsNil() && val.Type().Elem() == e.val.Type() {
						e.val.Set(val.Elem())
					}
					return true
				}
			}
		}
		if e.EndValue {
			en := reflect.ValueOf(item)
			if en.Type() == e.val.Type() {
				e.val.Set(en)
			}
		}
	}
	return false
}

func (e *Walker) Change(item any) {
	e.change(item)
}

func (e *Walker) Current() reflect.Value {
	return e.val
}
func (e *Walker) CurrentPtr() (reflect.Value, bool) {
	if e.val.CanAddr() {
		return e.val.Addr(), true
	}
	return e.val, false

}

func (e *Walker) CurrentPtrIface() (any, bool) {
	if e.val.CanAddr() {
		return e.val.Addr().Interface(), true
	}
	return nil, false

}

func (e *Walker) checkcansetnoend(val reflect.Value, nextItemName string) bool {
	switch val.Kind() {
	case reflect.Struct:
		if e.CanSetCheck != nil  {
			return e.CanSetCheck(val, e.Path +"." +nextItemName, e)
		}
		return false
	case reflect.Array, reflect.Slice, reflect.Pointer, reflect.Map, reflect.Func, reflect.Chan, reflect.Interface:
		if e.CanSetCheck != nil  {
			return e.CanSetCheck(val, e.Path +"." +nextItemName, e)
		}
		return false
	case reflect.Invalid, reflect.UnsafePointer:
		return false
	default:
		return true
	}
}
func (e *Walker) checkcanset(val reflect.Value, nextItemName string) bool {
	e.EndValue = false
	switch val.Kind() {
	case reflect.Struct:
		if e.CanSetCheck != nil  {
			return e.CanSetCheck(val, e.Path +"." +nextItemName, e)
		}
		return false
	case reflect.Array, reflect.Slice, reflect.Pointer, reflect.Map, reflect.Func, reflect.Chan, reflect.Interface:
		if e.CanSetCheck != nil  {
			return e.CanSetCheck(val, e.Path +"." +nextItemName, e)
		}
		return false
	case reflect.Invalid, reflect.UnsafePointer:
		return false
	default:
		e.EndValue = true
		return true
	}
}

func (e *Walker) Child(name string) (reflect.Value, bool) {
	
	switch e.val.Kind() {
	case reflect.Struct:
		return e.val.FieldByName(name), true
	}
	return e.val, false
}