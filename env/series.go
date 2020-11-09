// series.go
package env

//"fmt"

//"fmt"

//"fmt"

type TSeries struct {
	S   []Object
	pos int
}

func NewTSeries(ser []Object) *TSeries {
	ser1 := TSeries{ser, 0}
	return &ser1
}

func (ser TSeries) Ended() bool {
	return ser.pos > len(ser.S)
}

func (ser TSeries) AtLast() bool {
	return ser.pos > len(ser.S)-1
}

func (ser TSeries) Pos() int {
	return ser.pos
}

func (ser *TSeries) Next() {
	ser.pos++
}

func (ser *TSeries) Pop() Object {
	ser.pos++
	if len(ser.S) >= ser.pos {
		return ser.S[ser.pos-1]
	} else {
		return nil
	}
}

func (ser *TSeries) Put(obj Object) {
	ser.S[ser.pos-1] = obj // -1 ... because we already poped out the word .. if this works past the experiment improve this
}

func (ser *TSeries) Append(obj Object) *TSeries {
	ser.S = append(ser.S, obj) // -1 ... because we already poped out the word .. if this works past the experiment improve this
	return ser
}

func (ser *TSeries) AppendMul(objs []Object) *TSeries {
	ser.S = append(ser.S, objs...) // -1 ... because we already poped out the word .. if this works past the experiment improve this
	return ser
}

func (ser *TSeries) Reset() {
	//fmt.Println("RESET")
	ser.pos = 0
}

func (ser *TSeries) SetPos(pos int) {
	ser.pos = pos
}

func (ser *TSeries) GetPos() int {
	return ser.pos
}

func (ser *TSeries) GetAll() []Object {
	return ser.S
}

func (ser TSeries) Peek() Object {
	//fmt.Println(ser.pos)
	//fmt.Println(ser.s)
	if len(ser.S) > ser.pos { // maybe we could store len in object .. test later if it's faster
		return ser.S[ser.pos]
	}
	return nil
}

func (ser TSeries) Get(n int) Object {
	//ser.pos += n + 1
	return ser.S[n]
}

func (ser TSeries) Len() int {
	return len(ser.S)
}
