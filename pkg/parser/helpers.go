package parser

import (
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"

	infrabev1alpha1 "github.com/kuidio/kuid/apis/backend/infra/v1alpha1"
	"github.com/stoewer/go-strcase"
)

// TemplateHelperFunctions specifies a set of functions that are supplied as
// helpers to the templates that are used within this file.
var templateHelperFunctions = template.FuncMap{
	// inc provides a means to add 1 to a number, and is used within templates
	// to check whether the index of an element within a loop is the last one,
	// such that special handling can be provided for it (e.g., not following
	// it with a comma in a list of arguments).
	"inc":  func(i int) int { return i + 1 },
	"dec":  func(i int) int { return i - 1 },
	"mul":  func(p1 int, p2 int) int { return p1 * p2 },
	"mul3": func(p1, p2, p3 int) int { return p1 * p2 * p3 },
	"boolValue": func(b bool) int {
		if b {
			return 1
		} else {
			return 0
		}
	},
	"toUpperCamelCase": strcase.UpperCamelCase,
	"toLowerCamelCase": strcase.LowerCamelCase,
	"toKebabCase":      strcase.KebabCase,
	"toLower":          strings.ToLower,
	"toUpper":          strings.ToUpper,
	"mod":              func(i, j int) bool { return i%j == 0 },
	"deref":            func(s *string) string { return *s },
	"derefInt":         func(i *int) int { return *i },
	"list2string": func(s []*string) string {
		var str string
		for i, v := range s {
			if i < len(s)-1 {
				str = str + fmt.Sprintf("\"%s\", ", *v)
			} else {
				str = str + fmt.Sprintf("\"%s\"", *v)
			}
		}
		return str
	},
	"rtCommExpr": func(vrfUpId, lmgs int, wlShortname string) string {
		// if we come here there should be at least 1 element
		rtCommExpr := fmt.Sprintf("[rt-lmg%d-%d-%s]", 1, vrfUpId+1, wlShortname)
		for i := 2; i <= lmgs; i++ {
			rtCommExpr += fmt.Sprintf(" OR [rt-lmg%d-%d-%s]", i, vrfUpId+i, wlShortname)
		}
		return rtCommExpr
	},
	"lastmap": func(s string, x map[string][]*string) bool {
		i := 0
		for k := range x {
			if k == s {
				if i == len(x)-1 {
					return true
				}
			}
			i++
		}
		return false
	},
	"interfaceCounterMap": func() map[int]int {
		return map[int]int{}
	},
	"interfaceCounterMapIncrement": func(key int, counter map[int]int) string {
		if _, exists := counter[key]; !exists {
			counter[key] = 0
		}
		counter[key] = counter[key] + 1
		return ""
	},
	"interfaceCounterMapGet": func(key int, counter map[int]int) int {
		return counter[key]
	},
	"pad": func(mnc int) string {
		count := 0
		mnc_temp := mnc
		for mnc_temp != 0 {
			mnc_temp /= 10
			count = count + 1
		}
		if count == 2 {
			return "0" + strconv.Itoa(mnc)
		}
		return strconv.Itoa(mnc)
	},
	"isBreakoutPort": func(portName string) bool {
		return strings.Count(portName, "/") == 2
	},
	"ifName": func(portName string) string {
		// Compile the regular expression
		re := regexp.MustCompile(`^e(\d+)-(\d+)$`)

		// Check if the input string matches the pattern
		if !re.MatchString(portName) {
			return portName
		}

		// Perform the substitution
		return re.ReplaceAllString(portName, "ethernet-$1/$2")
	},
	"derefUint32": func(p *uint32) uint32 {
		if p == nil {
			return 0
		}
		return *p
	},
	"derefBool": func(p *bool) bool {
		if p == nil {
			return false
		}
		return *p
	},
	"isisLevel": func(level infrabev1alpha1.ISISLevel) string {
		return string(level)
	},
	"isisMetricStyle": func(metricStyle infrabev1alpha1.ISISMetricStyle) string {
		if metricStyle == infrabev1alpha1.ISISMetricStyleNarrow {
			return "narrow"
		} else {
			return "wise"
		}
	},
	"networkType": func(networkType infrabev1alpha1.NetworkType) string {
		switch networkType {
		case infrabev1alpha1.NetworkTypeBroadcast:
			return "broadcast"
		case infrabev1alpha1.NetworkTypeP2P:
			return "point-to-point"
		}
		return "point-to-point"
	},
}
