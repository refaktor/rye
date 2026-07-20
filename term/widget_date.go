package term

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/refaktor/rye/env"
)

// =============================================================================
// DateWidget - Date picker with YYYY-MM-DD format
// =============================================================================

// DateWidget handles date input with field navigation
type DateWidget struct {
	BaseWidget
	Year  int
	Month int
	Day   int
	Field int // 0=year, 1=month, 2=day
}

// NewDateWidget creates a new date widget with optional initial date
func NewDateWidget(initialDate string, idx *env.Idxs) *DateWidget {
	w := &DateWidget{
		BaseWidget: NewBaseWidget(idx),
		Year:       2025,
		Month:      1,
		Day:        1,
		Field:      0,
	}

	// Parse initial date if provided
	if initialDate != "" {
		parts := strings.Split(initialDate, "-")
		if len(parts) == 3 {
			if y, err := strconv.Atoi(parts[0]); err == nil && y > 0 {
				w.Year = y
			}
			if m, err := strconv.Atoi(parts[1]); err == nil && m >= 1 && m <= 12 {
				w.Month = m
			}
			if d, err := strconv.Atoi(parts[2]); err == nil && d >= 1 {
				w.Day = d
			}
		}
	}

	// Validate day for the month
	maxDays := w.daysInMonth(w.Year, w.Month)
	if w.Day > maxDays {
		w.Day = maxDays
	}

	return w
}

func (w *DateWidget) daysInMonth(year, month int) int {
	switch month {
	case 2:
		if (year%4 == 0 && year%100 != 0) || (year%400 == 0) {
			return 29
		}
		return 28
	case 4, 6, 9, 11:
		return 30
	default:
		return 31
	}
}

func (w *DateWidget) formatDate() string {
	return fmt.Sprintf("%04d-%02d-%02d", w.Year, w.Month, w.Day)
}

func (w *DateWidget) Render() {
	RestoreCurPos()

	parts := []string{
		fmt.Sprintf("%04d", w.Year),
		fmt.Sprintf("%02d", w.Month),
		fmt.Sprintf("%02d", w.Day),
	}

	for i, part := range parts {
		if i > 0 {
			termPrint("-")
		}
		if i == w.Field {
			w.Theme.Selected()
			termPrint(part)
			w.Theme.Normal()
		} else {
			termPrint(part)
		}
	}
}

func (w *DateWidget) HandleKey(key WidgetKey) (done bool, canceled bool) {
	if key.IsCancel() {
		termPrintln("")
		return true, true
	}

	if key.IsSubmit() {
		termPrintln("")
		return true, false
	}

	// Field navigation
	if key.IsLeft() {
		w.Field--
		if w.Field < 0 {
			w.Field = 2
		}
		return false, false
	}
	if key.IsRight() {
		w.Field++
		if w.Field > 2 {
			w.Field = 0
		}
		return false, false
	}

	// Value increment/decrement
	if key.IsUp() {
		w.increment()
		return false, false
	}
	if key.IsDown() {
		w.decrement()
		return false, false
	}

	// Digit input
	if key.ASCII >= '0' && key.ASCII <= '9' {
		w.handleDigit(int(key.ASCII - '0'))
		return false, false
	}

	return false, false
}

func (w *DateWidget) increment() {
	switch w.Field {
	case 0: // Year
		w.Year++
		if w.Year > 9999 {
			w.Year = 1
		}
	case 1: // Month
		w.Month++
		if w.Month > 12 {
			w.Month = 1
		}
	case 2: // Day
		w.Day++
		maxDays := w.daysInMonth(w.Year, w.Month)
		if w.Day > maxDays {
			w.Day = 1
		}
	}
	w.validateDay()
}

func (w *DateWidget) decrement() {
	switch w.Field {
	case 0: // Year
		w.Year--
		if w.Year < 1 {
			w.Year = 9999
		}
	case 1: // Month
		w.Month--
		if w.Month < 1 {
			w.Month = 12
		}
	case 2: // Day
		w.Day--
		if w.Day < 1 {
			w.Day = w.daysInMonth(w.Year, w.Month)
		}
	}
	w.validateDay()
}

func (w *DateWidget) handleDigit(digit int) {
	switch w.Field {
	case 0: // Year - shift digits left
		yearStr := fmt.Sprintf("%04d", w.Year)
		yearStr = yearStr[1:] + strconv.Itoa(digit)
		w.Year, _ = strconv.Atoi(yearStr)
	case 1: // Month
		newMonth := (w.Month%10)*10 + digit
		if newMonth >= 1 && newMonth <= 12 {
			w.Month = newMonth
		} else if digit >= 1 && digit <= 9 {
			w.Month = digit
		}
	case 2: // Day
		maxDays := w.daysInMonth(w.Year, w.Month)
		newDay := (w.Day%10)*10 + digit
		if newDay >= 1 && newDay <= maxDays {
			w.Day = newDay
		} else if digit >= 1 && digit <= maxDays {
			w.Day = digit
		}
	}
	w.validateDay()
}

func (w *DateWidget) validateDay() {
	maxDays := w.daysInMonth(w.Year, w.Month)
	if w.Day > maxDays {
		w.Day = maxDays
	}
}

func (w *DateWidget) GetValue() env.Object {
	return *env.NewString(w.formatDate())
}

func (w *DateWidget) GetHeight() int {
	return 1
}
