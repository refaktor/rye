// series.go
package env

//"fmt"

type TSeries struct {
	s   []Object
	pos int
}

func NewTSeries(ser []Object) *TSeries {
	ser1 := TSeries{ser, 0}
	return &ser1
}

func (ser TSeries) Ended() bool {
	return ser.pos < len(ser.s)
}

func (ser TSeries) AtLast() bool {
	return ser.pos < len(ser.s)-2
}

func (ser TSeries) Pos() int {
	return ser.pos
}

func (ser *TSeries) Next() {
	ser.pos++
}

func (ser *TSeries) Pop() Object {
	//fmt.Println(ser.pos)
	ser.pos++
	return ser.s[ser.pos-1]
}

func (ser *TSeries) Reset() {
	ser.pos = 0
}

func (ser *TSeries) SetPos(pos int) {
	ser.pos = pos
}

func (ser *TSeries) GetPos() int {
	return ser.pos
}

func (ser TSeries) Peek() Object {
	return ser.s[ser.pos-1]
}

func (ser TSeries) Get(n int) Object {
	//ser.pos += n + 1
	return ser.s[n]
}

func (ser TSeries) Len() int {
	return len(ser.s)
}
