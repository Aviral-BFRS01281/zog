package zjson

import (
	"encoding/json"
	"errors"
	"io"

	p "github.com/Aviral-BFRS01281/zog/internals"
	"github.com/Aviral-BFRS01281/zog/zconst"
)

// func Unmarshal(data []byte) p.DpFactory {
// 	return func() (p.DataProvider, p.ZogIssue) {
// 		var m map[string]any
// 		err := json.Unmarshal(data, &m)
// 		if err != nil {
// 			return nil, &p.ZogErr{C: zconst.IssueCodeInvalidJSON, Err: err}
// 		}
// 		if m == nil {
// 			return nil, &p.ZogErr{C: zconst.IssueCodeInvalidJSON, Err: errors.New("nill json body")}
// 		}
// 		return p.NewMapDataProvider(m), nil
// 	}
// }

var (
	jsonTag string = "json"
)

// Decodes JSON data. Does not support json arrays or primitives
/*
- "null" -> nil -> Not accepted by zhttp -> errs["$root"]-> required issue
- "{}" -> okay -> map[]{}
- "" -> parsing error -> errs["$root"]-> parsing error
- "1213" -> zhttp -> plain value
  - struct schema -> hey this valid input
  - "string is not an object"
*/
func Decode(r io.Reader) p.DpFactory {
	return func() (p.DataProvider, *p.ZogIssue) {
		closer, ok := r.(io.Closer)
		if ok {
			defer closer.Close()
		}
		var m map[string]any
		decod := json.NewDecoder(r)
		err := decod.Decode(&m)
		if err != nil {
			return nil, &p.ZogIssue{Code: zconst.IssueCodeInvalidJSON, Err: err}
		}
		if m == nil {
			return nil, &p.ZogIssue{Code: zconst.IssueCodeInvalidJSON, Err: errors.New("nill json body")}
		}
		return p.NewMapDataProvider(m, &jsonTag), nil
	}
}
